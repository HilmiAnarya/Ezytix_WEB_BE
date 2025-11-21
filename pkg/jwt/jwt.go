package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateAccessToken(userID uint, role, email, phone string) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Role:   role,
		Email:  email,
		Phone:  phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return "", errors.New("JWT_SECRET is missing")
	}

	return token.SignedString([]byte(secret))
}

func CreateRefreshToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_REFRESH_SECRET")

	if secret == "" {
		return "", errors.New("JWT_REFRESH_SECRET is missing")
	}

	return token.SignedString([]byte(secret))
}

func ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		return nil, errors.New("JWT_SECRET is missing")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || token == nil {
		return nil, errors.New("invalid or expired access token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid or malformed access token")
	}

	return claims, nil
}

func ValidateRefreshToken(tokenString string) (uint, error) {
	secret := []byte(os.Getenv("JWT_REFRESH_SECRET"))
	if len(secret) == 0 {
		return 0, errors.New("JWT_REFRESH_SECRET is missing")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid or expired refresh token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid refresh token claims")
	}

	rawUserID, ok := mapClaims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id in refresh token")
	}

	return uint(rawUserID), nil
}