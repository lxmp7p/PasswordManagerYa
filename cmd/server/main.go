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

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			nil,
		),
	)

	db, err := repository.New(context.Background(), "host=localhost port=5432 user=app password=testpass dbname=passwordmanager")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	userRepository := repository.NewUserRepository(db.Pool)
	authService := service.NewAuthService(userRepository)
	h := handler.NewHandler(authService)

	h.InitRoutes(r)

	server := &http.Server{
		Addr:    ":4443",
		Handler: r,
	}

	log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
}
