package jwt

import (
	"github.com/handokobeni/agent-openhands/internal/usecase/auth"
	"golang.org/x/crypto/bcrypt"
)

// passwordHasher implements auth.PasswordHasher using bcrypt
type passwordHasher struct {
	cost int
}

// NewPasswordHasher creates a new bcrypt password hasher
func NewPasswordHasher() auth.PasswordHasher {
	return &passwordHasher{cost: bcrypt.DefaultCost}
}

// Hash generates a bcrypt hash of the password
func (h *passwordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare compares a hashed password with a plain password
func (h *passwordHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
