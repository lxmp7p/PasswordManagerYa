package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	ServerPort string

	JWTSecret          string
	VaultEncryptionKey string
}

func Load() (*Config, error) {

	_ = godotenv.Load()

	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),

		ServerPort: os.Getenv("SERVER_PORT"),

		JWTSecret:          os.Getenv("JWT_SECRET"),
		VaultEncryptionKey: os.Getenv("ENCRYPTION_KEY"),
	}, nil
}

func (c Config) DBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
	)
}
