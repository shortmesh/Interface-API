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

See [SECURITY.md](./SECURITY.md) for production configuration.

## Authentication

- **Token/Credential Management** - Basic Auth with `CLIENT_ID:CLIENT_SECRET`
- **Device/Webhook Operations** - Bearer Auth with Matrix token (`mt_xxxxx`)

## Token Management

Tokens are Matrix authentication tokens that provide access to device and webhook operations. Each token is associated with the credential that created it.

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
> - Tokens are automatically associated with the credential used to create them

### List Tokens

View all tokens associated with your credential:

```bash
curl -X GET http://localhost:8080/api/v1/tokens \
  -u "$CLIENT_ID:$CLIENT_SECRET"
```

Super admin credentials can see all tokens they created plus legacy tokens (those without a credential_id).

### Delete Token

```bash
curl -X DELETE http://localhost:8080/api/v1/tokens/{token_id} \
  -u "$CLIENT_ID:$CLIENT_SECRET"
```

Replace `{token_id}` with the token ID from the List Tokens response.

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

Send messages to individual contacts or groups. At least one of `contact` or `group_url` must be provided.

#### Text Only (JSON)

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

#### Multiple Recipients (Comma-Separated)

Send the same message to multiple addresses by providing a comma-separated list:

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contact": "1234567890, 0987654321, another_address",
    "platform": "wa",
    "text": "Hello to multiple recipients"
  }'
```

#### Send to Group (using group_url)

If you don't have the group ID, use the group invite URL. You can also provide multiple group URLs comma-separated:

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "group_url": "https://chat.whatsapp.com/invite123",
    "platform": "wa",
    "text": "Hello group!"
  }'
```

Multiple groups:

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "group_url": "https://chat.whatsapp.com/invite123, https://chat.whatsapp.com/invite456",
    "platform": "wa",
    "text": "Hello to multiple groups!"
  }'
```

#### Reply to a Message

Use `reply_id` to reply to a specific message (works with both contacts and groups):

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contact": "1234567890",
    "platform": "wa",
    "text": "This is a reply",
    "reply_id": "$eCLf_5pzs8ZBhEs5ZojVRS0K3SBLf73cHs2q_Je4kfQ"
  }'
```

Or reply in a group:

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "group_url": "https://chat.whatsapp.com/invite123",
    "platform": "wa",
    "text": "This is a reply in group",
    "reply_id": "$eCLf_5pzs8ZBhEs5ZojVRS0K3SBLf73cHs2q_Je4kfQ"
  }'
```

#### Text + File (Multipart)

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -F "contact=1234567890" \
  -F "platform=wa" \
  -F "text=Check out this document" \
  -F "file=@/path/to/document.pdf"
```

For multiple recipients with files, use comma-separated addresses:

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -F "contact=1234567890, 0987654321" \
  -F "platform=wa" \
  -F "text=Check out this document" \
  -F "file=@/path/to/document.pdf"
```

Send file to a group:

```bash
curl -X POST http://localhost:8080/api/v1/devices/237123456789/message \
  -H "Authorization: Bearer $TOKEN" \
  -F "group_url=https://chat.whatsapp.com/invite123" \
  -F "platform=wa" \
  -F "text=Check out this document" \
  -F "file=@/path/to/document.pdf"
