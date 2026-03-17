#!/bin/bash
# Helper script to run Go commands with appropriate build tags

set -e

# Check if DB encryption is disabled
if grep -qE '^DISABLE_DB_ENCRYPTION=(false|False|FALSE)' .env 2>/dev/null; then
  BUILD_TAGS="-tags sqlcipher"
else
  BUILD_TAGS=""
fi

# Execute go run with the provided arguments
exec go run "$BUILD_TAGS" "$@"
