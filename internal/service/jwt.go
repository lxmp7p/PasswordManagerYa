package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	secret []byte
}

func NewTokenService(secret string) TokenManagerInterface {
	return &TokenService{
		secret: []byte(secret),
	}
}

func (m *TokenService) Generate(userID int64, login string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"login": login,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.secret)
}
