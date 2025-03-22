package entity

import "github.com/golang-jwt/jwt/v5"

var JwtContextKey = "token"
var UserContextKey = "user"

type TokenClaim struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthDao struct {
	Token string `json:"token"`
}
