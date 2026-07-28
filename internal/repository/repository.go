package repository

import (
	"context"
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

type UserRepositoryInterface interface {
	Create(ctx context.Context, login string, passwordHash string) error
	FindByLogin(ctx context.Context, login string) (*User, error)
}
