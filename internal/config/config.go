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
	Enabled          bool            `yaml:"enabled"`
	Port             string          `yaml:"port"`
	ProtocolVersions []string        `yaml:"protocol_versions"`
	TLS              TLSConfig       `yaml:"tls"`
	RateLimit        RateLimitConfig `yaml:"ratelimit"`
}

type TLSConfig struct {
	Enabled        bool     `yaml:"enabled"`
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

// Load загружает YAML-конфигурацию из файла.
// Если path пустой, загружает из "secretly.yaml" в корне приложения.
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
