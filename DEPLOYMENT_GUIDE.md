# Secretly Full-Stack Deployment Guide

This guide covers deploying the complete Secretly system with both the Go backend and React web dashboard.

## 🏗️ Architecture Overview

The full-stack deployment includes:
- **Go Backend**: API server with gRPC and HTTP endpoints
- **React Web Dashboard**: Modern web interface for secret management
- **Database**: SQLite (default) or PostgreSQL for data storage
- **Static Assets**: Web dashboard served by the Go server
- **Optional**: Redis for caching, Nginx for reverse proxy, monitoring stack

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Node.js 18+ (for building web assets)
- Go 1.21+ (for building server)
- Make (for build automation)

### 1. Clone and Build

```bash
# Clone the repository
git clone <repository-url>
cd secretly

# Build the web dashboard
cd web
npm install
npm run build
cd ..

# Build the Go server
cd server
go build -o secretly-server ./
cd ..
```

### 2. Quick Test

```bash
# Run integration test
./scripts/test-web-integration.sh

# Keep server running for manual testing
./scripts/test-web-integration.sh --keep-running
```

### 3. Production Deployment

```bash
# Using Docker Compose
docker-compose -f docker-compose.full-stack.yml up -d

# Or build and run manually
cd server
SECRETLY_CONFIG_PATH="./config/production.yaml" ./secretly-server
```

## 📋 Deployment Options

### Option 1: Docker Compose (Recommended)

**Advantages:**
- Easy to deploy and manage
- Includes monitoring and database
- Production-ready configuration
- Automatic health checks and restarts

```bash
# Basic deployment
docker-compose -f docker-compose.full-stack.yml up -d

# With monitoring stack
docker-compose -f docker-compose.full-stack.yml --profile monitoring up -d

# View logs
docker-compose -f docker-compose.full-stack.yml logs -f secretly
```

### Option 2: Manual Deployment

**Advantages:**
- Full control over configuration
- Can integrate with existing infrastructure
- Easier debugging and development

```bash
# 1. Build web assets
cd web && npm run build && cd ..

# 2. Build server
cd server && go build -o secretly-server ./ && cd ..

# 3. Create configuration
cp server/config/production.yaml ./secretly.yaml
# Edit configuration as needed

# 4. Run server
cd server
SECRETLY_CONFIG_PATH="../secretly.yaml" ./secretly-server
```

### Option 3: Kubernetes Deployment

See `k8s/` directory for Kubernetes manifests (to be created).

## ⚙️ Configuration

### Environment Variables

```bash
# Core settings
SECRETLY_CONFIG_PATH=/path/to/config.yaml
ENVIRONMENT=production
GIN_MODE=release

# Domain and CORS
SECRETLY_DOMAIN=your-domain.com

# Database (if using PostgreSQL)
DB_HOST=postgres
DB_PORT=5432
DB_NAME=secretly
DB_USER=secretly
DB_PASSWORD=your-password

# Optional services
REDIS_PASSWORD=your-redis-password
TELEMETRY_ENDPOINT=https://your-telemetry-endpoint
TELEMETRY_API_KEY=your-api-key
```

### Configuration File

The main configuration is in YAML format. Key sections:

```yaml
environment: "production"

server:
  http:
    enabled: true
    port: "8080"
    web_assets_path: "/app/web/dist"  # Path to built web assets
    domain: "your-domain.com"
    allowed_origins:
      - "https://your-domain.com"

storage:
  type: "local"  # or "remote"
  database:
    path: "/app/data/secretly.db"
  encryption:
    enabled: true
    use_kek: true
```

## 🔒 Security Configuration

### TLS/SSL Setup

#### Option 1: Reverse Proxy (Recommended)

Use Nginx or similar to handle SSL termination:

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://secretly:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### Option 2: Direct TLS

Configure TLS directly in the server:

```yaml
server:
  http:
    tls:
      enabled: true
      cert_file: "/app/certs/cert.pem"
      key_file: "/app/certs/key.pem"
```

### CORS Configuration

Configure allowed origins for web dashboard:

```yaml
server:
  http:
    allowed_origins:
      - "https://your-domain.com"
      - "https://www.your-domain.com"
```

### Rate Limiting

Enable rate limiting to prevent abuse:

```yaml
server:
  http:
    ratelimit:
      enabled: true
      requests_per_second: 50
      burst: 100
```

## 📊 Monitoring and Logging

### Health Checks

The server provides health check endpoints:

- `GET /health` - Basic health check
- `GET /api/v1/system/info` - System information (requires auth)
- `GET /api/v1/system/metrics` - System metrics (requires auth)

### Logging

Configure structured logging:

```yaml
telemetry:
  enabled: true
  log_file: "/app/logs/secretly.log"
  endpoint: "https://your-logging-service"
```

### Monitoring Stack

The Docker Compose setup includes optional monitoring:

```bash
# Deploy with monitoring
docker-compose -f docker-compose.full-stack.yml --profile monitoring up -d

# Access Grafana
open http://localhost:3001
```

## 🗄️ Database Management

### SQLite (Default)

- **Pros**: Simple, no additional setup required
- **Cons**: Not suitable for high-concurrency scenarios
- **Best for**: Small to medium deployments

