package serpapi

import (
	"fmt"
	"sync"

	serpapi "github.com/serpapi/google-search-results-golang"
)

var (
	apiKeys []string
	currentKeyIndex int
	mu sync.Mutex
)

// Offer represents a product offer from a merchant.

type Offer struct {
	Merchant string  `json:"merchant"`
	Price    float64 `json:"price"`
	Link     string  `json:"link"`
}

func InitSerpApi(keys []string) {
	apiKeys = keys
}

func getNextApiKey() string {
	mu.Lock()
	defer mu.Unlock()
	key := apiKeys[currentKeyIndex]
	currentKeyIndex = (currentKeyIndex + 1) % len(apiKeys)
	return key
}

func SearchProduct(query string) (string, error) {
	parameter := map[string]string{
		"engine": "google_product",
		"q":      query,
	}

	search := serpapi.NewGoogleSearch(parameter, getNextApiKey())
	result, err := search.GetJSON()
	if err != nil {
		return "", err
	}

	productResults, ok := result["product_results"].([]interface{})
	if !ok || len(productResults) == 0 {
		return "", fmt.Errorf("no product results found")
	}

	firstProduct, ok := productResults[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid product result format")
	}

	googleProductID, ok := firstProduct["product_id"].(string)
	if !ok {
		return "", fmt.Errorf("google_product_id not found")
	}

	return googleProductID, nil
}

func GetProductOffers(productID string) ([]Offer, error) {
	parameter := map[string]string{
		"engine":     "google_product",
		"product_id": productID,
		"offers":     "true",
	}

	search := serpapi.NewGoogleSearch(parameter, getNextApiKey())
	result, err := search.GetJSON()
	if err != nil {
		return nil, err
	}

	offersResults, ok := result["offers"].([]interface{})
	if !ok || len(offersResults) == 0 {
		return nil, fmt.Errorf("no offers found")
	}

	var offers []Offer
	for _, offerData := range offersResults {
		offerMap, ok := offerData.(map[string]interface{})
		if !ok {
			continue
		}

		offer := Offer{
			Merchant: offerMap["merchant"].(string),
			Price:    offerMap["price"].(float64),
			Link:     offerMap["link"].(string),
		}
		offers = append(offers, offer)
	}

	return offers, nil
}
