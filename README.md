# Blytz Cloud

Platform for deploying personalized OpenClaw AI assistants. Users sign up, configure their assistant, pay $29/month, and get a working AI assistant within 2 minutes via Telegram.

## Features

- 🤖 **AI Assistant Deployment** - Automated provisioning of OpenClaw instances
- 💳 **Stripe Integration** - Secure payment processing with subscriptions
- 🐳 **Docker-based** - Each customer gets isolated container
- 🌐 **Custom Subdomains** - Automatic Caddy reverse proxy configuration
- 📱 **Telegram Integration** - Bot token validation and messaging
- 🔄 **Lifecycle Management** - Suspend, resume, and terminate assistants
- 📊 **Monitoring** - Health checks and status tracking

## Quick Start

### Prerequisites

- Go 1.26+
- Docker
- SQLite3
- Caddy (for production)

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

### Production Deployment

```bash
# Install systemd service
sudo ./deployments/install.sh

# Configure environment
sudo nano /opt/blytz/.env

# Start service
sudo systemctl start blytz
sudo systemctl enable blytz
```

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│   User Browser  │────▶│  Blytz API   │────▶│   Stripe    │
└─────────────────┘     └──────┬───────┘     └─────────────┘
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

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Landing page |
| GET | `/configure` | Assistant configuration |
| POST | `/api/signup` | Create customer account |
| GET | `/api/status/:id` | Get customer status |
| POST | `/api/webhook/stripe` | Stripe webhooks |
| GET | `/api/health` | Health check |

## Environment Variables

```bash
# Required
OPENAI_API_KEY=sk-...
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID=price_...

# Optional
PORT=8080
MAX_CUSTOMERS=20
PORT_RANGE_START=30000
PORT_RANGE_END=30999
BASE_DOMAIN=blytz.cloud
DATABASE_PATH=./database.sqlite
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run E2E tests
RUN_E2E=true go test ./internal/e2e -tags=e2e
```

## Project Structure

```
blytz-cloud/
├── cmd/server/          # Entry point
├── internal/
│   ├── api/            # HTTP handlers
│   ├── config/         # Configuration
│   ├── db/             # Database operations
│   ├── provisioner/    # Docker management
│   ├── workspace/      # File generation
│   ├── telegram/       # Bot validation
│   ├── stripe/         # Payment processing
│   └── caddy/          # Reverse proxy
├── deployments/        # Systemd scripts
└── static/            # HTML templates
```

## License

MIT

## Author

Built by [gmsas95](https://github.com/gmsas95)
