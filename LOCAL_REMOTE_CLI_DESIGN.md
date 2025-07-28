# 🔄 Local + Remote CLI Architecture Design

## 🎯 **Vision: Universal CLI**

The CLI should work seamlessly in **all deployment scenarios**:

### **Scenario 1: Local Development** 
```bash
# Developer working alone on their machine
secretly server start --local
secretly secret create --name "dev-api-key" --value "abc123"
```

### **Scenario 2: Team Development**
```bash
# Multiple developers sharing a development server
secretly config set-remote --url http://dev-server:8080
secretly secret create --name "shared-db-password" --value "secret123"
# ↑ All team members can access this secret
```

### **Scenario 3: Production**
```bash
# Enterprise deployment with authentication
secretly config set-remote --url https://secrets.company.com
secretly auth login --api-key sk_prod_...
secretly secret create --name "prod-api-key" --value "xyz789"
```

## 🏗️ **Architecture Overview**

### **Unified Storage Interface**
```go
// Same interface for all storage types
type Storage interface {
    CreateSecret(ctx context.Context, secret *models.SecretNode) (*models.SecretNode, error)
    GetSecret(ctx context.Context, id uint) (*models.SecretNode, error)
    UpdateSecret(ctx context.Context, secret *models.SecretNode) error
    DeleteSecret(ctx context.Context, id uint) error
    // ... all other methods
}

// Implementations:
// 1. LocalStorage    - Direct database access
// 2. RemoteStorage   - HTTP API client
// 3. HybridStorage   - Local cache + remote sync
```

### **Configuration-Driven Selection**
```yaml
# secretly.yaml
storage:
  type: "auto"  # auto-detect based on context
  
  # Local configuration (default)
  local:
    database:
      path: "./secrets.db"
    
  # Remote configuration (when needed)
  remote:
    base_url: "http://localhost:8080"  # or remote server
    api_key: "${SECRETLY_API_KEY}"
    timeout: 30s
    
  # Hybrid configuration (advanced)
  hybrid:
    local_cache: true
    sync_interval: "5m"
    offline_mode: true
```

## 🔧 **Implementation Plan**

### **Phase 1: Storage Factory Pattern**

```go
// internal/storage/factory.go
type StorageFactory struct {
    config *config.Config
}

func (f *StorageFactory) CreateStorage() (Storage, error) {
    switch f.config.Storage.Type {
    case "local":
        return f.createLocalStorage()
    case "remote":
        return f.createRemoteStorage()
    case "hybrid":
        return f.createHybridStorage()
    case "auto":
        return f.autoDetectStorage()
    default:
        return f.createLocalStorage() // safe default
    }
}

func (f *StorageFactory) autoDetectStorage() (Storage, error) {
    // 1. Check if local database exists
    if f.hasLocalDatabase() {
        return f.createLocalStorage()
    }
    
    // 2. Check if remote server is configured and reachable
    if f.hasRemoteConfig() && f.isRemoteReachable() {
        return f.createRemoteStorage()
    }
    
    // 3. Default to local storage
    return f.createLocalStorage()
}
```

### **Phase 2: Remote Storage Implementation**

```go
// internal/storage/remote/remote.go
type RemoteStorage struct {
    client   *http.Client
    baseURL  string
    apiKey   string
    timeout  time.Duration
}

func (rs *RemoteStorage) CreateSecret(ctx context.Context, secret *models.SecretNode) (*models.SecretNode, error) {
    url := fmt.Sprintf("%s/api/v1/secrets", rs.baseURL)
    
    body, err := json.Marshal(secret)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+rs.apiKey)
    
    resp, err := rs.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to create secret: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
    }
    
    var result models.SecretNode
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &result, nil
}

// Implement all other Storage interface methods...
```

### **Phase 3: CLI Configuration Commands**

```go
// internal/cli/config/config.go
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Manage CLI configuration",
}

var setRemoteCmd = &cobra.Command{
    Use:   "set-remote",
    Short: "Configure remote server connection",
    RunE:  runSetRemote,
}

func runSetRemote(cmd *cobra.Command, args []string) error {
    url, _ := cmd.Flags().GetString("url")
    apiKey, _ := cmd.Flags().GetString("api-key")
    
    // Update configuration
    cfg, err := config.Load("secretly.yaml")
    if err != nil {
        return err
    }
    
    cfg.Storage.Type = "remote"
    cfg.Storage.Remote.BaseURL = url
    cfg.Storage.Remote.APIKey = apiKey
    
    return config.Save("secretly.yaml", cfg)
}
```

