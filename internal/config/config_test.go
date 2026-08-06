package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Success(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "password")
	t.Setenv("DB_NAME", "vault")

	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("JWT_SECRET", "secret")

	cfg, err := Load()

	assert.NoError(t, err)

	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Equal(t, "password", cfg.DBPassword)
	assert.Equal(t, "vault", cfg.DBName)

	assert.Equal(t, "8080", cfg.ServerPort)
	assert.Equal(t, "secret", cfg.JWTSecret)
}

func TestDBConnectionString(t *testing.T) {
	cfg := Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "password",
		DBName:     "vault",
	}

	result := cfg.DBConnectionString()

	expected := "host=localhost port=5432 user=postgres password=password dbname=vault"

	assert.Equal(
		t,
		expected,
		result,
	)
}

func TestLoad_EmptyEnv(t *testing.T) {
	envs := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"SERVER_PORT",
		"JWT_SECRET",
	}

	for _, env := range envs {
		os.Unsetenv(env)
	}

	cfg, err := Load()

	assert.NoError(t, err)

	assert.Empty(t, cfg.DBHost)
	assert.Empty(t, cfg.DBPort)
	assert.Empty(t, cfg.DBUser)
	assert.Empty(t, cfg.DBPassword)
	assert.Empty(t, cfg.DBName)
	assert.Empty(t, cfg.ServerPort)
	assert.Empty(t, cfg.JWTSecret)
}
