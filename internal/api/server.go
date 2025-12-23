package api

import (
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/api/rest/handlers"
	"go-ecommerce-app/internal/infra"

	"github.com/gofiber/fiber/v2"
)

func StartServer(config config.AppConfig) {
	app := fiber.New()

	restHandler := &rest.RestHandler{
		App: app,
		DB:  infra.GetDB(),
	}

	setupRoutes(restHandler)

	app.Listen(config.ServerPort)
}

func setupRoutes(restHandler *rest.RestHandler) {
	handlers.SetupUserRoutes(restHandler)
}
