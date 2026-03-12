# Admin CLI

Manage services and user service subscriptions.

## Usage

```bash
./bin/admin <command> <action> [options]
```

Built with `make build`

## Commands

### Service Management

```bash
# List services
./bin/admin service list

# Create service
./bin/admin service create -name=<name> -display-name="<display>" [-description="<desc>"]

# Activate/deactivate service
./bin/admin service activate -id=<id>
./bin/admin service deactivate -id=<id>

# Delete service (cascades to user subscriptions)
./bin/admin service delete -id=<id>
```

**Example:**

```bash
./bin/admin service create -name=telegram -display-name="Telegram Bot"
```

### User Service Management

```bash
# List user's services
./bin/admin user-service list -user-id=<id>

# Subscribe user to service
./bin/admin user-service create \
  -user-id=<id> \
  -service-id=<id> \
  -client-id=<id> \
  -client-secret=<secret>

# Enable/disable subscription
./bin/admin user-service enable -id=<id>
./bin/admin user-service disable -id=<id>

# Delete subscription
./bin/admin user-service delete -id=<id>
```

**Example:**

```bash
./bin/admin user-service create \
  -user-id=1 \
  -service-id=1 \
  -client-id=abc123 \
  -client-secret=secret456
```
