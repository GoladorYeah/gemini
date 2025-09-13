package search

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gemini/backend/internal/models"

	"github.com/meilisearch/meilisearch-go"
)

// NewClient creates and returns a new Meilisearch client.
func NewClient() meilisearch.ServiceManager {
	client := meilisearch.New("http://meilisearch:7700")
	return client
}

// IndexSampleProducts adds some sample products to the Meilisearch index.
func IndexSampleProducts(client meilisearch.ServiceManager) *meilisearch.TaskInfo {
	index := client.Index("products")

	documents := []models.Product{
		{
			ID:       "1",
			Title:    "New Balance Little Kid's 530 Bungee - Moonbeam/Phantom",
			Category: "Kids Shoes",
			Features: []string{"bungee", "easy on/off", "brand: New Balance"},
		},
		{
			ID:       "2",
			Title:    "New Balance Little Kid's 530 Bungee - Blue/White",
			Category: "Kids Shoes",
			Features: []string{"bungee", "easy on/off", "brand: New Balance"},
		},
		{
			ID:       "3",
			Title:    "Apple iPhone 16 Pro Max, 256GB Desert Titanium",
			Category: "Smartphones",
			Features: []string{"256GB", "color: Desert Titanium", "brand: Apple"},
		},
	}

	primaryKey := "ID"
	task, err := index.AddDocuments(documents, &primaryKey)
	if err != nil {
		log.Printf("Error adding documents: %v", err)
		return nil
	}

	log.Printf("Added documents to Meilisearch. Task ID: %d", task.TaskUID)
	return task
}

// Search performs a search on the products index.
func Search(client meilisearch.ServiceManager, query models.NormalizedResponse) (*meilisearch.SearchResponse, error) {
	index := client.Index("products")

	var filter string
	if query.Category != "" {
		filter = fmt.Sprintf("category = '%s'", query.Category)
	}

	if len(query.Features) > 0 {
		var featuresFilter []string
		for _, feature := range query.Features {
			featuresFilter = append(featuresFilter, fmt.Sprintf("features = '%s'", feature))
		}
		if filter != "" {
			filter = fmt.Sprintf("%s AND (%s)", filter, strings.Join(featuresFilter, " AND "))
		} else {
			filter = strings.Join(featuresFilter, " AND ")
		}
	}

	searchRes, err := index.Search(query.Title, &meilisearch.SearchRequest{
		Limit:  10,
		Filter: filter,
	})

	if err != nil {
		log.Printf("Error searching: %v", err)
		return nil, err
	}

	return searchRes, nil
}

// AddProduct adds a product to the Meilisearch index.
func AddProduct(client meilisearch.ServiceManager, product models.Product) (*meilisearch.TaskInfo, error) {
	index := client.Index("products")
	primaryKey := "ID"
	task, err := index.AddDocuments([]models.Product{product}, &primaryKey)
	if err != nil {
		log.Printf("Error adding document: %v", err)
		return nil, err
	}

	return task, nil
}

// UpdateProduct updates a product in the Meilisearch index.
func UpdateProduct(client meilisearch.ServiceManager, product models.Product) (*meilisearch.TaskInfo, error) {
	index := client.Index("products")
	primaryKey := "ID"
	task, err := index.UpdateDocuments([]models.Product{product}, &primaryKey)
	if err != nil {
		log.Printf("Error updating document: %v", err)
		return nil, err
	}

	return task, nil
}

// DeleteProduct deletes a product from the Meilisearch index.
func DeleteProduct(client meilisearch.ServiceManager, productID string) (*meilisearch.TaskInfo, error) {
	index := client.Index("products")
	task, err := index.DeleteDocument(productID)
	if err != nil {
		log.Printf("Error deleting document: %v", err)
		return nil, err
	}

	return task, nil
}

// GetProduct retrieves a product from the Meilisearch index.
func GetProduct(client meilisearch.ServiceManager, productID string) (*models.Product, error) {
	index := client.Index("products")

	var product models.Product
	err := index.GetDocument(productID, nil, &product)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		return nil, err
	}

	return &product, nil
}

// GetAllProducts retrieves all products from the Meilisearch index.
func GetAllProducts(client meilisearch.ServiceManager) ([]models.Product, error) {
	index := client.Index("products")

	var products []models.Product
	// This is not efficient for large datasets, but it's fine for this example.
	// A better approach would be to use pagination.
	searchRes, err := index.Search("", &meilisearch.SearchRequest{Limit: 1000})
	if err != nil {
		log.Printf("Error getting all documents: %v", err)
		return nil, err
	}

	for _, hit := range searchRes.Hits {
		var product models.Product
		jsonBytes, _ := json.Marshal(hit)
		json.Unmarshal(jsonBytes, &product)
		products = append(products, product)
	}

	return products, nil
}
