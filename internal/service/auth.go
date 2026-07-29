package service

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"passwordmanager/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	logger   *slog.Logger
	authRepo repository.UserRepositoryInterface
	tokens   TokenManagerInterface
}

func NewAuthService(
	logger *slog.Logger,
	authRepo repository.UserRepositoryInterface,
	tokens TokenManagerInterface,
) AuthServiceInterface {
	return &authService{
		logger:   logger,
		authRepo: authRepo,
		tokens:   tokens,
	}
}

func (s *authService) Register(ctx context.Context, login string, password string) error {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	err = s.authRepo.Create(ctx, login, string(passHash))
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return nil
}

func (s *authService) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := s.authRepo.FindByLogin(ctx, login)
	if err != nil {
		s.logger.Error(err.Error())
		return "", err
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		s.logger.Error(err.Error())
		return "", errors.New("invalid credentials")
	}
	jwt, err := s.tokens.Generate(user.ID, user.Login)
	if err != nil {
		s.logger.Error(err.Error())
		return "", errors.New(http.StatusText(http.StatusInternalServerError))
	}
	return jwt, nil
}
