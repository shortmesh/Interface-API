#!/bin/bash
# Development environment setup script

set -e

echo "Setting up development environment..."

# Create .env from example if it doesn't exist
if [ ! -f .env ]; then
  echo "Creating .env from example.env..."
  cp example.env .env
fi

# Function to generate and update key in .env
generate_key() {
  local key_name=$1
  local key_pattern=$2
  local generate_cmd=$3

  if ! grep -q "^${key_name}=${key_pattern}" .env 2>/dev/null; then
    echo "Generating ${key_name}..."
    local key_value
    key_value=$(eval "$generate_cmd")
    sed -i.bak "s|^${key_name}=.*|${key_name}=${key_value}|" .env && rm -f .env.bak
  else
    echo "${key_name} already set"
  fi
}

# Generate keys
generate_key "HASH_KEY" "[A-Za-z0-9+/=]\{40,\}" "openssl rand -base64 32"
generate_key "DB_ENCRYPTION_KEY" "[A-Fa-f0-9]\{64,\}" "openssl rand -hex 32"
generate_key "CLIENT_ID" "[A-Za-z0-9]\{20,\}" "openssl rand -hex 16"
generate_key "CLIENT_SECRET" "[A-Za-z0-9]\{40,\}" "openssl rand -hex 32"

echo "Setup complete! Run 'make migrate-up && make run' to start."
