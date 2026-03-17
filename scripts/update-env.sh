#!/bin/bash
# Script to update .env file with new fields from example.env while preserving existing values

set -e

# Detect if we should use production paths
if [ "$1" = "--production" ] || [ "$1" = "-p" ]; then
  PRODUCTION=true
  shift
  EXAMPLE_FILE="${1:-/opt/interface-api/default.env}"
  ENV_FILE="${2:-/opt/interface-api/.env}"
else
  PRODUCTION=false
  EXAMPLE_FILE="${1:-example.env}"
  ENV_FILE="${2:-.env}"
fi

BACKUP_FILE="${ENV_FILE}.backup.$(date +%Y%m%d_%H%M%S)"

echo "Updating ${ENV_FILE} from ${EXAMPLE_FILE}..."

# Check if example file exists
if [ ! -f "$EXAMPLE_FILE" ]; then
  echo "Error: Example file '${EXAMPLE_FILE}' not found"
  exit 1
fi

# Check if .env exists
if [ ! -f "$ENV_FILE" ]; then
  echo "No existing ${ENV_FILE} found. Creating from ${EXAMPLE_FILE}..."
  cp "$EXAMPLE_FILE" "$ENV_FILE"
  [ "$PRODUCTION" = true ] && chmod 600 "$ENV_FILE"
  echo "Created ${ENV_FILE}"
  exit 0
fi

# Create backup
echo "Creating backup: ${BACKUP_FILE}"
cp "$ENV_FILE" "$BACKUP_FILE"

# Create temporary file
TEMP_FILE=$(mktemp)

# Read current .env into associative array
declare -A current_vars
while IFS='=' read -r key value; do
  # Skip empty lines and comments
  [[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue
  # Remove leading/trailing whitespace from key
  key=$(echo "$key" | xargs)
  current_vars["$key"]="$value"
done <"$ENV_FILE"

# Process example.env line by line
while IFS= read -r line; do
  # If it's a comment or empty line, keep it
  if [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]]; then
    echo "$line" >>"$TEMP_FILE"
    continue
  fi

  # Extract key
  if [[ "$line" =~ ^([^=]+)= ]]; then
    key="${BASH_REMATCH[1]}"
    key=$(echo "$key" | xargs)

    # If key exists in current .env, use its value
    if [[ -n "${current_vars[$key]+isset}" ]]; then
      echo "${key}=${current_vars[$key]}" >>"$TEMP_FILE"
      unset 'current_vars[$key]'
    else
      # New key, use default from example
      echo "$line" >>"$TEMP_FILE"
      echo "  + Added new field: $key"
    fi
  else
    # Malformed line, keep as is
    echo "$line" >>"$TEMP_FILE"
  fi
done <"$EXAMPLE_FILE"

# Append any keys that were in old .env but not in example.env
if [ ${#current_vars[@]} -gt 0 ]; then
  echo "" >>"$TEMP_FILE"
  echo "# Legacy fields (not in ${EXAMPLE_FILE})" >>"$TEMP_FILE"
  for key in "${!current_vars[@]}"; do
    echo "${key}=${current_vars[$key]}" >>"$TEMP_FILE"
    echo "  ! Kept legacy field: $key"
  done
fi

# Replace old .env with new one
mv "$TEMP_FILE" "$ENV_FILE"
[ "$PRODUCTION" = true ] && chmod 600 "$ENV_FILE"

echo ""
echo "Update complete!"
echo "  Original saved to: ${BACKUP_FILE}"
echo "  Updated file: ${ENV_FILE}"
echo ""
echo "Please review the changes and update any new fields as needed."

# Remind to restart service in production
if [ "$PRODUCTION" = true ]; then
  echo ""
  echo "Remember to restart the service: systemctl restart interface-api"
fi
