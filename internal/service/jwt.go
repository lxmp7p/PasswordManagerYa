package service

import (
	"strconv"
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

type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

func (m *TokenService) Generate(userID int64, login string) (string, error) {
	claims := jwt.MapClaims{
		"Subject": strconv.FormatInt(userID, 10),
		"login":   login,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.secret)
}

func (m *TokenService) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString, claims,
		func(t *jwt.Token) (any, error) {
			return m.secret, nil
		},
	)
	if err != nil {
		return &Claims{}, err
	}
	if !token.Valid {
		return &Claims{}, err
	}
	return claims, err
}
