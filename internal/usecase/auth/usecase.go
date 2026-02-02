package auth

import (
	"context"
	"errors"
	"time"

	"github.com/handokobeni/agent-openhands/internal/domain/entity"
	"github.com/handokobeni/agent-openhands/internal/domain/repository"
)

var (
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrUserNotFound        = errors.New("user not found")
)

// Config holds the configuration for auth usecase
type Config struct {
	RefreshTokenExpiry time.Duration
}

// authUseCase implements the UseCase interface
// Following Open/Closed Principle (OCP) - open for extension, closed for modification
type authUseCase struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	passwordHasher   PasswordHasher
	tokenGenerator   TokenGenerator
	config           Config
}

// NewAuthUseCase creates a new instance of auth use case
// Following Dependency Injection pattern for better testability
func NewAuthUseCase(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	passwordHasher PasswordHasher,
	tokenGenerator TokenGenerator,
	config Config,
) UseCase {
	return &authUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordHasher:   passwordHasher,
		tokenGenerator:   tokenGenerator,
		config:           config,
	}
}

// Register creates a new user account
func (uc *authUseCase) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := uc.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &entity.User{
		Email:     req.Email,
		Password:  hashedPassword,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User: UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}

// Login authenticates a user and returns tokens
func (uc *authUseCase) Login(ctx context.Context, req LoginRequest, deviceInfo entity.DeviceInfo) (*AuthResponse, error) {
	// Find user by email
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := uc.passwordHasher.Compare(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	claims := entity.TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
	}

	accessToken, err := uc.tokenGenerator.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenGenerator.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	refreshTokenEntity := &entity.RefreshToken{
		UserID:     user.ID,
		Token:      refreshToken,
		DeviceInfo: deviceInfo.Device,
		IPAddress:  deviceInfo.IPAddress,
		UserAgent:  deviceInfo.UserAgent,
		ExpiresAt:  time.Now().Add(uc.config.RefreshTokenExpiry),
		CreatedAt:  time.Now(),
	}

	if err := uc.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User: UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken generates new access and refresh tokens
func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string, deviceInfo entity.DeviceInfo) (*TokenResponse, error) {
	// Find refresh token in database
	storedToken, err := uc.refreshTokenRepo.FindByToken(ctx, refreshToken)
	if err != nil || storedToken == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Validate token
	if !storedToken.IsValid() {
		return nil, ErrInvalidRefreshToken
	}

	// Get user
	user, err := uc.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	// Revoke old refresh token
	if err := uc.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	// Generate new tokens
	claims := entity.TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
	}

	newAccessToken, err := uc.tokenGenerator.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := uc.tokenGenerator.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	newRefreshTokenEntity := &entity.RefreshToken{
		UserID:     user.ID,
		Token:      newRefreshToken,
		DeviceInfo: deviceInfo.Device,
		IPAddress:  deviceInfo.IPAddress,
		UserAgent:  deviceInfo.UserAgent,
		ExpiresAt:  time.Now().Add(uc.config.RefreshTokenExpiry),
		CreatedAt:  time.Now(),
	}

	if err := uc.refreshTokenRepo.Create(ctx, newRefreshTokenEntity); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout revokes the specified refresh token
func (uc *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	return uc.refreshTokenRepo.RevokeByToken(ctx, refreshToken)
}

// LogoutAll revokes all refresh tokens for a user
func (uc *authUseCase) LogoutAll(ctx context.Context, userID int64) error {
	return uc.refreshTokenRepo.RevokeAllByUserID(ctx, userID)
}

// GetActiveSessions returns all active sessions for a user
func (uc *authUseCase) GetActiveSessions(ctx context.Context, userID int64) ([]*entity.RefreshToken, error) {
	tokens, err := uc.refreshTokenRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter only valid tokens
	var activeSessions []*entity.RefreshToken
	for _, token := range tokens {
		if token.IsValid() {
			activeSessions = append(activeSessions, token)
		}
	}

	return activeSessions, nil
}
