package handlers

import (
	"testing"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/core"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/internal/storage/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestShareHandlerCreation(t *testing.T) {
	// Initialize i18n for tests
	cfg := &config.Config{
		Locale: config.LocaleConfig{
			Language:         "en",
			FallbackLanguage: "en",
		},
	}
	err := i18n.Initialize(cfg)
	require.NoError(t, err)

	// Create an in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create storage and core service
	storage := local.NewLocalStorage(db)
	coreService := core.NewSecretlyCore(storage)

	// Create share handler
	handler, err := NewShareHandler(coreService)
	require.NoError(t, err)
	assert.NotNil(t, handler)
}

func TestShareHandlerBasicValidation(t *testing.T) {
	// Test that NewShareHandler accepts nil input (no validation currently)
	handler, err := NewShareHandler(nil)
	assert.NoError(t, err)
	assert.NotNil(t, handler)
	
	// The handler will have a nil coreService, which would cause issues in actual use
	// but the constructor doesn't validate this
}