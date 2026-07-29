package handler

import (
	"passwordmanager/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	auth  *AuthHandler
	vault *VaultHandler
}

func NewHandler(
	authService service.AuthServiceInterface,
	vaultService service.VaultServiceInterface,
	tokenService service.TokenManagerInterface) *Handler {
	return &Handler{
		auth:  NewAuthHandler(authService),
		vault: NewVaultHandler(vaultService, tokenService),
	}
}

func (h *Handler) InitRoutes(r chi.Router) {
	h.auth.InitRoutes(r)
	h.vault.InitRoutes(r)
}
