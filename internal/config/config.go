package config

import (
	"fmt"
	"path/filepath"

	"github.com/secretlyhq/secretly/internal/securefiles"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Locale     LocaleConfig     `yaml:"locale"`
	Server     ServerConfig     `yaml:"server"`
	Storage    StorageConfig    `yaml:"storage"`
	Secrets    SecretsConfig    `yaml:"secrets"`
	Telemetry  TelemetryConfig  `yaml:"telemetry"`
	Security   SecurityConfig   `yaml:"security"`
	SoftDelete SoftDeleteConfig `yaml:"soft_delete"`
	Purge      PurgeConfig      `yaml:"purge"`
}

type LocaleConfig struct {
	Language         string `yaml:"language"`
	FallbackLanguage string `yaml:"fallback_language"`
}

type ServerConfig struct {
	HTTP ServerInstanceConfig `yaml:"http"`
	GRPC ServerInstanceConfig `yaml:"grpc"`
}

type ServerInstanceConfig struct {
	Enabled           bool            `yaml:"enabled"`
	Port              string          `yaml:"port"`
	ProtocolVersions  []string        `yaml:"protocol_versions"`
	TLS               TLSConfig       `yaml:"tls"`
	RateLimit         RateLimitConfig `yaml:"ratelimit"`
	SwaggerEnabled    bool            `yaml:"swagger_enabled,omitempty"`
	ReflectionEnabled bool            `yaml:"reflection_enabled,omitempty"`
}

type TLSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AutoCert       bool     `yaml:"auto_cert,omitempty"`
	Domains        []string `yaml:"domains,omitempty"`
	CertFile       string   `yaml:"cert_file"`
	KeyFile        string   `yaml:"key_file"`
	AllowedCiphers []string `yaml:"allowed_ciphers"`
}

type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

type StorageConfig struct {
	Database   DatabaseConfig   `yaml:"database"`
	Encryption EncryptionConfig `yaml:"encryption"`
}

type DatabaseConfig struct {
	Path         string `yaml:"path"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type EncryptionConfig struct {
	Enabled bool   `yaml:"enabled"`
	UseKEK  bool   `yaml:"use_kek"`
	KEKPath string `yaml:"kek_path"`
	DEKPath string `yaml:"dek_path"`
}

type SecretsConfig struct {
	Chunking ChunkingConfig `yaml:"chunking"`
	Limits   LimitsConfig   `yaml:"limits"`
}

type ChunkingConfig struct {
	Enabled            bool `yaml:"enabled"`
	MaxChunkSizeKB     int  `yaml:"max_chunk_size_kb"`
	MaxChunksPerSecret int  `yaml:"max_chunks_per_secret"`
}

type LimitsConfig struct {
	MaxSecretsPerUser int `yaml:"max_secrets_per_user"`
}

type TelemetryConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	LogFile  string `yaml:"log_file"`
	APIKey   string `yaml:"api_key"`
}

type SecurityConfig struct {
	EnableFilePermissionCheck  bool `yaml:"enable_file_permission_check"`
	AutoFixFilePermissions     bool `yaml:"auto_fix_file_permissions"`
	AllowUnsafeFilePermissions bool `yaml:"allow_unsafe_file_permissions"`
}

type SoftDeleteConfig struct {
	Enabled       bool `yaml:"enabled"`
	RetentionDays int  `yaml:"retention_days"`
}

type PurgeConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Schedule string `yaml:"schedule"`
}

const appRootDir = "."

// Load loads the YAML configuration file.
// If path is empty, it will load "secretly.yaml" from the application root.
func Load(path string) (*Config, error) {
	if path == "" {
		path = filepath.Join(appRootDir, "secretly.yaml")
	}

	data, err := securefiles.SafeReadFile(appRootDir, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// LoadConfig loads configuration using the default path.
// Used for server module compatibility.
func LoadConfig() (*Config, error) {
	return Load("")
}

// Validate checks the configuration for required fields and correctness.
func (c *Config) Validate() error {
	// Validate server settings
	if c.Server.HTTP.Enabled && c.Server.HTTP.Port == "" {
		return fmt.Errorf("HTTP server is enabled but no port is specified")
	}

	if c.Server.GRPC.Enabled && c.Server.GRPC.Port == "" {
		return fmt.Errorf("gRPC server is enabled but no port is specified")
	}

	// Validate TLS settings
	if c.Server.HTTP.TLS.Enabled {
		if c.Server.HTTP.TLS.AutoCert {
			if len(c.Server.HTTP.TLS.Domains) == 0 {
				return fmt.Errorf("autocert is enabled but no domains are specified")
			}
		} else {
			if c.Server.HTTP.TLS.CertFile == "" || c.Server.HTTP.TLS.KeyFile == "" {
				return fmt.Errorf("TLS is enabled but cert_file or key_file is missing")
			}
		}
	}

	if c.Server.GRPC.TLS.Enabled {
		if !c.Server.GRPC.TLS.AutoCert {
			if c.Server.GRPC.TLS.CertFile == "" || c.Server.GRPC.TLS.KeyFile == "" {
				return fmt.Errorf("gRPC TLS is enabled but cert_file or key_file is missing")
			}
		}
	}

	// Validate database configuration
	if c.Storage.Database.Path == "" {
		return fmt.Errorf("database path is not specified")
	}

	// Validate locale configuration
	if c.Locale.Language == "" {
		c.Locale.Language = "en" // Default to English
	}
	if c.Locale.FallbackLanguage == "" {
		c.Locale.FallbackLanguage = "en" // Default fallback to English
	}

	// Validate supported language codes
	supportedLanguages := map[string]bool{
		"en": true, "ru": true, "es": true, "fr": true, "de": true,
	}
	if !supportedLanguages[c.Locale.Language] {
		return fmt.Errorf("unsupported language: %s. Supported languages: en, ru, es, fr, de", c.Locale.Language)
	}
	if !supportedLanguages[c.Locale.FallbackLanguage] {
		return fmt.Errorf("unsupported fallback language: %s. Supported languages: en, ru, es, fr, de", c.Locale.FallbackLanguage)
	}

	return nil
}