---
type: readme
title: Blytz Cloud
resource: blytz-cloud
description: "Platform for deploying personalized OpenClaw AI assistants. Users sign up, configure their assistant, pay $29/month, and get a working AI assistant within 2 minutes via Telegram."
tags: [docker, go, stripe]
updated: 2026-06-18
---

# Blytz Cloud

Platform for deploying personalized OpenClaw AI assistants. Users sign up, configure their assistant, pay $29/month, and get a working AI assistant within 2 minutes via Telegram.

**Status:** Production Ready ✅  
**Version:** 1.0.0  
**Last Updated:** February 19, 2026

## ✨ Features

- 🤖 **AI Assistant Deployment** - Automated provisioning of OpenClaw instances
- 💳 **Stripe Integration** - Secure payment processing with subscriptions
- 🐳 **Docker-based** - Each customer gets isolated container
- 🌐 **Custom Subdomains** - Automatic Caddy reverse proxy configuration
- 📱 **Telegram Integration** - Bot token validation and messaging
- 🔄 **Lifecycle Management** - Suspend, resume, and terminate assistants
- 📊 **Monitoring** - Health checks and structured logging
- 🔒 **Security** - Rate limiting, input sanitization, Docker secrets

## 🚀 Quick Start

### Prerequisites

- Go 1.26+
- Docker & Docker Compose
- SQLite3
- Caddy (for production with subdomains)

### Installation

```bash
# Clone the repository
git clone https://github.com/gmsas95/blytz-cloud.git
cd blytz-cloud

# Copy environment template
cp .env.example .env
# Edit .env with your API keys

# Build
go build -o blytz ./cmd/server

# Run
./blytz
```

The server will start on port 8080 by default.

### Production Deployment

```bash
# Install systemd service
sudo ./deployments/install.sh

# Configure environment
sudo nano /opt/blytz/.env

# Start service
sudo systemctl start blytz
sudo systemctl enable blytz

# Check status
sudo systemctl status blytz
sudo journalctl -u blytz -f
```

## 🏗️ Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│   User Browser  │────▶│  Blytz API   │────▶│   Stripe    │
└─────────────────┘     └──────┬───────┘     └─────────────┘
                               │
                               ▼
                        ┌──────────────┐
                        │   SQLite DB  │
                        └──────┬───────┘
                               │
                               ▼
                        ┌──────────────┐
                        │   Docker     │
                        │  Containers  │
                        └──────┬───────┘
                               │
                               ▼
                        ┌──────────────┐
                        │    Caddy     │
                        │  Reverse Proxy│
                        └──────────────┘
```

## 📡 API Endpoints

| Method | Path | Description | Rate Limit |
|--------|------|-------------|------------|
| GET | `/` | Landing page | None |
| GET | `/configure` | Assistant configuration | None |
| GET | `/success` | Success page | None |
| POST | `/api/signup` | Create customer account | 5/min |
| GET | `/api/status/:id` | Get customer status | None |
| POST | `/api/webhook/stripe` | Stripe webhooks | 100/min |
| GET | `/api/health` | Health check | None |

### Health Check Response

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "checks": {
    "database": {
      "status": "pass"
    },
    "docker": {
      "status": "pass"
    }
  },
  "timestamp": "2026-02-19T14:50:24.964Z"
}
```

## 🔧 Environment Variables

### Required

```bash
OPENAI_API_KEY=sk-...
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID=price_...
```

### Optional

```bash
# Server
PORT=8080
BASE_DOMAIN=blytz.cloud
PLATFORM_PORT=8080

# Database
DATABASE_PATH=./tmp/platform/database.sqlite

# Customer Limits
MAX_CUSTOMERS=20
PORT_RANGE_START=30000
PORT_RANGE_END=30999

# Paths
CUSTOMERS_DIR=./tmp/customers
TEMPLATES_DIR=./internal/workspace/templates

# Caddy (for production)
CADDY_ADMIN_URL=http://localhost:2019

# Security
OPENCLAW_GATEWAY_TOKEN_PREFIX=blytz_
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/api -v
go test ./internal/db -v
go test ./internal/provisioner -v

# Run E2E tests (requires Docker & Stripe credentials)
RUN_E2E=true go test ./internal/e2e -tags=e2e

# Run with race detector
go test -race ./...
```

## 📁 Project Structure

```
blytz-cloud/
├── cmd/
│   └── server/            # Application entry point
├── internal/
│   ├── api/               # HTTP handlers, middleware, routes
│   │   ├── handler.go
│   │   ├── router.go
│   │   ├── ratelimit.go   # Rate limiting middleware
│   │   └── *_test.go
│   ├── config/            # Configuration loading
│   ├── db/                # Database operations & migrations
│   ├── provisioner/       # Docker lifecycle management
│   │   ├── service.go
│   │   ├── compose.go     # Docker Compose generation
│   │   ├── ports.go       # Thread-safe port allocation
│   │   └── *_test.go
│   ├── workspace/         # File generation (AGENTS.md, etc.)
│   ├── telegram/          # Bot token validation
│   ├── stripe/            # Payment processing & webhooks
│   ├── caddy/             # Reverse proxy management
│   └── e2e/               # End-to-end tests
├── deployments/           # Systemd service files
├── static/                # HTML templates (embedded)
└── tmp/                   # Customer data directory
```

## 🔐 Security Features

- **Docker Secrets** - API keys stored in `.env.secret` files with 0600 permissions
- **Rate Limiting** - Prevents abuse (5 req/min for signup, 100 req/min for webhooks)
- **Input Sanitization** - Customer IDs sanitized to prevent directory traversal
- **Thread-Safe Operations** - Port allocation protected by mutex
- **Structured Logging** - JSON logs with Zap (no sensitive data)
- **SQL Injection Prevention** - All queries use prepared statements

## 📊 Monitoring & Observability

### Structured Logging

The application uses Zap for structured JSON logging:

```json
{
  "level": "info",
  "ts": 1234567890.123,
  "msg": "Server starting",
  "port": "8080"
}
```

### Request Logging

All requests are logged with method, path, and status:

```json
{
  "level": "info",
  "ts": 1234567890.123,
  "msg": "Request",
  "method": "POST",
  "path": "/api/signup",
  "status": 201
}
```

### Health Monitoring

Health check endpoint (`/api/health`) monitors:
- Database connectivity
- Docker availability
- Overall system status

## 🔄 Customer Lifecycle

1. **Sign Up** - User submits email, assistant config, Telegram token
2. **Validation** - System validates Telegram bot token
3. **Payment** - Stripe checkout session created
4. **Provisioning** - Webhook triggers container deployment
5. **Active** - Assistant running on assigned subdomain
6. **Management** - Can be suspended/resumed via Stripe events
7. **Termination** - Subscription cancellation removes all resources

## 🛠️ Development

### Building

```bash
# Development build
go build -o blytz ./cmd/server

# Production build
go build -ldflags="-s -w" -o blytz ./cmd/server

# Build with race detector
go build -race -o blytz ./cmd/server
```

### Code Quality

```bash
# Run linter
golangci-lint run ./...

# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

## 📚 Documentation

- [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - Implementation details
- [AGENTS.md](./AGENTS.md) - Development guidelines
- [KNOWLEDGE_BASE.md](./KNOWLEDGE_BASE.md) - Project knowledge

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](./LICENSE) for details

## 👤 Author

Built by [gmsas95](https://github.com/gmsas95)

---

**Made with ❤️ for the AI assistant community**
