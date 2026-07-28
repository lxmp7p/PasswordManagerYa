package handler

import (
	"passwordmanager/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	auth *AuthHandler
}

func NewHandler(authService service.AuthServiceInterface) *Handler {
	return &Handler{
		auth: NewAuthHandler(authService),
	}
}

func (h *Handler) InitRoutes(r chi.Router) {
	h.auth.InitRoutes(r)
}
