package services

import (
	"fmt"
	"time"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/encryption"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/secretlyhq/secretly/internal/storage/repository"
	"gorm.io/gorm"
)

// SecretService provides business logic for secret management
type SecretService struct {
	repo       repository.SecretRepository
	encryption *encryption.SecretEncryption
	config     *config.Config
}

// SecretCreateRequest represents a request to create a new secret
type SecretCreateRequest struct {
	Name          string
	Value         []byte
	Type          string
	NamespaceID   uint
	ZoneID        uint
	EnvironmentID uint
	MaxReads      *int
	Expiration    *time.Time
	Metadata      map[string]interface{}
	CreatedBy     string
}

// SecretUpdateRequest represents a request to update a secret
type SecretUpdateRequest struct {
	ID         uint
	Value      []byte
	Type       string
	MaxReads   *int
	Expiration *time.Time
	Metadata   map[string]interface{}
	UpdatedBy  string
}

// SecretResponse represents a secret response
type SecretResponse struct {
	ID            uint                   `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	NamespaceID   uint                   `json:"namespace_id"`
	ZoneID        uint                   `json:"zone_id"`
	EnvironmentID uint                   `json:"environment_id"`
	MaxReads      *int                   `json:"max_reads,omitempty"`
	Expiration    *time.Time             `json:"expiration,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Status        string                 `json:"status"`
	CreatedBy     string                 `json:"created_by"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	VersionCount  int                    `json:"version_count"`
}

// SecretValueResponse represents a secret with its decrypted value
type SecretValueResponse struct {
	SecretResponse
	Value []byte `json:"value"`
}

// ListOptions represents options for listing secrets
type ListOptions struct {
	NamespaceID   uint
	ZoneID        uint
	EnvironmentID uint
	Limit         int
	Offset        int
	Search        string
}

// NewSecretService creates a new secret service
func NewSecretService(repo repository.SecretRepository, encryptionCfg *config.EncryptionConfig, baseDir string, db *gorm.DB, cfg *config.Config) *SecretService {
	secretEncryption := encryption.NewSecretEncryption(encryptionCfg, baseDir, db)
	if err := secretEncryption.Initialize(); err != nil {
		fmt.Printf("Warning: Failed to initialize encryption: %v\n", err)
	}

	return &SecretService{
		repo:       repo,
		encryption: secretEncryption,
		config:     cfg,
	}
}

// CreateSecret creates a new secret with encryption
func (s *SecretService) CreateSecret(req *SecretCreateRequest) (*SecretResponse, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByName(req.Name, req.NamespaceID, req.ZoneID, req.EnvironmentID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%s: secret with name '%s' already exists", i18n.T("ErrorValidation", nil), req.Name)
	}

	secretNode := &models.SecretNode{
		Name:          req.Name,
		NamespaceID:   req.NamespaceID,
		ZoneID:        req.ZoneID,
		EnvironmentID: req.EnvironmentID,
		IsSecret:      true,
		Type:          req.Type,
		MaxReads:      req.MaxReads,
		Expiration:    req.Expiration,
		Status:        "active",
		CreatedBy:     req.CreatedBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(secretNode); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	if len(req.Value) > 0 {
		if _, err := s.encryption.StoreSecret(secretNode, req.Value); err != nil {
			if delErr := s.repo.Delete(secretNode.ID); delErr != nil {
				fmt.Printf("Warning: failed to rollback secret creation: %v\n", delErr)
			}
			return nil, fmt.Errorf("%s: %w", i18n.T("ErrorEncryptionFailed", nil), err)
		}
	}

	return s.secretToResponse(secretNode), nil
}

// GetSecret retrieves a secret by ID
func (s *SecretService) GetSecret(id uint) (*SecretResponse, error) {
	secret, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorSecretNotFound", nil), err)
	}
	return s.secretToResponse(secret), nil
}

// GetSecretValue retrieves a secret with its decrypted value
func (s *SecretService) GetSecretValue(id uint) (*SecretValueResponse, error) {
	secret, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorSecretNotFound", nil), err)
	}

	if secret.Expiration != nil && time.Now().After(*secret.Expiration) {
		return nil, fmt.Errorf("%s", i18n.T("ErrorSecretExpired", nil))
	}

	version, err := s.repo.GetLatestVersion(secret.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorSecretNotFound", nil), err)
	}

	value, err := s.encryption.RetrieveSecret(version.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorDecryptionFailed", nil), err)
	}

	// Increment read count
	version.ReadCount++
	_ = s.repo.UpdateVersion(version)

	return &SecretValueResponse{
		SecretResponse: *s.secretToResponse(secret),
		Value:          value,
	}, nil
}

// UpdateSecret updates an existing secret
func (s *SecretService) UpdateSecret(req *SecretUpdateRequest) (*SecretResponse, error) {
	secret, err := s.repo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorSecretNotFound", nil), err)
	}

	if req.Type != "" {
		secret.Type = req.Type
	}
	if req.MaxReads != nil {
		secret.MaxReads = req.MaxReads
	}
	if req.Expiration != nil {
		secret.Expiration = req.Expiration
	}
	secret.UpdatedAt = time.Now()

	if err := s.repo.Update(secret); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	if len(req.Value) > 0 {
		if _, err := s.encryption.StoreSecret(secret, req.Value); err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.T("ErrorEncryptionFailed", nil), err)
		}
	}

	return s.secretToResponse(secret), nil
}

// DeleteSecret deletes a secret
func (s *SecretService) DeleteSecret(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorSecretNotFound", nil), err)
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}
	return nil
}

// ListSecrets lists secrets with pagination and filtering
func (s *SecretService) ListSecrets(opts *ListOptions) ([]SecretResponse, int64, error) {
	var (
		secrets []models.SecretNode
		err     error
	)

	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	if opts.Limit > 1000 {
		opts.Limit = 1000
	}

	if opts.Search != "" {
		secrets, err = s.repo.Search(opts.Search, opts.NamespaceID, opts.ZoneID, opts.EnvironmentID)
	} else {
		secrets, err = s.repo.List(opts.NamespaceID, opts.ZoneID, opts.EnvironmentID, opts.Limit, opts.Offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	total, err := s.repo.Count(opts.NamespaceID, opts.ZoneID, opts.EnvironmentID)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}

	responses := make([]SecretResponse, len(secrets))
	for i, secret := range secrets {
		responses[i] = *s.secretToResponse(&secret)
	}

	return responses, total, nil
}

// GetSecretVersions returns all versions of a secret
func (s *SecretService) GetSecretVersions(secretID uint) ([]models.SecretVersion, error) {
	if _, err := s.repo.GetByID(secretID); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorSecretNotFound", nil), err)
	}
	versions, err := s.repo.GetVersions(secretID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("ErrorInternalServer", nil), err)
	}
	return versions, nil
}

// validateCreateRequest validates a create request
func (s *SecretService) validateCreateRequest(req *SecretCreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("%s: %s", i18n.T("ErrorValidation", nil), i18n.T("LabelName", nil))
	}
	if req.NamespaceID == 0 {
		return fmt.Errorf("%s: %s", i18n.T("ErrorValidation", nil), i18n.T("LabelNamespace", nil))
	}
	if req.ZoneID == 0 {
		return fmt.Errorf("%s: %s", i18n.T("ErrorValidation", nil), i18n.T("LabelZone", nil))
	}
	if req.EnvironmentID == 0 {
		return fmt.Errorf("%s: %s", i18n.T("ErrorValidation", nil), i18n.T("LabelEnvironment", nil))
	}
	if req.CreatedBy == "" {
		return fmt.Errorf("%s: %s", i18n.T("ErrorValidation", nil), i18n.T("ErrorRequiredField", nil))
	}
	if len(req.Value) == 0 {
		return fmt.Errorf("%s: %s", i18n.T("ErrorValidation", nil), i18n.T("LabelValue", nil))
	}
	return nil
}

// secretToResponse converts a secret model to response
func (s *SecretService) secretToResponse(secret *models.SecretNode) *SecretResponse {
	versions, _ := s.repo.GetVersions(secret.ID)
	return &SecretResponse{
		ID:            secret.ID,
		Name:          secret.Name,
		Type:          secret.Type,
		NamespaceID:   secret.NamespaceID,
		ZoneID:        secret.ZoneID,
		EnvironmentID: secret.EnvironmentID,
		MaxReads:      secret.MaxReads,
		Expiration:    secret.Expiration,
		Status:        secret.Status,
		CreatedBy:     secret.CreatedBy,
		CreatedAt:     secret.CreatedAt,
		UpdatedAt:     secret.UpdatedAt,
		VersionCount:  len(versions),
	}
}
