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

	// Configure MeiliSearch settings
	err := configureMeiliSearch(client)
	if err != nil {
		log.Printf("Error configuring MeiliSearch: %v", err)
	}

	return client
}

// configureMeiliSearch sets up the filterable and searchable attributes
func configureMeiliSearch(client meilisearch.ServiceManager) error {
	// Get or create the products index
	index := client.Index("products")

	// Configure filterable attributes
	filterableAttributes := []string{"category", "features"}
	filterableAttrsInterface := make([]interface{}, len(filterableAttributes))
	for i, v := range filterableAttributes {
		filterableAttrsInterface[i] = v
	}
	_, err := index.UpdateFilterableAttributes(&filterableAttrsInterface)
	if err != nil {
		log.Printf("Error setting filterable attributes: %v", err)
		return err
	}

	// Configure searchable attributes
	searchableAttributes := []string{"title", "category", "features"}
	_, err = index.UpdateSearchableAttributes(&searchableAttributes)
	if err != nil {
		log.Printf("Error setting searchable attributes: %v", err)
		return err
	}

	// Configure sortable attributes (optional)
	sortableAttributes := []string{"title", "category"}
	_, err = index.UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		log.Printf("Error setting sortable attributes: %v", err)
		return err
	}

	log.Println("MeiliSearch configuration updated successfully")
	return nil
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

// buildMeiliSearchFilter creates a proper filter string for MeiliSearch
func buildMeiliSearchFilter(query models.NormalizedResponse) string {
	var filters []string

	// Add category filter if not empty
	if query.Category != "" {
		filters = append(filters, fmt.Sprintf("category = '%s'", query.Category))
	}

	// Add features filters - use OR instead of AND for better results
	if len(query.Features) > 0 {
		var featureFilters []string
		for _, feature := range query.Features {
			if feature != "" {
				featureFilters = append(featureFilters, fmt.Sprintf("features = '%s'", feature))
			}
		}
		if len(featureFilters) > 0 {
			featuresFilter := strings.Join(featureFilters, " OR ")
			if len(featureFilters) > 1 {
				featuresFilter = "(" + featuresFilter + ")"
			}
			filters = append(filters, featuresFilter)
		}
	}

	return strings.Join(filters, " AND ")
}

// Search performs a search on the products index with improved error handling.
func Search(client meilisearch.ServiceManager, query models.NormalizedResponse) (*meilisearch.SearchResponse, error) {
	index := client.Index("products")

	// Build MeiliSearch filter
	filter := buildMeiliSearchFilter(query)

	searchReq := &meilisearch.SearchRequest{
		Filter: filter,
		Limit:  10,
	}

	searchRes, err := index.Search(query.Title, searchReq)
	if err != nil {
		log.Printf("Error searching with filter '%s': %v", filter, err)
		// Fallback to simple text search without filters
		searchRes, err = index.Search(query.Title, &meilisearch.SearchRequest{Limit: 10})
		if err != nil {
			log.Printf("Fallback search also failed: %v", err)
			return nil, fmt.Errorf("search failed: %v", err)
		}
		log.Printf("Fallback search succeeded, returned %d results", len(searchRes.Hits))
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
