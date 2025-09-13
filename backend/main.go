package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"gemini/backend/internal/cache"
	"gemini/backend/internal/models"
	"gemini/backend/internal/normalizer"
	"gemini/backend/internal/search"
	"gemini/backend/internal/serpapi"
	"gemini/backend/internal/storage"
	"github.com/meilisearch/meilisearch-go"
	_ "github.com/lib/pq"
)

var (
	meiliClient meilisearch.ServiceManager
	db          *sql.DB
)

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		fn(w, r)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received search query: %+v\n", req)

	normalizedResp, err := normalizer.Normalize(req)
	if err != nil {
		http.Error(w, "Error normalizing query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Normalized response: %+v\n", normalizedResp)

	if err := storage.LogSearchQuery(db, req.Query, normalizedResp.Title, normalizedResp.Category); err != nil {
		log.Printf("Error logging search query: %v", err)
	}

	searchResults, err := search.Search(meiliClient, *normalizedResp)
	if err != nil {
		http.Error(w, "Error searching: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var products []models.Product
	jsonBytes, _ := json.Marshal(searchResults.Hits)
	json.Unmarshal(jsonBytes, &products)

	for i := range products {
		if products[i].GoogleProductID == "" {
			log.Printf("Product %s does not have a Google Product ID. Calling SerpApi.", products[i].Title)
			googleProductID, err := serpapi.SearchProduct(products[i].Title)
			if err != nil {
				log.Printf("Error searching for product %s on SerpApi: %v", products[i].Title, err)
			} else {
				log.Printf("Found Google Product ID for %s: %s", products[i].Title, googleProductID)
				products[i].GoogleProductID = googleProductID
				_, err := search.UpdateProduct(meiliClient, products[i])
				if err != nil {
					log.Printf("Error updating product %s in Meilisearch: %v", products[i].Title, err)
				}
			}
		} else {
			log.Printf("Product %s has a Google Product ID: %s", products[i].Title, products[i].GoogleProductID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func productOffersHandler(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimPrefix(r.URL.Path, "/api/product/")
	productID = strings.TrimSuffix(productID, "/offers")

	// Check cache first
	var offers []serpapi.Offer
	if err := cache.Get(productID, &offers); err == nil {
		log.Printf("Cache hit for product %s", productID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(offers)
		return
	}

	log.Printf("Cache miss for product %s", productID)

	// Get product from Meilisearch
	product, err := search.GetProduct(meiliClient, productID)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	if product.GoogleProductID == "" {
		http.Error(w, "Product does not have a Google Product ID", http.StatusNotFound)
		return
	}

	// Get offers from SerpApi
	offers, err = serpapi.GetProductOffers(product.GoogleProductID)
	if err != nil {
		http.Error(w, "Error getting offers from SerpApi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache the offers for 1 day
	cache.Set(productID, offers, 24*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(offers)
}

func parserStartHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://parser:8082/start", "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Error calling parser service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func parserStopHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://parser:8082/stop", "application/json", nil)
	if err != nil {
		http.Error(w, "Error calling parser service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func parserStatusHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://parser:8082/status")
	if err != nil {
		http.Error(w, "Error calling parser service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func apiKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		geminiKeys := os.Getenv("GEMINI_API_KEYS")
		serpApiKeys := os.Getenv("SERPAPI_API_KEYS")

		maskedGeminiKeys := ""
		if len(geminiKeys) > 0 {
			maskedGeminiKeys = "..." + geminiKeys[len(geminiKeys)-4:]
		}

		maskedSerpApiKeys := ""
		if len(serpApiKeys) > 0 {
			maskedSerpApiKeys = "..." + serpApiKeys[len(serpApiKeys)-4:]
		}

		json.NewEncoder(w).Encode(map[string]string{
			"gemini_api_keys": maskedGeminiKeys,
			"serpapi_api_keys": maskedSerpApiKeys,
		})
	} else if r.Method == http.MethodPost {
		var keys struct {
			GeminiKeys  string `json:"gemini_api_keys"`
			SerpApiKeys string `json:"serpapi_api_keys"`
		}
		if err := json.NewDecoder(r.Body).Decode(&keys); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// In a real application, you would securely store these keys.
		// For now, we just log that we received them.
		log.Printf("Received new API keys. Gemini: %s, SerpApi: %s", keys.GeminiKeys, keys.SerpApiKeys)

		// You would also need to re-initialize the clients with the new keys.

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "API keys updated"})
	}
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	service := strings.TrimPrefix(r.URL.Path, "/api/admin/logs/")
	if service == "" {
		http.Error(w, "Service not specified", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("docker", "logs", "gemini-"+service+"-1")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	if err := cmd.Run(); err != nil {
		http.Error(w, fmt.Sprintf("Error getting logs for %s: %v\n%s", service, err, out.String()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(out.Bytes())
}

func productsAdminHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		products, err := storage.GetAllProducts(db)
		if err != nil {
			http.Error(w, "Error getting products from database", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	case http.MethodPost:
		var product models.Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := storage.AddProduct(db, product); err != nil {
			http.Error(w, "Error adding product to database", http.StatusInternalServerError)
			return
		}

		if _, err := search.AddProduct(meiliClient, product); err != nil {
			http.Error(w, "Error adding product to Meilisearch", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case http.MethodPut:
		var product models.Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := storage.UpdateProduct(db, product); err != nil {
			http.Error(w, "Error updating product in database", http.StatusInternalServerError)
			return
		}

		if _, err := search.UpdateProduct(meiliClient, product); err != nil {
			http.Error(w, "Error updating product in Meilisearch", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		productID := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
		if err := storage.DeleteProduct(db, productID); err != nil {
			http.Error(w, "Error deleting product from database", http.StatusInternalServerError)
			return
		}

		if _, err := search.DeleteProduct(meiliClient, productID); err != nil {
			http.Error(w, "Error deleting product from Meilisearch", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func statisticsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := storage.GetSearchStatistics(db)
	if err != nil {
		http.Error(w, "Error getting statistics from database", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func main() {
	geminiKeys := os.Getenv("GEMINI_API_KEYS")
	serpApiKeys := os.Getenv("SERPAPI_API_KEYS")

	if geminiKeys == "" || serpApiKeys == "" {
		log.Fatal("API keys are not set. Please set GEMINI_API_KEYS and SERPAPI_API_KEYS environment variables.")
	}

	// Initialize Gemini client
	normalizer.InitGemini(strings.Split(geminiKeys, ","))

	// Initialize SerpApi client
	serpapi.InitSerpApi(strings.Split(serpApiKeys, ","))

	// Initialize Redis client
	cache.InitRedis()

	// Initialize Database
	var err error
	db, err = storage.NewDB("postgres://user:password@postgres/geminidb?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Meilisearch client
	meiliClient = search.NewClient()
	// Index some sample products
	task := search.IndexSampleProducts(meiliClient)
	if task != nil {
		log.Printf("Waiting for indexing task %d to complete...", task.TaskUID)
		taskManager := meiliClient.TaskManager()
		_, err := taskManager.WaitForTask(task.TaskUID, time.Second*5)
		if err != nil {
			log.Printf("Error waiting for task: %v", err)
		}
		log.Println("Indexing task completed.")
	}

	http.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from backend!")
	}))

	http.HandleFunc("/api/search", withCORS(searchHandler))
	http.HandleFunc("/api/product/", withCORS(productOffersHandler))

	// Admin endpoints
	http.HandleFunc("/api/admin/parser/start", withCORS(parserStartHandler))
	http.HandleFunc("/api/admin/parser/stop", withCORS(parserStopHandler))
	http.HandleFunc("/api/admin/parser/status", withCORS(parserStatusHandler))
	http.HandleFunc("/api/admin/keys", withCORS(apiKeysHandler))
	http.HandleFunc("/api/admin/logs/", withCORS(logsHandler))
	http.HandleFunc("/api/admin/products/", withCORS(productsAdminHandler))
	http.HandleFunc("/api/admin/statistics", withCORS(statisticsHandler))

	log.Println("Starting server on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
