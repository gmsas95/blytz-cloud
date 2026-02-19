# AGENTS.md - BlytzCloud Development Guide

## Project Overview

BlytzCloud is a Go-based platform for deploying personalized OpenClaw AI assistants. It uses Gin web framework, SQLite, Docker, and Caddy reverse proxy.

## Build Commands

```bash
# Build the application
go build -o blytz ./cmd/server

# Build with race detection (for development)
go build -race -o blytz ./cmd/server

# Run the server
./blytz

# Or run directly
go run ./cmd/server
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Run a single test (specify package and test name)
go test ./internal/db -run TestCreateCustomer -v

# Run tests for a specific package
go test ./internal/config -v
go test ./internal/db -v
go test ./internal/workspace -v
go test ./internal/api -v

# Run integration tests only
go test -tags=integration ./...
```

## Code Quality Commands

```bash
# Format code
go fmt ./...

# Vet code for common issues
go vet ./...

# Tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

## Project Structure

```
blytz/
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/             # HTTP handlers, middleware, routes
│   ├── config/          # Configuration loading
│   ├── db/              # Database connection, migrations, CRUD
│   ├── provisioner/     # Container lifecycle management
│   ├── workspace/       # AGENTS.md, USER.md, SOUL.md generation
│   ├── telegram/        # Bot token validation
│   ├── stripe/          # Checkout and webhooks
│   └── caddy/           # Reverse proxy integration
├── static/              # HTML templates (embedded)
├── deployments/         # systemd service files
└── go.mod, go.sum       # Go module files
```

## Code Style Guidelines

### Imports

Group imports in this order (separated by blank lines):
1. Standard library packages
2. Third-party packages
3. Internal project packages

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"modernc.org/sqlite"

	"blytz/internal/config"
	"blytz/internal/db"
)
```

### Naming Conventions

- **Packages**: Lowercase, single word (e.g., `config`, `provisioner`)
- **Exported identifiers**: PascalCase (e.g., `CreateCustomer`, `ServerConfig`)
- **Unexported identifiers**: camelCase (e.g., `validateToken`, `dbConn`)
- **Interfaces**: End with "-er" or describe capability (e.g., `Provisioner`, `Handler`)
- **Test files**: `*_test.go` suffix
- **Test functions**: `Test` prefix + PascalCase (e.g., `TestCreateCustomer`)

### Types and Structs

```go
// Customer represents a platform customer
type Customer struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Provisioner manages container lifecycle
type Provisioner interface {
	Create(ctx context.Context, customerID string) error
	Start(ctx context.Context, customerID string) error
	Stop(ctx context.Context, customerID string) error
	Remove(ctx context.Context, customerID string) error
}
```

### Error Handling

Always check errors explicitly. Use wrapped errors with context:

```go
if err != nil {
	return fmt.Errorf("failed to create customer %s: %w", email, err)
}
```

Use sentinel errors for specific cases:

```go
var ErrCustomerNotFound = errors.New("customer not found")
var ErrCapacityExceeded = errors.New("maximum customer capacity reached")
```

### HTTP Handlers

Follow this pattern for Gin handlers:

```go
func (h *Handler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	customer, err := h.db.CreateCustomer(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to create customer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create customer"})
		return
	}

	c.JSON(http.StatusCreated, customer)
}
```

### Database Operations

Use context for all database operations:

```go
func (db *DB) CreateCustomer(ctx context.Context, c *Customer) error {
	query := `INSERT INTO customers (id, email, status, created_at) VALUES (?, ?, ?, ?)`
	_, err := db.conn.ExecContext(ctx, query, c.ID, c.Email, c.Status, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert customer: %w", err)
	}
	return nil
}
```

### Testing

Test coverage targets:
- `config`: 90%
- `db`: 85%
- `workspace`: 90%
- `provisioner`: 70%
- `api`: 80%

Use table-driven tests:

```go
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid", "user@example.com", false},
		{"invalid", "not-an-email", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

Use `testify/assert` for assertions if available, otherwise standard library.

### Configuration

Load from environment variables with sensible defaults:

```go
type Config struct {
	Port         int    `env:"PORT" envDefault:"8080"`
	DatabasePath string `env:"DATABASE_PATH" envDefault:"./database.sqlite"`
	MaxCustomers int    `env:"MAX_CUSTOMERS" envDefault:"20"`
}
```

### Logging

Use structured logging with zap:

```go
logger.Info("customer created",
	zap.String("customer_id", customer.ID),
	zap.String("email", customer.Email),
	zap.Duration("duration", time.Since(start)))
