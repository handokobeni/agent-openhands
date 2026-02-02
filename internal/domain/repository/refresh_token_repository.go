package repository

import (
	"context"

	"github.com/handokobeni/agent-openhands/internal/domain/entity"
)

// RefreshTokenRepository defines the interface for refresh token data operations
// Following Interface Segregation Principle (ISP) - only token-related methods
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entity.RefreshToken, error)
	RevokeByToken(ctx context.Context, token string) error
	RevokeAllByUserID(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context) error
}
