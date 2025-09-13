package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

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
			category TEXT,
			additional_images JSONB,
			google_product_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_products_price_eur ON products(price_eur)`,
		`CREATE INDEX IF NOT EXISTS idx_products_title ON products USING gin(to_tsvector('english', title))`,
		`CREATE INDEX IF NOT EXISTS idx_products_google_id ON products(google_product_id)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// SaveProducts saves products to database
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
			features, category, additional_images, google_product_id,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			image_url = EXCLUDED.image_url,
			image_local = EXCLUDED.image_local,
			price_gbp = EXCLUDED.price_gbp,
			price_eur = EXCLUDED.price_eur,
			offer_count = EXCLUDED.offer_count,
			features = EXCLUDED.features,
			category = EXCLUDED.category,
			additional_images = EXCLUDED.additional_images,
			google_product_id = EXCLUDED.google_product_id,
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
			product.Category,
			additionalImagesJSON,
			"", // google_product_id пока пустой
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

// Close closes the database connection
func (d *DatabaseStorage) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}
