package dto

type Category struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	ParentID     *uint  `json:"parent_id,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
	DisplayOrder int    `json:"display_order,omitempty"`
}

type Product struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	CategoryID  uint    `json:"category_id"`
	Stock       int     `json:"stock,omitempty"`
	ImageURL    string  `json:"image_url,omitempty"`
}
