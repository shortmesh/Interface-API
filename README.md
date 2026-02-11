# Interface API

The primary interface for user interaction.

## Requirements

- Go 1.24.0 or higher
- MySQL (optional - SQLite is used by default)

## Quick Start

1. Clone the repository
2. Copy `.env.example` to `.env` and configure as needed
3. Run the application:

   ```bash
   make run
   ```

## Configuration

Create a `.env` file in the root directory. The application supports two database modes:

**SQLite (Default):**

```env
PORT=8080
SQLITE_DB_PATH=shortmesh.db
```

**MySQL:**

```env
PORT=8080
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DATABASE=shortmesh
MYSQL_USERNAME=root
MYSQL_PASSWORD=yourpassword
```

> Note: If `SQLITE_DB_PATH` is set, SQLite will be used. Otherwise, MySQL configuration is required.

## Running

**Development:**

```bash
make run
```

**Build:**

```bash
make build
./main
```

**Tests:**

```bash
make test
```
