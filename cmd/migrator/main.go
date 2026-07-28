package main

import (
	"log"

	repo "passwordmanager/internal/repository/migrator"

	"github.com/pressly/goose/v3"
)

func main() {
	db, err := repo.NewPostgres(
		"localhost",
		5432,
		"app",
		"testpass",
		"passwordmanager")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(db, "../../migrations"); err != nil {
		log.Fatal(err)
	}

	log.Println("Migrations applied")
}
