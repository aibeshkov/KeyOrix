package secret

import (
	"fmt"
	"time"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/services"
	"github.com/secretlyhq/secretly/internal/storage/repository"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	getID        uint
	getName      string
	getShowValue bool
	getNamespace uint
	getZone      uint
	getEnv       uint
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a secret",
	Long: `Retrieve a secret by ID or name.

Examples:
  secretly secret get --id 123
  secretly secret get --name "db-password" --namespace 1 --zone 1 --environment 1
  secretly secret get --id 123 --show-value  # Show decrypted value`,
	RunE: runGet,
}

func init() {
	getCmd.Flags().UintVar(&getID, "id", 0, "Secret ID")
	getCmd.Flags().StringVar(&getName, "name", "", "Secret name")
	getCmd.Flags().UintVar(&getNamespace, "namespace", 1, "Namespace ID (required with --name)")
	getCmd.Flags().UintVar(&getZone, "zone", 1, "Zone ID (required with --name)")
	getCmd.Flags().UintVar(&getEnv, "environment", 1, "Environment ID (required with --name)")
	getCmd.Flags().BoolVar(&getShowValue, "show-value", false, "Show decrypted secret value")
}

func runGet(cmd *cobra.Command, args []string) error {
	if getID == 0 && getName == "" {
		return fmt.Errorf("either --id or --name is required")
	}

	// Load configuration
	cfg, err := config.Load("secretly.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to database
	db, err := gorm.Open(sqlite.Open(cfg.Storage.Database.Path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize service
	repo := repository.NewSecretRepository(db)
	service := services.NewSecretService(repo, &cfg.Storage.Encryption, ".", db, cfg)

	var secretID uint

	// Get secret ID
	if getID != 0 {
		secretID = getID
	} else {
		// Find by name
		secret, err := repo.GetByName(getName, getNamespace, getZone, getEnv)
		if err != nil {
			return fmt.Errorf("secret not found: %w", err)
		}
		secretID = secret.ID
	}

	if getShowValue {
		// Get secret with value
		response, err := service.GetSecretValue(secretID)
		if err != nil {
			return fmt.Errorf("failed to get secret value: %w", err)
		}

		displaySecretWithValue(response)
	} else {
		// Get secret metadata only
		response, err := service.GetSecret(secretID)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}

		displaySecret(response)
	}

	return nil
}

func displaySecret(secret *services.SecretResponse) {
	fmt.Printf("🔐 Secret Information\n")
	fmt.Printf("====================\n")
	fmt.Printf("ID: %d\n", secret.ID)
	fmt.Printf("Name: %s\n", secret.Name)
	fmt.Printf("Type: %s\n", secret.Type)
	fmt.Printf("Status: %s\n", secret.Status)
	fmt.Printf("Namespace: %d\n", secret.NamespaceID)
	fmt.Printf("Zone: %d\n", secret.ZoneID)
	fmt.Printf("Environment: %d\n", secret.EnvironmentID)
	fmt.Printf("Created By: %s\n", secret.CreatedBy)
	fmt.Printf("Created: %s\n", secret.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", secret.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("Versions: %d\n", secret.VersionCount)

	if secret.MaxReads != nil {
		fmt.Printf("Max Reads: %d\n", *secret.MaxReads)
	}

	if secret.Expiration != nil {
		fmt.Printf("Expires: %s\n", secret.Expiration.Format(time.RFC3339))
		if time.Now().After(*secret.Expiration) {
			fmt.Printf("⚠️  Status: EXPIRED\n")
		}
	}

	fmt.Printf("\n💡 Use --show-value to display the decrypted value\n")
}

func displaySecretWithValue(secret *services.SecretValueResponse) {
	displaySecret(&secret.SecretResponse)

	fmt.Printf("\n🔓 Decrypted Value\n")
	fmt.Printf("==================\n")

	// Check if value looks like binary data
	if isBinaryData(secret.Value) {
		fmt.Printf("Value: <binary data, %d bytes>\n", len(secret.Value))
		fmt.Printf("Hex: %x\n", secret.Value)
	} else {
		fmt.Printf("Value: %s\n", string(secret.Value))
	}

	fmt.Printf("Size: %d bytes\n", len(secret.Value))
}

func isBinaryData(data []byte) bool {
	// Simple heuristic: if more than 10% of bytes are non-printable, consider it binary
	nonPrintable := 0
	for _, b := range data {
		if b < 32 && b != 9 && b != 10 && b != 13 { // Not tab, newline, or carriage return
			nonPrintable++
		}
	}

	if len(data) == 0 {
		return false
	}

	return float64(nonPrintable)/float64(len(data)) > 0.1
}
