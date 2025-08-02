package common

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/core"
	"github.com/secretlyhq/secretly/internal/storage"
)

// InitializeCoreService creates a core service instance using the storage factory
// This function should be used by all CLI commands instead of directly creating storage
func InitializeCoreService() (*core.SecretlyCore, error) {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		// If no config file exists, use default local storage
		cfg = &config.Config{
			Storage: config.StorageConfig{
				Type: "local",
				Database: config.DatabaseConfig{
					Path: "./secrets.db",
				},
			},
		}
	}

	// Create storage using factory
	factory := storage.NewStorageFactory()
	storageImpl, err := factory.CreateStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Create and return core service
	return core.NewSecretlyCore(storageImpl), nil
}