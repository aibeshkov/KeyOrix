package storage

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/core/storage"
	"github.com/secretlyhq/secretly/internal/storage/local"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/secretlyhq/secretly/internal/storage/remote"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// StorageFactory creates storage instances based on configuration
type StorageFactory interface {
	CreateStorage(config *config.Config) (storage.Storage, error)
}

// DefaultStorageFactory is the default implementation of StorageFactory
type DefaultStorageFactory struct{}

// NewStorageFactory creates a new storage factory
func NewStorageFactory() StorageFactory {
	return &DefaultStorageFactory{}
}

// CreateStorage creates a storage instance based on the configuration
func (f *DefaultStorageFactory) CreateStorage(config *config.Config) (storage.Storage, error) {
	// Default to local storage if type is not specified or is "local"
	switch config.Storage.Type {
	case "remote":
		return f.createRemoteStorage(config)
	default: // "local", "", or any other value defaults to local
		return f.createLocalStorage(config)
	}
}

// createLocalStorage creates a local storage instance
func (f *DefaultStorageFactory) createLocalStorage(config *config.Config) (storage.Storage, error) {
	// Use default path if not specified
	dbPath := config.Storage.Database.Path
	if dbPath == "" {
		dbPath = "./secrets.db"
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models (import the actual models)
	if err := f.migrateDatabase(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return local.NewLocalStorage(db), nil
}

// createRemoteStorage creates a remote storage instance
func (f *DefaultStorageFactory) createRemoteStorage(config *config.Config) (storage.Storage, error) {
	if config.Storage.Remote == nil {
		return nil, fmt.Errorf("remote storage configuration is required")
	}

	remoteConfig := &remote.Config{
		BaseURL:        config.Storage.Remote.BaseURL,
		APIKey:         config.Storage.Remote.APIKey,
		TimeoutSeconds: config.Storage.Remote.TimeoutSeconds,
		RetryAttempts:  config.Storage.Remote.RetryAttempts,
		TLSVerify:      config.Storage.Remote.TLSVerify,
	}

	return remote.NewRemoteStorage(remoteConfig)
}

// migrateDatabase performs database migrations
func (f *DefaultStorageFactory) migrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Namespace{},
		&models.Zone{},
		&models.Environment{},
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.Group{},
		&models.UserGroup{},
		&models.GroupRole{},
		&models.SecretNode{},
		&models.SecretVersion{},
		&models.SecretAccessLog{},
		&models.SecretMetadataHistory{},
		&models.ShareRecord{},
		&models.Session{},
		&models.PasswordReset{},
		&models.Tag{},
		&models.SecretTag{},
		&models.Notification{},
		&models.AuditEvent{},
		&models.Setting{},
		&models.SystemMetadata{},
		&models.APIClient{},
		&models.APIToken{},
		&models.RateLimit{},
		&models.APICallLog{},
		&models.GRPCService{},
		&models.IdentityProvider{},
		&models.ExternalIdentity{},
	)
}