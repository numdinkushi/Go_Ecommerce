package domain

import "time"

type User struct {
	ID        uint    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Code      int `json:"code"`
	Expiry    time.Time `json:"expiry"`
	Verified  bool `json:"verified"`
	Password  string `json:"password"`
	UserType  string `json:"user_type"`
	// CreatedAt time.Time `json:"created_at"`
	// UpdatedAt time.Time `json:"updated_at"`
}
