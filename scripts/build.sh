#!/bin/bash
# V Panel Build Script

set -e

# Configuration
APP_NAME="v-panel"
VERSION="${VERSION:-dev}"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build directories
BUILD_DIR="./build"
DIST_DIR="./dist"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean build directories
clean() {
    log_info "Cleaning build directories..."
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    mkdir -p "$BUILD_DIR" "$DIST_DIR"
}

# Build frontend
build_frontend() {
    log_info "Building frontend..."
    
    if [ ! -d "web" ]; then
        log_warn "Frontend directory not found, skipping..."
        return
    fi
    
    cd web
    
    if [ ! -d "node_modules" ]; then
        log_info "Installing frontend dependencies..."
        npm ci --legacy-peer-deps
    fi
    
    npm run build
    cd ..
    
    # Copy frontend build to dist
    if [ -d "web/dist" ]; then
        cp -r web/dist "$DIST_DIR/web"
    fi
}

# Build backend
build_backend() {
    log_info "Building backend..."
    
    LDFLAGS="-s -w"
    LDFLAGS="$LDFLAGS -X main.version=$VERSION"
    LDFLAGS="$LDFLAGS -X main.buildTime=$BUILD_TIME"
    LDFLAGS="$LDFLAGS -X main.gitCommit=$GIT_COMMIT"
    
    # Build for current platform
    CGO_ENABLED=1 go build \
        -ldflags="$LDFLAGS" \
        -o "$BUILD_DIR/$APP_NAME" \
        ./cmd/v/main.go
    
    log_info "Backend built: $BUILD_DIR/$APP_NAME"
}

# Build for multiple platforms
build_all_platforms() {
    log_info "Building for all platforms..."
    
    LDFLAGS="-s -w"
    LDFLAGS="$LDFLAGS -X main.version=$VERSION"
    LDFLAGS="$LDFLAGS -X main.buildTime=$BUILD_TIME"
    LDFLAGS="$LDFLAGS -X main.gitCommit=$GIT_COMMIT"
    
    # Linux AMD64
    log_info "Building for linux/amd64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="$LDFLAGS" \
        -o "$BUILD_DIR/${APP_NAME}-linux-amd64" \
        ./cmd/v/main.go
    
    # Linux ARM64
    log_info "Building for linux/arm64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -ldflags="$LDFLAGS" \
        -o "$BUILD_DIR/${APP_NAME}-linux-arm64" \
        ./cmd/v/main.go
    
    # macOS AMD64
    log_info "Building for darwin/amd64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
        -ldflags="$LDFLAGS" \
        -o "$BUILD_DIR/${APP_NAME}-darwin-amd64" \
        ./cmd/v/main.go
    
    # macOS ARM64
    log_info "Building for darwin/arm64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
        -ldflags="$LDFLAGS" \
        -o "$BUILD_DIR/${APP_NAME}-darwin-arm64" \
        ./cmd/v/main.go
}

# Create distribution package
package() {
    log_info "Creating distribution package..."
    
    # Copy binary
    cp "$BUILD_DIR/$APP_NAME" "$DIST_DIR/"
    
    # Copy configs
    mkdir -p "$DIST_DIR/configs"
    cp configs/config.yaml.example "$DIST_DIR/configs/"
    cp configs/xray.json.example "$DIST_DIR/configs/"
    
    # Copy scripts
    mkdir -p "$DIST_DIR/scripts"
    cp scripts/*.sh "$DIST_DIR/scripts/" 2>/dev/null || true
    
    # Create data directories
    mkdir -p "$DIST_DIR/data" "$DIST_DIR/logs"
    
    # Create archive
    ARCHIVE_NAME="${APP_NAME}-${VERSION}.tar.gz"
    tar -czf "$ARCHIVE_NAME" -C "$DIST_DIR" .
    
    log_info "Distribution package created: $ARCHIVE_NAME"
}

# Run tests
test() {
    log_info "Running tests..."
    go test -v ./...
}

# Show help
help() {
    echo "V Panel Build Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  clean       Clean build directories"
    echo "  frontend    Build frontend only"
    echo "  backend     Build backend only"
    echo "  all         Build frontend and backend"
    echo "  platforms   Build for all platforms"
    echo "  package     Create distribution package"
    echo "  test        Run tests"
    echo "  help        Show this help"
    echo ""
    echo "Environment variables:"
    echo "  VERSION     Version string (default: dev)"
}

# Main
case "${1:-all}" in
    clean)
        clean
        ;;
    frontend)
        build_frontend
        ;;
    backend)
        clean
        build_backend
        ;;
    all)
        clean
        build_frontend
        build_backend
        ;;
    platforms)
        clean
        build_all_platforms
        ;;
    package)
        clean
        build_frontend
        build_backend
        package
        ;;
    test)
        test
        ;;
    help|--help|-h)
        help
        ;;
    *)
        log_error "Unknown command: $1"
        help
        exit 1
        ;;
esac

log_info "Done!"
