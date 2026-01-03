package dto

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserSignUp struct {
	UserLogin
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type UserUpdate struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Password  *string `json:"password,omitempty"`
}

type VerificationCodeInput struct {
	Code int `json:"code"`
}

type BecomeSellerInput struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	PhoneNumber       string `json:"phone_number"`
	BankAccountNumber string `json:"bank_account_number"`
	BankCode          string `json:"bank_code"`
	PaymentType       string `json:"payment_type"`
}

type AddressInput struct {
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	PostalCode   string `json:"postal_code"`
}

type ProfileInput struct {
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Address   AddressInput `json:"address"`
}

type AddressUpdateInput struct {
	AddressLine1 *string `json:"address_line1,omitempty"`
	AddressLine2 *string `json:"address_line2,omitempty"`
	City         *string `json:"city,omitempty"`
	State        *string `json:"state,omitempty"`
	Country      *string `json:"country,omitempty"`
	PostalCode   *string `json:"postal_code,omitempty"`
}

type ProfileUpdateInput struct {
	FirstName *string            `json:"first_name,omitempty"`
	LastName  *string            `json:"last_name,omitempty"`
	Address   AddressUpdateInput `json:"address,omitempty"`
}
