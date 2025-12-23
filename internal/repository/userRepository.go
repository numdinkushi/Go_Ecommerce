package repository

import (
	"errors"
	"go-ecommerce-app/internal/domain"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	CreateUser(user *domain.User) (*domain.User, error)
	FindUserByEmail(email string) (*domain.User, error)
	FindUserByID(id uint) (*domain.User, error)
	FindAllUsers() ([]domain.User, error)
	UpdateUser(id uint, u domain.User) (domain.User, error)
	DeleteUser(id uint) error
}

type userRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{DB: db}
}

func (r *userRepository) CreateUser(user *domain.User) (*domain.User, error) {
	err := r.DB.Create(user).Error

	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return nil, err
	}

	log.Println("User created successfully")
	return user, nil
}

func (r *userRepository) FindUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindUserByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAllUsers() ([]domain.User, error) {
	var users []domain.User
	err := r.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) UpdateUser(id uint, u domain.User) (domain.User, error) {
	var user domain.User
	err := r.DB.Model(&user).Clauses(clause.Returning{}).Where("id=?", id).Updates(u).Error
	if err != nil {
		log.Printf("error on update %v", err)
		return domain.User{}, errors.New("failed update user")
	}
	return user, nil
}

func (r *userRepository) DeleteUser(id uint) error {
	return r.DB.Delete(&domain.User{}, id).Error
}
