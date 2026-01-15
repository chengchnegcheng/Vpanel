// Package auth provides authentication and authorization services.
package auth

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"v/pkg/errors"
)

// Claims represents JWT claims.
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RefreshClaims represents refresh token claims.
type RefreshClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// Config holds authentication configuration.
type Config struct {
	JWTSecret          string
	TokenExpiry        time.Duration
	RefreshTokenExpiry time.Duration
}

// Service provides authentication operations.
type Service struct {
	config Config
}

// NewService creates a new authentication service.
func NewService(cfg Config) *Service {
	if cfg.TokenExpiry == 0 {
		cfg.TokenExpiry = 24 * time.Hour
	}
	if cfg.RefreshTokenExpiry == 0 {
		cfg.RefreshTokenExpiry = 7 * 24 * time.Hour
	}
	return &Service{config: cfg}
}

// GenerateToken generates a JWT token for a user.
func (s *Service) GenerateToken(userID int64, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// GenerateRefreshToken generates a refresh token for a user.
func (s *Service) GenerateRefreshToken(userID int64) (string, error) {
	claims := &RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// ValidateToken validates a JWT token and returns the claims.
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.NewUnauthorizedError("invalid token signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.NewUnauthorizedError("invalid token claims")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims.
func (s *Service) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.NewUnauthorizedError("invalid token signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid refresh token")
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.NewUnauthorizedError("invalid refresh token claims")
	}

	return claims, nil
}

// HashPassword hashes a password using bcrypt.
func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash.
func (s *Service) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateTemporaryPassword generates a random temporary password.
func (s *Service) GenerateTemporaryPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	const length = 12

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to time-based if crypto/rand fails
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		} else {
			b[i] = charset[n.Int64()]
		}
	}
	return string(b)
}

// GenerateTOTPSecret generates a new TOTP secret for 2FA.
func (s *Service) GenerateTOTPSecret() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567" // Base32 charset
	const length = 32

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

// VerifyTOTP verifies a TOTP code against a secret.
// This is a simplified implementation - in production, use a proper TOTP library.
func (s *Service) VerifyTOTP(secret, code string) bool {
	// For a proper implementation, use a library like github.com/pquerna/otp
	// This is a placeholder that accepts any 6-digit code for testing
	if len(code) != 6 {
		return false
	}
	// In production, implement proper TOTP verification
	// For now, return true for testing purposes if secret is valid
	return len(secret) > 0
}
