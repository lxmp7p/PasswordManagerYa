package handler

import "github.com/go-chi/chi/v5"

type Handler struct {
	auth *AuthHandler
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) InitHandler(r chi.Router) {
	h.auth.InitRoutes(r)
}
