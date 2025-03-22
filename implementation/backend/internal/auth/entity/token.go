package entity

import "github.com/golang-jwt/jwt/v5"

var JwtContextKey = "token"

type TokenClaim struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
