# Security Configuration

## Modes

- **development** (default): HTTP allowed, no TLS required
- **production**: HTTPS enforced for server and external services

Set via `APP_MODE=development` or `APP_MODE=production`

## Production Requirements

### Server TLS

Required unless behind a reverse proxy:

```bash
APP_MODE=production
TLS_CERT_FILE=/path/to/cert.crt
TLS_KEY_FILE=/path/to/key.key
```

If using reverse proxy for TLS termination:

```bash
APP_MODE=production
ALLOW_INSECURE_SERVER=true
```

### External Services

Must use HTTPS/WSS/AMQPS in production:

```bash
APP_MODE=production
MAS_URL=https://mas.example.com
MAS_ADMIN_URL=https://mas-admin.example.com
MATRIX_CLIENT_URL=https://matrix.example.com
RABBITMQ_URL=amqps://user:pass@rabbitmq.example.com:5671/
```

To allow HTTP/WS/AMQP (not recommended):

```bash
ALLOW_INSECURE_EXTERNAL=true
```

## Configuration Examples

### Development

```bash
APP_MODE=development
HOST=127.0.0.1
PORT=8080
MAS_URL=http://localhost:8000
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

### Production with TLS

```bash
APP_MODE=production
HOST=0.0.0.0
PORT=8443
TLS_CERT_FILE=/etc/ssl/certs/api.crt
TLS_KEY_FILE=/etc/ssl/private/api.key
MAS_URL=https://mas.example.com
RABBITMQ_URL=amqps://user:pass@rabbitmq.example.com:5671/
```

### Production Behind Reverse Proxy

```bash
APP_MODE=production
HOST=127.0.0.1
PORT=8080
ALLOW_INSECURE_SERVER=true
MAS_URL=https://mas.example.com
RABBITMQ_URL=amqps://user:pass@rabbitmq.example.com:5671/
```

## Common Errors

**"TLS_CERT_FILE and TLS_KEY_FILE must be set"**

- Set cert/key paths or use `ALLOW_INSECURE_SERVER=true` with reverse proxy

**"production mode requires HTTPS/WSS/AMQPS for external service"**

- Update URLs to secure protocols or set `ALLOW_INSECURE_EXTERNAL=true`
