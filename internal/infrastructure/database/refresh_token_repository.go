package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/handokobeni/agent-openhands/internal/domain/entity"
	"github.com/handokobeni/agent-openhands/internal/domain/repository"
)

// refreshTokenRepository implements repository.RefreshTokenRepository
type refreshTokenRepository struct {
	db *sql.DB
}

// NewRefreshTokenRepository creates a new refresh token repository instance
func NewRefreshTokenRepository(db *sql.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (user_id, token, device_info, ip_address, user_agent, expires_at, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		token.UserID,
		token.Token,
		token.DeviceInfo,
		token.IPAddress,
		token.UserAgent,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	token.ID = id
	return nil
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	query := `SELECT id, user_id, token, device_info, ip_address, user_agent, expires_at, revoked_at, created_at 
			  FROM refresh_tokens WHERE token = ?`

	rt := &entity.RefreshToken{}
	var revokedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.DeviceInfo,
		&rt.IPAddress,
		&rt.UserAgent,
		&rt.ExpiresAt,
		&revokedAt,
		&rt.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if revokedAt.Valid {
		rt.RevokedAt = &revokedAt.Time
	}

	return rt, nil
}

func (r *refreshTokenRepository) FindByUserID(ctx context.Context, userID int64) ([]*entity.RefreshToken, error) {
	query := `SELECT id, user_id, token, device_info, ip_address, user_agent, expires_at, revoked_at, created_at 
			  FROM refresh_tokens WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*entity.RefreshToken
	for rows.Next() {
		rt := &entity.RefreshToken{}
		var revokedAt sql.NullTime

		err := rows.Scan(
			&rt.ID,
			&rt.UserID,
			&rt.Token,
			&rt.DeviceInfo,
			&rt.IPAddress,
			&rt.UserAgent,
			&rt.ExpiresAt,
			&revokedAt,
			&rt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if revokedAt.Valid {
			rt.RevokedAt = &revokedAt.Time
		}

		tokens = append(tokens, rt)
	}

	return tokens, rows.Err()
}

func (r *refreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET revoked_at = ? WHERE token = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), token)
	return err
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < ?`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
