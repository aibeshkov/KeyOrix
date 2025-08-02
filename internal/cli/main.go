package cli

import (
	"fmt"
	"os"

	"github.com/secretlyhq/secretly/internal/cli/auth"
	"github.com/secretlyhq/secretly/internal/cli/config"
	"github.com/secretlyhq/secretly/internal/cli/connect"
	"github.com/secretlyhq/secretly/internal/cli/encryption"
	"github.com/secretlyhq/secretly/internal/cli/rbac"
	"github.com/secretlyhq/secretly/internal/cli/secret"
	"github.com/secretlyhq/secretly/internal/cli/share"
	"github.com/secretlyhq/secretly/internal/cli/status"
	"github.com/secretlyhq/secretly/internal/cli/system"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "secretly",
	Short: "Secretly - A secure secret management tool",
	Long:  `Secretly is a tool for securely storing, managing, and sharing secrets.`,
}

func init() {
	// Add all available commands
	rootCmd.AddCommand(secret.SecretCmd)
	rootCmd.AddCommand(share.ShareCmd)
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(connect.ConnectCmd)
	rootCmd.AddCommand(encryption.EncryptionCmd)
	rootCmd.AddCommand(rbac.RbacCmd)
	rootCmd.AddCommand(status.StatusCmd)
	rootCmd.AddCommand(system.SystemCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}