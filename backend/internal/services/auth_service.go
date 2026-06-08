package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
	"github.com/talesmasoero/mybooklist/backend/internal/repositories"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
	resetTokenDuration   = 15 * time.Minute
)

// bcryptCost is a var so tests can override it with bcrypt.MinCost for speed.
var bcryptCost = 12

type RegisterInput struct {
	Email           string
	Password        string
	Name            string
	SecurityAnswers []domain.SaveAnswerInput
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*domain.User, string, string, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, string, error)
	InitiatePasswordReset(ctx context.Context, email string) (*domain.PasswordResetSession, error)
	ValidatePasswordResetAnswers(ctx context.Context, userID uuid.UUID, answers []domain.AnswerInput) (string, error)
	ResetPasswordWithToken(ctx context.Context, rawToken, newPassword string) error
}

type authService struct {
	db             *sql.DB
	userRepo       repositories.UserRepository
	answerRepo     repositories.UserSecurityAnswerRepository
	sqRepo         repositories.SecurityQuestionRepository
	resetTokenRepo repositories.PasswordResetTokenRepository
	jwtSecret      []byte
}

func NewAuthService(
	db *sql.DB,
	userRepo repositories.UserRepository,
	answerRepo repositories.UserSecurityAnswerRepository,
	sqRepo repositories.SecurityQuestionRepository,
	resetTokenRepo repositories.PasswordResetTokenRepository,
	jwtSecret string,
) AuthService {
	return &authService{
		db:             db,
		userRepo:       userRepo,
		answerRepo:     answerRepo,
		sqRepo:         sqRepo,
		resetTokenRepo: resetTokenRepo,
		jwtSecret:      []byte(jwtSecret),
	}
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (*domain.User, string, string, error) {
	if err := validateRegisterInput(input.Email, input.Password, input.Name); err != nil {
		return nil, "", "", err
	}
	if err := validateAnswerCount(input.SecurityAnswers); err != nil {
		return nil, "", "", err
	}
	if err := validateDistinctQuestionIDs(input.SecurityAnswers); err != nil {
		return nil, "", "", err
	}
	if err := validateAnswerLengths(input.SecurityAnswers); err != nil {
		return nil, "", "", err
	}

	allQuestions, err := s.sqRepo.ListAll(ctx)
	if err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to validate security questions", Err: err}
	}
	if err := validateQuestionIDsExist(input.SecurityAnswers, allQuestions); err != nil {
		return nil, "", "", err
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
	if err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to process password", Err: err}
	}

	hashedAnswers := make([]domain.SaveAnswerInput, len(input.SecurityAnswers))
	for i, a := range input.SecurityAnswers {
		normalized := strings.ToLower(strings.TrimSpace(a.Answer))
		ah, err := bcrypt.GenerateFromPassword([]byte(normalized), bcryptCost)
		if err != nil {
			return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to process security answers", Err: err}
		}
		hashedAnswers[i] = domain.SaveAnswerInput{QuestionID: a.QuestionID, Answer: string(ah)}
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(input.Name),
		CreatedAt:    now,
		UpdatedAt:    now,
		ConsentedAt:  now,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to start transaction", Err: err}
	}
	defer tx.Rollback()

	if err := s.userRepo.CreateInTx(ctx, tx, user); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, "", "", &domain.AppError{Code: http.StatusConflict, ErrorCode: "ERR_EMAIL_ALREADY_EXISTS", Message: "email already registered", Err: err}
		}
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to create user", Err: err}
	}

	if err := s.answerRepo.SaveAnswersInTx(ctx, tx, user.ID, hashedAnswers); err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to save security answers", Err: err}
	}

	if err := tx.Commit(); err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to commit transaction", Err: err}
	}

	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to generate tokens", Err: err}
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", "", &domain.AppError{Code: http.StatusUnauthorized, ErrorCode: "ERR_INVALID_CREDENTIALS", Message: "invalid credentials", Err: domain.ErrInvalidCredentials}
		}
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to retrieve user", Err: err}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusUnauthorized, ErrorCode: "ERR_INVALID_CREDENTIALS", Message: "invalid credentials", Err: domain.ErrInvalidCredentials}
	}

	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to generate tokens", Err: err}
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) InitiatePasswordReset(ctx context.Context, email string) (*domain.PasswordResetSession, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	empty := &domain.PasswordResetSession{Questions: []domain.SecurityQuestion{}}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return empty, nil
	}

	questions, err := s.answerRepo.GetQuestionsForUser(ctx, user.ID)
	if err != nil || len(questions) < 2 {
		return empty, nil
	}

	return &domain.PasswordResetSession{
		UserID:    user.ID,
		Questions: questions,
	}, nil
}

