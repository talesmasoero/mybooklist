package services

import (
	"context"
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
)

// bcryptCost is a var so tests can override it with bcrypt.MinCost for speed.
var bcryptCost = 12

type AuthService interface {
	Register(ctx context.Context, email, password, name string) (*domain.User, string, string, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, string, error)
}

type authService struct {
	repo      repositories.UserRepository
	jwtSecret []byte
}

func NewAuthService(repo repositories.UserRepository, jwtSecret string) AuthService {
	return &authService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *authService) Register(ctx context.Context, email, password, name string) (*domain.User, string, string, error) {
	if err := validateRegisterInput(email, password, name); err != nil {
		return nil, "", "", err
	}

	email = strings.ToLower(strings.TrimSpace(email))

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to process password", Err: err}
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(name),
		CreatedAt:    now,
		UpdatedAt:    now,
		ConsentedAt:  now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, "", "", &domain.AppError{Code: http.StatusConflict, ErrorCode: "ERR_EMAIL_ALREADY_EXISTS", Message: "email already registered", Err: err}
		}
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to create user", Err: err}
	}

	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, "", "", &domain.AppError{Code: http.StatusInternalServerError, ErrorCode: "ERR_INTERNAL", Message: "failed to generate tokens", Err: err}
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.GetByEmail(ctx, email)
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
