package handler

import (
	"log/slog"
	"passwordmanager/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	auth  *AuthHandler
	vault *VaultHandler
}

func NewHandler(
	logger *slog.Logger,
	authService service.AuthServiceInterface,
	vaultService service.VaultServiceInterface,
	tokenService service.TokenManagerInterface) *Handler {
	return &Handler{
		auth:  NewAuthHandler(logger, authService),
		vault: NewVaultHandler(logger, vaultService, tokenService),
	}
}

func (h *Handler) InitRoutes(r chi.Router) {
	h.auth.InitRoutes(r)
	h.vault.InitRoutes(r)
}
