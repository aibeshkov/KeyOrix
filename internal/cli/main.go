package cli

import (
	"fmt"
	"os"

	"github.com/secretlyhq/secretly/internal/cli/share"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "secretly",
	Short: "Secretly - A secure secret management tool",
	Long:  `Secretly is a tool for securely storing, managing, and sharing secrets.`,
}

func init() {
	// Add commands
	rootCmd.AddCommand(share.ShareCmd)
	
	// TODO: Add other commands (secret, user, etc.)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}