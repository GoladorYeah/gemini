package storage

import (
	"database/sql"
	"fmt"

	"gemini/backend/internal/models"

	"github.com/lib/pq"
)

// NewDB creates and returns a new database connection.
func NewDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// AddProduct adds a new product to the database.
func AddProduct(db *sql.DB, product models.Product) error {
	query := `INSERT INTO products (id, title, category, features, google_product_id, image_url) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, product.ID, product.Title, product.Category, pq.Array(product.Features), product.GoogleProductID, product.ImageURL)
	return err
}

// GetProduct retrieves a product from the database by its ID.
func GetProduct(db *sql.DB, productID string) (*models.Product, error) {
	query := `SELECT id, title, category, features, google_product_id, image_url FROM products WHERE id = $1`
	row := db.QueryRow(query, productID)

	var product models.Product
	var features pq.StringArray
	if err := row.Scan(&product.ID, &product.Title, &product.Category, &features, &product.GoogleProductID, &product.ImageURL); err != nil {
		return nil, err
	}
	product.Features = []string(features)

	return &product, nil
}

// GetAllProducts retrieves all products from the database with proper error handling.
func GetAllProducts(db *sql.DB) ([]models.Product, error) {
	// First check what columns exist in the products table
	rows, err := db.Query(`
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = 'products' 
		AND table_schema = 'public'
	`)
	if err != nil {
		return nil, fmt.Errorf("error checking products table columns: %v", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			continue
		}
		columns = append(columns, column)
	}

	// Determine the correct column name for category
	categoryColumn := "category"
	for _, col := range columns {
		if col == "categories" {
			categoryColumn = "categories"
			break
		}
	}

	// Build the query with the correct column name
	query := fmt.Sprintf(`
		SELECT id, title, %s, features, 
		       COALESCE(google_product_id, '') as google_product_id,
		       COALESCE(image_url, '') as image_url 
		FROM products
	`, categoryColumn)

	rows, err = db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying products: %v", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		var features pq.StringArray

		err := rows.Scan(
			&product.ID,
			&product.Title,
			&product.Category,
			&features,
			&product.GoogleProductID,
			&product.ImageURL,
		)
		if err != nil {
			continue // Skip problematic rows
		}

		product.Features = []string(features)
		products = append(products, product)
	}

	return products, nil
}

// UpdateProduct updates a product in the database.
func UpdateProduct(db *sql.DB, product models.Product) error {
	query := `UPDATE products SET title = $2, category = $3, features = $4, google_product_id = $5, image_url = $6 WHERE id = $1`
	_, err := db.Exec(query, product.ID, product.Title, product.Category, pq.Array(product.Features), product.GoogleProductID, product.ImageURL)
	return err
}

// DeleteProduct deletes a product from the database.
func DeleteProduct(db *sql.DB, productID string) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := db.Exec(query, productID)
	return err
}

// LogSearchQuery logs a search query to the database with proper error handling.
func LogSearchQuery(db *sql.DB, query string, normalizedTitle string, category string) error {
	// Check if search_logs table exists first
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'search_logs'
		);
	`).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking if search_logs table exists: %v", err)
	}

	if !exists {
		// Create the table if it doesn't exist
		createTableQuery := `
			CREATE TABLE search_logs (
				id SERIAL PRIMARY KEY,
				query TEXT NOT NULL,
				normalized_title TEXT,
				category TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`
		_, err = db.Exec(createTableQuery)
		if err != nil {
			return fmt.Errorf("error creating search_logs table: %v", err)
		}
	}

	insertQuery := `INSERT INTO search_logs (query, normalized_title, category) VALUES ($1, $2, $3)`
	_, err = db.Exec(insertQuery, query, normalizedTitle, category)
	return err
}

// GetSearchStatistics retrieves search statistics from the database with proper error handling.
func GetSearchStatistics(db *sql.DB) (map[string]interface{}, error) {
	// Check if search_logs table exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'search_logs'
		);
	`).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("error checking if search_logs table exists: %v", err)
	}

	if !exists {
		return map[string]interface{}{
			"total_requests":     0,
			"unique_queries":     0,
			"most_popular_query": "No data available",
		}, nil
	}

	query := `
		SELECT
			COUNT(*) AS total_requests,
			COALESCE(
				(SELECT query FROM search_logs GROUP BY query ORDER BY COUNT(*) DESC LIMIT 1),
				'No queries yet'
			) AS most_popular_query,
			COALESCE(
				(SELECT COUNT(*) FROM (SELECT DISTINCT query FROM search_logs) AS distinct_queries),
				0
			) AS unique_queries
		FROM search_logs;
	`
	row := db.QueryRow(query)

	var totalRequests, uniqueQueries int
	var mostPopularQuery string
	if err := row.Scan(&totalRequests, &mostPopularQuery, &uniqueQueries); err != nil {
		return nil, fmt.Errorf("error getting search statistics: %v", err)
	}

	stats := map[string]interface{}{
		"total_requests":     totalRequests,
		"unique_queries":     uniqueQueries,
		"most_popular_query": mostPopularQuery,
	}

	return stats, nil
}
