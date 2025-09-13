package models

import "time"

// Product represents a product with all its details
type Product struct {
	ID          string            `json:"id" db:"id"`
	Title       string            `json:"title" db:"title"`
	URL         string            `json:"url" db:"url"`
	ImageURL    string            `json:"image_url,omitempty" db:"image_url"`
	ImageLocal  string            `json:"image_local,omitempty" db:"image_local"`
	Price       *PriceInfo        `json:"price_info,omitempty" db:"-"`
	Features    map[string]string `json:"features,omitempty" db:"-"`
	Categories  []string          `json:"categories,omitempty" db:"-"`
	ExtraImages []ImageInfo       `json:"additional_images,omitempty" db:"-"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// PriceInfo contains price information
type PriceInfo struct {
	PriceGBP   string  `json:"price_gbp,omitempty"`
	PriceEUR   float64 `json:"price_eur,omitempty"`
	OfferCount string  `json:"offer_count,omitempty"`
}

// ImageInfo contains information about additional images
type ImageInfo struct {
	URL   string `json:"url"`
	Local string `json:"local"`
}

// BasicProduct represents a product with basic information from list page
type BasicProduct struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ProductExists represents a check for product existence
type ProductExists struct {
	ID     string
	Exists bool
}
