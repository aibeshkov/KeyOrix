# Secretly System Setup Guide

This guide covers the complete setup and configuration of the Secretly secrets management system.

## 🚀 Quick Start

### 1. Initialize the System
```bash
# Initialize with default settings
secretly system init

# Or use interactive setup
secretly system init --interactive
```

### 2. Validate the Setup
```bash
# Check system configuration
secretly system validate

# Audit file permissions
secretly system audit

# Check encryption status
secretly encryption status
```

### 3. Start Using Secretly
Your system is now ready for secure secret management!

## 📋 Complete Setup Process

### Step 1: System Initialization

The `secretly system init` command creates all necessary files and directories:

```
📁 Project Structure After Init:
├── secretly.yaml          # Main configuration (0600)
├── secretly_template.yaml # Template file (0644)
├── keys/
│   ├── kek.key           # Key Encryption Key (0600)
│   └── dek.key           # Data Encryption Key (0600)
├── secretly.db           # SQLite database (0600)
├── secretly.log          # Application logs (0644)
└── certs/                # TLS certificates (if enabled)
    ├── server.crt        # Certificate (0600)
    └── server.key        # Private key (0600)
```

### Step 2: Configuration Overview

The `secretly.yaml` configuration includes:

```yaml
# Server settings
server:
  http:
    enabled: true
    port: "8080"
  grpc:
    enabled: true
    port: "9090"

# Storage and encryption
storage:
  database:
    path: "secretly.db"
  encryption:
    enabled: true
    kek_path: "keys/kek.key"
    dek_path: "keys/dek.key"

# Security policies
security:
  enable_file_permission_check: true
  auto_fix_file_permissions: false
  allow_unsafe_file_permissions: false
```

### Step 3: Security Validation

The system performs comprehensive security checks:

- ✅ **File Permissions**: All critical files have 0600 permissions
- ✅ **File Ownership**: Files are owned by the current user
- ✅ **Encryption Keys**: KEK/DEK files exist and are valid (32 bytes)
- ✅ **Database Access**: Database file is accessible
- ✅ **Configuration**: Config file is valid and complete

## 🔧 Advanced Configuration

### Selective Component Initialization

Initialize only specific components:

```bash
# Encryption only
secretly system init --encryption

# Database only
secretly system init --database

# Multiple components
secretly system init --encryption --database --logging
```

### Custom Configuration Paths

```bash
# Use custom config file location
secretly system init --config /path/to/my-config.yaml

# Validate custom config
secretly system validate --config /path/to/my-config.yaml
```

### Force Overwrite (Dangerous)

```bash
# Overwrite existing files
secretly system init --force

# ⚠️ WARNING: This will overwrite existing configuration and keys!
```

## 🔐 Encryption Management

### Initialize Encryption Separately
```bash
# Initialize encryption keys
secretly encryption init

# Check encryption status
secretly encryption status

# Rotate encryption keys
secretly encryption rotate

# Validate encryption setup
secretly encryption validate

# Fix key file permissions
secretly encryption fix-perms
```

### Encryption Features
- **AES-256-GCM**: Industry-standard authenticated encryption
- **Key Management**: Separate KEK and DEK with rotation support
- **Chunked Encryption**: Support for large secrets
- **Key Versioning**: Track key versions for rotation
- **Secure Storage**: Keys stored with 0600 permissions

## 🛡️ Security Best Practices

### 1. File Permissions
```bash
# Regular permission audits
secretly system audit

# Automatic permission fixing (if needed)
secretly system validate --fix
```

### 2. Key Management
```bash
# Regular key rotation
secretly encryption rotate

# Backup keys before rotation
cp keys/kek.key keys/kek.key.backup.$(date +%s)
cp keys/dek.key keys/dek.key.backup.$(date +%s)
```

### 3. System Validation
```bash
# Always validate before starting
secretly system validate

# Check encryption status
secretly encryption status
```

### 4. Production Deployment
- Enable file permission checks
- Use TLS for all network communications
- Store keys in secure, backed-up locations
- Monitor file permissions regularly
- Use strong authentication mechanisms

## 🔍 Troubleshooting

### Common Issues and Solutions

#### 1. Permission Denied Errors
```bash
# Check current permissions
secretly system audit

# Fix permissions automatically
secretly system validate --fix

# Manual permission fix
chmod 0600 secretly.yaml keys/*.key secretly.db
```

#### 2. Missing Configuration
```bash
# Recreate configuration
secretly system init

# Force overwrite corrupted config
secretly system init --force
```

#### 3. Encryption Key Issues
```bash
# Regenerate encryption keys
secretly encryption init

# Check key status
secretly encryption status

# Validate key files
secretly encryption validate
```

#### 4. Database Issues
```bash
# Reinitialize database
secretly system init --database

# Check database permissions
ls -la secretly.db
```

### Debug Mode
```bash
# Enable debug logging
export SECRETLY_DEBUG=true
secretly system init
```

## 📊 System Status Commands

### Comprehensive Status Check
```bash
# System validation
secretly system validate

# Encryption status
secretly encryption status

# File permission audit
secretly system audit
```

### Expected Output (Healthy System)
```
🔍 Validating Secretly System
============================
🔍 Startup Validation Results
============================
Configuration: ✅
Permissions:   ✅
Encryption:    ✅
Database:      ✅

🎉 All validations passed!
```

## 🚀 Next Steps

After successful system initialization:

1. **Start the Server**: Configure and start HTTP/gRPC servers
2. **Create Secrets**: Begin storing and managing secrets
3. **Set Up Users**: Configure authentication and authorization
4. **Monitor System**: Regular validation and auditing
5. **Backup Strategy**: Implement key and database backup procedures

## 📚 Additional Resources

- **Encryption Guide**: `internal/encryption/README.md`
- **System Commands**: `internal/cli/system/README.md`
- **Configuration Reference**: `secretly_template.yaml`
- **Examples**: `examples/system_init_example.go`

## 🆘 Support

If you encounter issues:

1. Run `secretly system validate` for detailed diagnostics
2. Check file permissions with `secretly system audit`
3. Verify encryption setup with `secretly encryption status`
4. Review configuration in `secretly.yaml`
5. Check logs in `secretly.log`

The system provides comprehensive error messages and recovery suggestions for most common issues.