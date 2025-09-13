package storage

import (
	"fmt"
	"pricerunner-parser/internal/config"
)

// NewStorage creates a storage instance based on configuration
func NewStorage(cfg *config.Config) (Storage, error) {
	switch cfg.Storage.Type {
	case "json":
		return NewJSONStorage(cfg.Storage.OutputDir), nil
	case "database":
		return NewDatabaseStorage(cfg.Storage.Database)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}
