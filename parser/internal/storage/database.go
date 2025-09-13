package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"pricerunner-parser/internal/config"
	"pricerunner-parser/internal/models"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DatabaseStorage implements Storage interface for PostgreSQL
type DatabaseStorage struct {
	db *sql.DB
}

// NewDatabaseStorage creates a new database storage instance
func NewDatabaseStorage(cfg config.DatabaseConfig) (*DatabaseStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &DatabaseStorage{db: db}

	// Создаем таблицы если их нет
	if err := storage.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return storage, nil
}

// createTables creates necessary database tables
func (d *DatabaseStorage) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id VARCHAR(50) PRIMARY KEY,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			image_url TEXT,
			image_local TEXT,
			price_gbp VARCHAR(20),
			price_eur DECIMAL(10,2),
			offer_count VARCHAR(10),
			features JSONB,
			categories TEXT[],
			additional_images JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_products_price_eur ON products(price_eur)`,
		`CREATE INDEX IF NOT EXISTS idx_products_title ON products USING gin(to_tsvector('english', title))`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// SaveProducts saves products to database (для совместимости, сохраняем все в основную таблицу)
func (d *DatabaseStorage) SaveProducts(products []models.Product, pageNumber int) error {
	if len(products) == 0 {
		return nil
	}

	// Используем транзакцию для атомарности
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO products (id, title, url, image_url, image_local, 
			price_gbp, price_eur, offer_count, 
			features, categories, additional_images, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			image_url = EXCLUDED.image_url,
			image_local = EXCLUDED.image_local,
			price_gbp = EXCLUDED.price_gbp,
			price_eur = EXCLUDED.price_eur,
			offer_count = EXCLUDED.offer_count,
			features = EXCLUDED.features,
			categories = EXCLUDED.categories,
			additional_images = EXCLUDED.additional_images,
			updated_at = EXCLUDED.updated_at
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, product := range products {
		// Конвертируем сложные поля в JSON
		featuresJSON, _ := json.Marshal(product.Features)
		categoriesArray := fmt.Sprintf("{%s}", strings.Join(product.Categories, ","))
		additionalImagesJSON, _ := json.Marshal(product.ExtraImages)

		var priceGBP string
		var priceEUR sql.NullFloat64
		var offerCount string

		if product.Price != nil {
			priceGBP = product.Price.PriceGBP
			if product.Price.PriceEUR > 0 {
				priceEUR = sql.NullFloat64{Float64: product.Price.PriceEUR, Valid: true}
			}
			offerCount = product.Price.OfferCount
		}

		_, err := stmt.Exec(
			product.ID,
			product.Title,
			product.URL,
			product.ImageURL,
			product.ImageLocal,
			priceGBP,
			priceEUR,
			offerCount,
			featuresJSON,
			categoriesArray,
			additionalImagesJSON,
			product.CreatedAt,
			product.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert product %s: %w", product.ID, err)
		}
	}

	return tx.Commit()
}

// SaveFinalData для БД это то же самое что и SaveProducts
func (d *DatabaseStorage) SaveFinalData(products []models.Product) error {
	return d.SaveProducts(products, 0)
}

// ProductExists checks if a product exists in database
func (d *DatabaseStorage) ProductExists(productID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)"
	err := d.db.QueryRow(query, productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}
	return exists, nil
}

// GetExistingProducts returns all existing product IDs
func (d *DatabaseStorage) GetExistingProducts() ([]string, error) {
	query := "SELECT id FROM products ORDER BY created_at DESC"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing products: %w", err)
	}
	defer rows.Close()

	var productIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		productIDs = append(productIDs, id)
	}

	return productIDs, nil
}

// GetProductsByPriceRange returns products within price range
func (d *DatabaseStorage) GetProductsByPriceRange(minPrice, maxPrice float64) ([]models.Product, error) {
	query := `
		SELECT id, title, url, image_url, image_local, 
		   price_gbp, price_eur, offer_count, 
			   features, categories, additional_images,
			   created_at, updated_at 
		FROM products 
		WHERE price_eur BETWEEN $1 AND $2 
		ORDER BY price_eur ASC
	`

	rows, err := d.db.Query(query, minPrice, maxPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to query products by price: %w", err)
	}
	defer rows.Close()

	return d.scanProducts(rows)
}

// SearchProducts searches products by title
func (d *DatabaseStorage) SearchProducts(searchTerm string) ([]models.Product, error) {
	query := `
		SELECT id, title, url, image_url, image_local, 
		   price_gbp, price_eur, offer_count, 
			   features, categories, additional_images,
			   created_at, updated_at 
		FROM products 
		WHERE to_tsvector('english', title) @@ plainto_tsquery('english', $1)
		ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	return d.scanProducts(rows)
}

// GetProductStats returns statistics about products
func (d *DatabaseStorage) GetProductStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Общее количество
	var totalCount int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalCount); err == nil {
		stats["total_products"] = totalCount
	}

	// С ценами
	var withPrices int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM products WHERE price_eur IS NOT NULL").Scan(&withPrices); err == nil {
		stats["products_with_prices"] = withPrices
	}

	// С изображениями
	var withImages int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM products WHERE image_local IS NOT NULL AND image_local != ''").Scan(&withImages); err == nil {
		stats["products_with_images"] = withImages
	}

	// Средняя цена
	var avgPrice sql.NullFloat64
	if err := d.db.QueryRow("SELECT AVG(price_eur) FROM products WHERE price_eur IS NOT NULL").Scan(&avgPrice); err == nil && avgPrice.Valid {
		stats["average_price"] = avgPrice.Float64
	}

	// Минимальная и максимальная цены
	var minPrice, maxPrice sql.NullFloat64
	if err := d.db.QueryRow("SELECT MIN(price_eur), MAX(price_eur) FROM products WHERE price_eur IS NOT NULL").Scan(&minPrice, &maxPrice); err == nil {
		if minPrice.Valid {
			stats["min_price"] = minPrice.Float64
		}
		if maxPrice.Valid {
			stats["max_price"] = maxPrice.Float64
		}
	}

	return stats, nil
}

// scanProducts helper function to scan rows into Product structs
func (d *DatabaseStorage) scanProducts(rows *sql.Rows) ([]models.Product, error) {
	var products []models.Product

	for rows.Next() {
		var product models.Product
		var featuresJSON, additionalImagesJSON []byte
		var categories string
		var priceGBP string
		var priceEUR sql.NullFloat64
		var offerCount string

		err := rows.Scan(
			&product.ID,
			&product.Title,
			&product.URL,
			&product.ImageURL,
			&product.ImageLocal,
			&priceGBP,
			&priceEUR,
			&offerCount,
			&featuresJSON,
			&categories,
			&additionalImagesJSON,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Восстанавливаем сложные поля
		if len(featuresJSON) > 0 {
			json.Unmarshal(featuresJSON, &product.Features)
		}

		if len(additionalImagesJSON) > 0 {
			json.Unmarshal(additionalImagesJSON, &product.ExtraImages)
		}

		// Парсим массив категорий
		if categories != "" {
			categories = strings.Trim(categories, "{}")
			if categories != "" {
				product.Categories = strings.Split(categories, ",")
			}
		}

		// Восстанавливаем цену
		if priceGBP != "" || priceEUR.Valid {
			product.Price = &models.PriceInfo{
				PriceGBP:   priceGBP,
				OfferCount: offerCount,
			}
			if priceEUR.Valid {
				product.Price.PriceEUR = priceEUR.Float64
			}
		}

		products = append(products, product)
	}

	return products, nil
}

// Close closes the database connection
func (d *DatabaseStorage) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}
