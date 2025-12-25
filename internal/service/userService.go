package service

import (
	"errors"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
	"log"
)

type UserService struct {
	Repo repository.UserRepository
	Auth helper.Auth
}

func NewUserService(repo repository.UserRepository, auth helper.Auth) UserService {
	return UserService{
		Repo: repo,
		Auth: auth,
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

func (s UserService) GetVerificationCode(id uint, code int) (int, error) {
	//perform some db operation
	//business logic
	return 0, nil
}

func (s UserService) VerifyCode(id uint, code int) (bool, error) {
	//perform some db operation
	//business logic
	return false, nil
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

func (s UserService) BecomeSeller(id uint, user interface{}) (*domain.User, error) {
	//perform some db operation
	//business logic
	return &domain.User{}, nil
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
