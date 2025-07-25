package share

import (
	"context"
	"fmt"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/core"
	"github.com/secretlyhq/secretly/internal/storage/local"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	revokeShareID uint
)

var revokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a share",
	RunE:  runRevoke,
}

func init() {
	revokeCmd.Flags().UintVar(&revokeShareID, "share-id", 0, "Share ID (required)")
	revokeCmd.MarkFlagRequired("share-id")
}

func runRevoke(cmd *cobra.Command, args []string) error {
	// Load config and connect to database
	cfg, err := config.Load("secretly.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.Storage.Database.Path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models (ensure tables exist)
	if err := db.AutoMigrate(&models.SecretNode{}, &models.SecretVersion{}, &models.ShareRecord{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize storage and service
	storage := local.NewLocalStorage(db)
	service := core.NewSecretlyCore(storage)

	// Call service
	ctx := context.Background()
	err = service.RevokeShare(ctx, revokeShareID, 1) // CLI user ID
	if err != nil {
		return fmt.Errorf("failed to revoke share: %w", err)
	}

	// Print result
	fmt.Printf("✅ Share revoked successfully!\n")
	fmt.Printf("Share ID: %d\n", revokeShareID)

	return nil
}