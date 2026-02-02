package entity

import "time"

// RefreshToken represents the refresh token entity stored in database
type RefreshToken struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Token        string     `json:"token"`
	DeviceInfo   string     `json:"device_info"`
	IPAddress    string     `json:"ip_address"`
	UserAgent    string     `json:"user_agent"`
	ExpiresAt    time.Time  `json:"expires_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked checks if the refresh token has been revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsValid checks if the refresh token is still valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked()
}
