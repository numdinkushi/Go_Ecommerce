package main

import (
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/api"
	"go-ecommerce-app/internal/infra"
	"log"
)

func main() {
	cfg, err := config.SetupEnv()
	if err != nil {
		log.Fatalf("Failed to setup environment: %v", err)
	}

	err = infra.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Starting server...")
	api.StartServer(cfg)
}
