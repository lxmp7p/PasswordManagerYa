package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"passwordmanager/internal/handler"
	"passwordmanager/internal/repository"
	"passwordmanager/internal/service"

	"passwordmanager/internal/config"

	"github.com/go-chi/chi/v5"
)

var secretToken string

func main() {
	r := chi.NewRouter()

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			nil,
		),
	)

	cfg, err := config.Load()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := repository.New(context.Background(), cfg.DBConnectionString())
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	userRepository := repository.NewUserRepository(db.Pool)
	vaultRepository := repository.NewVaultRepository(db.Pool)

	tokenService := service.NewTokenService(cfg.JWTSecret)
	authService := service.NewAuthService(logger, userRepository, tokenService)
	vaultService := service.NewVaultService(logger, vaultRepository)

	h := handler.NewHandler(authService, vaultService, tokenService)

	h.InitRoutes(r)

	server := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: r,
	}

	log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
}
