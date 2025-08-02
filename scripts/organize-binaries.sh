#!/bin/bash

# Binary Organization Script
# Moves all binary executables to ./bin directory for better organization

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

echo "📦 Organizing Binary Executables"
echo "================================"

# Create bin directory
log_info "Creating ./bin directory..."
mkdir -p bin

# Track moved files
MOVED_COUNT=0

# Function to move binary if it exists
move_binary() {
    local source="$1"
    local target="$2"
    
    if [ -f "$source" ]; then
        log_info "Moving $source -> bin/$target"
        mv "$source" "bin/$target"
        MOVED_COUNT=$((MOVED_COUNT + 1))
        return 0
    else
        log_warning "$source not found (skipping)"
        return 1
    fi
}

# Move main binaries
log_info "Moving main application binaries..."
move_binary "secretly" "secretly"
move_binary "server/secretly-server" "secretly-server"

# Move test binaries
log_info "Moving test binaries..."
move_binary "core-test" "core-test"
move_binary "secret_crud" "secret_crud"
move_binary "secret-crud-test" "secret-crud-test"
move_binary "new-architecture" "new-architecture"
move_binary "system_init" "system_init"

# Move any other executables found in root
log_info "Scanning for other executables in root directory..."
for file in *; do
    if [ -f "$file" ] && [ -x "$file" ] && [[ ! "$file" == *.* ]] && [[ ! "$file" == "bin" ]]; then
        # Check if it's likely a binary (executable without extension)
        if file "$file" | grep -q "executable"; then
            log_info "Found additional executable: $file"
            move_binary "$file" "$file"
        fi
    fi
done

# Create symlinks in root for main binaries (for backward compatibility)
log_info "Creating convenience symlinks..."
if [ -f "bin/secretly" ]; then
    ln -sf bin/secretly secretly
    log_success "Created symlink: secretly -> bin/secretly"
fi

if [ -f "bin/secretly-server" ]; then
    ln -sf bin/secretly-server secretly-server
    log_success "Created symlink: secretly-server -> bin/secretly-server"
fi

# Update scripts to use bin directory
log_info "Updating scripts to use bin directory..."

# Update start-server.sh
if [ -f "scripts/start-server.sh" ]; then
    sed -i.bak 's|./secretly|./bin/secretly|g' scripts/start-server.sh
    sed -i.bak 's|./server/secretly-server|./bin/secretly-server|g' scripts/start-server.sh
    log_success "Updated scripts/start-server.sh"
fi

# Update test scripts
if [ -f "scripts/test-real-usage.sh" ]; then
    sed -i.bak 's|./secretly|./bin/secretly|g' scripts/test-real-usage.sh
    log_success "Updated scripts/test-real-usage.sh"
fi

# Update deploy scripts
if [ -f "scripts/deploy-simple.sh" ]; then
    sed -i.bak 's|./secretly|./bin/secretly|g' scripts/deploy-simple.sh
    log_success "Updated scripts/deploy-simple.sh"
fi

# Update comprehensive test script
if [ -f "scripts/run-comprehensive-tests.sh" ]; then
    sed -i.bak 's|./secretly|./bin/secretly|g' scripts/run-comprehensive-tests.sh
    log_success "Updated scripts/run-comprehensive-tests.sh"
fi

# Create a comprehensive build script that outputs to bin
log_info "Creating improved build script..."
cat > scripts/build.sh << 'EOF'
#!/bin/bash

# Comprehensive Build Script - Outputs to ./bin directory
# Supports multiple build modes and platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Build configuration
BUILD_MODE="${BUILD_MODE:-release}"
TARGET_OS="${TARGET_OS:-$(go env GOOS)}"
TARGET_ARCH="${TARGET_ARCH:-$(go env GOARCH)}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
BUILD_TIME="$(date -u '+%Y-%m-%d_%H:%M:%S')"
GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

echo "🔨 Building Secretly Binaries"
echo "============================="
echo "Build Mode: $BUILD_MODE"
echo "Target OS: $TARGET_OS"
echo "Target Arch: $TARGET_ARCH"
echo "Version: $VERSION"
echo "Build Time: $BUILD_TIME"
echo "Git Commit: $GIT_COMMIT"
echo ""

# Create bin directory
log_info "Creating bin directory..."
mkdir -p bin

# Build flags based on mode
if [ "$BUILD_MODE" = "debug" ]; then
    BUILD_FLAGS="-gcflags='all=-N -l'"
    log_info "Debug mode: Including debug symbols"
