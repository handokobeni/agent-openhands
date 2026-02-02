package entity

// TokenPair represents the access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenClaims represents the claims stored in JWT
type TokenClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

// DeviceInfo represents device information for token tracking
type DeviceInfo struct {
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Device    string `json:"device"`
}
