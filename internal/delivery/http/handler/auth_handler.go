package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/handokobeni/agent-openhands/internal/delivery/http/middleware"
	"github.com/handokobeni/agent-openhands/internal/domain/entity"
	"github.com/handokobeni/agent-openhands/internal/usecase/auth"
	"github.com/handokobeni/agent-openhands/pkg/response"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authUseCase        auth.UseCase
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(authUseCase auth.UseCase, accessTokenExpiry, refreshTokenExpiry time.Duration) *AuthHandler {
	return &AuthHandler{
		authUseCase:        authUseCase,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "invalid request body")
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" || req.Name == "" {
		response.ValidationError(w, "email, password, and name are required")
		return
	}

	if len(req.Password) < 8 {
		response.ValidationError(w, "password must be at least 8 characters")
		return
	}

	result, err := h.authUseCase.Register(r.Context(), req)
	if err != nil {
		if err == auth.ErrUserAlreadyExists {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	response.Success(w, http.StatusCreated, "user registered successfully", result)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.ValidationError(w, "email and password are required")
		return
	}

	deviceInfo := h.extractDeviceInfo(r)

	result, err := h.authUseCase.Login(r.Context(), req, deviceInfo)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			response.Unauthorized(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	// Set access token in httpOnly cookie for security
	h.setAccessTokenCookie(w, result.AccessToken)

	// Set refresh token in httpOnly cookie
	h.setRefreshTokenCookie(w, result.RefreshToken)

	response.Success(w, http.StatusOK, "login successful", result)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Try to get refresh token from cookie first
	var refreshToken string
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		refreshToken = cookie.Value
	}

	// If not in cookie, try request body
	if refreshToken == "" {
		var req auth.RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		response.ValidationError(w, "refresh token is required")
		return
	}

	deviceInfo := h.extractDeviceInfo(r)

	result, err := h.authUseCase.RefreshToken(r.Context(), refreshToken, deviceInfo)
	if err != nil {
		if err == auth.ErrInvalidRefreshToken {
			// Clear cookies on invalid refresh token
			h.clearCookies(w)
			response.Unauthorized(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	// Set new access token in httpOnly cookie
	h.setAccessTokenCookie(w, result.AccessToken)

	// Set new refresh token in httpOnly cookie
	h.setRefreshTokenCookie(w, result.RefreshToken)

	response.Success(w, http.StatusOK, "token refreshed successfully", result)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get refresh token from cookie
	var refreshToken string
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		refreshToken = cookie.Value
	}

	if refreshToken != "" {
		_ = h.authUseCase.Logout(r.Context(), refreshToken)
	}

	// Clear cookies
	h.clearCookies(w)

	response.Success(w, http.StatusOK, "logged out successfully", nil)
}

// LogoutAll handles logout from all devices
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "")
		return
	}

	if err := h.authUseCase.LogoutAll(r.Context(), userID); err != nil {
		response.InternalError(w)
		return
	}

	// Clear cookies
	h.clearCookies(w)

	response.Success(w, http.StatusOK, "logged out from all devices", nil)
}

// GetSessions returns all active sessions for the current user
func (h *AuthHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "")
		return
	}

	sessions, err := h.authUseCase.GetActiveSessions(r.Context(), userID)
	if err != nil {
		response.InternalError(w)
		return
	}

	// Map to response (hide sensitive token data)
	type SessionResponse struct {
		ID         int64     `json:"id"`
		DeviceInfo string    `json:"device_info"`
		IPAddress  string    `json:"ip_address"`
		UserAgent  string    `json:"user_agent"`
		CreatedAt  time.Time `json:"created_at"`
		ExpiresAt  time.Time `json:"expires_at"`
	}

	var sessionResponses []SessionResponse
	for _, s := range sessions {
		sessionResponses = append(sessionResponses, SessionResponse{
			ID:         s.ID,
			DeviceInfo: s.DeviceInfo,
			IPAddress:  s.IPAddress,
			UserAgent:  s.UserAgent,
			CreatedAt:  s.CreatedAt,
			ExpiresAt:  s.ExpiresAt,
		})
	}

	response.Success(w, http.StatusOK, "active sessions retrieved", sessionResponses)
}

// Me returns the current authenticated user info
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "")
		return
	}

	email, _ := middleware.GetUserEmail(r.Context())

	response.Success(w, http.StatusOK, "user info retrieved", map[string]interface{}{
		"user_id": userID,
		"email":   email,
	})
}

// Helper methods

func (h *AuthHandler) extractDeviceInfo(r *http.Request) entity.DeviceInfo {
	// Get real IP (handle proxies)
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	userAgent := r.Header.Get("User-Agent")
	device := parseDeviceFromUserAgent(userAgent)

	return entity.DeviceInfo{
		IPAddress: ip,
		UserAgent: userAgent,
		Device:    device,
	}
}

func (h *AuthHandler) setAccessTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.accessTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/auth/refresh",
		MaxAge:   int(h.refreshTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) clearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func parseDeviceFromUserAgent(userAgent string) string {
	// Simple device detection (can be enhanced with a proper library)
	switch {
	case contains(userAgent, "Mobile"):
		return "Mobile"
	case contains(userAgent, "Tablet"):
		return "Tablet"
	default:
		return "Desktop"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsRune(s, substr))
}

func containsRune(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
