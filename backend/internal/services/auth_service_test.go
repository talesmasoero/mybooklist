package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

// ─── mockUserRepository ─────────────────────────────────────────────────────

type mockUserRepository struct {
	createFunc         func(ctx context.Context, user *domain.User) error
	getByEmailFunc     func(ctx context.Context, email string) (*domain.User, error)
	getByIDFunc        func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	updateNameFunc     func(ctx context.Context, id uuid.UUID, name string) error
	updatePasswordFunc func(ctx context.Context, id uuid.UUID, passwordHash string) error
	deleteFunc         func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) CreateInTx(_ context.Context, _ *sql.Tx, _ *domain.User) error {
	return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	if m.updateNameFunc != nil {
		return m.updateNameFunc(ctx, id, name)
	}
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	if m.updatePasswordFunc != nil {
		return m.updatePasswordFunc(ctx, id, passwordHash)
	}
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// ─── helpers ────────────────────────────────────────────────────────────────

func newAuthSvcForTest(repo *mockUserRepository) AuthService {
	return NewAuthService(nil, repo, nil, nil, nil, "test-secret-key-for-testing")
}

func TestMain(m *testing.M) {
	bcryptCost = bcrypt.MinCost
	os.Exit(m.Run())
}

// ─── Register (input validation only — DB path requires integration test) ───

func TestAuthService_Register_Validation(t *testing.T) {
	validAnswers := []domain.SaveAnswerInput{
		{QuestionID: 1, Answer: "resposta um"},
		{QuestionID: 2, Answer: "resposta dois"},
	}

	tests := []struct {
		name     string
		input    RegisterInput
		wantCode int
	}{
		{
			name:     "invalid email format",
			input:    RegisterInput{Email: "not-an-email", Password: "password123", Name: "Test User", SecurityAnswers: validAnswers},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "password too short",
			input:    RegisterInput{Email: "user@example.com", Password: "short", Name: "Test User", SecurityAnswers: validAnswers},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "empty name",
			input:    RegisterInput{Email: "user@example.com", Password: "password123", Name: "", SecurityAnswers: validAnswers},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "no security answers",
			input:    RegisterInput{Email: "user@example.com", Password: "password123", Name: "Test User", SecurityAnswers: nil},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "only one security answer",
			input: RegisterInput{
				Email: "user@example.com", Password: "password123", Name: "Test User",
				SecurityAnswers: []domain.SaveAnswerInput{{QuestionID: 1, Answer: "resposta"}},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "four security answers (too many)",
			input: RegisterInput{
				Email: "user@example.com", Password: "password123", Name: "Test User",
				SecurityAnswers: []domain.SaveAnswerInput{
					{QuestionID: 1, Answer: "r1"}, {QuestionID: 2, Answer: "r2"},
					{QuestionID: 3, Answer: "r3"}, {QuestionID: 4, Answer: "r4"},
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "duplicate question IDs",
			input: RegisterInput{
				Email: "user@example.com", Password: "password123", Name: "Test User",
				SecurityAnswers: []domain.SaveAnswerInput{
					{QuestionID: 1, Answer: "resposta um"}, {QuestionID: 1, Answer: "outra resposta"},
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "answer too short",
			input: RegisterInput{
				Email: "user@example.com", Password: "password123", Name: "Test User",
				SecurityAnswers: []domain.SaveAnswerInput{
					{QuestionID: 1, Answer: "x"}, {QuestionID: 2, Answer: "ok"},
				},
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newAuthSvcForTest(&mockUserRepository{})
			user, accessToken, refreshToken, err := svc.Register(context.Background(), tc.input)

			require.Error(t, err)
			var appErr *domain.AppError
			require.True(t, errors.As(err, &appErr), "error must be *domain.AppError")
			assert.Equal(t, tc.wantCode, appErr.Code)
			assert.Nil(t, user)
			assert.Empty(t, accessToken)
			assert.Empty(t, refreshToken)
		})
	}
}

// ─── Login ──────────────────────────────────────────────────────────────────

func TestAuthService_Login(t *testing.T) {
	validHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	require.NoError(t, err, "test setup: failed to generate bcrypt hash")

	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: string(validHash),
		Name:         "Test User",
	}

	tests := []struct {
		name     string
		email    string
		password string
		repoUser *domain.User
		repoErr  error
		wantErr  bool
		wantCode int
	}{
		{
			name:     "success",
			email:    "user@example.com",
			password: "correct-password",
			repoUser: existingUser,
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrong-password",
			repoUser: existingUser,
			wantErr:  true,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "user not found",
			email:    "unknown@example.com",
			password: "password123",
			repoErr:  domain.ErrUserNotFound,
			wantErr:  true,
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repoUser := tc.repoUser
			repoErr := tc.repoErr
			repo := &mockUserRepository{
				getByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
					if repoErr != nil {
						return nil, repoErr
					}
					return repoUser, nil
				},
			}
			svc := newAuthSvcForTest(repo)

			user, accessToken, refreshToken, err := svc.Login(context.Background(), tc.email, tc.password)

			if tc.wantErr {
				require.Error(t, err)
				var appErr *domain.AppError
				require.True(t, errors.As(err, &appErr), "error must be *domain.AppError")
				assert.Equal(t, tc.wantCode, appErr.Code)
				assert.Nil(t, user)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
				assert.Equal(t, existingUser.ID, user.ID)
				assert.Equal(t, existingUser.Email, user.Email)
			}
		})
	}
}
