package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"pricerunner-parser/internal/config"
	"pricerunner-parser/internal/parser"
	"pricerunner-parser/internal/storage"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

type ParserController struct {
	mu     sync.Mutex
	parser *parser.Parser
	status string
	cfg    *config.Config
}

func NewParserController(cfg *config.Config) *ParserController {
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	p := parser.New(cfg, store)
	return &ParserController{
		parser: p,
		status: "idle",
		cfg:    cfg,
	}
}

func (c *ParserController) runParsing() {
	c.mu.Lock()
	if c.status == "running" {
		c.mu.Unlock()
		log.Println("Parser is already running")
		return
	}
	c.status = "running"
	c.mu.Unlock()

	log.Println("Starting parser...")
	if err := c.parser.Parse(); err != nil {
		log.Printf("Parsing failed: %v", err)
	}
	log.Println("Parsing completed.")

	c.mu.Lock()
	c.status = "idle"
	c.mu.Unlock()
}

func (c *ParserController) startHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL      string `json:"url"`
		Category string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In a real application, you would update the config or pass these to the parser
	log.Printf("Manual start for URL: %s, Category: %s", req.URL, req.Category)

	go c.runParsing()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Parser started"})
}

func (c *ParserController) stopHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status != "running" {
		http.Error(w, "Parser is not running", http.StatusConflict)
		return
	}

	// In a real application, you would need a way to gracefully stop the parser.
	// For now, we'll just log a message.
	log.Println("Received request to stop parser. Graceful stop is not yet implemented.")
	c.status = "stopping"

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Parser stopping"})
}

func (c *ParserController) statusHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": c.status})
}

func (c *ParserController) startScheduler() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			log.Println("Starting scheduled parsing...")
			c.runParsing()
		}
	}()
}

func initPlaywright() error {
	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// Инициализируем Playwright в самом начале
	log.Println("Initializing Playwright...")
	if err := initPlaywright(); err != nil {
		log.Printf("Warning: Playwright initialization failed: %v", err)
		// Продолжаем работу, возможно браузеры уже установлены
	}

	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	controller := NewParserController(cfg)
	controller.startScheduler()

	http.HandleFunc("/start", controller.startHandler)
	http.HandleFunc("/stop", controller.stopHandler)
	http.HandleFunc("/status", controller.statusHandler)

	log.Println("Parser API server starting on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
