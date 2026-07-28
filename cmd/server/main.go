package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	server := &http.Server{
		Addr:    ":4443",
		Handler: r,
	}

	log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
}
