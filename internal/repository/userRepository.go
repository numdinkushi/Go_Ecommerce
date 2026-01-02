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
	CreateBankAccount(bankAccount *domain.BankAccount) (*domain.BankAccount, error)

	// Cart methods
	CreateCart(cart *domain.Cart) (*domain.Cart, error)
	FindCartByUserID(userID uint) ([]domain.Cart, error)
	FindCartByUserIDAndProductID(userID uint, productID uint) (*domain.Cart, error)
	UpdateCart(cart *domain.Cart) (*domain.Cart, error)
	DeleteCartItem(userID uint, productID uint) error
	DeleteAllCartItems(userID uint) error
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
		return domain.User{}, err // Return original error instead of wrapping
	}
	return user, nil
}

func (r *userRepository) DeleteUser(id uint) error {
	return r.DB.Delete(&domain.User{}, id).Error
}

func (r *userRepository) CreateBankAccount(bankAccount *domain.BankAccount) (*domain.BankAccount, error) {
	err := r.DB.Create(bankAccount).Error
	if err != nil {
		log.Printf("Failed to create bank account: %v", err)
		return nil, err
	}
	log.Println("Bank account created successfully")
	return bankAccount, nil
}

// Cart methods

func (r *userRepository) CreateCart(cart *domain.Cart) (*domain.Cart, error) {
	err := r.DB.Create(cart).Error
	if err != nil {
		log.Printf("Failed to create cart item: %v", err)
		return nil, err
	}
	log.Println("Cart item created successfully")
	return cart, nil
}

func (r *userRepository) FindCartByUserID(userID uint) ([]domain.Cart, error) {
	var cartItems []domain.Cart
	err := r.DB.Where("user_id = ?", userID).Find(&cartItems).Error
	if err != nil {
		return nil, err
	}
	return cartItems, nil
}

func (r *userRepository) FindCartByUserIDAndProductID(userID uint, productID uint) (*domain.Cart, error) {
	var cartItem domain.Cart
	err := r.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&cartItem).Error
	if err != nil {
		return nil, err
	}
	return &cartItem, nil
}

func (r *userRepository) UpdateCart(cart *domain.Cart) (*domain.Cart, error) {
	var updatedCart domain.Cart
	err := r.DB.Model(&updatedCart).Clauses(clause.Returning{}).Where("id=?", cart.ID).Updates(cart).Error
	if err != nil {
		log.Printf("Failed to update cart item: %v", err)
		return nil, err
	}
	return &updatedCart, nil
}

func (r *userRepository) DeleteCartItem(userID uint, productID uint) error {
	result := r.DB.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&domain.Cart{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cart item not found")
	}
	return nil
}

func (r *userRepository) DeleteAllCartItems(userID uint) error {
	result := r.DB.Where("user_id = ?", userID).Delete(&domain.Cart{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
