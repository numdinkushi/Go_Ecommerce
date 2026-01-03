package handlers

import (
	"strings"

	"go-ecommerce-app/config"
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
	config      config.AppConfig
}

func SetupUserRoutes(restHandler *rest.RestHandler, bankService *service.BankService) {
	app := restHandler.App

	//create an instance of user repository and inject to service
	userRepo := repository.NewUserRepository(restHandler.DB)
	catalogueRepo := repository.NewCatalogueRepository(restHandler.DB)
	userService := service.NewUserService(userRepo, catalogueRepo, restHandler.Auth, restHandler.Config, bankService)
	handler := UserHandler{
		userService: userService,
		auth:        restHandler.Auth,
		config:      restHandler.Config,
	}

	//public endpoints (no authentication required)
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)

	//private endpoints (authentication required)
	privateRoutes := app.Group("/", restHandler.Auth.Authorize)
	privateRoutes.Get("/users", handler.GetUsers)
	privateRoutes.Get("/users/profile", handler.GetProfile)
	privateRoutes.Post("/users/profile", handler.CreateProfile)
	privateRoutes.Patch("/users/profile", handler.UpdateProfile)
	privateRoutes.Get("/users/verify", handler.GetVerificationCode)
	privateRoutes.Post("/users/verify", handler.Verify)
	privateRoutes.Get("/users/:id", handler.FindUserByID)
	privateRoutes.Put("/users/:id", handler.UpdateUser)
	privateRoutes.Delete("/users/:id", handler.DeleteUser)
	privateRoutes.Get("/verify", handler.GetVerificationCode)
	privateRoutes.Post("/verify", handler.Verify)
	privateRoutes.Delete("/profile", handler.DeleteProfile)
	privateRoutes.Get("/orders", handler.Orders)
	privateRoutes.Get("/orders/:id", handler.GetOrder)
	privateRoutes.Post("/become-seller", handler.BecomeSeller)
	privateRoutes.Get("/addresses", handler.Addresses)
	privateRoutes.Get("/payments", handler.Payments)
	privateRoutes.Get("/reviews", handler.Reviews)
	privateRoutes.Get("/wishlist", handler.Wishlist)
	privateRoutes.Get("/cart/:product_id", handler.GetCartItem)
	privateRoutes.Get("/cart", handler.GetCartItems)
	privateRoutes.Post("/cart", handler.AddToCart)
	privateRoutes.Patch("/cart/:product_id/increment", handler.IncrementCartItem)
	privateRoutes.Patch("/cart/:product_id/decrement", handler.DecrementCartItem)
	privateRoutes.Put("/cart", handler.UpdateCart)
	privateRoutes.Delete("/cart/:product_id", handler.DeleteCartItem)
	privateRoutes.Delete("/cart", handler.ClearCart)
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
			"message": "User found",
			"user":    userResponse,
		})
	}

	// Otherwise, return all users
	users, err := h.userService.FindAllUsers()
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	usersResponse := make([]fiber.Map, len(users))
	for i, user := range users {
		usersResponse[i] = fiber.Map{
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
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Users retrieved successfully",
		"users":   usersResponse,
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
		"message": "User found",
		"user":    userResponse,
	})
}

func (h *UserHandler) UpdateUser(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid user ID")
	}

	var body struct {
		dto.UserUpdate
		UserType *string `json:"user_type,omitempty"`
	}
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if body.UserType != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "user_type cannot be updated through this endpoint",
			"error":   "To become a seller, please use the /become-seller endpoint which includes required verification steps",
		})
	}

	updatedUser, err := h.userService.UpdateUser(uint(id), body.UserUpdate)
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

	err := h.userService.GetVerificationCode(user.ID)
	if err != nil {
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "SMS") || strings.Contains(errorMessage, "Twilio") || strings.Contains(errorMessage, "phone") {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Failed to send verification code",
				"error":   errorMessage,
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Verification code sent successfully",
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
	profile, err := h.userService.GetProfile(user.ID)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	profileResponse := fiber.Map{
		"id":         profile.ID,
		"first_name": profile.FirstName,
		"last_name":  profile.LastName,
		"email":      profile.Email,
		"phone":      profile.Phone,
		"user_type":  profile.UserType,
		"verified":   profile.Verified,
		"created_at": profile.CreatedAt,
		"updated_at": profile.UpdatedAt,
	}

	if profile.Address.ID != 0 {
		profileResponse["address"] = fiber.Map{
			"id":            profile.Address.ID,
			"address_line1": profile.Address.AddressLine1,
			"address_line2": profile.Address.AddressLine2,
			"city":          profile.Address.City,
			"state":         profile.Address.State,
			"country":       profile.Address.Country,
			"postal_code":   profile.Address.PostalCode,
			"created_at":    profile.Address.CreatedAt,
			"updated_at":    profile.Address.UpdatedAt,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile retrieved successfully",
		"profile": profileResponse,
	})
}

