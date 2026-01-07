package service

import (
	"errors"
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/pkg/notification"
	"log"
	"strconv"
	"time"
)

type UserService struct {
	Repo          repository.UserRepository
	CatalogueRepo repository.CatalogueRepository
	Auth          helper.Auth
	Config        config.AppConfig
	BankService   *BankService
}

func NewUserService(repo repository.UserRepository, catalogueRepo repository.CatalogueRepository, auth helper.Auth, config config.AppConfig, bankService *BankService) UserService {
	return UserService{
		Repo:          repo,
		CatalogueRepo: catalogueRepo,
		Auth:          auth,
		Config:        config,
		BankService:   bankService,
	}
}

func (s UserService) Register(user dto.UserSignUp) (*domain.User, error) {
	log.Println("Registering user", user)

	hashedPassword, err := s.Auth.CreateHashedPassword(user.Password)
	if err != nil {
		return nil, errors.New("failed to create hashed password")
	}
	// user.Password = hashedPassword

	createdUser, err := s.Repo.CreateUser(&domain.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Password:  hashedPassword,
	})
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s UserService) Login(email string, password string) (*domain.User, string, error) {
	user, err := s.Repo.FindUserByEmail(email)
	if err != nil {
		return nil, "", errors.New("user does not exist with the provided email id")
	}

	// Verify plain text password against hashed password from database
	// password: plain text from login request
	// user.Password: bcrypt hash stored in database
	isValidPassword, err := s.Auth.VerifyPassword(password, user.Password)
	if err != nil {
		return nil, "", errors.New("invalid password")
	}

	if !isValidPassword {
		return nil, "", errors.New("invalid password")
	}

	token, err := s.Auth.GenerateToken(user.ID, user.Email, user.UserType)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return user, token, nil

	// compare password and generate token
}

func (s UserService) FindUserByEmail(email string) (*domain.User, error) {
	user, err := s.Repo.FindUserByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s UserService) FindUserByID(id uint) (*domain.User, error) {
	user, err := s.Repo.FindUserByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s UserService) FindAllUsers() ([]domain.User, error) {
	users, err := s.Repo.FindAllUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s UserService) UpdateUser(id uint, updateData dto.UserUpdate) (*domain.User, error) {

	// Check if user exists
	existingUser, err := s.Repo.FindUserByID(id)
	if err != nil {
		return nil, err
	}

	// If email is being updated, check if it's already taken by another user
	if updateData.Email != nil && *updateData.Email != existingUser.Email {
		userWithEmail, err := s.Repo.FindUserByEmail(*updateData.Email)
		if err == nil && userWithEmail != nil && userWithEmail.ID != id {
			return nil, errors.New("email already exists")
		}
	}

	updateUser := domain.User{}

	if updateData.FirstName != nil {
		updateUser.FirstName = *updateData.FirstName
	}
	if updateData.LastName != nil {
		updateUser.LastName = *updateData.LastName
	}
	if updateData.Email != nil {
		updateUser.Email = *updateData.Email
	}
	if updateData.Phone != nil {
		updateUser.Phone = *updateData.Phone
	}
	if updateData.Password != nil {
		updateUser.Password = *updateData.Password
	}

	user, err := s.Repo.UpdateUser(id, updateUser)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s UserService) DeleteUser(id uint) error {
	err := s.Repo.DeleteUser(id)
	if err != nil {
		return err
	}
	return nil
}

func (s UserService) isVerifiedUser(id uint) bool {
	currentUser, err := s.Repo.FindUserByID(id)

	return err == nil && currentUser.Verified
}

