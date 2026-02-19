#!/bin/bash
# Load Testing Script for BlytzCloud
# Tests concurrent user signups to simulate 30 users

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOTAL_USERS=${TOTAL_USERS:-30}
CONCURRENT=${CONCURRENT:-5}
DELAY_BETWEEN_BATCHES=${DELAY_BETWEEN_BATCHES:-10}

echo "=========================================="
echo "BlytzCloud Load Testing"
echo "=========================================="
echo "Target: $BASE_URL"
echo "Total Users: $TOTAL_USERS"
echo "Concurrent: $CONCURRENT"
echo "Delay between batches: ${DELAY_BETWEEN_BATCHES}s"
echo "=========================================="
echo ""

# Function to create a single user
create_user() {
    local user_num=$1
    local email="testuser${user_num}@example.com"
    local timestamp=$(date +%s%N)
    
    # Make unique email to avoid conflicts
    email="loadtest${user_num}_${timestamp}@test.com"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/signup" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$email\",
            \"assistant_name\": \"Test Assistant $user_num\",
            \"custom_instructions\": \"Load testing user $user_num\",
            \"telegram_bot_token\": \"123456:ABC${user_num}DEF\"
        }" 2>&1)
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" -eq 201 ]; then
        echo "✓ User $user_num created successfully"
        echo "  Email: $email"
        # Extract customer_id if present
        if echo "$body" | grep -q "customer_id"; then
            customer_id=$(echo "$body" | grep -o '"customer_id":"[^"]*"' | cut -d'"' -f4)
            echo "  Customer ID: $customer_id"
        fi
        return 0
    elif [ "$http_code" -eq 409 ]; then
        echo "⚠ User $user_num already exists (expected in some cases)"
        return 0
    elif [ "$http_code" -eq 503 ]; then
        echo "✗ User $user_num failed: Platform at capacity"
        return 1
    else
        echo "✗ User $user_num failed: HTTP $http_code"
        echo "  Response: $body"
        return 1
    fi
}

# Function to check system status
check_system_status() {
    echo ""
    echo "Checking system status..."
    response=$(curl -s "$BASE_URL/api/status/system" 2>&1)
    if [ $? -eq 0 ]; then
        echo "✓ System status:"
        echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    else
        echo "✗ Failed to get system status"
    fi
    echo ""
}

# Main test execution
main() {
    local success_count=0
    local fail_count=0
    local start_time=$(date +%s)
    
    echo "Starting load test..."
    echo ""
    
    # Process users in batches
    for ((batch=0; batch<TOTAL_USERS; batch+=CONCURRENT)); do
        batch_end=$((batch + CONCURRENT))
        if [ $batch_end -gt $TOTAL_USERS ]; then
            batch_end=$TOTAL_USERS
        fi
        
        echo "Processing batch $((batch/CONCURRENT + 1)): users $((batch+1))-$batch_end"
        
        # Launch concurrent requests
        for ((i=batch; i<batch_end; i++)); do
            create_user $i &
        done
        
        # Wait for all background jobs to complete
        wait
        
        # Check status after each batch
        check_system_status
        
        # Delay between batches (except for the last one)
        if [ $batch_end -lt $TOTAL_USERS ]; then
            echo "Waiting ${DELAY_BETWEEN_BATCHES}s before next batch..."
            sleep $DELAY_BETWEEN_BATCHES
        fi
    done
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo ""
    echo "=========================================="
    echo "Load Test Complete"
    echo "=========================================="
    echo "Duration: ${duration}s"
    echo "Users processed: $TOTAL_USERS"
    echo ""
    
    # Final system status
    check_system_status
    
    # Docker stats
    echo ""
    echo "Docker Container Stats:"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.Status}}" 2>/dev/null | grep blytz- || echo "No blytz containers running yet (provisioning may still be in progress)"
    
    echo ""
    echo "System Resources:"
    echo "CPU Usage:"
    top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print "  " $1 "%"}'
    echo "Memory Usage:"
    free -h | grep "Mem:" | awk '{print "  Used: " $3 " / Total: " $2}'
    echo ""
    echo "=========================================="
}

# Show usage
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Environment Variables:"
    echo "  BASE_URL                  Target URL (default: http://localhost:8080)"
    echo "  TOTAL_USERS               Number of users to create (default: 30)"
    echo "  CONCURRENT                Concurrent requests per batch (default: 5)"
    echo "  DELAY_BETWEEN_BATCHES     Seconds between batches (default: 10)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Test with defaults (30 users)"
    echo "  TOTAL_USERS=10 $0                     # Test with 10 users"
    echo "  BASE_URL=http://192.168.1.100:8080 TOTAL_USERS=50 $0"
    echo ""
    exit 0
fi

# Check dependencies
if ! command -v curl &> /dev/null; then
    echo "Error: curl is required but not installed"
    exit 1
fi

# Run main test
main
