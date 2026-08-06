package handler

import (
	"context"
	"passwordmanager/internal/dto"
	"passwordmanager/internal/service"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

var _ service.AuthServiceInterface = (*MockAuthService)(nil)

type MockVaultService struct{}

var _ service.VaultServiceInterface = (*MockVaultService)(nil)

func (m *MockVaultService) Create(
	ctx context.Context,
	userID int64,
	item service.VaultCreate,
) (int64, error) {
	return 0, nil
}

func (m *MockVaultService) Get(
	ctx context.Context,
	itemID int64,
	userID int64,
) (dto.VaultResponseGet, error) {
	return dto.VaultResponseGet{}, nil
}

func (m *MockVaultService) List(
	ctx context.Context,
	userID int64,
) ([]dto.VaultResponseGet, error) {
	return nil, nil
}

type MockTokenService struct{}

func (m *MockTokenService) Generate(
	userID int64,
	login string,
) (string, error) {
	return "test-token", nil
}

func (m *MockTokenService) Parse(
	login string,
) (*service.Claims, error) {
	return &service.Claims{
		Login: login,
	}, nil
}

var _ service.TokenManagerInterface = (*MockTokenService)(nil)

func TestNewHandler(t *testing.T) {
	authService := new(MockAuthService)
	vaultService := new(MockVaultService)
	tokenService := new(MockTokenService)

	h := NewHandler(
		authService,
		vaultService,
		tokenService,
	)

	require.NotNil(t, h)
	require.NotNil(t, h.auth)
	require.NotNil(t, h.vault)
}

func TestHandler_InitRoutes(t *testing.T) {
	authService := new(MockAuthService)
	vaultService := new(MockVaultService)
	tokenService := new(MockTokenService)

	h := NewHandler(
		authService,
		vaultService,
		tokenService,
	)

	r := chi.NewRouter()

	require.NotPanics(t, func() {
		h.InitRoutes(r)
	})
}
