package repository

import (
	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/gorm"
)

type SecretRepository interface {
	Create(secret *models.SecretNode) error
	GetByID(id uint) (*models.SecretNode, error)
	GetVersions(secretID uint) ([]models.SecretVersion, error)
	Delete(secretID uint) error
}

type secretRepo struct {
	db *gorm.DB
}

func NewSecretRepository(db *gorm.DB) SecretRepository {
	return &secretRepo{db}
}

func (r *secretRepo) Create(secret *models.SecretNode) error {
	return r.db.Create(secret).Error
}

func (r *secretRepo) GetByID(id uint) (*models.SecretNode, error) {
	var secret models.SecretNode
	err := r.db.First(&secret, id).Error
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

func (r *secretRepo) GetVersions(secretID uint) ([]models.SecretVersion, error) {
	var versions []models.SecretVersion
	err := r.db.Where("secret_node_id = ?", secretID).Find(&versions).Error
	return versions, err
}

func (r *secretRepo) Delete(secretID uint) error {
	return r.db.Delete(&models.SecretNode{}, secretID).Error
}
