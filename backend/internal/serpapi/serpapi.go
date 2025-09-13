package serpapi

import (
	"fmt"
	"log"
	"sync"

	serpapi "github.com/serpapi/google-search-results-golang"
)

var (
	apiKeys         []string
	currentKeyIndex int
	mu              sync.Mutex
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
		"engine": "google_shopping",
		"q":      query,
		"gl":     "ch",
		"hl":     "en",
	}

	search := serpapi.NewGoogleSearch(parameter, getNextApiKey())
	result, err := search.GetJSON()
	if err != nil {
		return "", err
	}

	log.Printf("SerpApi response keys: %v", getMapKeys(result))

	// Проверяем shopping_results
	if shoppingResults, ok := result["shopping_results"].([]interface{}); ok && len(shoppingResults) > 0 {
		if firstProduct, ok := shoppingResults[0].(map[string]interface{}); ok {
			if productID, ok := firstProduct["product_id"].(string); ok && productID != "" {
				log.Printf("Found product_id in shopping_results: %s", productID)
				return productID, nil
			}
		}
	}

	// Проверяем product_results (альтернативный путь)
	if productResults, ok := result["product_results"].([]interface{}); ok && len(productResults) > 0 {
		if firstProduct, ok := productResults[0].(map[string]interface{}); ok {
			if productID, ok := firstProduct["product_id"].(string); ok && productID != "" {
				log.Printf("Found product_id in product_results: %s", productID)
				return productID, nil
			}
		}
	}

	return "", fmt.Errorf("no product_id found in response")
}

func GetProductOffers(productID string) ([]Offer, error) {
	parameter := map[string]string{
		"engine":     "google_product",
		"product_id": productID,
		"gl":         "ch",
		"hl":         "en",
		"offers":     "1",
	}

	search := serpapi.NewGoogleSearch(parameter, getNextApiKey())
	result, err := search.GetJSON()
	if err != nil {
		return nil, err
	}

	log.Printf("Product offers response keys: %v", getMapKeys(result))

	var offers []Offer

	// Проверяем sellers (основной путь для офферов)
	if sellers, ok := result["sellers"].([]interface{}); ok && len(sellers) > 0 {
		log.Printf("Found %d sellers", len(sellers))
		for i, sellerData := range sellers {
			if i >= 10 { // Ограничиваем количество офферов
				break
			}
			if sellerMap, ok := sellerData.(map[string]interface{}); ok {
				offer := parseSellerOffer(sellerMap)
				if offer.Merchant != "" && offer.Price > 0 {
					offers = append(offers, offer)
				}
			}
		}
	}

	// Альтернативно проверяем offers
	if offersData, ok := result["offers"].([]interface{}); ok && len(offersData) > 0 {
		log.Printf("Found %d offers", len(offersData))
		for i, offerData := range offersData {
			if i >= 10 { // Ограничиваем количество офферов
				break
			}
			if offerMap, ok := offerData.(map[string]interface{}); ok {
				offer := parseOfferData(offerMap)
				if offer.Merchant != "" && offer.Price > 0 {
					offers = append(offers, offer)
				}
			}
		}
	}

	if len(offers) == 0 {
		return nil, fmt.Errorf("no offers found for product %s", productID)
	}

	log.Printf("Successfully parsed %d offers", len(offers))
	return offers, nil
}

// parseSellerOffer парсит данные продавца из sellers
func parseSellerOffer(sellerMap map[string]interface{}) Offer {
	offer := Offer{}

	// Извлекаем название магазина
	if name, ok := sellerMap["name"].(string); ok {
		offer.Merchant = name
	}

	// Извлекаем цену
	if priceStr, ok := sellerMap["price"].(string); ok {
		offer.Price = parsePrice(priceStr)
	} else if priceFloat, ok := sellerMap["price"].(float64); ok {
		offer.Price = priceFloat
	}

	// Извлекаем ссылку
	if link, ok := sellerMap["link"].(string); ok {
		offer.Link = link
	}

	return offer
}

// parseOfferData парсит данные из offers
func parseOfferData(offerMap map[string]interface{}) Offer {
	offer := Offer{}

	// Извлекаем название магазина
	if merchant, ok := offerMap["merchant"].(string); ok {
		offer.Merchant = merchant
	} else if source, ok := offerMap["source"].(string); ok {
		offer.Merchant = source
	}

	// Извлекаем цену
	if priceStr, ok := offerMap["price"].(string); ok {
		offer.Price = parsePrice(priceStr)
	} else if priceFloat, ok := offerMap["price"].(float64); ok {
		offer.Price = priceFloat
	}

	// Извлекаем ссылку
	if link, ok := offerMap["link"].(string); ok {
		offer.Link = link
	}

	return offer
}

// parsePrice извлекает числовое значение цены из строки
func parsePrice(priceStr string) float64 {
	// Простой парсинг цены - извлекаем числа
	var price float64
	fmt.Sscanf(priceStr, "%f", &price)
	return price
}

// getMapKeys возвращает ключи карты для отладки
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