```yaml
storage:
  database:
    path: "/app/data/secretly.db"
```

### PostgreSQL (Recommended for Production)

- **Pros**: Better performance, concurrent access, backup tools
- **Cons**: Additional complexity
- **Best for**: Production deployments

```yaml
storage:
  database:
    type: "postgres"
    host: "postgres"
    port: 5432
    name: "secretly"
    user: "secretly"
    password: "your-password"
```

### Backup Strategy

#### SQLite Backup

```bash
# Create backup
sqlite3 /app/data/secretly.db ".backup /backup/secretly-$(date +%Y%m%d).db"

# Restore backup
cp /backup/secretly-20240115.db /app/data/secretly.db
```

#### PostgreSQL Backup

```bash
# Create backup
pg_dump -h postgres -U secretly secretly > /backup/secretly-$(date +%Y%m%d).sql

# Restore backup
psql -h postgres -U secretly secretly < /backup/secretly-20240115.sql
```

## 🔧 Troubleshooting

### Common Issues

#### Web Assets Not Loading

**Problem**: Web dashboard shows blank page or 404 errors

**Solutions:**
1. Check web assets path in configuration
2. Ensure web assets are built: `cd web && npm run build`
3. Verify file permissions on web assets directory
4. Check server logs for asset serving errors

```bash
# Debug asset serving
curl -v http://localhost:8080/
curl -v http://localhost:8080/assets/index.js
```

#### CORS Errors

**Problem**: Browser shows CORS errors when accessing API

**Solutions:**
1. Check `allowed_origins` in configuration
2. Ensure web dashboard domain is included
3. Verify CORS headers in browser developer tools

```yaml
server:
  http:
    allowed_origins:
      - "https://your-actual-domain.com"  # Must match exactly
```

#### Database Connection Issues

**Problem**: Server fails to start with database errors

**Solutions:**
1. Check database file permissions (SQLite)
2. Verify database connection settings (PostgreSQL)
3. Ensure database directory exists and is writable

```bash
# Check SQLite database
ls -la /app/data/
sqlite3 /app/data/secretly.db ".tables"

# Check PostgreSQL connection
psql -h postgres -U secretly -d secretly -c "SELECT 1;"
```

#### Performance Issues

**Problem**: Slow response times or high resource usage

**Solutions:**
1. Enable rate limiting
2. Optimize database queries
3. Add caching layer (Redis)
4. Scale horizontally with load balancer

### Debug Mode

Enable debug logging for troubleshooting:

```yaml
environment: "development"  # Enables debug logging
```

Or set environment variable:

```bash
SECRETLY_DEBUG=true ./secretly-server
```

### Log Analysis

Check logs for common issues:

```bash
# Docker logs
docker-compose logs -f secretly

# File logs
tail -f /app/logs/secretly.log

# System logs
journalctl -u secretly -f
```

## 📈 Performance Optimization

### Server Optimization

```yaml
server:
  http:
    # Optimize connection handling
    read_timeout: "15s"
    write_timeout: "15s"
    idle_timeout: "60s"
    
    # Enable compression
    compression: true
    
    # Optimize rate limiting
    ratelimit:
      enabled: true
      requests_per_second: 100
      burst: 200
```

### Database Optimization

```yaml
storage:
  database:
    # Optimize connection pool
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_lifetime: "1h"
```

### Caching

Add Redis for caching:

```yaml
cache:
  enabled: true
  type: "redis"
  redis:
    host: "redis"
    port: 6379
    password: "your-password"
    db: 0
```

## 🚀 Scaling

### Horizontal Scaling

Deploy multiple instances behind a load balancer:

```yaml
# docker-compose.scale.yml
services:
  secretly:
    # ... configuration
    deploy:
      replicas: 3
  
  nginx:
    # Load balancer configuration
    upstream:
      - secretly:8080
```

### Vertical Scaling

Increase resource limits:

```yaml
services:
  secretly:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G
```

## 🔄 Updates and Maintenance

### Rolling Updates

```bash
# Build new version
docker build -t secretly:new .

# Update with zero downtime
docker-compose -f docker-compose.full-stack.yml up -d --no-deps secretly
```

### Backup Before Updates

```bash
# Backup data
docker-compose exec secretly sqlite3 /app/data/secretly.db ".backup /backup/pre-update.db"

# Backup configuration
cp server/config/production.yaml backup/
```

### Health Monitoring

Set up monitoring to detect issues:

```bash
# Simple health check script
#!/bin/bash
if ! curl -f http://localhost:8080/health; then
    echo "Health check failed"
    # Send alert
fi
```

## 📚 Additional Resources

- [API Documentation](server/openapi.yaml)
- [Web Dashboard User Guide](web/docs/USER_GUIDE.md)
- [Security Best Practices](docs/SECURITY.md)
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md)

## 🆘 Support

For issues and questions:

1. Check the troubleshooting section above
2. Review server logs for error messages
3. Test with the integration script: `./scripts/test-web-integration.sh`
4. Create an issue with detailed logs and configuration

---

**Deployment Status**: ✅ Production Ready  
**Last Updated**: January 2025  
**Version**: 1.0.0