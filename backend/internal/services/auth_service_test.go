package services

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

type mockUserRepository struct {
	createFunc     func(ctx context.Context, user *domain.User) error
	getByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	getByIDFunc    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
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

func TestMain(m *testing.M) {
	bcryptCost = bcrypt.MinCost
	os.Exit(m.Run())
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		userName string
		repoErr  error
		wantErr  bool
		wantCode int
	}{
		{
			name:     "success",
			email:    "user@example.com",
			password: "password123",
			userName: "Test User",
		},
		{
			name:     "email already exists",
			email:    "existing@example.com",
			password: "password123",
			userName: "Test User",
			repoErr:  domain.ErrUserAlreadyExists,
			wantErr:  true,
			wantCode: http.StatusConflict,
		},
		{
			name:     "invalid email format",
			email:    "not-an-email",
			password: "password123",
			userName: "Test User",
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "password too short",
			email:    "user@example.com",
			password: "short",
			userName: "Test User",
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "empty name",
			email:    "user@example.com",
			password: "password123",
			userName: "",
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repoErr := tc.repoErr
			repo := &mockUserRepository{
				createFunc: func(ctx context.Context, user *domain.User) error {
					return repoErr
				},
			}
			svc := NewAuthService(repo, "test-secret-key-for-testing")

			user, accessToken, refreshToken, err := svc.Register(context.Background(), tc.email, tc.password, tc.userName)

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
				assert.Equal(t, strings.ToLower(tc.email), user.Email)
				assert.Equal(t, strings.TrimSpace(tc.userName), user.Name)
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotZero(t, user.ConsentedAt)
				assert.NotEqual(t, uuid.Nil, user.ID)
			}
		})
	}
}

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
			svc := NewAuthService(repo, "test-secret-key-for-testing")

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
