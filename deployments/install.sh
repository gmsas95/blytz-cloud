#!/bin/bash
set -e

echo "Blytz Installation Script"
echo "========================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (use sudo)"
    exit 1
fi

# Configuration
BLYTZ_USER="blytz"
BLYTZ_DIR="/opt/blytz"

# Create blytz user if doesn't exist
if ! id "$BLYTZ_USER" &> /dev/null; then
    echo "Creating $BLYTZ_USER user..."
    useradd -r -s /bin/false -M "$BLYTZ_USER"
fi

# Add blytz user to docker group
usermod -aG docker "$BLYTZ_USER" 2>/dev/null || true

# Create directories
echo "Creating directories..."
mkdir -p "$BLYTZ_DIR"/{platform,customers,caddy,logs}
chown -R "$BLYTZ_USER:$BLYTZ_USER" "$BLYTZ_DIR"

# Check for binary
if [ ! -f "$BLYTZ_DIR/blytz" ]; then
    echo "Building blytz binary..."
    if command -v go > /dev/null; then
        cd /opt/blytz
        go build -o blytz ./cmd/server
    else
        echo "Error: Go not found. Please build the binary manually."
        echo "Run: go build -o blytz ./cmd/server"
        exit 1
    fi
fi

chmod +x "$BLYTZ_DIR/blytz"

# Create .env template if doesn't exist
if [ ! -f "$BLYTZ_DIR/.env" ]; then
    echo "Creating .env file..."
    cat > "$BLYTZ_DIR/.env" << 'EOF'
# API Keys (required)
OPENAI_API_KEY=your-openai-key-here
STRIPE_SECRET_KEY=your-stripe-key-here
STRIPE_WEBHOOK_SECRET=your-webhook-secret-here
STRIPE_PRICE_ID=your-price-id-here

# Platform Configuration
DATABASE_PATH=/opt/blytz/platform/database.sqlite
CUSTOMERS_DIR=/opt/blytz/customers
TEMPLATES_DIR=/opt/blytz/internal/workspace/templates
MAX_CUSTOMERS=20
PORT_RANGE_START=30000
PORT_RANGE_END=30999
BASE_DOMAIN=blytz.cloud
PLATFORM_PORT=8080
EOF
    chown "$BLYTZ_USER:$BLYTZ_USER" "$BLYTZ_DIR/.env"
    echo ""
    echo "⚠️  IMPORTANT: Please edit $BLYTZ_DIR/.env with your actual API keys"
    echo ""
fi

# Install systemd service
echo "Installing systemd service..."
cp deployments/blytz.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable blytz

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Edit $BLYTZ_DIR/.env with your API keys"
echo "2. Start the service: sudo systemctl start blytz"
echo "3. Check status: sudo systemctl status blytz"
echo "4. View logs: sudo journalctl -u blytz -f"
