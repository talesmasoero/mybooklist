package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":{"code":"ERR_UNAUTHORIZED","message":"missing or invalid authorization header"}}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			}, jwt.WithExpirationRequired())
			if err != nil || !token.Valid {
				http.Error(w, `{"error":{"code":"ERR_UNAUTHORIZED","message":"invalid or expired token"}}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":{"code":"ERR_UNAUTHORIZED","message":"invalid token claims"}}`, http.StatusUnauthorized)
				return
			}

			if tokenType, _ := claims["token_type"].(string); tokenType != "access" {
				http.Error(w, `{"error":{"code":"ERR_UNAUTHORIZED","message":"invalid token type"}}`, http.StatusUnauthorized)
				return
			}

			sub, _ := claims["sub"].(string)
			userID, err := uuid.Parse(sub)
			if err != nil {
				http.Error(w, `{"error":{"code":"ERR_UNAUTHORIZED","message":"invalid token subject"}}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}
