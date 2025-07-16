package repository

import (
	"errors"

	"gorm.io/gorm"
)

// Configuration представляет строку конфигурации
type Configuration struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

type ConfigRepository interface {
	Get(key string) (string, error)
	Set(key string, value string) error
}

type configRepo struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepo{db}
}

// Get возвращает значение конфигурации по ключу
func (r *configRepo) Get(key string) (string, error) {
	var config Configuration
	err := r.db.First(&config, "key = ?", key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // или вернуть ошибку — зависит от логики приложения
		}
		return "", err
	}
	return config.Value, nil
}

// Set устанавливает значение конфигурации по ключу
func (r *configRepo) Set(key string, value string) error {
	var config Configuration
	err := r.db.First(&config, "key = ?", key).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// создаём новую запись
		return r.db.Create(&Configuration{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}

	// обновляем существующую запись
	config.Value = value
	return r.db.Save(&config).Error
}
