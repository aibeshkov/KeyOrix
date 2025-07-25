# 🌐 Remote CLI Implementation Guide

## 🎯 **Implementation Roadmap**

Let me show you exactly how to implement the remote CLI scenario, which is one of the most immediately valuable use cases enabled by our new architecture.

## 🏗️ **Step 1: Create Remote Storage Implementation**

### **Remote Storage Interface**
```go
// internal/storage/remote/remote.go
package remote

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/secretlyhq/secretly/internal/core/storage"
    "github.com/secretlyhq/secretly/internal/storage/models"
)

type RemoteStorage struct {
    client    *http.Client
    baseURL   string
    apiKey    string
    userAgent string
}

type RemoteConfig struct {
    BaseURL   string `yaml:"base_url"`
    APIKey    string `yaml:"api_key"`
    Timeout   int    `yaml:"timeout_seconds"`
    UserAgent string `yaml:"user_agent"`
}

func NewRemoteStorage(config *RemoteConfig) *RemoteStorage {
    return &RemoteStorage{
        client: &http.Client{
            Timeout: time.Duration(config.Timeout) * time.Second,
        },
        baseURL:   config.BaseURL,
        apiKey:    config.APIKey,
        userAgent: config.UserAgent,
    }
}

// Implement storage.Storage interface
func (rs *RemoteStorage) CreateSecret(ctx context.Context, secret *models.SecretNode) (*models.SecretNode, error) {
    payload, err := json.Marshal(secret)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal secret: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", rs.baseURL+"/api/v1/secrets", bytes.NewBuffer(payload))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    rs.setHeaders(req)

    resp, err := rs.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("remote API call failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, rs.handleErrorResponse(resp)
    }

    var response struct {
        Data *models.SecretNode `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return response.Data, nil
}

func (rs *RemoteStorage) GetSecret(ctx context.Context, id uint) (*models.SecretNode, error) {
    url := fmt.Sprintf("%s/api/v1/secrets/%d", rs.baseURL, id)
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    rs.setHeaders(req)

    resp, err := rs.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("remote API call failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("secret not found")
    }
    if resp.StatusCode != http.StatusOK {
        return nil, rs.handleErrorResponse(resp)
    }

    var response struct {
        Data *models.SecretNode `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return response.Data, nil
}

func (rs *RemoteStorage) ListSecrets(ctx context.Context, filter *storage.SecretFilter) ([]*models.SecretNode, int64, error) {
    url := rs.buildListURL(filter)
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to create request: %w", err)
    }

    rs.setHeaders(req)

    resp, err := rs.client.Do(req)
    if err != nil {
        return nil, 0, fmt.Errorf("remote API call failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, 0, rs.handleErrorResponse(resp)
    }

    var response struct {
        Data struct {
            Secrets    []*models.SecretNode `json:"secrets"`
            Total      int64                `json:"total"`
            Page       int                  `json:"page"`
            PageSize   int                  `json:"page_size"`
            TotalPages int                  `json:"total_pages"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, 0, fmt.Errorf("failed to decode response: %w", err)
    }

    return response.Data.Secrets, response.Data.Total, nil
}

