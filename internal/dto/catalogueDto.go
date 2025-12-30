package dto

type Category struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Product struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	CategoryID  uint    `json:"category_id"`
	Stock       int     `json:"stock,omitempty"`
	ImageURL    string  `json:"image_url,omitempty"`
}

