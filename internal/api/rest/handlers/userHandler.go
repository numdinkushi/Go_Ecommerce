package handlers

import (
	"go-ecommerce-app/internal/api/rest"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	// service UserService
}


func SetupUserRoutes(restHandler *rest.RestHandler) {
	app := restHandler.App

	//create an instance of user service and inject to handler
	handler := UserHandler{}

	//public endpoints
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)

	//private endpoints
	app.Get("/verify", handler.Verify)
	app.Get("/profile", handler.Profile)
	app.Post("/profile", handler.CreateProfile)
	app.Put("/profile", handler.UpdateProfile)
	app.Delete("/profile", handler.DeleteProfile)
	app.Get("/orders", handler.Orders)
	app.Get("/orders/:id", handler.GetOrder)
	app.Post("/become-seller", handler.BecomeSeller)
	app.Get("verification-code", handler.VerificationCode)
	app.Get("/addresses", handler.Addresses)
	app.Get("/payments", handler.Payments)
	app.Get("/reviews", handler.Reviews)
	app.Get("/wishlist", handler.Wishlist)
	app.Get("/cart", handler.Cart)
	app.Get("/checkout", handler.Checkout)
	app.Get("/logout", handler.Logout)
}

func (h *UserHandler) Register(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Registered successfully",
	})
}

func (h *UserHandler) Login(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}


func (h *UserHandler) Verify(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Verified in successfully",
	})
}

func (h *UserHandler) Profile(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile fetched successfully",
	})
}

func (h *UserHandler) CreateProfile(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile created successfully",
	})
}

func (h *UserHandler) UpdateProfile(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile updated successfully",
	})
}

func (h *UserHandler) DeleteProfile(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile deleted successfully",
	})
}
func (h *UserHandler) Orders(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Orders fetched successfully",
	})
}

func (h *UserHandler) GetOrder(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order fetched successfully",
	})
}

func (h *UserHandler) BecomeSeller(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Become seller successfully",
	})
}

func (h *UserHandler) Addresses(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Addresses fetched successfully",
	})
}

func (h *UserHandler) Payments(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payments fetched successfully",
	})
}

func (h *UserHandler) Reviews(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Reviews fetched successfully",
	})
}

func (h *UserHandler) Wishlist(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Wishlist fetched successfully",
	})
}

func (h *UserHandler) Cart(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cart fetched successfully",
	})
}

func (h *UserHandler) Checkout(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Checkout fetched successfully",
	})
}

func (h *UserHandler) Logout(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func (h *UserHandler) VerificationCode(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Verification code fetched successfully",
	})
}