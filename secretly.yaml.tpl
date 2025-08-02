# secretly_template.yaml
# Default configuration template for Secretly configuration file secretly.yaml.
# Can be used as is or generated via `secretly system init`.

locale:
  # Primary language for the application interface
  # Supported languages: en (English), ru (Russian), es (Spanish), fr (French), de (German)
  language: "en"
  
  # Fallback language when translations are missing in the primary language
  # Should be one of the supported languages, typically "en" for maximum compatibility
  fallback_language: "en"

server:
  http:
    # Enable HTTP server
    enabled: true
    port: "8080"
    protocol_versions: ["1.1"]
    tls:
      # Enable TLS on HTTP
      enabled: false
      cert_file: "certs/server.crt"     # Path to TLS certificate
      key_file: "certs/server.key"      # Path to TLS key
      allowed_ciphers: []               # Optional cipher list
    ratelimit:
      # Enable rate limiting
      enabled: false
      requests_per_second: 10
      burst: 20

  grpc:
    # Enable gRPC server
    enabled: false
    port: "9090"
    protocol_versions: ["1.0"]
    tls:
      enabled: false
      cert_file: "certs/server.crt"
      key_file: "certs/server.key"
      allowed_ciphers: []
    ratelimit:
      enabled: false
      requests_per_second: 10
      burst: 20

storage:
  database:
    # Path to SQLite or PostgreSQL connection string
    path: "secretly.db"
    max_open_conns: 10
    max_idle_conns: 5

  encryption:
    # Enable envelope encryption
    enabled: true
    # Use Key Encryption Key (KEK) and Data Encryption Key (DEK)
    use_kek: true
    kek_path: "keys/kek.key"
    dek_path: "keys/dek.key"

secrets:
  chunking:
    # Enable chunking large secrets into smaller parts
    enabled: true
    max_chunk_size_kb: 64
    max_chunks_per_secret: 10

  limits:
    # Maximum number of secrets per user
    max_secrets_per_user: 1000

telemetry:
  # Enable anonymous usage telemetry
  enabled: false
  # Endpoint for sending telemetry
  endpoint: "https://telemetry.secretlyhq.com/v1/collect"
  # Local log file for telemetry events (optional)
  log_file: "telemetry.log"
  # API key or token if needed
  api_key: ""

security:
  # Check file permission safety on startup
  enable_file_permission_check: true
  auto_fix_file_permissions: true
  allow_unsafe_file_permissions: false

soft_delete:
  # Enable soft-deletion for all database entities
  enabled: true
  retention_days: 30

purge:
  # Enable periodic purge of expired/deleted database entities 
  enabled: true
  schedule: "0 0 * * *"

logging:
  # Enable logging
  enabled: true
  # Log level: debug, info, warn, error
  level: "info"
  # Path to log file
  file: "secretly.log"
  # Log format
  log_format: "text"  