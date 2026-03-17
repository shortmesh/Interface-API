#!/bin/bash
# Systemd service setup script

set -e

echo "Setting up systemd service..."

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
  echo "Error: This script must be run as root (use sudo)"
  exit 1
fi

# Create system user if it doesn't exist
if ! id interface-api >/dev/null 2>&1; then
  echo "Creating interface-api user..."
  useradd --system --no-create-home --shell /usr/sbin/nologin interface-api
else
  echo "User interface-api already exists"
fi

# Create application directory
echo "Creating application directory..."
mkdir -p /opt/interface-api

# Setup .env file
if [ ! -f /opt/interface-api/.env ]; then
  echo "Creating .env file from example.env..."
  cp example.env /opt/interface-api/.env

  echo "Setting production values..."
  sed -i "s|^APP_MODE=.*|APP_MODE=production|" /opt/interface-api/.env
  sed -i "s|^ALLOW_INSECURE_SERVER=.*|ALLOW_INSECURE_SERVER=true|" /opt/interface-api/.env
  sed -i "s|^ALLOW_INSECURE_EXTERNAL=.*|ALLOW_INSECURE_EXTERNAL=true|" /opt/interface-api/.env
  sed -i "s|^SQLITE_DB_PATH=.*|SQLITE_DB_PATH=/opt/interface-api/data/shortmesh.db|" /opt/interface-api/.env
  sed -i "s|^DISABLE_DB_ENCRYPTION=.*|DISABLE_DB_ENCRYPTION=false|" /opt/interface-api/.env
  sed -i "s|^AUTO_MIGRATE=.*|AUTO_MIGRATE=false|" /opt/interface-api/.env

  echo "Generating HASH_KEY..."
  HASH_KEY=$(openssl rand -base64 32)
  sed -i "s|^HASH_KEY=.*|HASH_KEY=$HASH_KEY|" /opt/interface-api/.env

  echo "Generating DB_ENCRYPTION_KEY..."
  DB_ENCRYPTION_KEY=$(openssl rand -hex 32)
  sed -i "s|^DB_ENCRYPTION_KEY=.*|DB_ENCRYPTION_KEY=$DB_ENCRYPTION_KEY|" /opt/interface-api/.env

  echo "Generating CLIENT_ID..."
  CLIENT_ID=$(openssl rand -hex 16)
  sed -i "s|^CLIENT_ID=.*|CLIENT_ID=$CLIENT_ID|" /opt/interface-api/.env

  echo "Generating CLIENT_SECRET..."
  CLIENT_SECRET=$(openssl rand -hex 32)
  sed -i "s|^CLIENT_SECRET=.*|CLIENT_SECRET=$CLIENT_SECRET|" /opt/interface-api/.env

  echo "WARNING: Review and update /opt/interface-api/.env with:"
  echo "  - TLS_CERT_FILE and TLS_KEY_FILE"
  echo "  - MAS_URL, MAS_ADMIN_URL, and credentials"
  echo "  - MATRIX_CLIENT_URL"
  echo "  - RABBITMQ_URL"
else
  echo ".env file already exists at /opt/interface-api/.env"
fi

# Copy default.env
echo "Copying default.env..."
cp default.env /opt/interface-api/default.env

# Create directories
echo "Creating data directory..."
mkdir -p /opt/interface-api/data

echo "Creating cache directories..."
mkdir -p /opt/interface-api/.cache/go-build /opt/interface-api/.cache/go-mod

# Set permissions
echo "Setting permissions..."
chown -R interface-api:interface-api /opt/interface-api
chmod 600 /opt/interface-api/.env

# Install systemd service
echo "Installing systemd service file..."
cp interface-api.service /etc/systemd/system/
systemctl daemon-reload

echo "Enabling service..."
systemctl enable interface-api

echo "Setup complete! Use 'systemctl start interface-api' to start the service."
