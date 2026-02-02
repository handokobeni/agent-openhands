package repository

import (
	"context"

	"github.com/handokobeni/agent-openhands/internal/domain/entity"
)

// UserRepository defines the interface for user data operations
// Following Interface Segregation Principle (ISP) - only user-related methods
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int64) error
}
