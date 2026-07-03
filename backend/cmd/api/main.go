package main

import (
	"log"

	"github.com/jabahum/keycloak-onboarder/backend/internal/config"
	"github.com/jabahum/keycloak-onboarder/backend/internal/database"
	"github.com/jabahum/keycloak-onboarder/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srv := server.New(cfg, db)

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