```

File must have an extension.

> [!NOTE]
>
> - If both `contact` and `group_url` are provided, `group_url` takes priority
> - The `reply_id` can be obtained from webhook payloads for incoming messages
> - Both `contact` and `group_url` support comma-separated values for multiple recipients

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

## Webhooks

Receive incoming message notifications via HTTP POST.

> [!IMPORTANT]
> Webhooks are tied to tokens. Every device stored under a token will publish incoming messages to that token's configured webhook. If multiple devices share the same token, all their incoming messages will be delivered to the same webhook URL.

### Add Webhook

```bash
curl -X POST http://localhost:8080/api/v1/webhooks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://your-server.com/webhook"}'
```

**Response:**

```json
{
  "id": 1,
  "url": "https://your-server.com/webhook",
  "active": true,
  "created_at": "2026-04-27T10:00:00Z",
  "updated_at": "2026-04-27T10:00:00Z"
}
```

### Webhook Payload Structure

When an incoming message is received, your webhook URL will receive a POST request with the following JSON structure:

```json
{
  "id": "$A-XXXXXXXXXXXXXXXXXXXXXXX",
  "is_contact": false,
  "type": "m.text",
  "from": "@whatsapp_2376987654321:matrix.example.com",
  "to": "@xxxxxxxxxxxx:matrix.example.com",
  "message": "Good morning",
  "device_id": "237123456789",
  "media": {
    "content": null,
    "info": {
      "size": 0,
      "mime_type": "",
      "width": 0,
      "height": 0,
      "blur_hash": ""
    }
  }
}
```

**Field descriptions:**

- `id`: Unique Matrix event ID for the message
- `is_contact`: `true` if message is from a saved contact, `false` if from unknown sender
- `type`: Matrix message type (e.g., "m.text", "m.image", "m.file", "m.video")
- `from`: Sender's Matrix user ID or contact name
- `to`: Recipient's Matrix user ID
- `message`: Text content of the message (empty for media-only messages)
- `device_id`: Device identifier that received the message
- `media`: Media object containing binary data and metadata (null if no media)
  - `content`: Base64 encoded or raw binary media data
  - `info`: Media metadata
    - `size`: File size in bytes
    - `mime_type`: MIME type of the media
    - `width`: Image/video width (0 if not applicable)
    - `height`: Image/video height (0 if not applicable)
    - `blur_hash`: BlurHash string for image preview (used by some clients)

### List Webhooks

```bash
curl -X GET http://localhost:8080/api/v1/webhooks \
  -H "Authorization: Bearer $TOKEN"
```

### Update Webhook

```bash
curl -X PUT http://localhost:8080/api/v1/webhooks/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://new-url.com/webhook", "active": false}'
```

Fields are optional. Omit to keep current value.

### Delete Webhook

```bash
curl -X DELETE http://localhost:8080/api/v1/webhooks/1 \
  -H "Authorization: Bearer $TOKEN"
```

## Credentials (Admin Only)

Manage API client credentials. Requires Basic Auth with admin credentials. Each credential can create and manage its own tokens.

### Create Credential

```bash
curl -X POST http://localhost:8080/api/v1/credentials \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "my-app",
    "description": "Production API client"
  }'
```

**Response:**

```json
{
  "message": "Credential created successfully",
  "credential": {
    "client_id": "my-app",
    "role": "user",
    "scopes": [],
    "description": "Production API client",
    "active": true,
    "created_at": "2026-04-27T10:00:00Z",
    "updated_at": "2026-04-27T10:00:00Z"
  },
  "client_secret": "xxxxxxxxxxxx"
}
```

> [!IMPORTANT]
> Save `client_secret` immediately. It's only shown once.

### List Credentials

```bash
curl -X GET http://localhost:8080/api/v1/credentials \
  -u "$CLIENT_ID:$CLIENT_SECRET"
```

### Update Credential

Update credential properties like description, active status, or scopes:

```bash
curl -X PUT http://localhost:8080/api/v1/credentials/my-app \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated description",
    "active": true,
    "scopes": ["tokens:read:*", "tokens:write:create", "tokens:write:delete"]
  }'
```

Fields are optional. Omit to keep current value.

### Delete Credential

```bash
curl -X DELETE http://localhost:8080/api/v1/credentials/my-app \
  -u "$CLIENT_ID:$CLIENT_SECRET"
```

Cannot delete super admin credentials.

> [!NOTE]
>
> - Each credential can create and manage its own tokens via the `/api/v1/tokens` endpoints
> - Tokens are automatically associated with the credential that created them
> - Deleting a credential will affect all tokens created by that credential

## API Reference

Swagger UI: **<http://localhost:8080/docs/index.html>**
