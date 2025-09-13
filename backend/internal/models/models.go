package models

// SearchRequest represents the expected JSON structure of a search query
type SearchRequest struct {
	Query  string `json:"query"`
	Lang   string `json:"lang"`
	Region string `json:"region"`
}

// NormalizedResponse represents the structured data returned by the AI normalizer service.
type NormalizedResponse struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Features []string `json:"features"`
}

// Product represents a product document in Meilisearch.
type Product struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Category        string   `json:"category"`
	Features        []string `json:"features"`
	GoogleProductID string   `json:"google_product_id"`
	ImageURL        string   `json:"image_url"`
}