func (s UserService) GetVerificationCode(id uint) error {
	//1. check if user is verified
	if s.isVerifiedUser(id) {
		return errors.New("user is already verified")
	}
	//2. if not verified, generate a verification code
	verificationCode, err := s.Auth.GenerateVerificationCode()
	if err != nil {
		return errors.New("failed to generate verification code")
	}
	user, err := s.Repo.FindUserByID(id)
	if err != nil {
		return errors.New("failed to find user")
	}
	user.Code = verificationCode
	user.Expiry = time.Now().Add(time.Minute * 10)
	_, err = s.Repo.UpdateUser(id, *user)
	if err != nil {
		return errors.New("failed to update user")
	}

	//send sms or email to user with verification code
	notificationClient := notification.NewNotificationClient(s.Config)
	formattedPhone := helper.FormatPhoneToE164(user.Phone)
	err = notificationClient.SendSMS(formattedPhone, strconv.Itoa(verificationCode))
	if err != nil {
		return errors.New("failed to send verification code: " + err.Error())
	}

	return nil
}

func (s UserService) VerifyCode(id uint, code int) (bool, error) {
	//1. check if user is verified
	if s.isVerifiedUser(id) {
		return false, errors.New("user is already verified")
	}
	//2. if not verified, verify the code
	user, err := s.Repo.FindUserByID(id)
	if err != nil {
		return false, errors.New("failed to find user")
	}
	if user.Code != code {
		return false, errors.New("invalid verification code")
	}
	if user.Expiry.Before(time.Now()) {
		return false, errors.New("verification code has expired")
	}
	user.Verified = true
	_, err = s.Repo.UpdateUser(id, *user)
	if err != nil {
		return false, errors.New("failed to update user")
	}
	return true, nil
}

func (s UserService) Profile(user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) GetProfile(userID uint) (*dto.ProfileResponse, error) {
	user, err := s.Repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Business logic: Transform to DTO
	profile := &dto.ProfileResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		UserType:  user.UserType,
		Verified:  user.Verified,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Business logic: Transform address if exists
	if user.Address.ID != 0 {
		profile.Address = &dto.AddressResponse{
			ID:           user.Address.ID,
			AddressLine1: user.Address.AddressLine1,
			AddressLine2: user.Address.AddressLine2,
			City:         user.Address.City,
			State:        user.Address.State,
			Country:      user.Address.Country,
			PostalCode:   user.Address.PostalCode,
			CreatedAt:    user.Address.CreatedAt,
			UpdatedAt:    user.Address.UpdatedAt,
		}
	}

	// Business logic: Transform cart items
	profile.Cart = make([]dto.CartItemResponse, len(user.Cart))
	for i, cartItem := range user.Cart {
		profile.Cart[i] = dto.CartItemResponse{
			ID:        cartItem.ID,
			ProductID: cartItem.ProductID,
			SellerID:  cartItem.SellerID,
			Name:      cartItem.Name,
			ImageURL:  cartItem.ImageURL,
			Price:     cartItem.Price,
			Quantity:  cartItem.Quantity,
			CreatedAt: cartItem.CreatedAt,
			UpdatedAt: cartItem.UpdatedAt,
		}
	}

	// Business logic: Transform orders
	profile.Orders = make([]dto.OrderResponse, len(user.Orders))
	for i, order := range user.Orders {
		profile.Orders[i] = s.toOrderResponse(&order)
	}

	return profile, nil
}

func (s UserService) CreateProfile(userID uint, profileInput dto.ProfileInput) (*domain.User, error) {
	// Check if profile already exists
	existingAddress, err := s.Repo.FindAddressByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existingAddress != nil {
		return nil, errors.New("profile already exists, use update endpoint instead")
	}

	user, err := s.Repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	user.FirstName = profileInput.FirstName
	user.LastName = profileInput.LastName
	_, err = s.Repo.UpdateUser(userID, *user)
	if err != nil {
		return nil, err
	}

	address := &domain.Address{
		UserID:       userID,
		AddressLine1: profileInput.Address.AddressLine1,
		AddressLine2: profileInput.Address.AddressLine2,
		City:         profileInput.Address.City,
		State:        profileInput.Address.State,
		Country:      profileInput.Address.Country,
		PostalCode:   profileInput.Address.PostalCode,
	}

	err = s.Repo.CreateProfile(address)
	if err != nil {
		return nil, err
	}

	// Reload user with address
	userWithAddress, err := s.Repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	return userWithAddress, nil
}

