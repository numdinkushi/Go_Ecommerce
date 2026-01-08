package dto

import "time"

type AddressResponse struct {
	ID          uint      `json:"id"`
	AddressLine1 string   `json:"address_line1"`
	AddressLine2 string   `json:"address_line2"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Country     string    `json:"country"`
	PostalCode  string    `json:"postal_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CartItemResponse struct {
	ID        uint      `json:"id"`
	ProductID uint      `json:"product_id"`
	SellerID  uint      `json:"seller_id"`
	Name      string    `json:"name"`
	ImageURL  string    `json:"image_url"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProfileResponse struct {
	ID        uint             `json:"id"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Email     string           `json:"email"`
	Phone     string           `json:"phone"`
	UserType  string           `json:"user_type"`
	Verified  bool             `json:"verified"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Address   *AddressResponse `json:"address,omitempty"`
	Cart      []CartItemResponse `json:"cart"`
	Orders    []OrderResponse   `json:"orders"`
}


