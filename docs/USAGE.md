# API Usage Guide

## Prerequisites

```bash
make setup && make migrate-up && make run
```

Get credentials from `.env`:

```bash
CLIENT_ID=$(grep '^CLIENT_ID=' .env | cut -d'=' -f2)
CLIENT_SECRET=$(grep '^CLIENT_SECRET=' .env | cut -d'=' -f2)
```

## Authentication

- **Token Creation** - Basic Auth with `CLIENT_ID:CLIENT_SECRET`
- **Device Operations** - Bearer Auth with Matrix token (`mt_xxxxx`)

## Token Management

### Create First Token (Admin)

First token is auto-marked as admin and creates the host Matrix identity:

```bash
curl -X POST http://localhost:8080/api/v1/tokens \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"use_host": false}'
```

**Response:**

```json
{
  "message": "Matrix token created successfully",
  "token": "mt_abc123..."
}
```

### Create Token

#### Use Host Identity (`use_host: true`)

Shares the admin's Matrix credentials and any linked devices:

```bash
curl -X POST http://localhost:8080/api/v1/tokens \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"use_host": true}'
```

#### New Matrix User (`use_host: false`)

Creates new Matrix credentials:

```bash
curl -X POST http://localhost:8080/api/v1/tokens \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"use_host": false, "expires_at": "2026-12-31T23:59:59Z"}'
```

> [!NOTE]
>
> - `use_host: true`: Token can access admin's linked devices
> - `use_host: false`: Token has its own empty device list, must link devices separately

## Device Management

Set your token:

```bash
TOKEN="mt_abc123..."
```

### Add Device

Requests device addition and returns QR code WebSocket URL:

```bash
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"platform": "wa"}'
```

**Response:**

```json
{
  "message": "Scan the QR code to link your device",
  "qr_code_url": "ws://localhost:8080/api/v1/devices/qr-code?token=mt_abc123..."
}
```

### Get QR Code (WebSocket)

Connect to WebSocket to receive QR code for device linking:

```bash
websocat "ws://localhost:8080/api/v1/devices/qr-code?token=$TOKEN"
```

### List Devices

```bash
curl -X GET http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**

```json
[
  {
    "platform": "wa",
    "device_id": "237123456789"
  }
]
```

### Send Message

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contact": "1234567890",
    "platform": "wa",
    "text": "Hello from API"
  }'
```

### Delete Device

```bash
curl -X DELETE http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "237123456789",
    "platform": "wa"
  }'
```

## API Reference

Swagger UI: **<http://localhost:8080/docs/index.html>**
