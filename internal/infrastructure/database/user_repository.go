package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/handokobeni/agent-openhands/internal/domain/entity"
	"github.com/handokobeni/agent-openhands/internal/domain/repository"
)

// userRepository implements repository.UserRepository
// Following Liskov Substitution Principle (LSP) - can be substituted for UserRepository interface
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (email, password, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, user.Email, user.Password, user.Name, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	query := `SELECT id, email, password, name, created_at, updated_at FROM users WHERE id = ?`
	user := &entity.User{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, email, password, name, created_at, updated_at FROM users WHERE email = ?`
	user := &entity.User{}

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `UPDATE users SET email = ?, password = ?, name = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, user.Email, user.Password, user.Name, user.UpdatedAt, user.ID)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
