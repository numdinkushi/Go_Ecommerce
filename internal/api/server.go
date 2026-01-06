package api

import (
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/api/rest/handlers"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/infra"
	"go-ecommerce-app/internal/service"
	"go-ecommerce-app/pkg/external/flutterwave"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	// Enable CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Length",
		MaxAge:           3600,
	}))

	db := infra.GetDB()

	// Run database migrations
	err := db.AutoMigrate(
		&domain.User{},
		&domain.BankAccount{},
		&domain.Category{},
		&domain.Product{},
		&domain.Cart{},
		&domain.Address{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.Transaction{},
		&domain.Payment{},
	)

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Clean up: Drop quantity column if it exists (we use qty instead)
	err = db.Exec(`
		DO $$ 
		BEGIN 
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'order_items' AND column_name = 'quantity'
			) THEN
				ALTER TABLE order_items DROP COLUMN quantity;
				RAISE NOTICE 'Column quantity dropped (using qty instead)';
			END IF;
		END $$;
	`).Error

	if err != nil {
		log.Printf("⚠️  Warning: Could not drop quantity column: %v", err)
	} else {
		log.Println("✅ Database cleanup: removed duplicate quantity column")
	}

	auth := helper.SetupAuth(config.JwtSecret)
	log.Println("✅ Database migration completed successfully")

	// Initialize external services
	var bankService *service.BankService
	if config.FlutterwaveSecretKey != "" {
		flutterwaveClient := flutterwave.NewClient(config.FlutterwaveSecretKey)
		bankService = service.NewBankService(flutterwaveClient)
		log.Println("✅ Bank verification service initialized (Flutterwave)")
	} else {
		log.Println("⚠️  FLUTTERWAVE_SECRET_KEY not set - bank verification features disabled")
	}

	restHandler := &rest.RestHandler{
		App:    app,
		DB:     db,
		Auth:   auth,
		Config: config,
	}

	setupRoutes(restHandler, bankService)

	log.Printf("🚀 Server starting on port %s", config.ServerPort)
	if err := app.Listen(config.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(restHandler *rest.RestHandler, bankService *service.BankService) {
	handlers.SetupUserRoutes(restHandler, bankService)
	handlers.SetupBankRoutes(restHandler, bankService)
	handlers.SetupCatalogueRoutes(restHandler, bankService)
	handlers.SetupTransactionRoutes(restHandler, bankService)
}
