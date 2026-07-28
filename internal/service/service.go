package service

import "context"

type AuthServiceInterface interface {
	Register(ctx context.Context, login string, password string) error
	Login(ctx context.Context, login string, password string) (string, error)
}