// Helper methods
func (rs *RemoteStorage) setHeaders(req *http.Request) {
    req.Header.Set("Authorization", "Bearer "+rs.apiKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("User-Agent", rs.userAgent)
}

func (rs *RemoteStorage) handleErrorResponse(resp *http.Response) error {
    var errorResp struct {
        Error   string `json:"error"`
        Message string `json:"message"`
        Code    int    `json:"code"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
        return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
    }
    
    return fmt.Errorf("API error (%d): %s - %s", errorResp.Code, errorResp.Error, errorResp.Message)
}

func (rs *RemoteStorage) buildListURL(filter *storage.SecretFilter) string {
    url := rs.baseURL + "/api/v1/secrets"
    params := make([]string, 0)
    
    if filter.NamespaceID != nil {
        params = append(params, fmt.Sprintf("namespace_id=%d", *filter.NamespaceID))
    }
    if filter.ZoneID != nil {
        params = append(params, fmt.Sprintf("zone_id=%d", *filter.ZoneID))
    }
    if filter.EnvironmentID != nil {
        params = append(params, fmt.Sprintf("environment_id=%d", *filter.EnvironmentID))
    }
    if filter.Type != nil {
        params = append(params, fmt.Sprintf("type=%s", *filter.Type))
    }
    if filter.Page > 0 {
        params = append(params, fmt.Sprintf("page=%d", filter.Page))
    }
    if filter.PageSize > 0 {
        params = append(params, fmt.Sprintf("page_size=%d", filter.PageSize))
    }
    
    if len(params) > 0 {
        url += "?" + strings.Join(params, "&")
    }
    
    return url
}

// Implement remaining storage.Storage interface methods...
// (UpdateSecret, DeleteSecret, User management, RBAC, etc.)
```

## 🔧 **Step 2: Update Configuration**

### **Enhanced Configuration Structure**
```go
// internal/config/config.go - Add remote storage config
type StorageConfig struct {
    Type       string           `yaml:"type"` // "local" or "remote"
    Database   DatabaseConfig   `yaml:"database"`
    Encryption EncryptionConfig `yaml:"encryption"`
    Remote     RemoteConfig     `yaml:"remote"` // New remote config
}

type RemoteConfig struct {
    BaseURL        string            `yaml:"base_url"`
    APIKey         string            `yaml:"api_key"`
    TimeoutSeconds int               `yaml:"timeout_seconds"`
    UserAgent      string            `yaml:"user_agent"`
    Headers        map[string]string `yaml:"headers"`
    TLS            RemoteTLSConfig   `yaml:"tls"`
}

type RemoteTLSConfig struct {
    InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
    CertFile           string `yaml:"cert_file"`
    KeyFile            string `yaml:"key_file"`
    CAFile             string `yaml:"ca_file"`
}
```

### **Configuration File Example**
```yaml
# secretly.yaml - Remote configuration
storage:
  type: "remote"  # Switch from "local" to "remote"
  remote:
    base_url: "https://api.secretly.company.com"
    api_key: "${SECRETLY_API_KEY}"  # Environment variable
    timeout_seconds: 30
    user_agent: "secretly-cli/1.0.0"
    headers:
      X-Client-Version: "1.0.0"
      X-Environment: "production"
    tls:
      insecure_skip_verify: false
      ca_file: "/etc/ssl/certs/company-ca.pem"
  
  # Local config still available for fallback
  database:
    type: "sqlite"
    path: "~/.secretly/cache.db"  # Local cache
```

## 🔄 **Step 3: Storage Factory Pattern**

### **Dynamic Storage Creation**
```go
// internal/storage/factory.go
package storage

import (
    "fmt"
    
    "github.com/secretlyhq/secretly/internal/config"
    "github.com/secretlyhq/secretly/internal/core/storage"
    "github.com/secretlyhq/secretly/internal/storage/local"
    "github.com/secretlyhq/secretly/internal/storage/remote"
    "gorm.io/gorm"
)

type StorageFactory struct {
    config *config.Config
    db     *gorm.DB
}

func NewStorageFactory(config *config.Config, db *gorm.DB) *StorageFactory {
    return &StorageFactory{
        config: config,
        db:     db,
    }
}

func (sf *StorageFactory) CreateStorage() (storage.Storage, error) {
    switch sf.config.Storage.Type {
    case "local":
        return local.NewLocalStorage(sf.db), nil
        
    case "remote":
        return remote.NewRemoteStorage(&sf.config.Storage.Remote), nil
        
    case "hybrid":
        return sf.createHybridStorage()
        
    default:
        return nil, fmt.Errorf("unsupported storage type: %s", sf.config.Storage.Type)
    }
}

func (sf *StorageFactory) createHybridStorage() (storage.Storage, error) {
    localStorage := local.NewLocalStorage(sf.db)
    remoteStorage := remote.NewRemoteStorage(&sf.config.Storage.Remote)
    
    return NewHybridStorage(localStorage, remoteStorage), nil
}
```

## 🔀 **Step 4: Hybrid Storage Implementation**

### **Smart Local/Remote Storage**
```go
// internal/storage/hybrid/hybrid.go
package hybrid

import (
    "context"
    "fmt"
    "time"

    "github.com/secretlyhq/secretly/internal/core/storage"
    "github.com/secretlyhq/secretly/internal/storage/models"
)

type HybridStorage struct {
    local      storage.Storage
    remote     storage.Storage
    cacheTime  time.Duration
    syncQueue  chan SyncOperation
}

type SyncOperation struct {
    Operation string
    Secret    *models.SecretNode
    Timestamp time.Time
}

func NewHybridStorage(local, remote storage.Storage) *HybridStorage {
    hs := &HybridStorage{
        local:     local,
        remote:    remote,
        cacheTime: 5 * time.Minute,
        syncQueue: make(chan SyncOperation, 100),
    }
    
    // Start background sync worker
    go hs.syncWorker()
    
    return hs
}

func (hs *HybridStorage) CreateSecret(ctx context.Context, secret *models.SecretNode) (*models.SecretNode, error) {
    // Try remote first for consistency
    result, err := hs.remote.CreateSecret(ctx, secret)
    if err != nil {
        // Fallback to local storage
        result, err = hs.local.CreateSecret(ctx, secret)
        if err != nil {
            return nil, err
        }
        
        // Queue for sync when remote is available
        hs.queueSync(SyncOperation{
            Operation: "create",
            Secret:    result,
            Timestamp: time.Now(),
        })
    } else {
        // Cache locally for fast access
        hs.local.CreateSecret(ctx, result)
    }
    
    return result, nil
}

func (hs *HybridStorage) GetSecret(ctx context.Context, id uint) (*models.SecretNode, error) {
    // Try local cache first
    secret, err := hs.local.GetSecret(ctx, id)
    if err == nil && hs.isCacheFresh(secret) {
        return secret, nil
    }
    
    // Fetch from remote
    secret, err = hs.remote.GetSecret(ctx, id)
    if err != nil {
        // Fallback to local (even if stale)
        if secret, localErr := hs.local.GetSecret(ctx, id); localErr == nil {
            return secret, nil
        }
        return nil, err
    }
    
    // Update local cache
    hs.local.CreateSecret(ctx, secret) // Upsert
    
    return secret, nil
}

func (hs *HybridStorage) isCacheFresh(secret *models.SecretNode) bool {
    return time.Since(secret.UpdatedAt) < hs.cacheTime
}

func (hs *HybridStorage) queueSync(op SyncOperation) {
    select {
    case hs.syncQueue <- op:
    default:
        // Queue full, drop oldest
        <-hs.syncQueue
        hs.syncQueue <- op
    }
}

func (hs *HybridStorage) syncWorker() {
    for op := range hs.syncQueue {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        
        switch op.Operation {
        case "create":
            hs.remote.CreateSecret(ctx, op.Secret)
        case "update":
            hs.remote.UpdateSecret(ctx, op.Secret)
        case "delete":
            hs.remote.DeleteSecret(ctx, op.Secret.ID)
        }
        
        cancel()
        
        // Rate limit sync operations
        time.Sleep(100 * time.Millisecond)
    }
}
```

## 🖥️ **Step 5: Update CLI Commands**

### **Zero-Change CLI Implementation**
```go
// internal/cli/secret/create.go - NO CHANGES NEEDED!
func runCreate(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load("secretly.yaml")
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // Storage factory handles local vs remote automatically
    db, err := gorm.Open(sqlite.Open(cfg.Storage.Database.Path), &gorm.Config{})
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }

    factory := storage.NewStorageFactory(cfg, db)
    storageImpl, err := factory.CreateStorage()
    if err != nil {
        return fmt.Errorf("failed to create storage: %w", err)
    }

    service := core.NewSecretlyCore(storageImpl)

    // Rest of the code remains identical!
    var req *core.CreateSecretRequest
    if createInteractive {
        req, err = interactiveCreate()
    } else {
        req, err = buildCreateRequest()
    }
    if err != nil {
        return err
    }

    ctx := context.Background()
    secret, err := service.CreateSecret(ctx, req)
    if err != nil {
        return fmt.Errorf("failed to create secret: %w", err)
    }

    // Same output formatting
    fmt.Printf("✅ Secret created successfully!\n")
    fmt.Printf("ID: %d\n", secret.ID)
    fmt.Printf("Name: %s\n", secret.Name)
    // ... rest unchanged
}
```

## 🔐 **Step 6: Authentication & Security**

### **API Key Management**
```go
// internal/auth/apikey.go
package auth

import (
    "fmt"
    "os"
    "path/filepath"
)

type APIKeyManager struct {
    configDir string
}

func NewAPIKeyManager() *APIKeyManager {
    homeDir, _ := os.UserHomeDir()
    return &APIKeyManager{
        configDir: filepath.Join(homeDir, ".secretly"),
    }
}

func (akm *APIKeyManager) StoreAPIKey(endpoint, apiKey string) error {
    // Store encrypted API key locally
    keyFile := filepath.Join(akm.configDir, "credentials")
    
    credentials := map[string]string{
        endpoint: apiKey,
    }
    
    // Encrypt and store credentials
    return akm.encryptAndStore(keyFile, credentials)
}

func (akm *APIKeyManager) GetAPIKey(endpoint string) (string, error) {
    keyFile := filepath.Join(akm.configDir, "credentials")
    
    credentials, err := akm.loadAndDecrypt(keyFile)
    if err != nil {
        return "", err
    }
    
    apiKey, exists := credentials[endpoint]
    if !exists {
        return "", fmt.Errorf("no API key found for endpoint: %s", endpoint)
    }
    
    return apiKey, nil
}

// CLI command for authentication
// secretly auth login --endpoint https://api.secretly.company.com --api-key <key>
```

## 🚀 **Step 7: Usage Examples**

### **Local to Remote Migration**
```bash
# 1. Current local usage (no changes)
secretly secret create --name "db-password" --value "secret123"

# 2. Configure remote endpoint
secretly auth login --endpoint https://api.secretly.company.com --api-key sk_live_...

# 3. Update config to use remote
echo "storage:
  type: remote
  remote:
    base_url: https://api.secretly.company.com
    api_key: \${SECRETLY_API_KEY}" > ~/.secretly/config.yaml

# 4. Same commands now work remotely!
secretly secret create --name "db-password" --value "secret123"
secretly secret list
secretly secret get --id 1
```

### **Hybrid Mode Usage**
```bash
# Configure hybrid mode for best of both worlds
echo "storage:
  type: hybrid
  remote:
    base_url: https://api.secretly.company.com
    api_key: \${SECRETLY_API_KEY}
  database:
    path: ~/.secretly/cache.db" > ~/.secretly/config.yaml

# Commands work offline (uses cache) and online (syncs to remote)
secretly secret create --name "offline-secret" --value "works-offline"
```

## 🌟 **Benefits Achieved**

### **For Users**
- **Zero Learning Curve**: Same CLI commands work remotely
- **Offline Support**: Hybrid mode works without internet
- **Team Collaboration**: Multiple users share same secret store
- **Enterprise Integration**: Connects to company infrastructure

### **For Organizations**
- **Centralized Management**: All secrets in one place
- **Audit Trail**: Complete audit logging on server
- **Access Control**: Server-side RBAC enforcement
- **Compliance**: Meets enterprise security requirements

### **For Developers**
- **No Code Changes**: Existing CLI code works unchanged
- **Flexible Deployment**: Local, remote, or hybrid modes
- **Easy Testing**: Switch between local and remote for testing
- **Future-Proof**: Architecture supports any storage backend

## 🎯 **Next Steps**

1. **Implement Remote Storage**: Start with basic CRUD operations
2. **Add Authentication**: Implement API key management
3. **Create Hybrid Mode**: Add local caching with remote sync
4. **Add Configuration**: Support multiple endpoints and profiles
5. **Enhance Security**: Add TLS, certificate pinning, etc.
6. **Add Monitoring**: Connection health, sync status, etc.

This implementation transforms the CLI from a local-only tool into a **flexible, enterprise-ready client** that can work in any environment while maintaining the same user experience.