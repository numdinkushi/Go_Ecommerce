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
	Repo   repository.UserRepository
	Auth   helper.Auth
	Config config.AppConfig
}

func NewUserService(repo repository.UserRepository, auth helper.Auth, config config.AppConfig) UserService {
	return UserService{
		Repo:   repo,
		Auth:   auth,
		Config: config,
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

func (s UserService) CreateProfile(user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) UpdateProfile(id uint, user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) DeleteProfile(id uint) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) Orders(user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
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

	// update user
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

	log.Printf("Attempting to create bank account for user %d: %+v", id, bankAccount)
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

func (s UserService) CreateCart(id uint, user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
}

func (s UserService) FindCart(id uint) ([]interface{}, error) {
	//perform some db operation
	//business logic
	return []interface{}{}, nil
}

func (s UserService) CreateOrder(user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
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
