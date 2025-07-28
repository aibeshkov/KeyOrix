# Remote CLI Setup Guide

This guide explains how to configure the Secretly CLI to work with remote servers, enabling team collaboration and enterprise deployment.

## Overview

The Secretly CLI supports two storage modes:
- **Local Mode**: Stores secrets in a local SQLite database (default)
- **Remote Mode**: Connects to a remote Secretly server via HTTP API

## Quick Start

### 1. Check Current Status

```bash
secretly config status
```

This shows your current configuration and storage type.

### 2. Configure Remote Server

```bash
secretly config set-remote --url https://api.secretly.company.com --api-key your-api-key
```

Or configure interactively:

```bash
secretly config set-remote --url https://api.secretly.company.com
# You'll be prompted for the API key
```

### 3. Authenticate

```bash
secretly auth login
```

This will prompt for your API key and store it securely.

### 4. Test Connection

```bash
secretly status
```

Or test connectivity:

```bash
secretly ping
```

## Configuration Options

### Environment Variables

You can use environment variables in your configuration:

```yaml
# secretly.yaml
storage:
  type: "remote"
  remote:
    base_url: "https://api.secretly.company.com"
    api_key: "${SECRETLY_API_KEY}"
    timeout_seconds: 30
    retry_attempts: 3
    tls_verify: true
```

Supported environment variables:
- `SECRETLY_API_KEY`
- `SECRETLY_TOKEN`
- `API_KEY`

### Configuration File

The CLI uses `secretly.yaml` for configuration:

```yaml
storage:
  type: "remote"  # "local" or "remote"
  
  # Local storage configuration
  database:
    path: "./secrets.db"
  
  # Remote storage configuration
  remote:
    base_url: "https://api.secretly.company.com"
    api_key: "${SECRETLY_API_KEY}"
    timeout_seconds: 30
    retry_attempts: 3
    tls_verify: true
```

## Commands Reference

### Configuration Commands

- `secretly config status` - Show current configuration
- `secretly config set-remote` - Configure remote server
- `secretly config use-local` - Switch to local storage
- `secretly config test-connection` - Test storage connection

### Authentication Commands

- `secretly auth login` - Set up API key authentication
- `secretly auth logout` - Clear authentication credentials
- `secretly auth status` - Check authentication status

### Status Commands

- `secretly status` - Check system health and connection
- `secretly ping` - Test remote server connectivity

## Deployment Scenarios

### Development Environment

```bash
# Use local storage for development
secretly config use-local
```

### Staging Environment

```bash
# Configure for staging server
secretly config set-remote --url https://staging-api.secretly.company.com
secretly auth login
```

### Production Environment

```bash
# Configure for production server
secretly config set-remote --url https://api.secretly.company.com
secretly auth login
```

## Troubleshooting

### Connection Issues

1. **Check network connectivity:**
   ```bash
   secretly ping
   ```

2. **Verify server URL:**
   ```bash
   secretly config status
   ```

3. **Test authentication:**
   ```bash
   secretly auth status
   ```

### Common Error Messages

#### "circuit breaker is open"
The CLI has detected multiple connection failures and temporarily stopped trying to connect. Wait 30 seconds and try again.

#### "failed to create storage"
Check your configuration file and ensure all required fields are present.

#### "health check failed"
The remote server is not responding. Check if the server is running and accessible.

### Offline Mode

If the remote server is unavailable, the CLI can automatically switch to local mode:

```bash
# This will temporarily switch to local storage
secretly config use-local
```

To switch back when connectivity is restored:

```bash
secretly config set-remote --url your-server-url
```

## Security Considerations

### API Key Storage

API keys are stored in the configuration file. Ensure proper file permissions:

```bash
chmod 600 secretly.yaml
```

### TLS/HTTPS

Always use HTTPS in production:

```yaml
storage:
  remote:
    base_url: "https://api.secretly.company.com"  # Use HTTPS
    tls_verify: true  # Verify certificates
```

### Network Security

- Use VPN or private networks when possible
- Configure firewall rules to restrict access
- Monitor API key usage and rotate regularly

## Performance Optimization

### Caching

The CLI automatically caches GET requests for 5 minutes to improve performance.

### Connection Pooling

HTTP connections are reused when possible to reduce latency.

### Retry Logic

Failed requests are automatically retried with exponential backoff:
- Initial retry after 1 second
- Second retry after 4 seconds  
- Third retry after 9 seconds

## Examples

### Basic Remote Setup

```bash
# Configure remote server
secretly config set-remote --url https://api.example.com --api-key abc123

# Verify configuration
secretly config status

# Test connection
secretly status

# Use normally
secretly secret create --name "api-key" --type "api_key"
secretly secret list
```

### Environment-Based Configuration

```bash
# Set environment variable
export SECRETLY_API_KEY="your-api-key-here"

# Configure with environment variable
secretly config set-remote --url https://api.example.com --api-key '${SECRETLY_API_KEY}'

# The API key will be read from the environment variable
secretly status
```

### Switching Between Environments

```bash
# Development (local)
secretly config use-local
secretly secret list

# Staging (remote)
secretly config set-remote --url https://staging-api.example.com
secretly auth login
secretly secret list

# Production (remote)
secretly config set-remote --url https://api.example.com  
secretly auth login
secretly secret list
```

## Migration Guide

### From Local to Remote

1. **Backup your local data:**
   ```bash
   cp secrets.db secrets.db.backup
   ```

2. **Configure remote server:**
   ```bash
   secretly config set-remote --url https://your-server.com
   secretly auth login
   ```

3. **Verify connection:**
   ```bash
   secretly status
   ```

4. **Migrate secrets manually or use export/import tools**

### From Remote to Local

1. **Switch to local mode:**
   ```bash
   secretly config use-local
   ```

2. **Verify local operation:**
   ```bash
   secretly status
   ```

The CLI will automatically create a local database and you can start using it immediately.