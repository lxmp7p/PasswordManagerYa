package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	//authService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) InitRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

}
