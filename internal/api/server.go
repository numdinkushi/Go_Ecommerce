package api

import (
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/api/rest/handlers"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/infra"
	"log"

	"github.com/gofiber/fiber/v2"
)

func StartServer(config config.AppConfig) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			if code == fiber.StatusNotFound {
				return c.Status(code).JSON(fiber.Map{
					"message": "Route not found",
					"error":   "The requested endpoint does not exist",
					"path":    c.Path(),
					"method":  c.Method(),
				})
			}

			return c.Status(code).JSON(fiber.Map{
				"message": "An error occurred",
				"error":   err.Error(),
			})
		},
	})

	db := infra.GetDB()

	// Run database migrations
	err := db.AutoMigrate(&domain.User{})

	auth := helper.SetupAuth(config.JwtSecret)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migration completed successfully")

	restHandler := &rest.RestHandler{
		App: app,
		DB:  db,
		Auth: auth,
	}

	setupRoutes(restHandler)

	app.Listen(config.ServerPort)
}

func setupRoutes(restHandler *rest.RestHandler) {
	handlers.SetupUserRoutes(restHandler)
}