elif [ "$BUILD_MODE" = "release" ]; then
    BUILD_FLAGS="-ldflags='-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT'"
    log_info "Release mode: Optimized build with version info"
else
    BUILD_FLAGS=""
    log_info "Standard build mode"
fi

# Set cross-compilation environment
export GOOS="$TARGET_OS"
export GOARCH="$TARGET_ARCH"

# Build main CLI
log_info "Building CLI binary (secretly)..."
if eval "go build $BUILD_FLAGS -o bin/secretly ./cmd/secretly"; then
    log_success "CLI binary built successfully"
else
    log_error "Failed to build CLI binary"
    exit 1
fi

# Build server
log_info "Building server binary (secretly-server)..."
if eval "go build $BUILD_FLAGS -o bin/secretly-server ./server"; then
    log_success "Server binary built successfully"
else
    log_error "Failed to build server binary"
    exit 1
fi

# Build additional tools if they exist
if [ -d "examples/secret_crud" ]; then
    log_info "Building secret_crud example..."
    if eval "go build $BUILD_FLAGS -o bin/secret_crud ./examples/secret_crud"; then
        log_success "secret_crud built successfully"
    else
        log_warning "Failed to build secret_crud"
    fi
fi

if [ -d "examples/new-architecture" ]; then
    log_info "Building new-architecture example..."
    if eval "go build $BUILD_FLAGS -o bin/new-architecture ./examples/new-architecture"; then
        log_success "new-architecture built successfully"
    else
        log_warning "Failed to build new-architecture"
    fi
fi

# Build validation tools if they exist
if [ -f "cmd/validate-translations/main.go" ]; then
    log_info "Building translation validator..."
    if eval "go build $BUILD_FLAGS -o bin/validate-translations ./cmd/validate-translations"; then
        log_success "validate-translations built successfully"
    else
        log_warning "Failed to build validate-translations"
    fi
fi