func (s UserService) UpdateProfile(userID uint, profileInput dto.ProfileUpdateInput) (*domain.User, error) {
	user, err := s.Repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Update user fields if provided
	if profileInput.FirstName != nil {
		user.FirstName = *profileInput.FirstName
	}
	if profileInput.LastName != nil {
		user.LastName = *profileInput.LastName
	}
	_, err = s.Repo.UpdateUser(userID, *user)
	if err != nil {
		return nil, err
	}

	// Check if address exists, update if exists, create if not (upsert pattern)
	existingAddress, err := s.Repo.FindAddressByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Only update address if address fields are provided
	hasAddressFields := profileInput.Address.AddressLine1 != nil ||
		profileInput.Address.AddressLine2 != nil ||
		profileInput.Address.City != nil ||
		profileInput.Address.State != nil ||
		profileInput.Address.Country != nil ||
		profileInput.Address.PostalCode != nil

	if hasAddressFields {
		address := &domain.Address{
			UserID: userID,
		}

		// Only set fields that are provided
		if profileInput.Address.AddressLine1 != nil {
			address.AddressLine1 = *profileInput.Address.AddressLine1
		}
		if profileInput.Address.AddressLine2 != nil {
			address.AddressLine2 = *profileInput.Address.AddressLine2
		}
		if profileInput.Address.City != nil {
			address.City = *profileInput.Address.City
		}
		if profileInput.Address.State != nil {
			address.State = *profileInput.Address.State
		}
		if profileInput.Address.Country != nil {
			address.Country = *profileInput.Address.Country
		}
		if profileInput.Address.PostalCode != nil {
			address.PostalCode = *profileInput.Address.PostalCode
		}

		if existingAddress != nil {
			// Update existing address - need to preserve existing values if not provided
			address.ID = existingAddress.ID
			if profileInput.Address.AddressLine1 == nil {
				address.AddressLine1 = existingAddress.AddressLine1
			}
			if profileInput.Address.AddressLine2 == nil {
				address.AddressLine2 = existingAddress.AddressLine2
			}
			if profileInput.Address.City == nil {
				address.City = existingAddress.City
			}
			if profileInput.Address.State == nil {
				address.State = existingAddress.State
			}
			if profileInput.Address.Country == nil {
				address.Country = existingAddress.Country
			}
			if profileInput.Address.PostalCode == nil {
				address.PostalCode = existingAddress.PostalCode
			}
			err = s.Repo.UpdateProfile(address)
			if err != nil {
				return nil, err
			}
		} else {
			// Create new address if it doesn't exist (but all required fields must be provided)
			if profileInput.Address.AddressLine1 == nil || profileInput.Address.City == nil ||
				profileInput.Address.State == nil || profileInput.Address.Country == nil ||
				profileInput.Address.PostalCode == nil {
				return nil, errors.New("all address fields are required when creating a new address")
			}
			// Set defaults for optional fields
			if profileInput.Address.AddressLine2 == nil {
				address.AddressLine2 = ""
			}
			err = s.Repo.CreateProfile(address)
			if err != nil {
				return nil, err
			}
		}
	}

	// Reload user with address
	userWithAddress, err := s.Repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	return userWithAddress, nil
}

func (s UserService) DeleteProfile(id uint) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) FindOrdersByUserID(userID uint) ([]dto.OrderResponse, error) {
	orders, err := s.Repo.FindOrdersByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Business logic: Transform to DTOs
	responses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = s.toOrderResponse(&order)
	}

	return responses, nil
}

func (s UserService) GetOrderByID(orderID uint, userID uint) (*dto.OrderResponse, error) {
	order, err := s.Repo.FindOrderByID(orderID)
	if err != nil {
		return nil, err
	}

	// Business logic: Verify ownership
	if order.UserID != userID {
		return nil, errors.New("order not found")
	}

	// Business logic: Transform to DTO
	response := s.toOrderResponse(order)
	return &response, nil
}

