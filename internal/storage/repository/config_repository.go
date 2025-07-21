package repository

import (
	"errors"

	"gorm.io/gorm"
)

// Configuration represents a key-value configuration record
type Configuration struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// ConfigRepository defines the interface for interacting with configuration storage
type ConfigRepository interface {
	Get(key string) (string, error)
	Set(key string, value string) error
}

type configRepo struct {
	db *gorm.DB
}

// NewConfigRepository creates a new instance of ConfigRepository
func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepo{db}
}

// Get returns the configuration value for the given key
func (r *configRepo) Get(key string) (string, error) {
	var config Configuration
	err := r.db.First(&config, "key = ?", key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // or return an error — depending on application logic
		}
		return "", err
	}
	return config.Value, nil
}

// Set sets the configuration value for the given key
func (r *configRepo) Set(key string, value string) error {
	var config Configuration
	err := r.db.First(&config, "key = ?", key).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// create a new record
		return r.db.Create(&Configuration{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}

	// update existing record
	config.Value = value
	return r.db.Save(&config).Error
}