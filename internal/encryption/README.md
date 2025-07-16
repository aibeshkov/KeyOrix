# Encryption Layer

This package provides a comprehensive encryption layer for the Secretly secrets management system. It implements AES-256-GCM encryption with proper key management and secure storage practices.

## Features

- **AES-256-GCM Encryption**: Industry-standard authenticated encryption
- **Key Management**: Separate Key Encryption Key (KEK) and Data Encryption Key (DEK)
- **Chunked Encryption**: Support for large secrets with automatic chunking
- **Key Rotation**: Safe key rotation with version tracking
- **Secure File Operations**: Path validation and permission management
- **Database Integration**: Seamless integration with GORM models

## Architecture

### Components

1. **EncryptionService** (`encryption.go`): Core encryption/decryption operations
2. **KeyManager** (`keymanager.go`): Key lifecycle and storage management
3. **Service** (`service.go`): High-level encryption service wrapper
4. **SecretEncryption** (`integration.go`): Database integration layer
5. **CLI Commands** (`cli/encryption/`): Command-line interface

### Key Management

The system uses a two-tier key architecture:

- **KEK (Key Encryption Key)**: Master key for encrypting other keys
- **DEK (Data Encryption Key)**: Key used for encrypting actual data

Keys are stored in separate files with strict permissions (0600) and are validated on startup.

## Configuration

Add encryption settings to your `config.yaml`:

```yaml
encryption:
  enabled: true
  use_kek: true
  kek_path: "keys/kek.key"
  dek_path: "keys/dek.key"
```

## Usage

### Initialize Encryption

```bash
# Initialize encryption keys
secretly encryption init

# Check encryption status
secretly encryption status

# Validate encryption setup
secretly encryption validate
```

### Programmatic Usage

```go
package main

import (
    "github.com/secretlyhq/secretly/internal/encryption"
    "github.com/secretlyhq/secretly/internal/config"
)

func main() {
    // Load configuration
    cfg, _ := config.Load("config.yaml")
    
    // Create encryption service
    service := encryption.NewService(&cfg.Storage.Encryption, ".")
    
    // Initialize
    service.Initialize()
    
    // Encrypt data
    plaintext := []byte("my secret data")
    encrypted, metadata, _ := service.EncryptSecret(plaintext)
    
    // Decrypt data
    decrypted, _ := service.DecryptSecret(encrypted)
}
```

### Database Integration

```go
// Create secret encryption handler
secretEncryption := encryption.NewSecretEncryption(&cfg.Storage.Encryption, baseDir, db)
secretEncryption.Initialize()

// Store encrypted secret
secretNode := &models.SecretNode{...}
plaintext := []byte("secret value")
version, _ := secretEncryption.StoreSecret(secretNode, plaintext)

// Retrieve and decrypt secret
decrypted, _ := secretEncryption.RetrieveSecret(version.ID)
```

## CLI Commands

### `secretly encryption init`
Initialize encryption keys if they don't exist.

### `secretly encryption status`
Display current encryption configuration and status.

### `secretly encryption rotate`
Rotate encryption keys and update key version.

### `secretly encryption validate`
Validate encryption setup and key file permissions.

### `secretly encryption fix-perms`
Automatically fix key file permissions.

## Security Features

### File Security
- Path validation prevents directory traversal attacks
- Automatic permission setting (0600) for key files
- Secure file operations with base directory validation

### Encryption Security
- AES-256-GCM authenticated encryption
- Cryptographically secure random key generation
- Proper nonce generation for each encryption operation
- Key versioning for rotation support

### Memory Security
- Secure key wiping from memory on shutdown
- Thread-safe operations with proper locking
- Explicit error handling for all cryptographic operations

## Error Handling

The encryption layer provides comprehensive error handling:

- Configuration validation
- Key file validation and creation
- Encryption/decryption error handling
- Database operation error handling
- Permission and security error handling

## Performance Considerations

### Chunking
Large secrets are automatically chunked for better performance:
- Default chunk size: 64KB
- Configurable chunk size
- Parallel chunk processing support
- Automatic chunk reassembly

### Caching
- In-memory key caching for performance
- Thread-safe key access
- Lazy initialization support

## Testing

Run the encryption example:

```bash
cd examples
go run encryption_example.go
```

This will demonstrate:
- Basic secret encryption/decryption
- Large secret chunking
- Status checking
- Validation

## Troubleshooting

### Common Issues

1. **Permission Denied**: Run `secretly encryption fix-perms`
2. **Key Not Found**: Run `secretly encryption init`
3. **Invalid Configuration**: Check `config.yaml` encryption settings
4. **Database Errors**: Ensure proper database migration

### Debug Mode

Enable debug logging by setting environment variable:
```bash
export SECRETLY_DEBUG=true
```

## Migration from Unencrypted

To migrate existing unencrypted secrets:

1. Enable encryption in configuration
2. Run `secretly encryption init`
3. Use the rotation command to re-encrypt existing secrets
4. Validate the setup with `secretly encryption validate`

## Best Practices

1. **Key Storage**: Store keys outside the application directory in production
2. **Backups**: Backup keys securely before rotation
3. **Permissions**: Regularly validate key file permissions
4. **Rotation**: Rotate keys periodically based on your security policy
5. **Monitoring**: Monitor encryption status and key versions