func (s UserService) GetOrdersBySellerID(sellerID uint) ([]dto.SellerOrderResponse, error) {
	orders, err := s.Repo.FindOrdersBySellerID(sellerID)
	if err != nil {
		return nil, err
	}

	// Business logic: Filter and transform orders
	sellerOrders := make([]dto.SellerOrderResponse, 0)
	for _, order := range orders {
		// Business logic: Filter items belonging to seller
		sellerItems := s.filterOrderItemsBySeller(order.Items, sellerID)

		// Business rule: Skip orders with no seller items
		if len(sellerItems) == 0 {
			continue
		}

		// Business logic: Transform to DTO
		sellerOrders = append(sellerOrders, dto.SellerOrderResponse{
			ID:             order.ID,
			UserID:         order.UserID,
			Status:         order.Status,
			Amount:         order.Amount,
			TransactionID:  order.TransactionId,
			OrderRefNumber: order.OrderRefNumber,
			PaymentID:      order.PaymentId,
			Items:          sellerItems,
			CreatedAt:      order.CreatedAt,
			UpdatedAt:      order.UpdatedAt,
		})
	}

	return sellerOrders, nil
}

func (s UserService) GetOrderDetailsBySellerID(orderID uint, sellerID uint) (*dto.SellerOrderResponse, error) {
	order, err := s.Repo.FindOrderByID(orderID)
	if err != nil {
		return nil, err
	}

	// Business logic: Filter items belonging to seller
	sellerItems := s.filterOrderItemsBySeller(order.Items, sellerID)

	// Business logic: Verify that the order has items belonging to the seller
	if len(sellerItems) == 0 {
		return nil, errors.New("order not found")
	}

	// Business logic: Transform to DTO
	response := dto.SellerOrderResponse{
		ID:             order.ID,
		UserID:         order.UserID,
		Status:         order.Status,
		Amount:         order.Amount,
		TransactionID:  order.TransactionId,
		OrderRefNumber: order.OrderRefNumber,
		PaymentID:      order.PaymentId,
		Items:          sellerItems,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}

	return &response, nil
}

