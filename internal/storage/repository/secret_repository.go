package repository

import (
	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/gorm"
)

type SecretRepository interface {
	Create(secret *models.SecretNode) error
	GetByID(id uint) (*models.SecretNode, error)
	GetByName(name string, namespaceID, zoneID, environmentID uint) (*models.SecretNode, error)
	List(namespaceID, zoneID, environmentID uint, limit, offset int) ([]models.SecretNode, error)
	Update(secret *models.SecretNode) error
	Delete(secretID uint) error
	GetVersions(secretID uint) ([]models.SecretVersion, error)
	GetLatestVersion(secretID uint) (*models.SecretVersion, error)
	CreateVersion(version *models.SecretVersion) error
	UpdateVersion(version *models.SecretVersion) error
	Search(query string, namespaceID, zoneID, environmentID uint) ([]models.SecretNode, error)
	Count(namespaceID, zoneID, environmentID uint) (int64, error)
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

func (r *secretRepo) GetByName(name string, namespaceID, zoneID, environmentID uint) (*models.SecretNode, error) {
	var secret models.SecretNode
	err := r.db.Where("name = ? AND namespace_id = ? AND zone_id = ? AND environment_id = ?",
		name, namespaceID, zoneID, environmentID).First(&secret).Error
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

func (r *secretRepo) List(namespaceID, zoneID, environmentID uint, limit, offset int) ([]models.SecretNode, error) {
	var secrets []models.SecretNode
	query := r.db.Where("namespace_id = ? AND zone_id = ? AND environment_id = ?",
		namespaceID, zoneID, environmentID)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("created_at DESC").Find(&secrets).Error
	return secrets, err
}

func (r *secretRepo) Update(secret *models.SecretNode) error {
	return r.db.Save(secret).Error
}

func (r *secretRepo) Delete(secretID uint) error {
	return r.db.Delete(&models.SecretNode{}, secretID).Error
}

func (r *secretRepo) GetLatestVersion(secretID uint) (*models.SecretVersion, error) {
	var version models.SecretVersion
	err := r.db.Where("secret_node_id = ?", secretID).
		Order("version_number DESC").
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *secretRepo) CreateVersion(version *models.SecretVersion) error {
	return r.db.Create(version).Error
}

func (r *secretRepo) UpdateVersion(version *models.SecretVersion) error {
	return r.db.Save(version).Error
}

func (r *secretRepo) Search(query string, namespaceID, zoneID, environmentID uint) ([]models.SecretNode, error) {
	var secrets []models.SecretNode
	searchPattern := "%" + query + "%"

	err := r.db.Where("namespace_id = ? AND zone_id = ? AND environment_id = ? AND (name LIKE ? OR type LIKE ?)",
		namespaceID, zoneID, environmentID, searchPattern, searchPattern).
		Order("created_at DESC").
		Find(&secrets).Error

	return secrets, err
}

func (r *secretRepo) Count(namespaceID, zoneID, environmentID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.SecretNode{}).
		Where("namespace_id = ? AND zone_id = ? AND environment_id = ?",
			namespaceID, zoneID, environmentID).
		Count(&count).Error
	return count, err
}
