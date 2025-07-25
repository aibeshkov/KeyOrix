package system

import (
	"fmt"
	"log"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/securefiles"
	"github.com/spf13/cobra"
)

var FixFilePermCmd = &cobra.Command{
	Use:   "fixfileperm",
	Short: "Fix file permissions on critical files (config, KEK, DEK)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("secretly.yaml")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		files := []securefiles.FilePermSpec{
			{Path: cfg.Storage.Encryption.KEKPath, Mode: 0600},
			{Path: cfg.Storage.Encryption.DEKPath, Mode: 0600},
			{Path: "secretly.yaml", Mode: 0600},
		}

		// fix permissions: autofix = true
		err = securefiles.FixFilePerms(files, true)
		if err != nil {
			fmt.Printf("❌ Failed to fix file permissions: %v\n", err)
		} else {
			fmt.Println("✅ All file permissions fixed successfully")
		}
	},
}
