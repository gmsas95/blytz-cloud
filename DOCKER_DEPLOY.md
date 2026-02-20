# 🐳 Docker Deployment Guide

Deploy Blytz Cloud using Docker Compose for easy replication and scaling.

## 🚀 Quick Start (3 Commands)

```bash
# 1. Clone and enter directory
git clone https://github.com/gmsas95/blytz-cloud.git
cd blytz-cloud

# 2. Copy and edit environment variables
cp .env.example .env
nano .env  # Add your API keys

# 3. Deploy!
./deploy.sh up
```

That's it! Access at http://localhost

## 📋 Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- Git

### Install Docker (Ubuntu/Debian)

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker --version
docker-compose --version
```

## 🔧 Configuration

### 1. Environment Variables

Create `.env` file:

```bash
cp .env.example .env
```

**Required variables:**
```bash
# LLM Provider API Keys
OPENAI_API_KEY=sk-your-openai-key

# Stripe (for payments)
STRIPE_SECRET_KEY=sk_test_your-key
STRIPE_WEBHOOK_SECRET=whsec_your-secret
STRIPE_PRICE_ID=price_your-price-id

# Platform Settings
MAX_CUSTOMERS=20
PORT_RANGE_START=30000
PORT_RANGE_END=30020
BASE_DOMAIN=localhost
```

### 2. Optional Configuration

Edit `docker-compose.yml` to adjust:
- Port mappings
- Resource limits
- Volume mounts

## 🎮 Deploy Commands

```bash
# Start/redeploy
./deploy.sh up

# Stop everything
./deploy.sh down

# Restart services
./deploy.sh restart

# View logs
./deploy.sh logs

# Update to latest code
./deploy.sh update

# Check status
./deploy.sh status
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Network                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Caddy      │  │  Frontend    │  │   Backend    │  │
│  │   :80        │  │   :3000      │  │   :8080      │  │
│  │  (Proxy)     │  │  (Next.js)   │  │   (Go API)   │  │
│  └──────┬───────┘  └──────────────┘  └──────────────┘  │
│         │                                               │
│  ┌──────┴───────┐                                      │
│  │  Customer    │                                      │
│  │ Containers   │  (Spawned by Backend)                │
│  └──────────────┘                                      │
└─────────────────────────────────────────────────────────┘
```

## 📂 Volumes & Data

Persistent data is stored in:
- `./data/` - SQLite database
- `./customers/` - Customer container data
- `caddy-data/` - Caddy certificates and config

**Backup:**
```bash
# Backup data
tar -czf backup-$(date +%Y%m%d).tar.gz data/ customers/

# Restore
tar -xzf backup-20240220.tar.gz
```

## 🔍 Troubleshooting

### Containers won't start
```bash
# Check logs
./deploy.sh logs

# Check specific service
docker-compose logs backend
docker-compose logs frontend

# Verify environment
cat .env | grep -v "^#" | grep -v "^$"
```

### Port conflicts
```bash
# Check what's using port 80
sudo lsof -i :80

# Change port in docker-compose.yml
# Change '80:80' to '8080:80' for example
```

### Permission denied
```bash
# Fix permissions
sudo chown -R $USER:$USER data/ customers/
chmod +x deploy.sh
```

### Backend can't spawn containers
The backend needs Docker socket access:
```bash
# Verify docker.sock is mounted
docker-compose exec backend ls -la /var/run/docker.sock

# Should show: /var/run/docker.sock
```

## 🔄 Updates

```bash
# Pull latest code and redeploy
./deploy.sh update

# Or manually:
git pull origin main
./deploy.sh restart
```

## 🌐 Production Deployment

For production with domain:

1. **Update .env:**
```bash
BASE_DOMAIN=yourdomain.com
```

2. **Update Caddyfile:**
```caddy
yourdomain.com {
    reverse_proxy frontend:3000
    
    handle /api/* {
        reverse_proxy backend:8080
    }
}
```

3. **Enable HTTPS:**
```bash
# Caddy auto-handles SSL with Let's Encrypt
# Just ensure ports 80/443 are open
```

4. **Set resource limits in docker-compose.yml:**
```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

## 📊 Monitoring

```bash
# Container stats
docker stats

# Disk usage
docker system df

# Clean up unused images
docker system prune -a
```

## 🧹 Cleanup

```bash
# Stop and remove everything
./deploy.sh down

# Remove all data (WARNING: Destructive!)
./deploy.sh down
rm -rf data/ customers/
docker volume prune
```

## 💡 Tips

- **Ryzen 7 + 32GB RAM** can handle 50+ customers easily
- Use `docker-compose up -d` instead of deploy.sh for manual control
- Database is SQLite - for 100+ customers, consider PostgreSQL
- Customer containers auto-spawn on their assigned ports
- Logs are in `docker-compose logs` - persist them for production

## 🆘 Need Help?

```bash
# Check all services
./deploy.sh status

# View backend logs
./deploy.sh logs | grep backend

# Test API
curl http://localhost/api/health
```
