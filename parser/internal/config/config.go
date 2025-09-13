package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Parser   ParserConfig   `yaml:"parser"`
	Currency CurrencyConfig `yaml:"currency"`
	Storage  StorageConfig  `yaml:"storage"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ParserConfig contains parsing-related settings
type ParserConfig struct {
	BaseURL   string          `yaml:"base_url"`
	Browser   BrowserConfig   `yaml:"browser"`
	Parsing   ParsingConfig   `yaml:"parsing"`
	Selectors SelectorsConfig `yaml:"selectors"`
}

// BrowserConfig contains browser settings
type BrowserConfig struct {
	Headless  bool           `yaml:"headless"`
	Timeout   int            `yaml:"timeout"`
	Viewport  ViewportConfig `yaml:"viewport"`
	UserAgent string         `yaml:"user_agent"`
}

// ViewportConfig contains viewport settings
type ViewportConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

// ParsingConfig contains parsing behavior settings
type ParsingConfig struct {
	MaxPages             int `yaml:"max_pages"`
	DelayBetweenRequests int `yaml:"delay_between_requests"`
	ScrollDelay          int `yaml:"scroll_delay"`
	MaxScrolls           int `yaml:"max_scrolls"`
}

// SelectorsConfig contains CSS selectors
type SelectorsConfig struct {
	ProductCards     string `yaml:"product_cards"`
	Price            string `yaml:"price"`
	MainImage        string `yaml:"main_image"`
	AdditionalImages string `yaml:"additional_images"`
	FeatureTables    string `yaml:"feature_tables"`
	NextPageButton   string `yaml:"next_page_button"`
}

// CurrencyConfig contains currency conversion settings
type CurrencyConfig struct {
	GBPToEUR float64 `yaml:"gbp_to_eur"`
}

// StorageConfig contains storage settings
type StorageConfig struct {
	Type      string         `yaml:"type"`
	OutputDir string         `yaml:"output_dir"`
	ImagesDir string         `yaml:"images_dir"`
	Database  DatabaseConfig `yaml:"database"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// Load reads and parses the configuration file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Создаем необходимые директории
	if err := os.MkdirAll(config.Storage.OutputDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(config.Storage.ImagesDir, 0755); err != nil {
		return nil, err
	}

	return &config, nil
}
