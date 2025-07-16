package encryption

import (
	"fmt"
	"os"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/encryption"
	"github.com/spf13/cobra"
)

// EncryptionCmd is the root command for encryption operations
var EncryptionCmd = &cobra.Command{
	Use:   "encryption",
	Short: "Manage encryption keys and settings",
	Long:  "Commands for managing encryption keys, rotating keys, and validating encryption setup",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize encryption keys",
	Long:  "Generate new encryption keys (KEK and DEK) if they don't exist",
	RunE:  runInit,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show encryption status",
	Long:  "Display current encryption configuration and key status",
	RunE:  runStatus,
}

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate encryption keys",
	Long:  "Generate new encryption keys and update key version",
	RunE:  runRotate,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate encryption setup",
	Long:  "Check encryption configuration and key file permissions",
	RunE:  runValidate,
}

var fixPermsCmd = &cobra.Command{
	Use:   "fix-perms",
	Short: "Fix key file permissions",
	Long:  "Automatically fix permissions on encryption key files",
	RunE:  runFixPerms,
}

func init() {
	EncryptionCmd.AddCommand(initCmd)
	EncryptionCmd.AddCommand(statusCmd)
	EncryptionCmd.AddCommand(rotateCmd)
	EncryptionCmd.AddCommand(validateCmd)
	EncryptionCmd.AddCommand(fixPermsCmd)
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load("")
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return cfg, nil
}

func runInit(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if !cfg.Storage.Encryption.Enabled {
		fmt.Println("❌ Encryption is disabled in configuration")
		return nil
	}

	baseDir, _ := os.Getwd()
	service := encryption.NewService(&cfg.Storage.Encryption, baseDir)

	fmt.Println("🔐 Initializing encryption...")
	if err := service.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}

	fmt.Println("✅ Encryption initialized successfully")
	fmt.Printf("📋 Key version: %s\n", service.GetKeyVersion())
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("🔐 Encryption Status")
	fmt.Println("==================")
	fmt.Printf("Enabled: %v\n", cfg.Storage.Encryption.Enabled)
	fmt.Printf("Use KEK: %v\n", cfg.Storage.Encryption.UseKEK)
	fmt.Printf("KEK Path: %s\n", cfg.Storage.Encryption.KEKPath)
	fmt.Printf("DEK Path: %s\n", cfg.Storage.Encryption.DEKPath)

	if !cfg.Storage.Encryption.Enabled {
		return nil
	}

	baseDir, _ := os.Getwd()
	service := encryption.NewService(&cfg.Storage.Encryption, baseDir)

	if err := service.Initialize(); err != nil {
		fmt.Printf("❌ Initialization failed: %v\n", err)
		return nil
	}

	fmt.Printf("Initialized: ✅\n")
	fmt.Printf("Key Version: %s\n", service.GetKeyVersion())

	return nil
}

func runRotate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if !cfg.Storage.Encryption.Enabled {
		return fmt.Errorf("encryption is disabled in configuration")
	}

	baseDir, _ := os.Getwd()
	service := encryption.NewService(&cfg.Storage.Encryption, baseDir)

	if err := service.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}

	fmt.Println("🔄 Rotating encryption keys...")
	if err := service.RotateKeys(); err != nil {
		return fmt.Errorf("failed to rotate keys: %w", err)
	}

	fmt.Println("✅ Keys rotated successfully")
	fmt.Printf("📋 New key version: %s\n", service.GetKeyVersion())
	fmt.Println("⚠️  Note: Existing secrets will need to be re-encrypted with the new keys")
	return nil
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if !cfg.Storage.Encryption.Enabled {
		fmt.Println("ℹ️  Encryption is disabled - nothing to validate")
		return nil
	}

	baseDir, _ := os.Getwd()
	service := encryption.NewService(&cfg.Storage.Encryption, baseDir)

	fmt.Println("🔍 Validating encryption setup...")

	if err := service.Initialize(); err != nil {
		fmt.Printf("❌ Initialization failed: %v\n", err)
		return err
	}

	if err := service.ValidateKeyFiles(); err != nil {
		fmt.Printf("❌ Key file validation failed: %v\n", err)
		fmt.Println("💡 Run 'secretly encryption fix-perms' to fix permissions")
		return err
	}

	fmt.Println("✅ Encryption setup is valid")
	return nil
}

func runFixPerms(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if !cfg.Storage.Encryption.Enabled {
		return fmt.Errorf("encryption is disabled in configuration")
	}

	baseDir, _ := os.Getwd()
	service := encryption.NewService(&cfg.Storage.Encryption, baseDir)

	if err := service.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}

	fmt.Println("🔧 Fixing key file permissions...")
	if err := service.FixKeyFilePermissions(); err != nil {
		return fmt.Errorf("failed to fix permissions: %w", err)
	}

	fmt.Println("✅ Key file permissions fixed")
	return nil
}
