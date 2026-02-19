# Interface API

The primary interface for user interaction.

## Requirements

- Go 1.24.0 or higher
- MySQL (optional - SQLite is used by default)

## Quick Start

```bash
git clone https://github.com/shortmesh/Interface-API.git
cd Interface-API
cp .env.example .env
make migrate-up
make run
```

Server runs on `http://localhost:8080`

**Running with Worker:**

```bash
# Terminal 1 - API Server
make run

# Terminal 2 - Message Worker
make worker

# Or run both in background (Unix-like systems)
make run & make worker &
```

## Configuration

All configuration via environment variables in `.env` file:

### Server

```env
PORT=8080                    # Server port
LOG_LEVEL=info               # Logging level: debug, info, warn, error, fatal
```

### Database

Choose SQLite (default) or MySQL:

**SQLite:**

```env
SQLITE_DB_PATH=./shortmesh.db
```

**MySQL:**

```env
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DATABASE=shortmesh
MYSQL_USERNAME=root
MYSQL_PASSWORD=yourpassword
```

Note: If `SQLITE_DB_PATH` is set, SQLite will be used. Otherwise, MySQL is required.

### Migrations

```env
AUTO_MIGRATE=true            # Run pending migrations on app start
AUTO_CREATE_TABLES=true      # Create all tables on first run using GORM AutoMigrate
```

**Production:** Set both to `false` and run migrations manually via `make migrate-up`

### Security

#### Cryptographic Keys

Generate secure 32-byte base64-encoded keys:

```bash
# Linux/Mac
openssl rand -base64 32
```

```env
ENCRYPTION_KEY=<base64-key-here>  # For AES-256-GCM encryption
HASH_KEY=<base64-key-here>        # For HMAC-SHA256 hashing
```

**Important:** All three keys must be base64-encoded 32-byte keys. Generate three separate keys.

#### Session Management

```env
MAX_SESSIONS_PER_USER=1           # Maximum concurrent sessions per user
SESSION_DURATION_HOURS=720        # Session duration (default: 30 days)
SESSION_TOKEN_PREFIX=sk_          # Prefix for session tokens
```

#### Password Policy

```env
PASSWORD_POLICY_ENABLED=true      # Enable password validation
PASSWORD_MIN_LENGTH=8             # Minimum password length
PASSWORD_MAX_LENGTH=64            # Maximum password length
PASSWORD_CHECK_PWNED=true         # Check against Pwned Passwords API
PASSWORD_CHECK_SPACES=true        # Reject passwords with leading/trailing spaces
PASSWORD_PWNED_TIMEOUT=5          # API timeout in seconds
PASSWORD_SKIP_PWNED_ON_ERROR=true # Continue if pwned check fails
```

### External Services

#### Matrix Authentication Service (MAS)

```env
MAS_URL=                          # MAS service URL
MAS_ADMIN_URL=                    # MAS admin API URL
ADMIN_CLIENT_ID=                  # Admin client ID for MAS
ADMIN_CLIENT_SECRET=              # Admin client secret for MAS
```

#### Matrix Client

```env
MATRIX_CLIENT_URL=                # Matrix client URL
```

#### Argon2id Password Hashing

```env
ARGON2_MEMORY=65536               # Memory usage in KiB (64 MB)
ARGON2_ITERATIONS=3               # Number of iterations
ARGON2_PARALLELISM=4              # Parallel threads
ARGON2_SALT_LENGTH=16             # Salt length in bytes
ARGON2_KEY_LENGTH=32              # Hash output length in bytes
```

#### RabbitMQ

```env
RABBITMQ_URL=amqp://guest:guest@localhost:5672/       # RabbitMQ connection URL
MESSAGE_EXCHANGE_NAME=shortmesh.messages              # Message exchange name (default: shortmesh.messages)
MESSAGE_QUEUE_NAME=shortmesh-messages-queue           # Main message queue (default: shortmesh-messages-queue)
MESSAGE_DELAY_QUEUE_NAME=shortmesh-messages-delay-queue # Delay queue for throttled messages (default: shortmesh-messages-delay-queue)
```

## Commands

```bash
make run              # Start API server
make worker           # Start message worker (single instance)
make worker-n N=5     # Start message worker with N concurrent workers
make build            # Build binaries (api + migrate + worker)
make test             # Run tests
make docs             # Generate Swagger documentation

# Migrations
make migrate-up       # Run pending migrations
make migrate-down     # Rollback last migration
make migrate-status   # Show migration status
make migrate-fresh    # Drop all and recreate
```

## Worker Service

The worker service processes outbound messages from RabbitMQ using a shared queue pattern with delayed retries for rate-limited messages.

### Architecture

- **Main Queue**: All workers consume from a single shared queue for round-robin message distribution
- **Delay Queue**: Throttled messages are temporarily held with TTL, then automatically routed back to the main queue via Dead Letter Exchange (DLX)

### Starting the Worker

```bash
# Single worker
make worker

# Multiple concurrent workers
make worker-n N=10
```

### Message Format

Messages published to the exchange must follow this JSON structure:

```json
{
  "device_id": "device123",
  "contact": "+1234567890",
  "platform_name": "platform1",
  "text": "Message content",
  "username": "user1"
}
```

### Configuration

Queue names can be customized via environment variables (see RabbitMQ Configuration section above):

- `MESSAGE_QUEUE_NAME` - Main worker queue
- `MESSAGE_DELAY_QUEUE_NAME` - Delay queue for throttled messages

### Throttling Configuration

See [pkg/throttler/README.md](pkg/throttler/README.md) for throttler configuration and examples.

## API Documentation

Swagger UI is available at `http://localhost:8080/swagger/index.html` when the server is running.

To regenerate Swagger documentation after making changes to API endpoints:

```bash
make docs
```

## References

- [Architecture Documentation](docs/architecture/) - System architecture and sequence diagrams
- [Migration Guide](migrations/README.md) - Creating and managing database migrations
- [Throttler documentation](pkg/throttler/README.md) - Guide on message rate limiting implementation
