package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/handokobeni/agent-openhands/internal/domain/entity"
	"github.com/handokobeni/agent-openhands/internal/usecase/auth"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Config holds JWT configuration
type Config struct {
	SecretKey         string
	AccessTokenExpiry time.Duration
}

// tokenGenerator implements auth.TokenGenerator
type tokenGenerator struct {
	config Config
}

// NewTokenGenerator creates a new JWT token generator
func NewTokenGenerator(config Config) auth.TokenGenerator {
	return &tokenGenerator{config: config}
}

// Claims represents JWT claims
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new JWT access token
func (t *tokenGenerator) GenerateAccessToken(claims entity.TokenClaims) (string, error) {
	now := time.Now()

	tokenClaims := Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(t.config.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString([]byte(t.config.SecretKey))
}

// GenerateRefreshToken generates a cryptographically secure refresh token
func (t *tokenGenerator) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateAccessToken validates and parses a JWT access token
func (t *tokenGenerator) ValidateAccessToken(tokenString string) (*entity.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(t.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &entity.TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}
