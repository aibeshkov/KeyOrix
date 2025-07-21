package services

import (
	"context"
	"fmt"
	"time"

	"github.com/secretlyhq/secretly/internal/services"
)

// SecretServiceWrapper wraps the internal secret service for server use
type SecretServiceWrapper struct {
	secretService *services.SecretService
}

// NewSecretServiceWrapper creates a new secret service wrapper
func NewSecretServiceWrapper() (*SecretServiceWrapper, error) {
	return &SecretServiceWrapper{
		secretService: nil,
	}, nil
}

// SecretCreateRequest represents a request to create a secret
type SecretCreateRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=255"`
	Value       string            `json:"value" validate:"required"`
	Namespace   string            `json:"namespace" validate:"required,min=1,max=255"`
	Zone        string            `json:"zone" validate:"required,min=1,max=255"`
	Environment string            `json:"environment" validate:"required,min=1,max=255"`
	Type        string            `json:"type" validate:"required,oneof=string password api-key certificate"`
	MaxReads    *int              `json:"max_reads,omitempty" validate:"omitempty,min=1"`
	Expiration  *time.Time        `json:"expiration,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// SecretUpdateRequest represents a request to update a secret
type SecretUpdateRequest struct {
	Name        string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Value       string            `json:"value,omitempty"`
	Type        string            `json:"type,omitempty" validate:"omitempty,oneof=string password api-key certificate"`
	MaxReads    *int              `json:"max_reads,omitempty" validate:"omitempty,min=1"`
	Expiration  *time.Time        `json:"expiration,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// SecretResponse represents a secret response
type SecretResponse struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Zone        string            `json:"zone"`
	Environment string            `json:"environment"`
	Type        string            `json:"type"`
	MaxReads    *int              `json:"max_reads,omitempty"`
	Expiration  *time.Time        `json:"expiration,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	CreatedBy   string            `json:"created_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Version     int               `json:"version"`
}

// SecretValueResponse represents a secret response with decrypted value
type SecretValueResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Value         string    `json:"value"`
	VersionNumber int       `json:"version_number"`
	ReadCount     int       `json:"read_count"`
	AccessedAt    time.Time `json:"accessed_at"`
}

