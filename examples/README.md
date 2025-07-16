# Secretly Examples

This directory contains comprehensive examples demonstrating the key features of the Secretly secrets management system.

## Available Examples

### 1. 🚀 [System Initialization](system_init/)
**File**: `system_init/main.go`

Demonstrates the complete system setup and validation process:
- System initialization and validation
- File structure verification
- Available commands overview
- Configuration structure
- Security best practices

**Run**: `go run examples/system_init/main.go`

### 2. 🔐 [Encryption](encryption/)
**File**: `encryption/main.go`

Demonstrates the complete encryption functionality:
- Basic secret encryption/decryption
- Large secret chunking
- Key management
- Database integration
- Encryption validation

**Run**: `go run examples/encryption/main.go`

## Getting Started

### Prerequisites

1. **Initialize Secretly** (for system_init example):
   ```bash
   secretly system init
   ```

2. **Go Environment**: Ensure you have Go installed and the project dependencies:
   ```bash
   go mod tidy
   ```

### Running Examples

Each example is self-contained and can be run independently:

```bash
# System initialization example
go run examples/system_init/main.go

# Encryption example (self-contained)
go run examples/encryption/main.go
```

## Example Structure

```
examples/
├── README.md                    # This file
├── system_init/
│   ├── main.go                 # System initialization demo
│   └── README.md               # Detailed documentation
└── encryption/
    ├── main.go                 # Encryption functionality demo
    └── README.md               # Detailed documentation
```

## What You'll Learn

### System Management
- How to initialize a Secretly system
- Configuration file structure and options
- File permission management
- System validation and auditing
- Security best practices

### Encryption
- AES-256-GCM encryption implementation
- Key management (KEK/DEK)
- Chunked encryption for large secrets
- Database integration with GORM
- Encryption status and validation

### CLI Usage
- System management commands
- Encryption management commands
- Validation and auditing tools
- Interactive setup options

## Real-World Usage Patterns

### Development Workflow
1. **Initialize**: `secretly system init --interactive`
2. **Validate**: `secretly system validate`
3. **Use**: Start storing and retrieving secrets
4. **Monitor**: Regular `secretly system audit`

### Production Deployment
1. **Secure Setup**: Enable all security features
2. **TLS Configuration**: Set up certificates
3. **Key Management**: Secure key storage and rotation
4. **Monitoring**: Regular validation and auditing

### Maintenance Tasks
1. **Key Rotation**: `secretly encryption rotate`
2. **Permission Audits**: `secretly system audit`
3. **System Validation**: `secretly system validate`
4. **Backup Procedures**: Secure key and database backups

## Troubleshooting

### Common Issues

1. **Permission Errors**:
   ```bash
   secretly system audit
   secretly system validate --fix
   ```

2. **Missing Files**:
   ```bash
   secretly system init --force
   ```

3. **Encryption Issues**:
   ```bash
   secretly encryption validate
   secretly encryption init
   ```

### Debug Mode

Enable detailed logging:
```bash
export SECRETLY_DEBUG=true
go run examples/system_init/main.go
```

## Additional Resources

- **System Setup Guide**: `../SYSTEM_SETUP.md`
- **Encryption Documentation**: `../internal/encryption/README.md`
- **System Commands**: `../internal/cli/system/README.md`
- **Configuration Reference**: `../secretly_template.yaml`

## Contributing

When adding new examples:
1. Create a new directory under `examples/`
2. Include a `main.go` file with the example code
3. Add a `README.md` with detailed documentation
4. Update this main README with the new example
5. Ensure examples are self-contained and well-documented

## Support

If you encounter issues with the examples:
1. Check the individual example README files
2. Run `secretly system validate` for system issues
3. Run `secretly encryption validate` for encryption issues
4. Review the main documentation files
5. Check file permissions with `secretly system audit`