# Make binaries executable
log_info "Setting executable permissions..."
chmod +x bin/*

# Build web assets if web directory exists
if [ -d "web" ] && [ -f "web/package.json" ]; then
    log_info "Building web assets..."
    cd web
    if command -v npm &> /dev/null; then
        if [ ! -d "node_modules" ]; then
            log_info "Installing web dependencies..."
            npm install
        fi
        log_info "Building web production bundle..."
        if npm run build; then
            log_success "Web assets built successfully"
        else
            log_warning "Failed to build web assets"
        fi
    else
        log_warning "npm not found, skipping web build"
    fi
    cd ..
fi

# Create build info file
log_info "Creating build information file..."
cat > bin/build-info.txt << EOF
Secretly Build Information
==========================
Version: $VERSION
Build Mode: $BUILD_MODE
Build Time: $BUILD_TIME
Git Commit: $GIT_COMMIT
Target OS: $TARGET_OS
Target Arch: $TARGET_ARCH
Go Version: $(go version)
EOF

# Display build results
echo ""
log_success "Build completed successfully!"
echo ""
echo "📦 Built Binaries:"
echo "=================="
ls -la bin/ | grep -E '^-.*x.*' || true
echo ""
echo "📊 Binary Sizes:"
echo "================"
du -h bin/* 2>/dev/null || true
echo ""
echo "🔧 Usage:"
echo "  ./bin/secretly --help        # CLI help"
echo "  ./bin/secretly-server --help # Server help"
echo "  ./bin/secretly version       # Show version"
echo ""
log_success "All binaries ready in ./bin/ directory! 🎉"
EOF

chmod +x scripts/build.sh

# Create a clean script
log_info "Creating clean script..."
cat > scripts/clean.sh << 'EOF'
#!/bin/bash

# Clean Script - Removes all built binaries

echo "🧹 Cleaning built binaries..."

# Remove bin directory
if [ -d "bin" ]; then
    rm -rf bin
    echo "✅ Removed ./bin directory"
fi

# Remove symlinks
if [ -L "secretly" ]; then
    rm secretly
    echo "✅ Removed secretly symlink"
fi

if [ -L "secretly-server" ]; then
    rm secretly-server
    echo "✅ Removed secretly-server symlink"
fi

# Remove any remaining binaries in root
for file in core-test secret_crud secret-crud-test new-architecture system_init; do
    if [ -f "$file" ]; then
        rm "$file"
        echo "✅ Removed $file"
    fi
done

echo "🧹 Clean completed!"
EOF

chmod +x scripts/clean.sh

# Update Makefile if it exists
if [ -f "Makefile" ]; then
    log_info "Updating Makefile..."
    cat > Makefile << 'EOF'
# Secretly Makefile - Comprehensive Build System

.PHONY: build build-debug build-release build-all-platforms build-docker clean test test-integration test-coverage run server web install uninstall dev docker-up docker-down help

# Default target
all: build

# Variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')
BUILD_MODE ?= release

# Build all binaries (default release mode)
build:
	@echo "🔨 Building Secretly ($(BUILD_MODE) mode)..."
	@BUILD_MODE=$(BUILD_MODE) ./scripts/build.sh

# Build in debug mode
build-debug:
	@echo "🔨 Building Secretly (debug mode)..."
	@BUILD_MODE=debug ./scripts/build.sh

# Build in release mode (optimized)
build-release:
	@echo "🔨 Building Secretly (release mode)..."
	@BUILD_MODE=release ./scripts/build.sh

# Build for all platforms
build-all-platforms:
	@echo "🌍 Building for all platforms..."
	@./scripts/build-all-platforms.sh

# Build Docker images
build-docker:
	@echo "🐳 Building Docker images..."
	@./scripts/build-docker.sh

# Clean all built artifacts
clean:
	@echo "🧹 Cleaning all build artifacts..."
	@./scripts/clean.sh
	@rm -rf dist/
	@docker rmi -f $(shell docker images -q secretly* 2>/dev/null) 2>/dev/null || true
	@echo "✅ Clean completed!"

# Run all tests
test:
	@echo "🧪 Running unit tests..."
	@go test ./... -v -short

# Run integration tests
test-integration:
	@echo "🧪 Running integration tests..."
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

# Run comprehensive test suite
test-all:
	@echo "🧪 Running comprehensive test suite..."
	@./scripts/run-comprehensive-tests.sh

# Run CLI with arguments
run:
	@./bin/secretly $(ARGS)

# Run server with arguments
server:
	@./bin/secretly-server $(ARGS)

# Build and serve web dashboard
web:
	@echo "🌐 Building and serving web dashboard..."
	@./scripts/setup-web-dashboard.sh

# Development mode (build and run server)
dev: build
	@echo "🚀 Starting development server..."
	@./bin/secretly-server --config secretly-simple.yaml

# Install binaries to system PATH
install: build
	@echo "📦 Installing binaries to /usr/local/bin/..."
	@sudo cp bin/secretly /usr/local/bin/
	@sudo cp bin/secretly-server /usr/local/bin/
	@sudo chmod +x /usr/local/bin/secretly /usr/local/bin/secretly-server
	@echo "✅ Installed successfully!"

# Uninstall binaries from system PATH
uninstall:
	@echo "🗑️  Uninstalling binaries..."
	@sudo rm -f /usr/local/bin/secretly /usr/local/bin/secretly-server
	@echo "✅ Uninstalled successfully!"

# Docker Compose operations
docker-up:
	@echo "🐳 Starting Docker services..."
	@docker-compose -f docker-compose.full-stack.yml up -d

docker-down:
	@echo "🐳 Stopping Docker services..."
	@docker-compose -f docker-compose.full-stack.yml down

# Format code
fmt:
	@echo "🎨 Formatting Go code..."
	@go fmt ./...
	@echo "✅ Code formatted!"

# Lint code
lint:
	@echo "🔍 Linting code..."
	@golangci-lint run || echo "⚠️  Install golangci-lint for better linting"

# Security scan
security:
	@echo "🔒 Running security scan..."
	@./scripts/security-hardening-simple.sh

# Generate documentation
docs:
	@echo "📚 Generating documentation..."
	@./scripts/create-documentation.sh

# Setup development environment
setup:
	@echo "⚙️  Setting up development environment..."
	@go mod download
	@go mod tidy
	@mkdir -p bin data logs
	@echo "✅ Development environment ready!"

# Show version information
version:
	@echo "Secretly Build Information"
	@echo "=========================="
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(shell go version)"

# Show comprehensive help
help:
	@echo "🔧 Secretly Makefile Commands"
	@echo "=============================="
	@echo ""
	@echo "📦 Build Commands:"
	@echo "  build              - Build all binaries (default: release mode)"
	@echo "  build-debug        - Build with debug symbols"
	@echo "  build-release      - Build optimized release binaries"
	@echo "  build-all-platforms- Build for all supported platforms"
	@echo "  build-docker       - Build Docker images"
	@echo ""
	@echo "🧪 Test Commands:"
	@echo "  test               - Run unit tests"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-all           - Run comprehensive test suite"
	@echo ""
	@echo "🚀 Run Commands:"
	@echo "  run ARGS='...'     - Run CLI with arguments"
	@echo "  server ARGS='...'  - Run server with arguments"
	@echo "  web                - Build and serve web dashboard"
	@echo "  dev                - Development mode (build + run server)"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  docker-up          - Start Docker services"
	@echo "  docker-down        - Stop Docker services"
	@echo ""
	@echo "🔧 Utility Commands:"
	@echo "  clean              - Remove all build artifacts"
	@echo "  install            - Install binaries to system PATH"
	@echo "  uninstall          - Remove binaries from system PATH"
	@echo "  fmt                - Format Go code"
	@echo "  lint               - Lint code"
	@echo "  security           - Run security hardening"
	@echo "  docs               - Generate documentation"
	@echo "  setup              - Setup development environment"
	@echo "  version            - Show version information"
	@echo "  help               - Show this help"
	@echo ""
	@echo "📖 Examples:"
	@echo "  make build                    # Build all binaries"
	@echo "  make run ARGS='--help'        # Show CLI help"
	@echo "  make server ARGS='--config=my.yaml'  # Run server with config"
	@echo "  make test-coverage            # Run tests with coverage"
	@echo "  BUILD_MODE=debug make build   # Debug build"
EOF
    log_success "Updated comprehensive Makefile"
fi

# Update .gitignore to include bin directory properly
log_info "Updating .gitignore..."
if [ -f ".gitignore" ]; then
    # Remove old binary entries and add bin directory
    grep -v "^secretly$\|^secretly-server$\|^core-test$\|^secret_crud$\|^secret-crud-test$\|^new-architecture$\|^system_init$" .gitignore > .gitignore.tmp || true
    echo "" >> .gitignore.tmp
    echo "# Built binaries" >> .gitignore.tmp
    echo "bin/" >> .gitignore.tmp
    echo "secretly" >> .gitignore.tmp
    echo "secretly-server" >> .gitignore.tmp
    mv .gitignore.tmp .gitignore
    log_success "Updated .gitignore"
fi

# Create bin directory README
log_info "Creating bin directory documentation..."
cat > bin/README.md << 'EOF'
# Binary Directory

This directory contains all compiled binary executables for the Secretly project.

## Binaries

### Main Applications
- **`secretly`** - Main CLI application for secret management
- **`secretly-server`** - Server application for web and API access

### Development Tools
- **`secret_crud`** - Example CRUD operations tool
- **`new-architecture`** - Architecture testing tool
- **`core-test`** - Core functionality testing tool

## Usage

### CLI Usage
```bash
# Run CLI directly
./bin/secretly --help

# Or use the convenience symlink
./secretly --help
```

### Server Usage
```bash
# Run server directly
./bin/secretly-server --config config.yaml

# Or use the convenience symlink
./secretly-server --config config.yaml
```

## Building

To rebuild all binaries:
```bash
# Using the build script
./scripts/build.sh

# Or using make
make build
```

## Cleaning

To remove all binaries:
```bash
# Using the clean script
./scripts/clean.sh

# Or using make
make clean
```

## Installation

To install binaries system-wide:
```bash
make install
```

This will copy binaries to `/usr/local/bin/` for system-wide access.
EOF

# Summary
echo ""
log_success "Binary organization completed!"
echo ""
echo "📦 Organization Summary:"
echo "======================="
echo "✅ Moved $MOVED_COUNT binary files to ./bin/"
echo "✅ Created convenience symlinks for main binaries"
echo "✅ Updated scripts to use ./bin/ directory"
echo "✅ Created improved build and clean scripts"
echo "✅ Updated Makefile with bin-aware targets"
echo "✅ Updated .gitignore for proper binary handling"
echo "✅ Created bin directory documentation"
echo ""
echo "📁 New Structure:"
echo "├── bin/                    # All binary executables"
echo "│   ├── secretly           # Main CLI binary"
echo "│   ├── secretly-server    # Server binary"
echo "│   └── ...                # Other tools"
echo "├── secretly -> bin/secretly           # Convenience symlink"
echo "├── secretly-server -> bin/secretly-server  # Convenience symlink"
echo "└── scripts/"
echo "    ├── build.sh           # Build to bin/"
echo "    └── clean.sh           # Clean bin/"
echo ""
echo "🔧 Usage:"
echo "  ./scripts/build.sh       # Build all binaries"
echo "  ./scripts/clean.sh       # Clean all binaries"
echo "  make build               # Alternative build"
echo "  make clean               # Alternative clean"
echo "  ./bin/secretly --help    # Direct binary access"
echo "  ./secretly --help        # Symlink access"
echo ""
log_success "Project organization improved! 🎉"