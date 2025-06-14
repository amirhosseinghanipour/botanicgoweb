package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrConfigError  = errors.New("configuration error")
)

type Config struct {
	JWTSecret     string
	TokenDuration time.Duration
	Issuer        string
}

var config Config

// Initialize sets up the auth configuration
func Initialize() error {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return fmt.Errorf("%w: JWT_SECRET environment variable is not set", ErrConfigError)
	}

	// Default to 24 hours if not specified
	tokenDuration := 24 * time.Hour
	if duration := os.Getenv("JWT_DURATION"); duration != "" {
		if parsed, err := time.ParseDuration(duration); err == nil {
			tokenDuration = parsed
		}
	}

	config = Config{
		JWTSecret:     jwtSecret,
		TokenDuration: tokenDuration,
		Issuer:        getEnvOrDefault("JWT_ISSUER", "botanic"),
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string) (string, error) {
	if config.JWTSecret == "" {
		return "", fmt.Errorf("%w: auth not initialized", ErrConfigError)
	}

	expirationTime := time.Now().Add(config.TokenDuration)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    config.Issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

func VerifyToken(tokenString string) (string, error) {
	if config.JWTSecret == "" {
		return "", fmt.Errorf("%w: auth not initialized", ErrConfigError)
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", ErrInvalidToken
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	return claims.UserID, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	if config.JWTSecret == "" {
		return nil, fmt.Errorf("%w: auth not initialized", ErrConfigError)
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