```

## Environment Setup

Required environment variables:
```bash
OPENAI_API_KEY=sk-xxx
STRIPE_SECRET_KEY=sk_live_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
STRIPE_PRICE_ID=price_xxx
DATABASE_PATH=/opt/blytz/platform/database.sqlite
MAX_CUSTOMERS=20
PORT_RANGE_START=30000
PORT_RANGE_END=30999
BASE_DOMAIN=blytz.cloud
PLATFORM_PORT=8080
```

## Dependencies

Key dependencies (latest versions):
- `github.com/gin-gonic/gin v1.11.0` - Web framework
- `modernc.org/sqlite v1.46.1` - SQLite driver
- `github.com/docker/docker v29.2.1` - Docker SDK
- `github.com/stripe/stripe-go v84.3.0` - Stripe integration
- `go.uber.org/zap v1.27.1` - Structured logging
- `github.com/google/uuid v1.6.0` - UUID generation

**Go Version:** 1.26.0 (minimum recommended)

Always run `go mod tidy` after adding/removing imports.

## Docker Guidelines

- Container names: `blytz-{customer-id}`
- Port range: 30000-30999
- Memory limit: 1GB per container
- Always use health checks in compose files

## Security Notes

- Never log sensitive data (API keys, tokens, passwords)
- Validate all inputs at API boundaries
- Use prepared statements for all SQL queries
- Store secrets in environment variables, never in code

---

## Development Practices

### Local Development Setup

**Prerequisites:**
- Go 1.26.0+
- Docker (with daemon accessible)
- SQLite3 (for local database inspection)
- Make (for running tasks)

**Initial Setup:**
```bash
# 1. Clone and enter directory
cd /home/gmsas95/blytz.cloud

# 2. Copy environment template
cp .env.example .env
# Edit .env with your values

# 3. Initialize Go module
go mod init blytz 2>/dev/null || go mod tidy

# 4. Create local directories
mkdir -p tmp/customers tmp/platform

# 5. Run tests to verify setup
go test ./...
```

**Environment Variables Template (.env.example):**
```bash
# API Keys (required for full functionality)
OPENAI_API_KEY=sk-your-key-here
STRIPE_SECRET_KEY=sk_test_your-key-here
STRIPE_WEBHOOK_SECRET=whsec_your-secret-here
STRIPE_PRICE_ID=price_your-price-id

# Platform Configuration
DATABASE_PATH=./tmp/platform/database.sqlite
CUSTOMERS_DIR=./tmp/customers
TEMPLATES_DIR=./internal/workspace/templates
MAX_CUSTOMERS=20
PORT_RANGE_START=30000
PORT_RANGE_END=30999
BASE_DOMAIN=localhost
PLATFORM_PORT=8080

# Security
OPENCLAW_GATEWAY_TOKEN_PREFIX=blytz_
```

**Hot Reload Development:**
```bash
# Install air for hot reload
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

### Testing Standards

**Coverage Enforcement:**
- All packages must maintain coverage targets defined in "Testing" section
- PRs cannot be merged if they decrease coverage by >5%
- Integration tests must pass before deployment

**Test Organization:**
```
internal/
├── config/
│   ├── config.go
│   └── config_test.go          # Unit tests
├── db/
│   ├── db.go
│   ├── customer.go
│   └── customer_test.go        # Unit + integration tests
```

**Test Data Management:**
- Use `testify/require` for assertions
- Clean up test data in `t.Cleanup()`
- Use `t.Parallel()` for isolated tests
- Mock external dependencies (Docker, Stripe, Telegram)

**Test Naming:**
- `Test<FunctionName>` - basic functionality
- `Test<FunctionName>_Error` - error cases
- `Test<FunctionName>_Integration` - integration tests (build tag)

### Linting and Code Quality

**Required Tools:**
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run ./...
```

**Linting Configuration (.golangci.yml):**
```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - typecheck
    - gosec
    - unconvert
    - goconst
    - gocyclo
    - misspell
    - lll
    - goimports
    - nakedret
    - prealloc
    - dogsled
    - bodyclose

linters-settings:
  lll:
    line-length: 120
  gocyclo:
    min-complexity: 15
  gosec:
    excludes:
      - G104  # Errcheck handled separately

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - lll
```

**Pre-commit Workflow:**
```bash
#!/bin/bash
# .git/hooks/pre-commit or use pre-commit tool

echo "Running pre-commit checks..."

# Format
go fmt ./... || exit 1

# Vet
go vet ./... || exit 1

# Lint
golangci-lint run ./... || exit 1

