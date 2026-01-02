package dto

type CreateCartRequest struct {
	Quantity  int     `json:"quantity"`
	ProductID uint    `json:"product_id"`
}

type UpdateCartRequest struct {
	Quantity *int `json:"quantity,omitempty"`
	Price    *float64 `json:"price,omitempty"`
	ProductID *uint `json:"product_id,omitempty"`
}

type DeleteCartRequest struct {
	ProductID uint `json:"product_id"`
}

type GetCartRequest struct {
	ProductID uint `json:"product_id"`
}

type GetCartResponse struct {	
	ID        uint    `json:"id"`
	SellerID  uint    `json:"seller_id"`
	Name      string  `json:"name"`
	ImageURL  string  `json:"image_url"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	ProductID uint    `json:"product_id"`
}