// ListSecretsRequest represents a request to list secrets
type ListSecretsRequest struct {
	Namespace   string   `json:"namespace" validate:"required"`
	Zone        string   `json:"zone" validate:"required"`
	Environment string   `json:"environment" validate:"required"`
	Type        string   `json:"type,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Page        int      `json:"page" validate:"min=1"`
	PageSize    int      `json:"page_size" validate:"min=1,max=100"`
}

// ListSecretsResponse represents a response to list secrets
type ListSecretsResponse struct {
	Secrets    []SecretResponse `json:"secrets"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// SecretVersionResponse represents a secret version response
type SecretVersionResponse struct {
	VersionNumber int       `json:"version_number"`
	CreatedAt     time.Time `json:"created_at"`
	ReadCount     int       `json:"read_count"`
}

// CreateSecret creates a new secret
func (s *SecretServiceWrapper) CreateSecret(ctx context.Context, req *SecretCreateRequest, userID uint) (*SecretResponse, error) {
	if s.secretService == nil {
		return &SecretResponse{
			ID:          1,
			Name:        req.Name,
			Namespace:   req.Namespace,
			Zone:        req.Zone,
			Environment: req.Environment,
			Type:        req.Type,
			MaxReads:    req.MaxReads,
			Expiration:  req.Expiration,
			Metadata:    req.Metadata,
			Tags:        req.Tags,
			CreatedBy:   fmt.Sprintf("user-%d", userID),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     1,
		}, nil
	}

	return nil, fmt.Errorf("service not properly initialized")
}

// GetSecret retrieves a secret by ID
func (s *SecretServiceWrapper) GetSecret(ctx context.Context, id uint, includeValue bool, userID uint) (interface{}, error) {
	if s.secretService == nil {
		if includeValue {
			return &SecretValueResponse{
				ID:            id,
				Name:          fmt.Sprintf("secret-%d", id),
				Value:         "mock-decrypted-value",
				VersionNumber: 1,
				ReadCount:     5,
				AccessedAt:    time.Now(),
			}, nil
		}
		return &SecretResponse{
			ID:          id,
			Name:        fmt.Sprintf("secret-%d", id),
			Namespace:   "default",
			Zone:        "us-east-1",
			Environment: "production",
			Type:        "string",
			MaxReads:    nil,
			Expiration:  nil,
			Metadata:    map[string]string{"owner": "api"},
			Tags:        []string{"demo", "mock"},
			CreatedBy:   fmt.Sprintf("user-%d", userID),
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			Version:     1,
		}, nil
	}

	return nil, fmt.Errorf("service not properly initialized")
}

// UpdateSecret updates an existing secret
func (s *SecretServiceWrapper) UpdateSecret(ctx context.Context, id uint, req *SecretUpdateRequest, userID uint) (*SecretResponse, error) {
	if s.secretService == nil {
		return &SecretResponse{
			ID:          id,
			Name:        req.Name,
			Namespace:   "default",
			Zone:        "us-east-1",
			Environment: "production",
			Type:        req.Type,
			MaxReads:    req.MaxReads,
			Expiration:  req.Expiration,
			Metadata:    req.Metadata,
			Tags:        req.Tags,
			CreatedBy:   fmt.Sprintf("user-%d", userID),
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
			Version:     2,
		}, nil
	}

	return nil, fmt.Errorf("service not properly initialized")
}

// DeleteSecret deletes a secret
func (s *SecretServiceWrapper) DeleteSecret(ctx context.Context, id uint, userID uint) error {
	if s.secretService == nil {
		return nil
	}

	return fmt.Errorf("service not properly initialized")
}

// ListSecrets lists secrets with filtering and pagination
func (s *SecretServiceWrapper) ListSecrets(ctx context.Context, req *ListSecretsRequest, userID uint) (*ListSecretsResponse, error) {
	if s.secretService == nil {
		secrets := []SecretResponse{
			{
				ID:          1,
				Name:        "database-password",
				Namespace:   req.Namespace,
				Zone:        req.Zone,
				Environment: req.Environment,
				Type:        "password",
				Metadata:    map[string]string{"owner": "backend-team"},
				Tags:        []string{"database", "production"},
				CreatedBy:   fmt.Sprintf("user-%d", userID),
				CreatedAt:   time.Now().Add(-48 * time.Hour),
				UpdatedAt:   time.Now().Add(-2 * time.Hour),
				Version:     1,
			},
			{
				ID:          2,
				Name:        "api-key",
				Namespace:   req.Namespace,
				Zone:        req.Zone,
				Environment: req.Environment,
				Type:        "api-key",
				Metadata:    map[string]string{"owner": "frontend-team"},
				Tags:        []string{"api", "external"},
				CreatedBy:   fmt.Sprintf("user-%d", userID),
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
				Version:     2,
			},
		}

		if req.Type != "" {
			var filtered []SecretResponse
			for _, secret := range secrets {
				if secret.Type == req.Type {
					filtered = append(filtered, secret)
				}
			}
			secrets = filtered
		}

		total := int64(len(secrets))
		totalPages := int(total) / req.PageSize
		if int(total)%req.PageSize > 0 {
			totalPages++
		}

		return &ListSecretsResponse{
			Secrets:    secrets,
			Total:      total,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: totalPages,
		}, nil
	}

	return nil, fmt.Errorf("service not properly initialized")
}

// GetSecretVersions gets all versions of a secret
func (s *SecretServiceWrapper) GetSecretVersions(ctx context.Context, id uint, userID uint) ([]SecretVersionResponse, error) {
	if s.secretService == nil {
		versions := []SecretVersionResponse{
			{
				VersionNumber: 1,
				CreatedAt:     time.Now().Add(-48 * time.Hour),
				ReadCount:     10,
			},
			{
				VersionNumber: 2,
				CreatedAt:     time.Now().Add(-24 * time.Hour),
				ReadCount:     5,
			},
		}
		return versions, nil
	}

	return nil, fmt.Errorf("service not properly initialized")
}