## 🚀 **Developer Experience**

### **Scenario 1: Local Development**
```bash
# Start local server (optional - CLI can work without it)
secretly server start --port 8080 --local

# CLI automatically uses local database
secretly secret create --name "dev-secret" --value "local123"
secretly secret list
```

### **Scenario 2: Connect to Development Server**
```bash
# One-time setup to connect to team server
secretly config set-remote --url http://dev-server.local:8080

# Now CLI uses remote server
secretly secret create --name "team-secret" --value "shared123"
secretly secret list  # Shows secrets from remote server
```

### **Scenario 3: Switch Between Local and Remote**
```bash
# Work locally
secretly config use-local
secretly secret list  # Shows local secrets

# Switch to remote
secretly config use-remote
secretly secret list  # Shows remote secrets

# Auto-detect mode (default)
secretly config use-auto
secretly secret list  # Uses best available option
```

## 🔧 **Server Deployment Options**

### **Option 1: Local Development Server**
```bash
# Simple local server for development
secretly server start --local --port 8080
# - Uses SQLite database
# - No authentication required
# - Perfect for team development
```

### **Option 2: Docker Compose for Teams**
```yaml
# docker-compose.yml
version: '3.8'
services:
  secretly-server:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/secretly
    volumes:
      - ./config:/app/config
      
  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=secretly
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### **Option 3: Production Deployment**
```bash
# Production server with full authentication
secretly server start \
  --port 8080 \
  --database-url postgres://... \
  --auth-required \
  --tls-cert /path/to/cert.pem \
  --tls-key /path/to/key.pem
```

## 📋 **Configuration Examples**

### **Local Development (Default)**
```yaml
# secretly.yaml
storage:
  type: "local"
  local:
    database:
      path: "./secrets.db"
      
server:
  port: 8080
  host: "localhost"
```

### **Team Development**
```yaml
# secretly.yaml
storage:
  type: "remote"
  remote:
    base_url: "http://dev-server.local:8080"
    timeout: "30s"
    
# No API key needed for development server
```

### **Production**
```yaml
# secretly.yaml
storage:
  type: "remote"
  remote:
    base_url: "https://secrets.company.com"
    api_key: "${SECRETLY_API_KEY}"
    timeout: "30s"
    tls_verify: true
```

### **Hybrid Mode (Advanced)**
```yaml
# secretly.yaml
storage:
  type: "hybrid"
  local:
    database:
      path: "./secrets.db"
  remote:
    base_url: "https://secrets.company.com"
    api_key: "${SECRETLY_API_KEY}"
  hybrid:
    cache_duration: "1h"
    sync_interval: "5m"
    offline_mode: true
```

## 🎯 **Benefits of This Approach**

### **For Developers:**
- **Seamless Experience**: Same CLI commands work everywhere
- **Flexible Deployment**: Start local, scale to team/production
- **No Lock-in**: Easy to switch between local and remote
- **Offline Capable**: Can work without network connection

### **For Teams:**
- **Easy Collaboration**: Share secrets via development server
- **Gradual Adoption**: Start with one developer, add team members
- **Development Parity**: Same secrets across team environments
- **Simple Setup**: Docker Compose for instant team server

### **For Enterprises:**
- **Scalable Architecture**: From development to production
- **Security Options**: Authentication, TLS, audit logging
- **Hybrid Deployment**: Local cache with remote sync
- **Compliance Ready**: Full audit trails and access control

## 🚀 **Implementation Priority**

### **Week 1: Storage Factory**
- Create storage factory pattern
- Implement auto-detection logic
- Update CLI to use factory

### **Week 2: Remote Storage**
- Implement HTTP client storage
- Add configuration management
- Test with local server

### **Week 3: CLI Commands**
- Add config management commands
- Implement connection switching
- Add status/health checks

### **Week 4: Server Enhancements**
- Add local development mode
- Improve Docker deployment
- Add health endpoints

## 🤔 **Questions for You**

1. **Default Behavior**: Should CLI default to local or auto-detect?
2. **Authentication**: For development servers, should we skip auth entirely?
3. **Configuration**: Should config be per-project or global?
4. **Server Modes**: Do you want a simple "development mode" flag?

This approach gives you maximum flexibility - developers can start locally and seamlessly transition to team collaboration without changing their workflow!

Would you like me to start implementing the storage factory pattern?