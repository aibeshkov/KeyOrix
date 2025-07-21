package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/secretlyhq/secretly/internal/cli/encryption"
	"github.com/secretlyhq/secretly/internal/cli/rbac"
	"github.com/secretlyhq/secretly/internal/cli/secret"
	"github.com/secretlyhq/secretly/internal/cli/system"
	"github.com/spf13/cobra"
)

var version = "dev" // will be overwritten via -ldflags

func main() {
	rootCmd := &cobra.Command{
		Use:     "secretly",
		Short:   "Secretly - Secure secrets management CLI",
		Version: version, // automatically includes --version flag
	}

	rootCmd.AddCommand(system.SystemCmd)
	rootCmd.AddCommand(encryption.EncryptionCmd)
	rootCmd.AddCommand(secret.SecretCmd)
	rootCmd.AddCommand(rbac.RbacCmd)

	// Add short alias -v
	rootCmd.PersistentFlags().BoolP("version-short", "v", false, "Print version")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		v, _ := cmd.Flags().GetBool("version-short")
		if v {
			fmt.Println(version)
			os.Exit(0)
		}
	}

	// Suppress internal help and error output for custom handling
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)

		if strings.Contains(err.Error(), "unknown command") {
			fmt.Fprintf(os.Stderr, "Run '%s --help' for usage.\n", rootCmd.CommandPath())
		}

		os.Exit(1)
	}
}