# Test
go test ./... || exit 1

echo "All checks passed!"
```

### Security Practices

**Secrets Management:**
- NEVER commit `.env` files
- NEVER log secrets (use `zap.String("key", "[REDACTED]")`)
- Use environment variables exclusively
- Rotate API keys quarterly

**Input Validation:**
- Validate at API boundaries (use Gin's binding)
- Sanitize user inputs before database storage
- Use prepared statements for all SQL queries
- Validate file paths to prevent directory traversal

**SQL Injection Prevention:**
```go
// GOOD: Prepared statement
db.Exec("INSERT INTO customers (email) VALUES (?)", email)

// BAD: String concatenation
db.Exec("INSERT INTO customers (email) VALUES ('" + email + "')")
```

**Dependency Security:**
```bash
# Check for vulnerabilities
go list -json -deps ./... | nancy sleuth

# Or use govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### Database Migrations

**Migration Strategy:**
- Use sequential migration files: `001_initial_schema.sql`, `002_add_index.sql`
- Store migrations in `internal/db/migrations/`
- Support both up and down migrations
- Never modify existing migration files after commit

**Migration File Format:**
```sql
-- 001_initial_schema.up.sql
CREATE TABLE customers (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    -- ...
);

-- 001_initial_schema.down.sql
DROP TABLE customers;
```

**Migration Runner:**
```go
func (db *DB) Migrate() error {
    // Apply pending migrations
    // Track applied migrations in schema_migrations table
}
```

### Observability and Monitoring

**Structured Logging Standards:**
- Use `zap` for all logging
- Always include `customer_id` in customer-related operations
- Include timing for operations >100ms
- Use appropriate log levels:
  - `Debug`: Detailed debugging info
  - `Info`: Normal operations
  - `Warn`: Recoverable issues
  - `Error`: Failures requiring attention
  - `Fatal`: Unrecoverable errors (avoid when possible)

**Log Format:**
```go
logger.Info("operation completed",
    zap.String("customer_id", customerID),
    zap.String("operation", "create_container"),
    zap.Duration("duration", elapsed),
    zap.Int("port", port))
```

**Health Checks:**
- `/api/health` - Basic platform health
- Check database connectivity
- Check Docker daemon accessibility
- Return 503 if any critical dependency is down

**Metrics to Track:**
- Request latency (p50, p95, p99)
- Active customer count
- Container start/stop/failure rates
- Stripe webhook processing latency
- Database query duration

### CI/CD Pipeline

**GitHub Actions Workflow (.github/workflows/ci.yml):**
```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.26'
      
      - name: Cache dependencies
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run linter
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
      
      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
      
      - name: Build
        run: go build -o blytz ./cmd/server
      
      - name: Vulnerability scan
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
```

**Branch Protection Rules:**
- Require PR reviews (1 approval minimum)
- Require status checks to pass (CI, tests, lint)
- Require branches to be up to date before merging
- Restrict pushes to main branch

**Release Process:**
1. Create release branch: `release/v1.2.0`
2. Update version in code
3. Run full test suite
4. Create PR to main
5. Merge after review
6. Tag release: `git tag v1.2.0`
7. Push tag: `git push origin v1.2.0`
8. GitHub Actions builds and attaches binaries

### Code Review Checklist

**For Authors:**
- [ ] Tests added for new functionality
- [ ] All tests pass locally
- [ ] Code is formatted (`go fmt`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] No secrets in code
- [ ] Documentation updated (if needed)
- [ ] PR description explains what and why

**For Reviewers:**
- [ ] Code follows project conventions
- [ ] Error handling is appropriate
- [ ] No obvious security issues
- [ ] Test coverage is adequate
- [ ] No unnecessary complexity
- [ ] Commit messages are clear

### Debugging and Troubleshooting

**Common Issues:**

1. **Docker permission denied:**
   ```bash
   sudo usermod -aG docker $USER
   # Log out and back in
   ```

2. **Port already in use:**
   ```bash
   # Find process using port
   lsof -i :8080
   # Kill if necessary
   kill -9 <PID>
   ```

3. **Database locked:**
   ```bash
   # Check for hanging connections
   lsof | grep database.sqlite
   # Remove WAL files if corrupted
   rm -f database.sqlite-shm database.sqlite-wal
   ```

**Debug Logging:**
```bash
# Enable debug logging
LOG_LEVEL=debug go run ./cmd/server

# Run with race detector
 go run -race ./cmd/server
```

**Profiling:**
```bash
# Enable pprof endpoints
# Access at http://localhost:8080/debug/pprof/

# Generate CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```
