#!/bin/bash
# Real-time Monitoring Script for BlytzCloud
# Monitors containers, database, and system resources

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
REFRESH_RATE="${REFRESH_RATE:-5}"

echo "=========================================="
echo "BlytzCloud Real-Time Monitor"
echo "=========================================="
echo "Target: $BASE_URL"
echo "Refresh: Every ${REFRESH_RATE}s"
echo "Press Ctrl+C to stop"
echo "=========================================="
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "Monitoring stopped."
    exit 0
}
trap cleanup SIGINT SIGTERM

# Function to get container count
get_container_count() {
    docker ps --filter "name=blytz-" --format "{{.Names}}" 2>/dev/null | wc -l
}

# Function to get system status from API
get_api_status() {
    curl -s "$BASE_URL/api/status/system" 2>/dev/null | python3 -c "import sys, json; d=json.load(sys.stdin); print(f\"Active: {d['capacity']['active_customers']}/{d['capacity']['max_capacity']} ({d['capacity']['usage_percentage']})\")" 2>/dev/null || echo "API unavailable"
}

# Function to get resource usage
get_resource_usage() {
    # CPU
    cpu_idle=$(top -bn1 | grep "Cpu(s)" | awk '{print $8}' | cut -d'%' -f1)
    cpu_used=$(echo "100 - $cpu_idle" | bc)
    
    # Memory
    mem_info=$(free | grep Mem)
    mem_total=$(echo $mem_info | awk '{print $2}')
    mem_used=$(echo $mem_info | awk '{print $3}')
    mem_percent=$(echo "scale=1; $mem_used * 100 / $mem_total" | bc)
    
    # Docker containers
    container_count=$(get_container_count)
    
    echo "CPU: ${cpu_used}% | Memory: ${mem_percent}% | Containers: $container_count"
}

# Function to get top containers by memory
get_top_containers() {
    docker stats --no-stream --format "{{.Name}}: {{.MemPerc}}" 2>/dev/null | grep blytz- | sort -t':' -k2 -nr | head -5
}

# Main monitoring loop
while true; do
    clear
    echo "=========================================="
    echo "BlytzCloud Monitor - $(date '+%Y-%m-%d %H:%M:%S')"
    echo "=========================================="
    echo ""
    
    # API Status
    echo "📊 API Status:"
    get_api_status
    echo ""
    
    # Resource Usage
    echo "💻 System Resources:"
    get_resource_usage
    echo ""
    
    # Container Count
    container_count=$(get_container_count)
    echo "🐳 Docker Containers: $container_count running"
    echo ""
    
    # Top containers
    if [ $container_count -gt 0 ]; then
        echo "🔝 Top 5 Containers by Memory:"
        get_top_containers | while read line; do
            echo "  $line"
        done
        echo ""
    fi
    
    # Quick stats
    echo "📈 Quick Stats:"
    echo "  Platform: $(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/health" 2>/dev/null)"
    echo "  Database: $(sqlite3 /opt/blytz/platform/database.sqlite "SELECT COUNT(*) FROM customers;" 2>/dev/null || echo "N/A") customers"
    echo ""
    
    echo "Press Ctrl+C to stop..."
    sleep $REFRESH_RATE
done
