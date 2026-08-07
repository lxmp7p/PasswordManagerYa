package service

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestNewTokenService(t *testing.T) {
	service := NewTokenService("secret")

	require.NotNil(t, service)
}

func TestTokenService_GenerateAndParse(t *testing.T) {
	service := NewTokenService("secret")

	token, err := service.Generate(
		1,
		"admin",
	)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := service.Parse(token)

	require.NoError(t, err)
	require.Equal(
		t,
		"admin",
		claims.Login,
	)
}

func TestTokenService_Parse_InvalidToken(t *testing.T) {
	service := NewTokenService("secret")

	claims, err := service.Parse(
		"broken.token.value",
	)

	require.Error(t, err)
	require.Nil(t, claims)
}

func TestTokenService_Parse_WrongSecret(t *testing.T) {
	service1 := NewTokenService("secret1")
	service2 := NewTokenService("secret2")

	token, err := service1.Generate(
		1,
		"admin",
	)

	require.NoError(t, err)

	claims, err := service2.Parse(token)

	require.Error(t, err)
	require.Nil(t, claims)
}

func TestTokenService_Generate_UserID(t *testing.T) {
	service := NewTokenService("secret")

	token, err := service.Generate(
		42,
		"admin",
	)

	require.NoError(t, err)

	claims := &Claims{}

	_, err = jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) {
			return []byte("secret"), nil
		},
	)

	require.NoError(t, err)

	require.Equal(
		t,
		"admin",
		claims.Login,
	)
}
