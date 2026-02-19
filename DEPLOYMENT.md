# Blytz Cloud - Production Deployment Guide

Complete guide for deploying Blytz Cloud on your infrastructure.

## 🏗️ Architecture Overview

```
Internet
    │
    ▼
┌────────────────────────────────────────────────────────────┐
│                        Caddy Server                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ blytz.cloud  │  │ app.blytz    │  │ *.blytz.cloud    │  │
│  │ (Landing)    │  │ (Dashboard)  │  │ (Tenant Agents)  │  │
│  │              │  │              │  │                  │  │
│  │ Next.js      │  │ Next.js      │  │ Docker           │  │
│  │ Port 3000    │  │ Port 3000    │  │ Containers       │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │   Go Backend API     │
                │   Port 8080          │
                │                      │
                │  • Signup            │
                │  • Webhooks          │
                │  • Status            │
                └──────────────────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │   SQLite Database    │
                └──────────────────────┘
```

## 📋 Prerequisites

### System Requirements
- **OS**: Ubuntu 22.04 LTS (recommended)
- **CPU**: 4+ cores (Ryzen 7 5700x perfect!)
- **RAM**: 16GB+ (32GB ideal for multi-tenancy)
- **Disk**: 100GB SSD
- **Network**: Public IP, ports 80/443 open

### Software
- Go 1.26+
- Node.js 18+
- Docker & Docker Compose
- Caddy
- Git

## 🚀 Quick Deploy (Local Testing)

Perfect for testing on your Ryzen 7 machine!

### 1. Clone and Setup

```bash
cd /opt
git clone https://github.com/gmsas95/blytz-cloud.git
cd blytz-cloud

# Copy environment template
cp .env.example .env
nano .env  # Edit with your API keys
```

### 2. Environment Configuration

Edit `.env`:

```bash
# Required API Keys
OPENAI_API_KEY=sk-your-openai-key
STRIPE_SECRET_KEY=sk_live_your-stripe-key
STRIPE_WEBHOOK_SECRET=whsec_your-webhook-secret
STRIPE_PRICE_ID=price_your-price-id

# Platform Configuration
DATABASE_PATH=/opt/blytz/data/database.sqlite
CUSTOMERS_DIR=/opt/blytz/customers
TEMPLATES_DIR=./internal/workspace/templates
MAX_CUSTOMERS=50
PORT_RANGE_START=30000
PORT_RANGE_END=30100
BASE_DOMAIN=blytz.cloud
PLATFORM_PORT=8080

# Caddy (optional for local testing)
CADDY_ADMIN_URL=http://localhost:2019
```

### 3. Build Go Backend

```bash
cd /opt/blytz-cloud

# Download dependencies
go mod download

# Build production binary
go build -ldflags="-s -w" -o blytz ./cmd/server

# Test build
./blytz --version
```

### 4. Build Frontend

```bash
cd /opt/blytz-cloud/frontend

# Install dependencies
npm install

# Build for production
npm run build

# Test production build
npm start
```

### 5. Setup Caddy (Subdomain Routing)

Install Caddy:

```bash
# Install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

Create Caddy configuration:

```bash
sudo mkdir -p /etc/caddy
sudo nano /etc/caddy/Caddyfile
```

Add this configuration:

```caddy
{
    # Enable admin API for dynamic subdomain management
    admin localhost:2019
    
    # Global options
    auto_https off  # Disable for local testing, enable for production
}

# Main website (Next.js frontend)
localhost:80 {
    reverse_proxy localhost:3000
    
    # Security headers
    header {
        X-Frame-Options DENY
        X-Content-Type-Options nosniff
        X-XSS-Protection "1; mode=block"
    }
}

