package handlers

import (
	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/internal/service"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	// service UserService
	userService service.UserService
}

func SetupUserRoutes(restHandler *rest.RestHandler) {
	app := restHandler.App

	//create an instance of user repository and inject to service
	userRepo := repository.NewUserRepository(restHandler.DB)
	userService := service.NewUserService(userRepo)
	handler := UserHandler{userService: userService}

	//public endpoints
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)

	//private endpoints
	app.Get("/users", handler.GetUsers)
	app.Get("/users/:id", handler.FindUserByID)
	app.Put("/users/:id", handler.UpdateUser)
	app.Delete("/users/:id", handler.DeleteUser)
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
	user := dto.UserSignUp{}
	err := ctx.BodyParser(&user)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Please provide all the required fields",
		})
	}

	createdUser, err := h.userService.Register(user)

	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    createdUser,
	})
}

func (h *UserHandler) GetUsers(ctx *fiber.Ctx) error {
	email := ctx.Query("email")

	// If email query parameter is provided, find user by email
	if email != "" {
		user, err := h.userService.FindUserByEmail(email)
		if err != nil {
			return helper.HandleDBError(ctx, err)
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "User found",
			"user":    user,
		})
	}

	// Otherwise, return all users
	users, err := h.userService.FindAllUsers()
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Users retrieved successfully",
		"users":   users,
		"count":   len(users),
	})
}

func (h *UserHandler) FindUserByID(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid user ID")
	}

	user, err := h.userService.FindUserByID(uint(id))
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User found",
		"user":    user,
	})
}

func (h *UserHandler) UpdateUser(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid user ID")
	}

	updateData := dto.UserUpdate{}
	if err := ctx.BodyParser(&updateData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	updatedUser, err := h.userService.UpdateUser(uint(id), updateData)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User updated successfully",
		"user":    updatedUser,
	})
}

func (h *UserHandler) DeleteUser(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid user ID")
	}

	err = h.userService.DeleteUser(uint(id))
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

func (h *UserHandler) Login(ctx *fiber.Ctx) error {
	loginData := dto.UserLogin{}
	if err := ctx.BodyParser(&loginData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "please provide valid inputs",
		})
	}

	token, err := h.userService.Login(loginData.Email, loginData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "please provide correct user id password",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "login",
		"token":   token,
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
