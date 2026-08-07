package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	defer db.Close()

	userRepository := repository.NewUserRepository(db.Pool)
	vaultRepository := repository.NewVaultRepository(db.Pool)

	tokenService := service.NewTokenService(cfg.JWTSecret)
	authService := service.NewAuthService(logger, userRepository, tokenService)
	vaultService := service.NewVaultService(logger, vaultRepository)

	h := handler.NewHandler(logger, authService, vaultService, tokenService)

	h.InitRoutes(r)

	server := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServeTLS("server.crt", "server.key"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error:", "error", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
