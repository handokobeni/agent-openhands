package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/handokobeni/agent-openhands/internal/usecase/auth"
	"github.com/handokobeni/agent-openhands/pkg/response"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	tokenGenerator auth.TokenGenerator
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(tokenGenerator auth.TokenGenerator) *AuthMiddleware {
	return &AuthMiddleware{tokenGenerator: tokenGenerator}
}

// Authenticate verifies the JWT token from cookie or Authorization header
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// First, try to get token from httpOnly cookie
		cookie, err := r.Cookie("access_token")
		if err == nil && cookie.Value != "" {
			tokenString = cookie.Value
		}

		// If not in cookie, try Authorization header as fallback
		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			response.Unauthorized(w, "missing access token")
			return
		}

		// Validate token
		claims, err := m.tokenGenerator.ValidateAccessToken(tokenString)
		if err != nil {
			response.Unauthorized(w, "invalid or expired access token")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetUserEmail extracts user email from context
func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}
