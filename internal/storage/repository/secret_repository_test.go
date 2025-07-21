package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/secretlyhq/secretly/internal/storage/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

func TestSecretRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSecretRepository(db)

	secret := &models.SecretNode{
		Name:          "test-secret",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		IsSecret:      true,
		Type:          "test",
		Status:        "active",
		CreatedBy:     "test-user",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	if secret.ID == 0 {
		t.Error("Expected non-zero ID after creation")
	}
}

func TestSecretRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSecretRepository(db)

	// Create a secret first
	secret := &models.SecretNode{
		Name:          "test-secret",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		IsSecret:      true,
		Type:          "test",
		Status:        "active",
		CreatedBy:     "test-user",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Retrieve the secret
	retrieved, err := repo.GetByID(secret.ID)
	if err != nil {
		t.Fatalf("Failed to get secret by ID: %v", err)
	}

	if retrieved.Name != secret.Name {
		t.Errorf("Expected name %s, got %s", secret.Name, retrieved.Name)
	}

	if retrieved.Type != secret.Type {
		t.Errorf("Expected type %s, got %s", secret.Type, retrieved.Type)
	}
}

func TestSecretRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSecretRepository(db)

	_, err := repo.GetByID(999)
	if err == nil {
		t.Error("Expected error for non-existent secret")
	}
}

func TestSecretRepository_CreateVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSecretRepository(db)

	// Create a secret first
	secret := &models.SecretNode{
		Name:          "test-secret",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		IsSecret:      true,
		Type:          "test",
		Status:        "active",
		CreatedBy:     "test-user",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Create a version
	version := &models.SecretVersion{
		SecretNodeID:   secret.ID,
		VersionNumber:  1,
		EncryptedValue: []byte("encrypted-test-value"),
		ReadCount:      0,
		CreatedAt:      time.Now(),
	}

	err = repo.CreateVersion(version)
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}

	if version.ID == 0 {
		t.Error("Expected non-zero ID after version creation")
	}
}

func TestSecretRepository_GetVersions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSecretRepository(db)

	// Create a secret first
	secret := &models.SecretNode{
		Name:          "test-secret",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		IsSecret:      true,
		Type:          "test",
		Status:        "active",
		CreatedBy:     "test-user",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Create multiple versions
	for i := 1; i <= 3; i++ {
		version := &models.SecretVersion{
			SecretNodeID:   secret.ID,
			VersionNumber:  i,
			EncryptedValue: []byte(fmt.Sprintf("encrypted-value-%d", i)),
			ReadCount:      0,
			CreatedAt:      time.Now(),
		}
		err = repo.CreateVersion(version)
		if err != nil {
			t.Fatalf("Failed to create version %d: %v", i, err)
		}
	}

	// Get versions
	versions, err := repo.GetVersions(secret.ID)
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(versions))
	}

	// Verify version numbers
	for i, version := range versions {
		expectedVersion := i + 1
		if version.VersionNumber != expectedVersion {
			t.Errorf("Expected version number %d, got %d", expectedVersion, version.VersionNumber)
		}
	}
}

func TestSecretRepository_GetLatestVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSecretRepository(db)

	// Create a secret first
	secret := &models.SecretNode{
		Name:          "test-secret",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		IsSecret:      true,
		Type:          "test",
		Status:        "active",
		CreatedBy:     "test-user",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Create multiple versions
	for i := 1; i <= 3; i++ {
		version := &models.SecretVersion{
			SecretNodeID:   secret.ID,
			VersionNumber:  i,
			EncryptedValue: []byte(fmt.Sprintf("encrypted-value-%d", i)),
			ReadCount:      0,
			CreatedAt:      time.Now(),
		}
		err = repo.CreateVersion(version)
		if err != nil {
			t.Fatalf("Failed to create version %d: %v", i, err)
		}
	}

	// Get latest version
	latest, err := repo.GetLatestVersion(secret.ID)
	if err != nil {
		t.Fatalf("Failed to get latest version: %v", err)
	}

	if latest.VersionNumber != 3 {
		t.Errorf("Expected latest version number 3, got %d", latest.VersionNumber)
	}
}
