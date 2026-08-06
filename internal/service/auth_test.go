package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"passwordmanager/internal/repository"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(
	ctx context.Context,
	login string,
	passwordHash string,
) error {
	args := m.Called(ctx, login, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) FindByLogin(
	ctx context.Context,
	login string,
) (*repository.User, error) {
	args := m.Called(ctx, login)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*repository.User), args.Error(1)
}

type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) Generate(
	id int64,
	login string,
) (string, error) {
	args := m.Called(id, login)

	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) Parse(
	token string,
) (*Claims, error) {
	args := m.Called(token)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*Claims), args.Error(1)
}

func TestAuthService_Register_Success(t *testing.T) {

	repo := new(MockUserRepository)
	tokens := new(MockTokenManager)

	service := NewAuthService(
		slog.New(
			slog.NewTextHandler(
				os.Stdout,
				nil,
			),
		),
		repo,
		tokens,
	)

	repo.
		On(
			"Create",
			mock.Anything,
			"admin",
			mock.Anything,
		).
		Return(nil)

	err := service.Register(
		context.Background(),
		"admin",
		"password",
	)

	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := new(MockUserRepository)
	tokens := new(MockTokenManager)

	service := NewAuthService(
		slog.Default(),
		repo,
		tokens,
	)

	repo.
		On(
			"FindByLogin",
			mock.Anything,
			"admin",
		).
		Return(
			nil,
			errors.New("not found"),
		)

	token, err := service.Login(
		context.Background(),
		"admin",
		"password",
	)

	require.Error(t, err)
	require.Empty(t, token)

	repo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := new(MockUserRepository)
	tokens := new(MockTokenManager)

	service := NewAuthService(
		slog.Default(),
		repo,
		tokens,
	)

	hash, _ := bcrypt.GenerateFromPassword(
		[]byte("correct"),
		bcrypt.DefaultCost,
	)

	repo.
		On(
			"FindByLogin",
			mock.Anything,
			"admin",
		).
		Return(
			&repository.User{
				ID:           1,
				Login:        "admin",
				PasswordHash: string(hash),
			},
			nil,
		)

	token, err := service.Login(
		context.Background(),
		"admin",
		"wrong",
	)

	require.Error(t, err)
	require.Empty(t, token)
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := new(MockUserRepository)
	tokens := new(MockTokenManager)

	service := NewAuthService(
		slog.Default(),
		repo,
		tokens,
	)

	hash, _ := bcrypt.GenerateFromPassword(
		[]byte("password"),
		bcrypt.DefaultCost,
	)

	repo.
		On(
			"FindByLogin",
			mock.Anything,
			"admin",
		).
		Return(
			&repository.User{
				ID:           1,
				Login:        "admin",
				PasswordHash: string(hash),
			},
			nil,
		)

	tokens.
		On(
			"Generate",
			int64(1),
			"admin",
		).
		Return(
			"jwt-token",
			nil,
		)

	token, err := service.Login(
		context.Background(),
		"admin",
		"password",
	)

	require.NoError(t, err)
	require.Equal(t, "jwt-token", token)

	tokens.AssertExpectations(t)
}
