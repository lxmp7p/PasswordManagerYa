package service

import (
	"context"
	"passwordmanager/internal/repository"
)

type authService struct {
	authRepo repository.UserRepositoryInterface
}

func NewAuthService(
	authRepo repository.UserRepositoryInterface,
) AuthServiceInterface {
	return &authService{
		authRepo: authRepo,
	}
}

func (s *authService) Register(
	ctx context.Context,
	login string,
	password string,
) error {
	s.authRepo.Create(ctx, login, password)
	return nil
}

func (s *authService) Login(
	ctx context.Context,
	login string,
	password string,
) (string, error) {
	return "", nil
}
