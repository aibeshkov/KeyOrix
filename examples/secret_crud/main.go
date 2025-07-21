package main

import (
	"fmt"
	"log"
	"time"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/services"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/secretlyhq/secretly/internal/storage/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔐 Secretly Secret CRUD Example")
	fmt.Println("===============================")

	// Load configuration
	cfg, err := config.Load("secretly.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := gorm.Open(sqlite.Open(cfg.Storage.Database.Path), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models (ensure tables exist)
	if err := db.AutoMigrate(&models.SecretNode{}, &models.SecretVersion{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize service
	repo := repository.NewSecretRepository(db)
	service := services.NewSecretService(repo, &cfg.Storage.Encryption, ".", db, cfg)

	// Example 1: Create a secret
	fmt.Println("\n📝 Example 1: Create Secret")
	createReq := &services.SecretCreateRequest{
		Name:          "example-api-key",
		Value:         []byte("sk-1234567890abcdef"),
		Type:          "api-key",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		CreatedBy:     "example-user",
	}

	secret, err := service.CreateSecret(createReq)
	if err != nil {
		log.Printf("Failed to create secret (might already exist): %v", err)
		// Try to get existing secret
		existing, err := repo.GetByName("example-api-key", 1, 1, 1)
		if err != nil {
			log.Fatalf("Failed to get existing secret: %v", err)
		}
		secret = &services.SecretResponse{
			ID:            existing.ID,
			Name:          existing.Name,
			Type:          existing.Type,
			NamespaceID:   existing.NamespaceID,
			ZoneID:        existing.ZoneID,
			EnvironmentID: existing.EnvironmentID,
			Status:        existing.Status,
			CreatedBy:     existing.CreatedBy,
			CreatedAt:     existing.CreatedAt,
			UpdatedAt:     existing.UpdatedAt,
		}
	}

	fmt.Printf("✅ Secret created/found: %s (ID: %d)\n", secret.Name, secret.ID)

	// Example 2: Get secret metadata
	fmt.Println("\n📖 Example 2: Get Secret Metadata")
	retrieved, err := service.GetSecret(secret.ID)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	fmt.Printf("Secret: %s\n", retrieved.Name)
	fmt.Printf("Type: %s\n", retrieved.Type)
	fmt.Printf("Status: %s\n", retrieved.Status)
	fmt.Printf("Created: %s\n", retrieved.CreatedAt.Format(time.RFC3339))

	// Example 3: Get secret value (decrypted)
	fmt.Println("\n🔓 Example 3: Get Secret Value")
	secretValue, err := service.GetSecretValue(secret.ID)
	if err != nil {
		log.Fatalf("Failed to get secret value: %v", err)
	}

	fmt.Printf("Decrypted value: %s\n", string(secretValue.Value))
	fmt.Printf("Value size: %d bytes\n", len(secretValue.Value))

	// Example 4: Update secret
	fmt.Println("\n🔄 Example 4: Update Secret")
	updateReq := &services.SecretUpdateRequest{
		ID:        secret.ID,
		Value:     []byte("sk-updated-9876543210fedcba"),
		Type:      "api-key-v2",
		UpdatedBy: "example-user",
	}

	updated, err := service.UpdateSecret(updateReq)
	if err != nil {
		log.Fatalf("Failed to update secret: %v", err)
	}

	fmt.Printf("✅ Secret updated: %s\n", updated.Name)
	fmt.Printf("New type: %s\n", updated.Type)
	fmt.Printf("Version count: %d\n", updated.VersionCount)

	// Example 5: List secrets
	fmt.Println("\n📋 Example 5: List Secrets")
	listOpts := &services.ListOptions{
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		Limit:         10,
		Offset:        0,
	}

	secrets, total, err := service.ListSecrets(listOpts)
	if err != nil {
		log.Fatalf("Failed to list secrets: %v", err)
	}

	fmt.Printf("Found %d secrets (total: %d)\n", len(secrets), total)
	for _, s := range secrets {
		fmt.Printf("- %s (ID: %d, Type: %s, Versions: %d)\n",
			s.Name, s.ID, s.Type, s.VersionCount)
	}

	// Example 6: Search secrets
	fmt.Println("\n🔍 Example 6: Search Secrets")
	searchOpts := &services.ListOptions{
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		Search:        "api",
		Limit:         10,
	}

	searchResults, searchTotal, err := service.ListSecrets(searchOpts)
	if err != nil {
		log.Fatalf("Failed to search secrets: %v", err)
	}

	fmt.Printf("Search results for 'api': %d secrets (total: %d)\n", len(searchResults), searchTotal)
	for _, s := range searchResults {
		fmt.Printf("- %s (Type: %s)\n", s.Name, s.Type)
	}

	// Example 7: Get secret versions
	fmt.Println("\n📚 Example 7: Get Secret Versions")
	versions, err := service.GetSecretVersions(secret.ID)
	if err != nil {
		log.Fatalf("Failed to get versions: %v", err)
	}

	fmt.Printf("Secret has %d versions:\n", len(versions))
	for _, version := range versions {
		fmt.Printf("- Version %d: %d bytes, %d reads, created %s\n",
			version.VersionNumber,
			len(version.EncryptedValue),
			version.ReadCount,
			version.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Example 8: Create secret with expiration
	fmt.Println("\n⏰ Example 8: Create Secret with Expiration")
	expiration := time.Now().Add(24 * time.Hour) // Expires in 24 hours
	maxReads := 5

	tempSecretReq := &services.SecretCreateRequest{
		Name:          "temp-token",
		Value:         []byte("temporary-access-token-12345"),
		Type:          "temp-token",
		NamespaceID:   1,
		ZoneID:        1,
		EnvironmentID: 1,
		MaxReads:      &maxReads,
		Expiration:    &expiration,
		CreatedBy:     "example-user",
	}

	tempSecret, err := service.CreateSecret(tempSecretReq)
	if err != nil {
		log.Printf("Failed to create temp secret (might already exist): %v", err)
	} else {
		fmt.Printf("✅ Temporary secret created: %s\n", tempSecret.Name)
		fmt.Printf("Expires: %s\n", tempSecret.Expiration.Format(time.RFC3339))
		fmt.Printf("Max reads: %d\n", *tempSecret.MaxReads)
	}

	// Example 9: CLI Command Examples
	fmt.Println("\n💻 Example 9: CLI Command Examples")
	fmt.Println("==================================")

	cliExamples := []struct {
		description string
		command     string
	}{
		{"Create a secret", "secretly secret create --name 'db-password' --value 'secret123' --type 'password'"},
		{"Get secret metadata", "secretly secret get --id " + fmt.Sprintf("%d", secret.ID)},
		{"Get secret value", "secretly secret get --id " + fmt.Sprintf("%d", secret.ID) + " --show-value"},
		{"List all secrets", "secretly secret list --namespace 1 --zone 1 --environment 1"},
		{"Search secrets", "secretly secret search --query 'api' --namespace 1"},
		{"Update secret", "secretly secret update --id " + fmt.Sprintf("%d", secret.ID) + " --type 'new-type'"},
		{"Get versions", "secretly secret versions --id " + fmt.Sprintf("%d", secret.ID)},
		{"Interactive create", "secretly secret create --interactive"},
		{"Create from file", "secretly secret create --name 'cert' --from-file ./certificate.pem"},
		{"Delete secret", "secretly secret delete --id " + fmt.Sprintf("%d", secret.ID) + " --force"},
	}

	fmt.Println("Available CLI commands:")
	for _, example := range cliExamples {
		fmt.Printf("%-20s: %s\n", example.description, example.command)
	}

	// Example 10: Best Practices
	fmt.Println("\n🏆 Example 10: Best Practices")
	fmt.Println("=============================")

	bestPractices := []string{
		"Use descriptive secret names with consistent naming conventions",
		"Set appropriate secret types for better organization",
		"Use expiration dates for temporary secrets",
		"Set max reads for one-time use secrets",
		"Regularly rotate long-lived secrets",
		"Use namespaces, zones, and environments for proper isolation",
		"Monitor secret access through audit logs",
		"Use interactive mode for sensitive secret creation",
		"Store large secrets (certificates, keys) from files",
		"Always validate secret retrieval before using values",
	}

	fmt.Println("Secret management best practices:")
	for i, practice := range bestPractices {
		fmt.Printf("%d. %s\n", i+1, practice)
	}

	fmt.Println("\n✅ Secret CRUD example completed successfully!")
	fmt.Println("💡 Try the CLI commands shown above to interact with secrets")
}