func (h *UserHandler) CreateProfile(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	profileInput := dto.ProfileInput{}
	if err := ctx.BodyParser(&profileInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if profileInput.FirstName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"error":   "first_name is required",
		})
	}

	if profileInput.LastName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"error":   "last_name is required",
		})
	}

	if profileInput.Address.AddressLine1 == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"error":   "address.address_line1 is required",
		})
	}

	if profileInput.Address.City == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"error":   "address.city is required",
		})
	}

	if profileInput.Address.State == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"error":   "address.state is required",
		})
	}

	if profileInput.Address.Country == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"error":   "address.country is required",
		})
	}

	if profileInput.Address.PostalCode == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"error":   "address.postal_code is required",
		})
	}

	profile, err := h.userService.CreateProfile(user.ID, profileInput)
	if err != nil {
		if err.Error() == "profile already exists, use update endpoint instead" {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Profile already exists",
				"error":   "Use PATCH /users/profile to update your profile",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	profileResponse := fiber.Map{
		"id":         profile.ID,
		"first_name": profile.FirstName,
		"last_name":  profile.LastName,
		"email":      profile.Email,
		"phone":      profile.Phone,
		"user_type":  profile.UserType,
		"verified":   profile.Verified,
		"created_at": profile.CreatedAt,
		"updated_at": profile.UpdatedAt,
	}

	if profile.Address.ID != 0 {
		profileResponse["address"] = fiber.Map{
			"id":            profile.Address.ID,
			"address_line1": profile.Address.AddressLine1,
			"address_line2": profile.Address.AddressLine2,
			"city":          profile.Address.City,
			"state":         profile.Address.State,
			"country":       profile.Address.Country,
			"postal_code":   profile.Address.PostalCode,
			"created_at":    profile.Address.CreatedAt,
			"updated_at":    profile.Address.UpdatedAt,
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Profile created successfully",
		"profile": profileResponse,
	})
}

func (h *UserHandler) UpdateProfile(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	profileInput := dto.ProfileUpdateInput{}
	if err := ctx.BodyParser(&profileInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	profile, err := h.userService.UpdateProfile(user.ID, profileInput)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	profileResponse := fiber.Map{
		"id":         profile.ID,
		"first_name": profile.FirstName,
		"last_name":  profile.LastName,
		"email":      profile.Email,
		"phone":      profile.Phone,
		"user_type":  profile.UserType,
		"verified":   profile.Verified,
		"created_at": profile.CreatedAt,
		"updated_at": profile.UpdatedAt,
	}

	if profile.Address.ID != 0 {
		profileResponse["address"] = fiber.Map{
			"id":            profile.Address.ID,
			"address_line1": profile.Address.AddressLine1,
			"address_line2": profile.Address.AddressLine2,
			"city":          profile.Address.City,
			"state":         profile.Address.State,
			"country":       profile.Address.Country,
			"postal_code":   profile.Address.PostalCode,
			"created_at":    profile.Address.CreatedAt,
			"updated_at":    profile.Address.UpdatedAt,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile updated successfully",
		"profile": profileResponse,
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
	user := h.auth.GetCurrentUser(ctx)

	becomeSellerInput := dto.BecomeSellerInput{}
	if err := ctx.BodyParser(&becomeSellerInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	updatedUser, token, err := h.userService.BecomeSeller(user.ID, becomeSellerInput)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	// Create user response without password
	userResponse := fiber.Map{
		"id":         updatedUser.ID,
		"first_name": updatedUser.FirstName,
		"last_name":  updatedUser.LastName,
		"email":      updatedUser.Email,
		"phone":      updatedUser.Phone,
		"user_type":  updatedUser.UserType,
		"verified":   updatedUser.Verified,
		"created_at": updatedUser.CreatedAt,
		"updated_at": updatedUser.UpdatedAt,
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Become seller successfully",
		"user":    userResponse,
		"token":   token,
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

func (h *UserHandler) GetCartItems(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	cartItems, err := h.userService.FindCartItems(user.ID)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cart items fetched successfully",
		"cart":    cartItems,
	})
}

func (h *UserHandler) GetCartItem(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	productID, err := ctx.ParamsInt("product_id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid product ID")
	}

	cartItem, err := h.userService.GetCartItem(user.ID, uint(productID))
	if err != nil {
		if err.Error() == "cart item not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Cart item not found",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cart item retrieved successfully",
		"cart":    cartItem,
	})
}

func (h *UserHandler) AddToCart(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	var request dto.CreateCartRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	cartItem, err := h.userService.AddToCart(user.ID, request)
	if err != nil {
		if err.Error() == "product not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Product not found",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Item added to cart successfully",
		"cart":    cartItem,
	})
}

func (h *UserHandler) UpdateCart(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	var request dto.UpdateCartRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	cartItem, err := h.userService.UpdateCart(user.ID, request)
	if err != nil {
		if err.Error() == "product ID is required" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Product ID is required",
			})
		}
		if err.Error() == "cart item not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Cart item not found",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cart item updated successfully",
		"cart":    cartItem,
	})
}

func (h *UserHandler) DeleteCartItem(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	productID, err := ctx.ParamsInt("product_id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid product ID")
	}

	err = h.userService.DeleteCartItem(user.ID, uint(productID))
	if err != nil {
		if err.Error() == "cart item not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Cart item not found",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cart item deleted successfully",
	})
}

func (h *UserHandler) ClearCart(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	err := h.userService.ClearCart(user.ID)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cart cleared successfully",
	})
}

func (h *UserHandler) IncrementCartItem(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	productID, err := ctx.ParamsInt("product_id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid product ID")
	}

	cartItem, err := h.userService.IncrementCartItem(user.ID, uint(productID))
	if err != nil {
		if err.Error() == "cart item not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Cart item not found",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cart item quantity incremented successfully",
		"cart":    cartItem,
	})
}

func (h *UserHandler) DecrementCartItem(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	productID, err := ctx.ParamsInt("product_id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid product ID")
	}

	cartItem, err := h.userService.DecrementCartItem(user.ID, uint(productID))
	if err != nil {
		if err.Error() == "cart item not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Cart item not found",
			})
		}
		if err.Error() == "quantity cannot be less than 1. Use delete to remove item" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Quantity cannot be less than 1. Use delete to remove item",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cart item quantity decremented successfully",
		"cart":    cartItem,
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
