#!/bin/bash
# E2E Test Runner for BlytzCloud

set -e

echo "=== BlytzCloud E2E Test Runner ==="
echo ""

# Check if we should run full e2e tests
if [ "$1" == "--full" ]; then
    export RUN_E2E=true
    echo "Mode: FULL E2E (including external services)"
else
    echo "Mode: Local E2E (mocked external services)"
fi

# Check for Docker tests
if [ "$2" == "--docker" ] || [ "$DOCKER_TEST" == "true" ]; then
    export DOCKER_TEST=true
    echo "Docker tests: ENABLED"
    
    # Verify Docker is available
    if ! docker version > /dev/null 2>&1; then
        echo "ERROR: Docker is not available"
        exit 1
    fi
    echo "Docker: Available"
else
    echo "Docker tests: SKIPPED (use --docker or set DOCKER_TEST=true)"
fi

echo ""
echo "=== Running Tests ==="
echo ""

# Run the tests
go test -v -tags=e2e ./internal/e2e/... -timeout 10m

echo ""
echo "=== E2E Tests Complete ==="
