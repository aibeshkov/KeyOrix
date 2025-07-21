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
	listNamespace uint
	listZone      uint
	listEnv       uint
	listLimit     int
	listOffset    int
	listSearch    string
	listFormat    string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
	Long: `List secrets with filtering and pagination.

Examples:
  secretly secret list
  secretly secret list --namespace 1 --zone 1 --environment 1
  secretly secret list --search "password" --limit 10
  secretly secret list --format table  # table or json`,
	RunE: runList,
}

func init() {
	listCmd.Flags().UintVar(&listNamespace, "namespace", 1, "Namespace ID")
	listCmd.Flags().UintVar(&listZone, "zone", 1, "Zone ID")
	listCmd.Flags().UintVar(&listEnv, "environment", 1, "Environment ID")
	listCmd.Flags().IntVar(&listLimit, "limit", 50, "Maximum number of results")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "Number of results to skip")
	listCmd.Flags().StringVar(&listSearch, "search", "", "Search query")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format (table, json)")
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Build list options
	opts := &services.ListOptions{
		NamespaceID:   listNamespace,
		ZoneID:        listZone,
		EnvironmentID: listEnv,
		Limit:         listLimit,
		Offset:        listOffset,
		Search:        listSearch,
	}

	// Get secrets
	secrets, total, err := service.ListSecrets(opts)
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	// Display results
	switch listFormat {
	case "json":
		displaySecretsJSON(secrets, total, opts)
	case "table":
		displaySecretsTable(secrets, total, opts)
	default:
		return fmt.Errorf("unsupported format: %s (use 'table' or 'json')", listFormat)
	}

	return nil
}

func displaySecretsTable(secrets []services.SecretResponse, total int64, opts *services.ListOptions) {
	fmt.Printf("🔐 Secrets List\n")
	fmt.Printf("===============\n")

	if opts.Search != "" {
		fmt.Printf("Search: %s\n", opts.Search)
	}
	fmt.Printf("Namespace: %d, Zone: %d, Environment: %d\n", opts.NamespaceID, opts.ZoneID, opts.EnvironmentID)
	fmt.Printf("Total: %d, Showing: %d (offset: %d, limit: %d)\n\n", total, len(secrets), opts.Offset, opts.Limit)

	if len(secrets) == 0 {
		fmt.Printf("No secrets found.\n")
		return
	}

	// Table header
	fmt.Printf("%-5s %-20s %-12s %-8s %-8s %-20s %-20s\n",
		"ID", "NAME", "TYPE", "VERSIONS", "STATUS", "CREATED", "EXPIRES")
	fmt.Printf("%-5s %-20s %-12s %-8s %-8s %-20s %-20s\n",
		"-----", "--------------------", "------------", "--------", "--------", "--------------------", "--------------------")

	// Table rows
	for _, secret := range secrets {
		expires := "Never"
		if secret.Expiration != nil {
			expires = secret.Expiration.Format("2006-01-02 15:04")
			if time.Now().After(*secret.Expiration) {
				expires += " (EXPIRED)"
			}
		}

		fmt.Printf("%-5d %-20s %-12s %-8d %-8s %-20s %-20s\n",
			secret.ID,
			truncateString(secret.Name, 20),
			truncateString(secret.Type, 12),
			secret.VersionCount,
			secret.Status,
			secret.CreatedAt.Format("2006-01-02 15:04"),
			truncateString(expires, 20))
	}

	// Pagination info
	if total > int64(opts.Limit) {
		fmt.Printf("\n📄 Pagination: Showing %d-%d of %d total\n",
			opts.Offset+1,
			min(opts.Offset+len(secrets), int(total)),
			total)

		if opts.Offset+opts.Limit < int(total) {
			fmt.Printf("💡 Use --offset %d to see more results\n", opts.Offset+opts.Limit)
		}
	}
}

func displaySecretsJSON(secrets []services.SecretResponse, total int64, opts *services.ListOptions) {
	fmt.Printf("{\n")
	fmt.Printf("  \"total\": %d,\n", total)
	fmt.Printf("  \"offset\": %d,\n", opts.Offset)
	fmt.Printf("  \"limit\": %d,\n", opts.Limit)
	fmt.Printf("  \"count\": %d,\n", len(secrets))
	fmt.Printf("  \"secrets\": [\n")

	for i, secret := range secrets {
		fmt.Printf("    {\n")
		fmt.Printf("      \"id\": %d,\n", secret.ID)
		fmt.Printf("      \"name\": \"%s\",\n", secret.Name)
		fmt.Printf("      \"type\": \"%s\",\n", secret.Type)
		fmt.Printf("      \"status\": \"%s\",\n", secret.Status)
		fmt.Printf("      \"namespace_id\": %d,\n", secret.NamespaceID)
		fmt.Printf("      \"zone_id\": %d,\n", secret.ZoneID)
		fmt.Printf("      \"environment_id\": %d,\n", secret.EnvironmentID)
		fmt.Printf("      \"created_by\": \"%s\",\n", secret.CreatedBy)
		fmt.Printf("      \"created_at\": \"%s\",\n", secret.CreatedAt.Format(time.RFC3339))
		fmt.Printf("      \"updated_at\": \"%s\",\n", secret.UpdatedAt.Format(time.RFC3339))
		fmt.Printf("      \"version_count\": %d", secret.VersionCount)

		if secret.MaxReads != nil {
			fmt.Printf(",\n      \"max_reads\": %d", *secret.MaxReads)
		}

		if secret.Expiration != nil {
			fmt.Printf(",\n      \"expiration\": \"%s\"", secret.Expiration.Format(time.RFC3339))
		}

		fmt.Printf("\n    }")
		if i < len(secrets)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("  ]\n")
	fmt.Printf("}\n")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
