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
	auth        helper.Auth
}

func SetupUserRoutes(restHandler *rest.RestHandler) {
	app := restHandler.App

	//create an instance of user repository and inject to service
	userRepo := repository.NewUserRepository(restHandler.DB)
	userService := service.NewUserService(userRepo, restHandler.Auth)
	handler := UserHandler{
		userService: userService,
		auth:        restHandler.Auth,
	}

	//public endpoints (no authentication required)
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)

	//private endpoints (authentication required)
	privateRoutes := app.Group("/", restHandler.Auth.Authorize)
	privateRoutes.Get("/users", handler.GetUsers)
	privateRoutes.Get("/users/profile", handler.GetProfile)
	privateRoutes.Get("/users/verify", handler.GetVerificationCode)
	privateRoutes.Post("/users/verify", handler.Verify)
	privateRoutes.Get("/users/:id", handler.FindUserByID)
	privateRoutes.Put("/users/:id", handler.UpdateUser)
	privateRoutes.Delete("/users/:id", handler.DeleteUser)
	privateRoutes.Get("/verify", handler.GetVerificationCode)
	privateRoutes.Post("/verify", handler.Verify)
	// privateRoutes.Get("/profile", handler.GetProfile)
	privateRoutes.Post("/profile", handler.CreateProfile)
	privateRoutes.Put("/profile", handler.UpdateProfile)
	privateRoutes.Delete("/profile", handler.DeleteProfile)
	privateRoutes.Get("/orders", handler.Orders)
	privateRoutes.Get("/orders/:id", handler.GetOrder)
	privateRoutes.Post("/become-seller", handler.BecomeSeller)
	privateRoutes.Get("/addresses", handler.Addresses)
	privateRoutes.Get("/payments", handler.Payments)
	privateRoutes.Get("/reviews", handler.Reviews)
	privateRoutes.Get("/wishlist", handler.Wishlist)
	privateRoutes.Get("/cart", handler.Cart)
	privateRoutes.Get("/checkout", handler.Checkout)
	privateRoutes.Get("/logout", handler.Logout)
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

	// Generate token for the newly registered user
	token, err := h.auth.GenerateToken(createdUser.ID, createdUser.Email, createdUser.UserType)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "User registered but failed to generate token",
		})
	}

	// Create user response without password
	userResponse := fiber.Map{
		"id":         createdUser.ID,
		"first_name": createdUser.FirstName,
		"last_name":  createdUser.LastName,
		"email":      createdUser.Email,
		"phone":      createdUser.Phone,
		"user_type":  createdUser.UserType,
		"verified":   createdUser.Verified,
		"created_at": createdUser.CreatedAt,
		"updated_at": createdUser.UpdatedAt,
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    userResponse,
		"token":   token,
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

	user, token, err := h.userService.Login(loginData.Email, loginData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid email or password",
		})
	}

	// Create user response without password
	userResponse := fiber.Map{
		"id":         user.ID,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"phone":      user.Phone,
		"user_type":  user.UserType,
		"verified":   user.Verified,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "login",
		"user":    userResponse,
		"token":   token,
	})
}

func (h *UserHandler) GetVerificationCode(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	verificationCode, err := h.userService.GetVerificationCode(user.ID, user.Code)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "get verification code",
		"data":    verificationCode,
	})
}

func (h *UserHandler) Verify(ctx *fiber.Ctx) error {

	user := h.auth.GetCurrentUser(ctx)
	verificationCodeInput := dto.VerificationCodeInput{}
	if err := ctx.BodyParser(&verificationCodeInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	isVerified, err := h.userService.VerifyCode(user.ID, verificationCodeInput.Code)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	if !isVerified {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid verification code",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Verified in successfully",
	})
}

func (h *UserHandler) GetProfile(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "get profile",
		"user":    user,
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
