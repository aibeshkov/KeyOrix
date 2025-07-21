package secret

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/services"
	"github.com/secretlyhq/secretly/internal/storage/repository"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	deleteID    uint
	deleteName  string
	deleteNS    uint
	deleteZone  uint
	deleteEnv   uint
	deleteForce bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret",
	Long: `Delete a secret permanently.

Examples:
  secretly secret delete --id 123
  secretly secret delete --name "db-password" --namespace 1 --zone 1 --environment 1
  secretly secret delete --id 123 --force  # Skip confirmation`,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().UintVar(&deleteID, "id", 0, "Secret ID")
	deleteCmd.Flags().StringVar(&deleteName, "name", "", "Secret name")
	deleteCmd.Flags().UintVar(&deleteNS, "namespace", 1, "Namespace ID (required with --name)")
	deleteCmd.Flags().UintVar(&deleteZone, "zone", 1, "Zone ID (required with --name)")
	deleteCmd.Flags().UintVar(&deleteEnv, "environment", 1, "Environment ID (required with --name)")
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteID == 0 && deleteName == "" {
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
	var secretName string

	// Get secret information
	if deleteID != 0 {
		secret, err := service.GetSecret(deleteID)
		if err != nil {
			return fmt.Errorf("secret not found: %w", err)
		}
		secretID = secret.ID
		secretName = secret.Name
	} else {
		// Find by name
		secret, err := repo.GetByName(deleteName, deleteNS, deleteZone, deleteEnv)
		if err != nil {
			return fmt.Errorf("secret not found: %w", err)
		}
		secretID = secret.ID
		secretName = secret.Name
	}

	// Show what we're about to delete
	fmt.Printf("🗑️  About to delete secret:\n")
	fmt.Printf("ID: %d\n", secretID)
	fmt.Printf("Name: %s\n", secretName)

	// Get versions count
	versions, err := service.GetSecretVersions(secretID)
	if err != nil {
		return fmt.Errorf("failed to get versions: %w", err)
	}
	fmt.Printf("Versions: %d\n", len(versions))

	// Confirmation
	if !deleteForce {
		fmt.Printf("\n⚠️  This action cannot be undone!\n")
		fmt.Printf("All versions and metadata will be permanently deleted.\n\n")

		if !confirmDeletion(secretName) {
			fmt.Printf("❌ Deletion cancelled\n")
			return nil
		}
	}

	// Delete the secret
	if err := service.DeleteSecret(secretID); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	fmt.Printf("✅ Secret '%s' (ID: %d) deleted successfully\n", secretName, secretID)
	fmt.Printf("🗑️  %d versions were also deleted\n", len(versions))

	return nil
}

func confirmDeletion(secretName string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Type the secret name '%s' to confirm deletion: ", secretName)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input != secretName {
		fmt.Printf("❌ Name mismatch. Expected '%s', got '%s'\n", secretName, input)
		return false
	}

	fmt.Printf("Are you absolutely sure? (yes/no): ")
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))

	return confirmation == "yes"
}
