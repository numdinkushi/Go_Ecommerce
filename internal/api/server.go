package api

import (
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/api/rest/handlers"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/infra"
	"log"

	"github.com/gofiber/fiber/v2"
)

func StartServer(config config.AppConfig) {
	app := fiber.New()

	db := infra.GetDB()

	// Run database migrations
	err := db.AutoMigrate(&domain.User{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migration completed successfully")

	restHandler := &rest.RestHandler{
		App: app,
		DB:  db,
	}

	setupRoutes(restHandler)

	app.Listen(config.ServerPort)
}

func setupRoutes(restHandler *rest.RestHandler) {
	handlers.SetupUserRoutes(restHandler)
}
