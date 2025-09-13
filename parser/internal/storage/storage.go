package storage

import "pricerunner-parser/internal/models"

// Storage defines the interface for data storage
type Storage interface {
	// SaveProducts saves a slice of products
	SaveProducts(products []models.Product, pageNumber int) error

	// SaveFinalData saves all products to a final file/table
	SaveFinalData(products []models.Product) error

	// ProductExists checks if a product already exists
	ProductExists(productID string) (bool, error)

	// GetExistingProducts returns a list of existing product IDs
	GetExistingProducts() ([]string, error)

	// Close closes any open connections
	Close() error
}