# API endpoints (Go backend)
localhost:80/api/* {
    reverse_proxy localhost:8080
    
    # CORS headers for API
    header {
        Access-Control-Allow-Origin "*"
        Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
        Access-Control-Allow-Headers "Content-Type, Authorization"
    }
}

# Health check
localhost:80/api/health {
    reverse_proxy localhost:8080
}

# Wildcard subdomain for tenant agents
*.localhost:80 {
    # Extract subdomain
    @hasSubdomain expression `{http.request.host.labels.1} != "localhost"`
    
    reverse_proxy @hasSubdomain localhost:8080 {
        header_up Host {http.request.host}
    }
}
```

Start Caddy:

```bash
sudo systemctl start caddy
sudo systemctl enable caddy

# Check status
sudo systemctl status caddy
```

### 6. Create Systemd Service

Create service file:

```bash
sudo nano /etc/systemd/system/blytz.service
```

Add:

```ini
[Unit]
Description=Blytz Cloud Platform
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/opt/blytz-cloud
ExecStart=/opt/blytz-cloud/blytz
Restart=always
RestartSec=5
Environment="PATH=/usr/local/bin:/usr/bin:/bin"

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable blytz
sudo systemctl start blytz

# Check status
sudo systemctl status blytz
sudo journalctl -u blytz -f
```

### 7. Create Frontend Service

```bash
sudo nano /etc/systemd/system/blytz-frontend.service
```

Add:

```ini
[Unit]
Description=Blytz Cloud Frontend
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/opt/blytz-cloud/frontend
ExecStart=/usr/bin/npm start
Restart=always
RestartSec=5
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
Environment="NODE_ENV=production"
Environment="PORT=3000"

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable blytz-frontend
sudo systemctl start blytz-frontend
```

## 🔧 Production Deployment (With Domain)

### 1. DNS Configuration

Set up DNS records for your domain:

```
Type    Name            Value                TTL
A       @               YOUR_SERVER_IP       300
A       *               YOUR_SERVER_IP       300
A       app             YOUR_SERVER_IP       300
```

### 2. Production Caddyfile

```caddy
{
    admin localhost:2019
    email your-email@example.com  # For Let's Encrypt
}

# Main domain → Next.js frontend
blytz.cloud {
    reverse_proxy localhost:3000
    
    header {
        X-Frame-Options DENY
        X-Content-Type-Options nosniff
        X-XSS-Protection "1; mode=block"
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
    }
}

# Dashboard subdomain
app.blytz.cloud {
    reverse_proxy localhost:3000
    
    header {
        X-Frame-Options DENY
        X-Content-Type-Options nosniff
        X-XSS-Protection "1; mode=block"
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
    }
}

# API endpoints
blytz.cloud/api/* {
    reverse_proxy localhost:8080
    
    header {
        Access-Control-Allow-Origin "https://blytz.cloud https://app.blytz.cloud"
        Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
        Access-Control-Allow-Headers "Content-Type, Authorization"
    }
}

# Wildcard subdomains for tenant agents
*.blytz.cloud {
    # Rate limiting
    rate_limit {
        zone static_limit {
            key static
            events 100
            window 1m
        }
    }
    
    reverse_proxy localhost:8080 {
        header_up Host {http.request.host}
        header_up X-Real-IP {http.request.remote}
        header_up X-Forwarded-For {http.request.remote}
        header_up X-Forwarded-Proto {http.request.scheme}
    }
    
    header {
        X-Frame-Options DENY
        X-Content-Type-Options nosniff
        X-XSS-Protection "1; mode=block"
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
    }
}
```

### 3. SSL Certificates

Caddy automatically handles SSL with Let's Encrypt. Just ensure:

- Port 80 and 443 are open
- DNS is properly configured
- Email is set in Caddyfile

### 4. Security Hardening

```bash
# Setup firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# Create dedicated user (optional but recommended)
sudo useradd -r -s /bin/false blytz
sudo chown -R blytz:blytz /opt/blytz-cloud

# Update systemd services to use blytz user
```

## 📊 Monitoring & Logs

### View Logs

```bash
# Backend logs
sudo journalctl -u blytz -f

# Frontend logs
sudo journalctl -u blytz-frontend -f

# Caddy logs
sudo journalctl -u caddy -f

# All logs
sudo tail -f /var/log/syslog | grep -E "(blytz|caddy)"
```

### Health Checks

```bash
# Check backend health
curl http://localhost:8080/api/health

# Check Caddy admin
curl http://localhost:2019/config/

# Test subdomain routing
curl -H "Host: test.blytz.cloud" http://localhost/
```

## 🔄 Updates

### Update Code

```bash
cd /opt/blytz-cloud
git pull origin main

# Rebuild backend
go build -ldflags="-s -w" -o blytz ./cmd/server
sudo systemctl restart blytz

# Rebuild frontend
cd frontend
npm install
npm run build
sudo systemctl restart blytz-frontend

# Reload Caddy
sudo systemctl reload caddy
```

## 🆘 Troubleshooting

### Issue: Services won't start

```bash
# Check for errors
sudo journalctl -u blytz --no-pager | tail -50

# Check port conflicts
sudo lsof -i :8080
sudo lsof -i :3000

# Check permissions
ls -la /opt/blytz-cloud
```

### Issue: Subdomains not working

```bash
# Test Caddy config
sudo caddy validate --config /etc/caddy/Caddyfile

# Check Caddy logs
sudo journalctl -u caddy -f

# Test manually
curl -H "Host: demo.blytz.cloud" http://localhost/api/health
```

### Issue: Frontend can't connect to backend

```bash
# Verify backend is running
curl http://localhost:8080/api/health

# Check Next.js proxy config
cat /opt/blytz-cloud/frontend/next.config.ts

# Test from frontend container
curl http://localhost:3000/api/health
```

## 📈 Scaling

### Vertical Scaling (More Power)
- Upgrade CPU/RAM
- Increase `MAX_CUSTOMERS` in `.env`
- Expand `PORT_RANGE_END`

### Horizontal Scaling (Multiple Servers)
1. Use PostgreSQL instead of SQLite
2. Use Redis for port allocation
3. Use shared storage for customer data
4. Use load balancer

## 🔒 Security Checklist

- [ ] Change default ports
- [ ] Enable firewall
- [ ] Use strong API keys
- [ ] Enable SSL (Caddy auto-handles)
- [ ] Regular security updates
- [ ] Backup database
- [ ] Monitor logs for anomalies
- [ ] Rate limiting enabled
- [ ] Input validation working

## 📞 Support

For issues:
1. Check logs: `sudo journalctl -u blytz -f`
2. Verify config: `cat /opt/blytz-cloud/.env`
3. Test endpoints: `curl http://localhost:8080/api/health`
4. Review docs: `cat /opt/blytz-cloud/README.md`

---

**Your Ryzen 7 5700x with 32GB RAM can easily handle 50+ concurrent tenants!** 🚀
