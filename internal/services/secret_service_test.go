package services

import (
	"fmt"
	"strings"
	"testing"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/secretlyhq/secretly/internal/storage/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
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
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestService(t *testing.T) *SecretService {
	// Initialize i18n system for tests
	cfg := &config.Config{
		Locale: config.LocaleConfig{
			Language:         "en",
			FallbackLanguage: "en",
		},
		Storage: config.StorageConfig{
			Encryption: config.EncryptionConfig{
				Enabled: false, // Disable encryption for testing
			},
		},
	}
	
	// Reset global state and initialize i18n
	i18n.ResetForTesting()
	err := i18n.Initialize(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize i18n: %v", err)
	}

	db := setupServiceTestDB(t)
	repo := repository.NewSecretRepository(db)

	return NewSecretService(repo, &cfg.Storage.Encryption, ".", db, cfg)
}

func TestSecretService_CreateSecret(t *testing.T) {
	service := createTestService(t)

	req := &SecretCreateRequest{
		Name:          "test-secret",
		Value:         []byte("test-value"),
		Type:          "test",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		CreatedBy:     "test-user",
	}

	response, err := service.CreateSecret(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, response.Name)
	}

	if response.Type != req.Type {
		t.Errorf("Expected type %s, got %s", req.Type, response.Type)
	}

	if response.ID == 0 {
		t.Error("Expected non-zero ID")
	}

	if response.VersionCount != 1 {
		t.Errorf("Expected 1 version, got %d", response.VersionCount)
	}
}

func TestSecretService_CreateSecret_ValidationErrors(t *testing.T) {
	service := createTestService(t)

	tests := []struct {
		name string
		req  *SecretCreateRequest
		want string
	}{
		{
			name: "missing name",
			req: &SecretCreateRequest{
				Value:         []byte("test-value"),
				NamespaceID:   1,
				ZoneID:        1,
				EnvironmentID: 1,
				CreatedBy:     "test-user",
			},
			want: "Validation error: Name",
		},
		{
			name: "missing value",
			req: &SecretCreateRequest{
				Name:          "test-secret",
				NamespaceID:   1,
				ZoneID:        1,
				EnvironmentID: 1,
				CreatedBy:     "test-user",
			},
			want: "Validation error: Value",
		},
		{
			name: "missing namespace",
			req: &SecretCreateRequest{
				Name:          "test-secret",
				Value:         []byte("test-value"),
				ZoneID:        1,
				EnvironmentID: 1,
				CreatedBy:     "test-user",
			},
			want: "Validation error: Namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateSecret(tt.req)
			if err == nil {
				t.Error("Expected validation error")
			}
			if err != nil && err.Error() != tt.want {
				t.Errorf("Expected error '%s', got '%s'", tt.want, err.Error())
			}
		})
	}
}

func TestSecretService_GetSecret(t *testing.T) {
	service := createTestService(t)

	// Create a secret first
	req := &SecretCreateRequest{
		Name:          "test-secret",
		Value:         []byte("test-value"),
		Type:          "test",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		CreatedBy:     "test-user",
	}

	created, err := service.CreateSecret(req)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Get the secret
	response, err := service.GetSecret(created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, response.Name)
	}

	if response.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, response.ID)
	}
}

func TestSecretService_GetSecret_NotFound(t *testing.T) {
	service := createTestService(t)

	_, err := service.GetSecret(999)
	if err == nil {
		t.Error("Expected error for non-existent secret")
	}
}

func TestSecretService_UpdateSecret(t *testing.T) {
	service := createTestService(t)

	// Create a secret first
	req := &SecretCreateRequest{
		Name:          "test-secret",
		Value:         []byte("original-value"),
		Type:          "test",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		CreatedBy:     "test-user",
	}

	created, err := service.CreateSecret(req)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Update the secret
	updateReq := &SecretUpdateRequest{
		ID:        created.ID,
		Value:     []byte("updated-value"),
		Type:      "updated-test",
		UpdatedBy: "test-user",
	}

	updated, err := service.UpdateSecret(updateReq)
	if err != nil {
		t.Fatalf("Failed to update secret: %v", err)
	}

	if updated.Type != "updated-test" {
		t.Errorf("Expected type 'updated-test', got %s", updated.Type)
	}

	if updated.VersionCount != 2 {
		t.Errorf("Expected 2 versions after update, got %d", updated.VersionCount)
	}
}

func TestSecretService_DeleteSecret(t *testing.T) {
	service := createTestService(t)

	// Create a secret first
	req := &SecretCreateRequest{
		Name:          "test-secret",
		Value:         []byte("test-value"),
		Type:          "test",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		CreatedBy:     "test-user",
	}

	created, err := service.CreateSecret(req)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Delete the secret
	err = service.DeleteSecret(created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify deletion
	_, err = service.GetSecret(created.ID)
	if err == nil {
		t.Error("Expected error after deletion")
	}
}

func TestSecretService_ListSecrets(t *testing.T) {
	service := createTestService(t)

	// Create multiple secrets
	for i := 1; i <= 3; i++ {
		req := &SecretCreateRequest{
			Name:          fmt.Sprintf("secret-%d", i),
			Value:         []byte(fmt.Sprintf("value-%d", i)),
			Type:          "test",
			NamespaceID:   1,
			ZoneID:        1,
			EnvironmentID: 1,
			CreatedBy:     "test-user",
		}
		_, err := service.CreateSecret(req)
		if err != nil {
			t.Fatalf("Failed to create secret %d: %v", i, err)
		}
	}

	// List secrets
	opts := &ListOptions{
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		Limit:         10,
		Offset:        0,
	}

	secrets, total, err := service.ListSecrets(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(secrets) != 3 {
		t.Errorf("Expected 3 secrets, got %d", len(secrets))
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
}

func TestSecretService_GetSecretVersions(t *testing.T) {
	service := createTestService(t)

	// Create a secret
	req := &SecretCreateRequest{
		Name:          "test-secret",
		Value:         []byte("original-value"),
		Type:          "test",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		CreatedBy:     "test-user",
	}

	created, err := service.CreateSecret(req)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Update the secret to create more versions
	for i := 2; i <= 3; i++ {
		updateReq := &SecretUpdateRequest{
			ID:        created.ID,
			Value:     []byte(fmt.Sprintf("value-%d", i)),
			UpdatedBy: "test-user",
		}
		_, err := service.UpdateSecret(updateReq)
		if err != nil {
			t.Fatalf("Failed to update secret: %v", err)
		}
	}

	// Get versions
	versions, err := service.GetSecretVersions(created.ID)
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(versions))
	}

	// Verify version numbers are correct
	for i, version := range versions {
		expectedVersion := i + 1
		if version.VersionNumber != expectedVersion {
			t.Errorf("Expected version number %d, got %d", expectedVersion, version.VersionNumber)
		}
	}
}

func TestSecretService_CreateSecret_DuplicateName(t *testing.T) {
	service := createTestService(t)

	req := &SecretCreateRequest{
		Name:          "duplicate-secret",
		Value:         []byte("test-value"),
		Type:          "test",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		CreatedBy:     "test-user",
	}

	// Create first secret
	_, err := service.CreateSecret(req)
	if err != nil {
		t.Fatalf("Failed to create first secret: %v", err)
	}

	// Try to create duplicate
	_, err = service.CreateSecret(req)
	if err == nil {
		t.Error("Expected error for duplicate secret name")
	}

	if err != nil {
		expectedMsg := "secret with name 'duplicate-secret' already exists"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
		}
	}
}
