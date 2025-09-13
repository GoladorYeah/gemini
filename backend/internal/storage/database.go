package storage

import (
	"database/sql"

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

// GetAllProducts retrieves all products from the database.
func GetAllProducts(db *sql.DB) ([]models.Product, error) {
	query := `SELECT id, title, category, features, google_product_id, image_url FROM products`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		var features pq.StringArray
		if err := rows.Scan(&product.ID, &product.Title, &product.Category, &features, &product.GoogleProductID, &product.ImageURL); err != nil {
			return nil, err
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

// LogSearchQuery logs a search query to the database.
func LogSearchQuery(db *sql.DB, query string, normalizedTitle string, category string) error {
	insertQuery := `INSERT INTO search_logs (query, normalized_title, category) VALUES ($1, $2, $3)`
	_, err := db.Exec(insertQuery, query, normalizedTitle, category)
	return err
}

// GetSearchStatistics retrieves search statistics from the database.
func GetSearchStatistics(db *sql.DB) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) AS total_requests,
			(SELECT query FROM search_logs GROUP BY query ORDER BY COUNT(*) DESC LIMIT 1) AS most_popular_query,
			(SELECT COUNT(*) FROM (SELECT DISTINCT query FROM search_logs) AS distinct_queries) AS unique_queries
		FROM search_logs;
	`
	row := db.QueryRow(query)

	var totalRequests, uniqueQueries int
	var mostPopularQuery string
	if err := row.Scan(&totalRequests, &mostPopularQuery, &uniqueQueries); err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_requests":     totalRequests,
		"unique_queries":     uniqueQueries,
		"most_popular_query": mostPopularQuery,
	}

	return stats, nil
}
