# Secret CRUD Example

This example demonstrates the complete CRUD (Create, Read, Update, Delete) functionality for secrets in the Secretly system.

## What This Example Shows

1. **Create Secrets** - Store encrypted secrets with metadata
2. **Read Secrets** - Retrieve secret metadata and decrypted values
3. **Update Secrets** - Modify existing secrets and create new versions
4. **Delete Secrets** - Remove secrets permanently
5. **List & Search** - Find secrets with filtering and pagination
6. **Version Management** - Handle multiple versions of secrets
7. **Advanced Features** - Expiration, max reads, and security features
8. **CLI Commands** - Complete command-line interface examples
9. **Best Practices** - Security and management recommendations

## Running the Example

```bash
# From the project root directory
go run examples/secret_crud/main.go
```

## Prerequisites

Ensure your Secretly system is initialized:

```bash
# Initialize the system
secretly system init

# Verify setup
secretly system validate
```

## Features Demonstrated

### 1. Basic CRUD Operations
- **Create**: Store new encrypted secrets
- **Read**: Retrieve secret metadata and values
- **Update**: Modify secrets and create new versions
- **Delete**: Remove secrets permanently

### 2. Advanced Secret Management
- **Expiration**: Set automatic expiration dates
- **Max Reads**: Limit the number of times a secret can be accessed
- **Types**: Categorize secrets (password, api-key, certificate, etc.)
- **Metadata**: Store additional information with secrets

### 3. Organization & Discovery
- **Namespaces**: Logical grouping of secrets
- **Zones**: Geographic or logical zones
- **Environments**: Development, staging, production separation
- **Search**: Find secrets by name or type
- **Pagination**: Handle large numbers of secrets

### 4. Version Control
- **Multiple Versions**: Keep history of secret changes
- **Version Metadata**: Track creation time, size, access count
- **Latest Version**: Always retrieve the most recent value

### 5. Security Features
- **Encryption**: All secret values are encrypted at rest
- **Access Logging**: Track who accessed what and when
- **Audit Trail**: Complete history of secret operations
- **Secure Input**: Hidden password input for CLI

## CLI Commands Available

### Create Secrets
```bash
# Basic creation
secretly secret create --name "db-password" --value "secret123" --type "password"

# From file
secretly secret create --name "ssl-cert" --from-file ./certificate.pem --type "certificate"

# Interactive mode (secure input)
secretly secret create --interactive

# With expiration and limits
secretly secret create --name "temp-token" --value "abc123" --expires "2024-12-31T23:59:59Z" --max-reads 5
```

### Read Secrets
```bash
# Get metadata only
secretly secret get --id 123

# Get with decrypted value
secretly secret get --id 123 --show-value

# Get by name
secretly secret get --name "db-password" --namespace 1 --zone 1 --environment 1
```

### Update Secrets
```bash
# Update value (creates new version)
secretly secret update --id 123 --value "new-secret"

# Update metadata
secretly secret update --id 123 --type "new-type" --max-reads 10

# Interactive update
secretly secret update --id 123 --interactive
```

### List & Search
```bash
# List all secrets
secretly secret list --namespace 1 --zone 1 --environment 1

# Search by name/type
secretly secret list --search "password" --limit 10

# JSON output
secretly secret list --format json
```

### Version Management
```bash
# List all versions
secretly secret versions --id 123

# JSON format
secretly secret versions --id 123 --format json
```

### Delete Secrets
```bash
# Delete with confirmation
secretly secret delete --id 123

# Force delete (skip confirmation)
secretly secret delete --id 123 --force

# Delete by name
secretly secret delete --name "old-secret" --namespace 1 --zone 1 --environment 1
```

## Expected Output

The example will demonstrate:
- ✅ Secret creation with encryption
- 📖 Metadata retrieval
- 🔓 Value decryption
- 🔄 Secret updates and versioning
- 📋 Listing and searching
- 🔍 Search functionality
- 📚 Version history
- ⏰ Expiration and limits
- 💻 CLI command examples
- 🏆 Best practices

## Security Features

### Encryption
- **AES-256-GCM**: Industry-standard authenticated encryption
- **Key Versioning**: Support for key rotation
- **Chunked Storage**: Large secrets split into encrypted chunks

### Access Control
- **Namespace Isolation**: Secrets isolated by namespace/zone/environment
- **Audit Logging**: All operations logged for security
- **Expiration**: Automatic secret expiration
- **Read Limits**: Limit secret access attempts

### Secure Operations
- **Hidden Input**: CLI passwords hidden during input
- **Confirmation**: Destructive operations require confirmation
- **Validation**: Input validation and error handling

## Integration Points

### Database
- **GORM Models**: Full ORM integration
- **Transactions**: Atomic operations
- **Migrations**: Automatic schema management

### Encryption Service
- **Transparent**: Automatic encryption/decryption
- **Metadata**: Encryption details stored with secrets
- **Key Management**: Integrated with key rotation

### CLI Interface
- **Consistent**: Uniform command structure
- **Flexible**: Multiple input methods (flags, files, interactive)
- **User-Friendly**: Clear output and error messages

## Best Practices Demonstrated

1. **Naming Conventions**: Use descriptive, consistent names
2. **Type Classification**: Categorize secrets appropriately
3. **Expiration Management**: Set expiration for temporary secrets
4. **Access Limits**: Use max reads for one-time secrets
5. **Regular Rotation**: Update long-lived secrets regularly
6. **Environment Separation**: Use namespaces/zones/environments
7. **Audit Monitoring**: Track secret access patterns
8. **Secure Input**: Use interactive mode for sensitive data
9. **File Storage**: Store large secrets from files
10. **Validation**: Always validate operations

## What You'll Learn

- Complete secret lifecycle management
- Encryption and security best practices
- CLI usage patterns and workflows
- Database integration and modeling
- Version control for sensitive data
- Search and discovery techniques
- Security monitoring and auditing
- Production deployment considerations

## Next Steps

After running this example:
1. Try creating your own secrets with the CLI
2. Experiment with different secret types and metadata
3. Test the search and filtering capabilities
4. Practice version management workflows
5. Implement your own secret management policies
6. Integrate with your applications and services

This example provides a complete foundation for enterprise-grade secret management with the Secretly system.