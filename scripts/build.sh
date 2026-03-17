#!/bin/bash
# Build script for all binaries

set -e

echo "Building binaries..."

# Check if DB encryption is disabled
if grep -qE '^DISABLE_DB_ENCRYPTION=(false|False|FALSE)' .env 2>/dev/null; then
  echo "  - With SQLCipher encryption"
  BUILD_TAGS="-tags sqlcipher"
else
  echo "  - With standard SQLite (unencrypted)"
  BUILD_TAGS=""
fi

# Build binaries
mkdir -p bin
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build "$BUILD_TAGS" -o bin/api cmd/api/main.go
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build "$BUILD_TAGS" -o bin/migrate cmd/migrate/main.go
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build "$BUILD_TAGS" -o bin/qr-worker cmd/qr-worker/main.go
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build "$BUILD_TAGS" -o bin/worker cmd/worker/main.go

echo "Build complete!"
