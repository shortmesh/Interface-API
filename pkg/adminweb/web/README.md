# ShortMesh Admin - React Vite App

This is the admin interface for ShortMesh, built with React, Vite, and Material-UI (MUI).

## Development

1. Install dependencies:
   ```bash
   make install
   # or
   yarn install
   ```

2. Start the development server:
   ```bash
   make dev
   # or
   yarn dev
   ```

   The dev server will proxy API requests to `http://localhost:8080`.

3. Build for production:
   ```bash
   make build
   # or
   yarn build
   ```

   The build output will be in the `dist/` directory, which is embedded into the Go binary.

## Project Structure

```
web/
├── src/
│   ├── main.jsx          # Application entry point
│   ├── App.jsx           # Main app with routing
│   ├── theme.js          # MUI theme configuration
│   ├── components/
│   │   └── Layout.jsx    # Main layout with sidebar
│   ├── pages/
│   │   ├── Login.jsx     # Login page
│   │   ├── Tokens.jsx    # Matrix tokens management
│   │   ├── Devices.jsx   # Device management with QR codes
│   │   └── Webhooks.jsx  # Webhook configuration
│   └── utils/
│       └── api.js        # API utility functions
├── index.html            # HTML entry point
├── vite.config.js        # Vite configuration
├── package.json          # Dependencies
└── dist/                 # Build output (embedded in Go binary)
```

## Features

- **Dark Mode UI**: Built with Material-UI dark theme
- **Responsive Design**: Works on mobile and desktop
- **Token Management**: Create and manage Matrix identity tokens
- **Device Management**: Add devices with QR code scanning via WebSocket
- **Webhook Management**: Configure webhook endpoints
- **Real-time Updates**: WebSocket support for device pairing

## Building for Production

The production build is automatically embedded in the Go binary via `go:embed`. Make sure to build the React app before building the Go application:

```bash
cd pkg/adminweb/web
make build
cd ../../..
go build ./cmd/api
```
