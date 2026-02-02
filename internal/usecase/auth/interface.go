package auth

import (
	"context"

	"github.com/handokobeni/agent-openhands/internal/domain/entity"
)

// UseCase defines the interface for authentication use cases
// Following Dependency Inversion Principle (DIP) - high-level modules depend on abstractions
type UseCase interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest, deviceInfo entity.DeviceInfo) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string, deviceInfo entity.DeviceInfo) (*TokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID int64) error
	GetActiveSessions(ctx context.Context, userID int64) ([]*entity.RefreshToken, error)
}

// PasswordHasher defines the interface for password hashing operations
// Following Single Responsibility Principle (SRP) - only handles password hashing
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// TokenGenerator defines the interface for JWT token operations
// Following Single Responsibility Principle (SRP) - only handles token generation
type TokenGenerator interface {
	GenerateAccessToken(claims entity.TokenClaims) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(token string) (*entity.TokenClaims, error)
}
