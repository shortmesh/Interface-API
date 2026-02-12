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
JWT_SECRET=<base64-key-here>      # For JWT token signing
```

**Important:** All three keys must be base64-encoded 32-byte keys. Generate three separate keys.

#### Session Management

```env
MAX_SESSIONS_PER_USER=1           # Maximum concurrent sessions per user
SESSION_DURATION_HOURS=720        # Session duration (default: 30 days)
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

#### Argon2id Password Hashing

```env
ARGON2_MEMORY=65536               # Memory usage in KiB (64 MB)
ARGON2_ITERATIONS=3               # Number of iterations
ARGON2_PARALLELISM=4              # Parallel threads
ARGON2_SALT_LENGTH=16             # Salt length in bytes
ARGON2_KEY_LENGTH=32              # Hash output length in bytes
```

## Commands

```bash
make run              # Start server
make build            # Build binaries (api + migrate)
make test             # Run tests
make swagger          # Generate Swagger documentation

# Migrations
make migrate-up       # Run pending migrations
make migrate-down     # Rollback last migration
make migrate-status   # Show migration status
make migrate-fresh    # Drop all and recreate
```

## API Documentation

Swagger UI is available at `http://localhost:8080/swagger/index.html` when the server is running.

To regenerate Swagger documentation after making changes to API endpoints:

```bash
make swagger
```

## References

- [Migration Guide](migrations/README.md) - Creating and managing database migrations
