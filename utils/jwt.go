package utils

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims defines custom JWT claims including user ID, email, and role
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GetJWTSecret returns the secret key from environment or fallback
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("⚠️ WARNING: JWT_SECRET environment variable is not set. Using default secret for development.")
		secret = "findit_jwt_secret_key_default_change_me_in_prod"
	}
	return []byte(secret)
}

// GenerateToken creates a signed JWT access token for a user
func GenerateToken(userID uint, email string, role string) (string, error) {
	expHours := 24
	if envExp := os.Getenv("JWT_EXPIRATION_HOURS"); envExp != "" {
		if val, err := strconv.Atoi(envExp); err == nil && val > 0 {
			expHours = val
		}
	}

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "findit-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecret())
}

// ValidateToken parses and validates a JWT token string
func ValidateToken(tokenStr string) (*JWTClaims, error) {
	if tokenStr == "" {
		return nil, errors.New("token string is empty")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return GetJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid or expired token")
}
