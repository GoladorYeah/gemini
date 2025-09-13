package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pricerunner-parser/internal/models"
)

// JSONStorage implements Storage interface for JSON files
type JSONStorage struct {
	outputDir string
}

// NewJSONStorage creates a new JSON storage instance
func NewJSONStorage(outputDir string) *JSONStorage {
	return &JSONStorage{
		outputDir: outputDir,
	}
}

// SaveProducts saves products to a page-specific JSON file
func (j *JSONStorage) SaveProducts(products []models.Product, pageNumber int) error {
	if len(products) == 0 {
		return nil
	}

	filename := fmt.Sprintf("products_page_%d_detailed.json", pageNumber)
	filepath := filepath.Join(j.outputDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filepath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(products); err != nil {
		return fmt.Errorf("failed to encode products: %w", err)
	}

	return nil
}

// SaveFinalData saves all products to the final JSON file
func (j *JSONStorage) SaveFinalData(products []models.Product) error {
	if len(products) == 0 {
		return nil
	}

	filepath := filepath.Join(j.outputDir, "all_products_detailed.json")

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create final file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(products); err != nil {
		return fmt.Errorf("failed to encode final products: %w", err)
	}

	return nil
}

// ProductExists checks if a product exists in any of the JSON files
func (j *JSONStorage) ProductExists(productID string) (bool, error) {
	existingProducts, err := j.GetExistingProducts()
	if err != nil {
		return false, err
	}

	for _, id := range existingProducts {
		if id == productID {
			return true, nil
		}
	}

	return false, nil
}

// GetExistingProducts returns all product IDs from existing JSON files
func (j *JSONStorage) GetExistingProducts() ([]string, error) {
	var allIDs []string

	// Читаем все JSON файлы в директории
	files, err := filepath.Glob(filepath.Join(j.outputDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob JSON files: %w", err)
	}

	for _, file := range files {
		// Пропускаем системные файлы
		if strings.Contains(filepath.Base(file), "temp") {
			continue
		}

		products, err := j.readProductsFromFile(file)
		if err != nil {
			continue // Пропускаем файлы с ошибками
		}

		for _, product := range products {
			allIDs = append(allIDs, product.ID)
		}
	}

	return allIDs, nil
}

// readProductsFromFile reads products from a specific JSON file
func (j *JSONStorage) readProductsFromFile(filePath string) ([]models.Product, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var products []models.Product
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&products); err != nil {
		return nil, err
	}

	return products, nil
}

// Close implements Storage interface (no-op for JSON)
func (j *JSONStorage) Close() error {
	return nil
}