func (s UserService) toOrderResponse(order *domain.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      item.Name,
			ImageURL:  item.ImageUrl,
			SellerID:  item.SellerId,
			Price:     item.Price,
			Quantity:  item.Quantity,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}

	return dto.OrderResponse{
		ID:             order.ID,
		UserID:         order.UserID,
		Status:         order.Status,
		Amount:         order.Amount,
		TransactionID:  order.TransactionId,
		OrderRefNumber: order.OrderRefNumber,
		PaymentID:      order.PaymentId,
		Items:          items,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

func (s UserService) filterOrderItemsBySeller(items []domain.OrderItem, sellerID uint) []dto.OrderItemResponse {
	sellerItems := make([]dto.OrderItemResponse, 0)
	for _, item := range items {
		if item.SellerId == sellerID {
			sellerItems = append(sellerItems, dto.OrderItemResponse{
				ID:        item.ID,
				ProductID: item.ProductID,
				Name:      item.Name,
				ImageURL:  item.ImageUrl,
				SellerID:  item.SellerId,
				Price:     item.Price,
				Quantity:  item.Quantity,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			})
		}
	}
	return sellerItems
}

func (s UserService) GetOrder(user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) BecomeSeller(id uint, seller dto.BecomeSellerInput) (*domain.User, string, error) {
	// find existing user
	user, err := s.Repo.FindUserByID(id)
	if err != nil {
		return nil, "", errors.New("failed to find user")
	}

	// check if already a seller and return error
	if user.UserType == "seller" {
		return nil, "", errors.New("user is already a seller")
	}

	// verify bank account BEFORE updating user type - this must succeed first
	if s.BankService == nil {
		return nil, "", errors.New("bank verification service is not available")
	}

	log.Printf("Verifying bank account for user %d before becoming seller: AccountNumber=%s, BankCode=%s", id, seller.BankAccountNumber, seller.BankCode)
	_, err = s.BankService.VerifyAccount(seller.BankAccountNumber, seller.BankCode)
	if err != nil {
		log.Printf("Bank account verification failed for user %d: %v", id, err)
		return nil, "", errors.New("bank account verification failed: " + err.Error())
	}
	log.Printf("Bank account verified successfully for user %d, proceeding with seller status update", id)

	// update user - only happens after successful bank verification
	user.UserType = "seller"
	user.FirstName = seller.FirstName
	user.LastName = seller.LastName
	user.Phone = seller.PhoneNumber

	updatedUser, err := s.Repo.UpdateUser(id, *user)
	if err != nil {
		log.Printf("Error updating user to seller: %v", err)
		return nil, "", errors.New("failed to update user: " + err.Error())
	}

	// create bank account information
	bankAccount := &domain.BankAccount{
		UserId:            id,
		BankName:          seller.PaymentType,
		BankAccountNumber: seller.BankAccountNumber,
		BankCode:          seller.BankCode,
	}

	createdBankAccount, err := s.Repo.CreateBankAccount(bankAccount)
	if err != nil {
		log.Printf("Error creating bank account: %v", err)
		return nil, "", errors.New("failed to create bank account: " + err.Error())
	}
	log.Printf("Bank account created successfully: ID=%d", createdBankAccount.ID)

	// generate new token
	token, err := s.Auth.GenerateToken(updatedUser.ID, updatedUser.Email, updatedUser.UserType)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	// return updated user and new token
	return &updatedUser, token, nil
}

func (s UserService) AddToCart(userID uint, request dto.CreateCartRequest) (*domain.Cart, error) {
	product, err := s.CatalogueRepo.GetProductByID(request.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	existingCart, err := s.Repo.FindCartByUserIDAndProductID(userID, request.ProductID)
	if err == nil {
		existingCart.Quantity += request.Quantity
		updatedCart, err := s.Repo.UpdateCart(existingCart)
		if err != nil {
			return nil, err
		}
		return updatedCart, nil
	}

	cartItem := &domain.Cart{
		UserID:    userID,
		SellerID:  product.SellerID,
		Name:      product.Name,
		ImageURL:  product.ImageURL,
		Price:     product.Price,
		Quantity:  request.Quantity,
		ProductID: request.ProductID,
	}

	createdCart, err := s.Repo.CreateCart(cartItem)
	if err != nil {
		return nil, err
	}

	return createdCart, nil
}

func (s UserService) FindCartItems(userID uint) ([]domain.Cart, error) {
	cartItems, err := s.Repo.FindCartByUserID(userID)
	if err != nil {
		return nil, err
	}
	return cartItems, nil
}

func (s UserService) GetCartItem(userID uint, productID uint) (*domain.Cart, error) {
	cartItem, err := s.Repo.FindCartByUserIDAndProductID(userID, productID)
	if err != nil {
		return nil, errors.New("cart item not found")
	}
	return cartItem, nil
}

func (s UserService) UpdateCart(userID uint, request dto.UpdateCartRequest) (*domain.Cart, error) {
	if request.ProductID == nil {
		return nil, errors.New("product ID is required")
	}

	cartItem, err := s.Repo.FindCartByUserIDAndProductID(userID, *request.ProductID)
	if err != nil {
		return nil, errors.New("cart item not found")
	}

	if request.Quantity != nil {
		cartItem.Quantity = *request.Quantity
	}
	if request.Price != nil {
		cartItem.Price = *request.Price
	}

	updatedCart, err := s.Repo.UpdateCart(cartItem)
	if err != nil {
		return nil, err
	}

	return updatedCart, nil
}

func (s UserService) DeleteCartItem(userID uint, productID uint) error {
	return s.Repo.DeleteCartItem(userID, productID)
}

func (s UserService) ClearCart(userID uint) error {
	return s.Repo.DeleteAllCartItems(userID)
}

func (s UserService) IncrementCartItem(userID uint, productID uint) (*domain.Cart, error) {
	cartItem, err := s.Repo.FindCartByUserIDAndProductID(userID, productID)
	if err != nil {
		return nil, errors.New("cart item not found")
	}

	cartItem.Quantity += 1
	updatedCart, err := s.Repo.UpdateCart(cartItem)
	if err != nil {
		return nil, err
	}

	return updatedCart, nil
}

func (s UserService) DecrementCartItem(userID uint, productID uint) (*domain.Cart, error) {
	cartItem, err := s.Repo.FindCartByUserIDAndProductID(userID, productID)
	if err != nil {
		return nil, errors.New("cart item not found")
	}

	if cartItem.Quantity <= 1 {
		return nil, errors.New("quantity cannot be less than 1. Use delete to remove item")
	}

	cartItem.Quantity -= 1
	updatedCart, err := s.Repo.UpdateCart(cartItem)
	if err != nil {
		return nil, err
	}

	return updatedCart, nil
}

func (s UserService) CreateOrder(u domain.User) (int, error) {
	// find cart items for the user
	cartItems, err := s.Repo.FindCartByUserID(u.ID)
	if err != nil {
		return 0, errors.New("failed to find cart items")
	}
	if len(cartItems) == 0 {
		return 0, errors.New("cart is empty")
	}

	// find success payment reference status
	// TODO: Implement payment verification logic
	paymentID := "PAY12345"
	transactionID := "12345"
	_ = paymentID
	_ = transactionID

	// generate order reference
	orderRef, err := helper.GenerateOrderRef(8)
	if err != nil {
		return 0, errors.New("failed to generate order reference")
	}

	// extract numeric part from orderRef (remove "0RD" prefix)
	orderRefNumberUint, err := strconv.ParseUint(orderRef[3:], 10, 32)
	if err != nil {
		return 0, errors.New("failed to parse order reference number")
	}
	orderRefNumber := uint(orderRefNumberUint)

	// calculate total amount from cart items
	var totalAmount float64
	var orderItems []domain.OrderItem
	for _, cartItem := range cartItems {
		totalAmount += cartItem.Price * float64(cartItem.Quantity)
		orderItems = append(orderItems, domain.OrderItem{
			ProductID: cartItem.ProductID,
			Name:      cartItem.Name,
			ImageUrl:  cartItem.ImageURL,
			SellerId:  cartItem.SellerID,
			Price:     cartItem.Price,
			Quantity:  uint(cartItem.Quantity),
		})
	}

	// create order with generated OrderRefNumber
	order := &domain.Order{
		UserID:         u.ID,
		Status:         "pending",
		Amount:         totalAmount,
		TransactionId:  transactionID,
		OrderRefNumber: orderRefNumber,
		PaymentId:      paymentID,
		Items:          orderItems,
	}

	createdOrder, err := s.Repo.CreateOrder(order)
	if err != nil {
		return 0, errors.New("failed to create order")
	}

	// Payment-First Flow: Email is sent after order creation with verified payment
	// Flow: POST /payment → User pays → Webhook verifies → POST /orders → Email sent here
	// TODO: Implement SendEmail() method in notification client
	// ✅ SEND EMAIL HERE - Order Confirmation
	// notificationClient := notification.NewNotificationClient(s.Config)
	// emailBody := buildOrderConfirmationEmail(createdOrder, u)
	// err = notificationClient.SendEmail(u.Email, "Order Confirmation", emailBody)
	// if err != nil {
	// 	log.Printf("Warning: failed to send order confirmation email: %v", err)
	// 	// Don't fail order creation if email fails
	// }

	// remove cart items from the cart
	err = s.Repo.DeleteAllCartItems(u.ID)
	if err != nil {
		log.Printf("Warning: failed to clear cart after order creation: %v", err)
		// Don't fail the order creation if cart clearing fails
	}

	// return order number (ID)
	return int(createdOrder.ID), nil
}

func (s UserService) FindOrder(id uint) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) GetOrderById(id uint, user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) GetOrders(user interface{}) ([]interface{}, error) {
	//perform some db operation
	//business logic
	return []interface{}{}, nil
}