func (s *authService) ValidatePasswordResetAnswers(ctx context.Context, userID uuid.UUID, answers []domain.AnswerInput) (string, error) {
	stored, err := s.answerRepo.GetByUserID(ctx, userID)
	if err != nil || len(stored) == 0 {
		return "", &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_INVALID_ANSWERS", Message: "incorrect security answers"}
	}

	storedByID := make(map[int]string, len(stored))
	for _, a := range stored {
		storedByID[a.QuestionID] = a.AnswerHash
	}

	correct := 0
	for _, incoming := range answers {
		hash, ok := storedByID[incoming.QuestionID]
		if !ok {
			continue
		}
		normalized := strings.ToLower(strings.TrimSpace(incoming.Answer))
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(normalized)) == nil {
			correct++
		}
	}

	if correct < 2 {
		return "", &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_INVALID_ANSWERS", Message: "incorrect security answers"}
	}

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to generate reset token", Err: err}
	}
	rawToken := hex.EncodeToString(rawBytes)
	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	now := time.Now().UTC()
	token := &domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(resetTokenDuration),
		CreatedAt: now,
	}

	if err := s.resetTokenRepo.Create(ctx, token); err != nil {
		return "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to create reset token", Err: err}
	}

	return rawToken, nil
}

func (s *authService) ResetPasswordWithToken(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "password must be at least 8 characters"}
	}

	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	token, err := s.resetTokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_TOKEN_INVALID", Message: "invalid or expired reset token", Err: err}
	}

	if token.UsedAt != nil {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_TOKEN_USED", Message: "reset token has already been used"}
	}

	if time.Now().UTC().After(token.ExpiresAt) {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_TOKEN_EXPIRED", Message: "reset token has expired"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to process password", Err: err}
	}

	if err := s.userRepo.UpdatePassword(ctx, token.UserID, string(hash)); err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to update password", Err: err}
	}

	if err := s.resetTokenRepo.MarkAsUsed(ctx, token.ID); err != nil {
		return &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to invalidate reset token", Err: err}
	}

	s.resetTokenRepo.InvalidateAllForUser(ctx, token.UserID)

	return nil
}

func (s *authService) generateTokens(userID uuid.UUID) (string, string, error) {
	now := time.Now()
	sub := userID.String()

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        sub,
		"token_type": domain.TokenTypeAccess,
		"iat":        now.Unix(),
		"exp":        now.Add(accessTokenDuration).Unix(),
	}).SignedString(s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        sub,
		"token_type": domain.TokenTypeRefresh,
		"iat":        now.Unix(),
		"exp":        now.Add(refreshTokenDuration).Unix(),
	}).SignedString(s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func validateRegisterInput(email, password, name string) error {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)

	if !strings.Contains(email, "@") || !strings.Contains(strings.SplitN(email, "@", 2)[1], ".") {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "invalid email format"}
	}
	if len(password) < 8 {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "password must be at least 8 characters"}
	}
	if name == "" {
		return &domain.AppError{Code: http.StatusBadRequest, ErrorCode: "ERR_VALIDATION", Message: "name is required"}
	}
	return nil
}
