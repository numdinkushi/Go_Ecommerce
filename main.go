package main

import (
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/api"
	"log"
)

func main() {
	config, err := config.SetupEnv()
	if err != nil {
		log.Fatalf("Failed to setup environment: %v", err)
	}
	
	api.StartServer(config)
}