package jwt

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`
	jwt.RegisteredClaims
}
