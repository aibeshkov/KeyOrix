# 🎯 Simple CLI Architecture: Config-Driven Storage

## **Core Principle: No Server Required for Local Development**

The CLI should work **directly with the database** by default, and only use remote mode when explicitly configured.

## **User Experience**

### **Scenario 1: Individual Developer (Default)**
```bash
# Works immediately - no server setup needed
secretly secret create --name "dev-secret" --value "abc123"
secretly secret list
secretly share create --secret-id 1 --recipient-id 2 --permission read

# Uses local SQLite database directly (current MVP behavior)
```

### **Scenario 2: Team Collaboration**
```bash
# One team member runs the server
secretly server start --port 8080

# Other team members switch to remote mode
secretly config set-remote --url http://team-server:8080

# Now CLI uses remote server for all operations
secretly secret create --name "team-secret" --value "shared123"
secretly secret list  # Shows secrets from remote server
```

### **Scenario 3: Switch Back to Local**
```bash
# Switch back to local database
secretly config use-local

# Now works with local database again
secretly secret list  # Shows local secrets
```

## **Configuration Design**

### **Default: Local Mode**
```yaml
# secretly.yaml (default)
storage:
  type: "local"
  local:
    database:
      path: "./secrets.db"
```

### **Remote Mode**
```yaml
# secretly.yaml (when using remote server)
storage:
  type: "remote"
  remote:
    base_url: "http://localhost:8080"
    api_key: ""  # Optional for development servers
    timeout: "30s"
```

## **CLI Commands for Configuration**

```bash
# Check current configuration
secretly config status

# Switch to remote mode
secretly config set-remote --url http://team-server:8080

# Switch back to local mode  
secretly config use-local

# Test connection (for remote mode)
secretly config test-connection
```

## **Implementation**

### **1. Storage Factory (Simple)**
```go
func CreateStorage(config *config.Config) (storage.Storage, error) {
    switch config.Storage.Type {
    case "remote":
        return remote.NewRemoteStorage(config.Storage.Remote)
    default: // "local" or empty
        return local.NewLocalStorage(config.Storage.Local.Database.Path)
    }
}
```

### **2. CLI Integration**
```go
// All CLI commands use the same pattern
func initializeCore() (*core.SecretlyCore, error) {
    config := loadConfig()
    storage := CreateStorage(config)
    return core.NewSecretlyCore(storage), nil
}
```

### **3. Configuration Commands**
```go
// secretly config set-remote --url http://server:8080
func setRemoteConfig(url string) error {
    config := loadConfig()
    config.Storage.Type = "remote"
    config.Storage.Remote.BaseURL = url
    return saveConfig(config)
}

// secretly config use-local
func useLocalConfig() error {
    config := loadConfig()
    config.Storage.Type = "local"
    return saveConfig(config)
}
```

## **Benefits**

### **For Individual Developers:**
- ✅ **No setup required** - CLI works immediately
- ✅ **No server needed** - direct database access
- ✅ **Same as current MVP** - zero breaking changes

### **For Teams:**
- ✅ **Simple collaboration** - one config change
- ✅ **Flexible deployment** - any team member can run server
- ✅ **Easy switching** - local/remote as needed

### **For Architecture:**
- ✅ **Clean separation** - storage abstraction
- ✅ **Simple configuration** - just a type switch
- ✅ **No complexity** - no auto-detection needed

## **Deployment Scenarios**

### **Individual Development**
```bash
# No server needed - works out of the box
secretly secret create --name "test" --value "local"
```

### **Team Development**
```bash
# One person runs server
secretly server start --port 8080

# Others configure remote access
secretly config set-remote --url http://dev-server:8080
```

### **Production**
```bash
# Production server with authentication
secretly config set-remote --url https://secrets.company.com
secretly config set-api-key sk_prod_...
```

This approach is **much cleaner** because:
1. **Default behavior unchanged** - CLI works locally like current MVP
2. **Simple configuration** - just a type switch in config
3. **No server dependency** - for local development
4. **Easy team adoption** - one config command to switch
5. **Clean architecture** - storage abstraction without complexity