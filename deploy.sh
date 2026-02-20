#!/bin/bash

# Blytz Cloud - Quick Deploy Script
# Usage: ./deploy.sh [up|down|restart|logs|update]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="blytz"
COMPOSE_FILE="docker-compose.yml"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if .env exists
check_env() {
    if [ ! -f .env ]; then
        log_warning ".env file not found!"
        log_info "Creating from .env.example..."
        if [ -f .env.example ]; then
            cp .env.example .env
            log_warning "Please edit .env with your API keys before starting!"
            exit 1
        else
            log_error ".env.example not found. Please create .env manually."
            exit 1
        fi
    fi
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found! Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose not found! Please install Docker Compose."
        exit 1
    fi
}

# Create necessary directories
setup_directories() {
    log_info "Setting up directories..."
    mkdir -p data customers
    log_success "Directories created"
}

# Deploy/up function
deploy_up() {
    log_info "Starting Blytz Cloud deployment..."
    
    check_env
    check_docker
    setup_directories
    
    log_info "Building and starting containers..."
    docker-compose -p $PROJECT_NAME -f $COMPOSE_FILE up --build -d
    
    log_info "Waiting for services to be healthy..."
    sleep 5
    
    # Check health
    if curl -f http://localhost:8080/api/health > /dev/null 2>&1; then
        log_success "Backend is healthy!"
    else
        log_warning "Backend health check failed. Check logs with: ./deploy.sh logs"
    fi
    
    log_success "Blytz Cloud is running!"
    log_info "Access the application at: http://localhost"
    log_info "API available at: http://localhost:8080"
}

# Stop/down function
deploy_down() {
    log_info "Stopping Blytz Cloud..."
    docker-compose -p $PROJECT_NAME -f $COMPOSE_FILE down
    log_success "Blytz Cloud stopped"
}

# Restart function
deploy_restart() {
    log_info "Restarting Blytz Cloud..."
    deploy_down
    deploy_up
}

# View logs
view_logs() {
    log_info "Viewing logs (Ctrl+C to exit)..."
    docker-compose -p $PROJECT_NAME -f $COMPOSE_FILE logs -f
}

# Update/redeploy
update_deploy() {
    log_info "Updating Blytz Cloud..."
    
    log_info "Pulling latest code..."
    git pull origin main
    
    log_info "Rebuilding and restarting..."
    docker-compose -p $PROJECT_NAME -f $COMPOSE_FILE down
    docker-compose -p $PROJECT_NAME -f $COMPOSE_FILE up --build -d
    
    log_success "Update complete!"
}

# Show status
show_status() {
    log_info "Container status:"
    docker-compose -p $PROJECT_NAME -f $COMPOSE_FILE ps
    
    echo ""
    log_info "Health checks:"
    
    if curl -f http://localhost:8080/api/health > /dev/null 2>&1; then
        log_success "Backend: HEALTHY"
    else
        log_error "Backend: UNHEALTHY"
    fi
    
    if curl -f http://localhost:3000 > /dev/null 2>&1; then
        log_success "Frontend: HEALTHY"
    else
        log_error "Frontend: UNHEALTHY"
    fi
}

# Main command handler
case "${1:-up}" in
    up|start|deploy)
        deploy_up
        ;;
    down|stop)
        deploy_down
        ;;
    restart)
        deploy_restart
        ;;
    logs)
        view_logs
        ;;
    update)
        update_deploy
        ;;
    status)
        show_status
        ;;
    *)
        echo "Usage: $0 {up|down|restart|logs|update|status}"
        echo ""
        echo "Commands:"
        echo "  up       - Deploy/start Blytz Cloud (default)"
        echo "  down     - Stop Blytz Cloud"
        echo "  restart  - Restart all services"
        echo "  logs     - View real-time logs"
        echo "  update   - Pull latest code and redeploy"
        echo "  status   - Check service health"
        exit 1
        ;;
esac
