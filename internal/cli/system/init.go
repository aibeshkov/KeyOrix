package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/securefiles"
	"github.com/spf13/cobra"
)

var (
	configPath     string
	interactive    bool
	initAll        bool
	initEncryption bool
	initDatabase   bool
	initLogging    bool
	initTelemetry  bool
	force          bool
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Secretly system with config and keys",
	Long: `Initialize the Secretly system by creating configuration files and setting up required components.

Supports selective initialization of different components:
- Configuration file (secretly.yaml)
- Encryption keys (KEK/DEK)
- Database setup
- Logging setup
- Telemetry setup

Examples:
  secretly system init                    # Initialize all components
  secretly system init --interactive     # Interactive setup wizard
  secretly system init --encryption      # Initialize encryption only
  secretly system init --force           # Overwrite existing files
  secretly system init --config ./my.yaml # Custom config path`,
	RunE: runInit,
}

func init() {
	InitCmd.Flags().StringVar(&configPath, "config", "./secretly.yaml", "Path to output config file")
	InitCmd.Flags().BoolVar(&interactive, "interactive", false, "Launch interactive setup wizard")
	InitCmd.Flags().BoolVar(&initAll, "all", true, "Initialize all components")
	InitCmd.Flags().BoolVar(&initEncryption, "encryption", false, "Initialize encryption keys")
	InitCmd.Flags().BoolVar(&initDatabase, "database", false, "Initialize database")
	InitCmd.Flags().BoolVar(&initLogging, "logging", false, "Initialize logging")
	InitCmd.Flags().BoolVar(&initTelemetry, "telemetry", false, "Initialize telemetry")
	InitCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files (dangerous)")
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Secretly System Initialization")
	fmt.Println("=================================")

	if initEncryption || initDatabase || initLogging || initTelemetry {
		initAll = false
	}

	if err := generateConfigFile(); err != nil {
		return fmt.Errorf("failed to generate config file: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if initAll || initEncryption {
		if err := initializeEncryption(cfg); err != nil {
			return fmt.Errorf("failed to initialize encryption: %w", err)
		}
	}

	if initAll || initDatabase {
		if err := initializeDatabase(cfg); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	if initAll || initLogging {
		if err := initializeLogging(cfg); err != nil {
			return fmt.Errorf("failed to initialize logging: %w", err)
		}
	}

	if initAll || initTelemetry {
		if err := initializeTelemetry(cfg); err != nil {
			return fmt.Errorf("failed to initialize telemetry: %w", err)
		}
	}

	fmt.Println("\n✅ Secretly system initialization completed successfully!")
	fmt.Printf("📋 Config file: %s\n", configPath)
	fmt.Println("🔐 Run 'secretly encryption status' to check encryption setup")
	fmt.Println("🛡️  Run 'secretly system audit' to validate file permissions")

	return nil
}

func generateConfigFile() error {
	fmt.Printf("📄 Generating config file: %s\n", configPath)

	if _, err := os.Stat(configPath); err == nil && !force {
		fmt.Printf("⚠️  Config file already exists: %s\n", configPath)
		fmt.Println("   Use --force to overwrite")
		return nil
	}

	templateData, err := securefiles.SafeReadFile(".", "secretly_template.yaml")
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := securefiles.SecureWriteFile(".", configPath, templateData, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Config file created: %s\n", configPath)
	return nil
}

func initializeDatabase(cfg *config.Config) error {
	dbPath := filepath.Clean(cfg.Storage.Database.Path)
	if strings.Contains(dbPath, "..") {
		return fmt.Errorf("invalid path for database: %s", dbPath)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.OpenFile(dbPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
		if cerr := file.Close(); cerr != nil {
			return fmt.Errorf("failed to close database file: %w", cerr)
		}
	}
	return nil
}

func initializeTelemetry(cfg *config.Config) error {
	telemetryFile := filepath.Clean(cfg.Telemetry.LogFile)
	if strings.Contains(telemetryFile, "..") {
		return fmt.Errorf("invalid path for telemetry log: %s", telemetryFile)
	}
	if err := os.MkdirAll(filepath.Dir(telemetryFile), 0750); err != nil {
		return fmt.Errorf("failed to create telemetry directory: %w", err)
	}
	if _, err := os.Stat(telemetryFile); os.IsNotExist(err) {
		file, err := os.OpenFile(telemetryFile, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to create telemetry log file: %w", err)
		}
		if cerr := file.Close(); cerr != nil {
			return fmt.Errorf("failed to close telemetry file: %w", cerr)
		}
	}
	return nil
}

func initializeEncryption(cfg *config.Config) error {
	kekDir := filepath.Dir(cfg.Storage.Encryption.KEKPath)
	dekDir := filepath.Dir(cfg.Storage.Encryption.DEKPath)
	if err := os.MkdirAll(kekDir, 0750); err != nil {
		return fmt.Errorf("failed to create KEK directory: %w", err)
	}
	if err := os.MkdirAll(dekDir, 0750); err != nil {
		return fmt.Errorf("failed to create DEK directory: %w", err)
	}
	return nil
}

func initializeLogging(cfg *config.Config) error {
	logPath := filepath.Clean("secretly.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0750); err != nil {
		return fmt.Errorf("failed to create logging directory: %w", err)
	}
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
		if cerr := file.Close(); cerr != nil {
			return fmt.Errorf("failed to close log file: %w", cerr)
		}
	}
	return nil
}
