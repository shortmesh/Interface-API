# Interface API

The primary interface for user interaction.

## Table of Contents

- [Quick Start](#quick-start)
- [Requirements](#requirements)
- [Configuration](#configuration)
- [Development](#development)
- [API Documentation](#api-documentation)
- [Resources](#resources)

## Quick Start

```bash
git clone https://github.com/shortmesh/Interface-API.git
cd Interface-API
make setup
make migrate-up
make run
```

Server: `http://localhost:8080`

## Requirements

- Go 1.24.0+
- SQLite (default) or MySQL
- RabbitMQ (for worker service)

## Configuration

> [!NOTE]
>
> `.env.default` contains operational default values. Only modify if you know what you're doing.

Copy `.env.example` to `.env` and configure as needed:

```bash
cp .env.example .env
# Or use: make setup (auto-generates keys)
```

### Server Configuration

- `HOST` - Host address the server will bind to (default: `127.0.0.1`)
- `PORT` - Port the server will listen on (default: `8080`)

### Required Environment Variables

The following environment variables **must** be set for the application to function properly:

#### Cryptographic Keys

> [!NOTE]
>
> If you already used `make setup`, these keys are auto-generated and set in your `.env` file. Do not change them unless you know what you're doing, as changing these keys will invalidate existing data.

- `ENCRYPTION_KEY` - Base64-encoded 32-byte key for encrypting sensitive data
- `HASH_KEY` - Base64-encoded 32-byte key for hashing data

Generate with: `openssl rand -base64 32`

#### Matrix Services

- `MAS_ADMIN_URL` - MAS admin API URL
- `ADMIN_CLIENT_ID` - Admin client ID for MAS
- `ADMIN_CLIENT_SECRET` - Admin client secret for MAS
- `MATRIX_CLIENT_URL` - Matrix client URL

See `.env.example` for all available options.

> [!WARNING]
>
> **Production:** Set `AUTO_MIGRATE=false` and `AUTO_CREATE_TABLES=false`, then run `make migrate-up`.

## Development

### Commands

```bash
make setup            # Setup .env with auto-generated keys
make run              # Start API server
make worker           # Start message worker
make build            # Build binaries
make test             # Run tests
make docs             # Generate Swagger docs
```

### Migrations

```bash
make migrate-up       # Run pending
make migrate-down     # Rollback last
make migrate-status   # Show status
```

See [Migration Guide](docs/MIGRATIONS.md) for details.

## API Documentation

Swagger UI: `http://localhost:8080/docs/index.html`

Regenerate: `make docs`

## Resources

- [Migration Guide](docs/MIGRATIONS.md)
- [Throttler Documentation](docs/THROTTLER.md)
- [QR Worker Documentation](docs/QR_WORKER.md)
- [Architecture Documentation](docs/architecture/)
