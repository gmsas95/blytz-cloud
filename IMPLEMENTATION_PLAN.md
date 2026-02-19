# BlytzCloud Implementation Plan
## Post-Audit Remediation Roadmap

**Generated:** February 19, 2026  
**Based on:** 2026-02-19 CROSS AUDIT.md  
**Estimated Duration:** 3-4 weeks  
**Team Size:** 1-2 developers

---

## Executive Summary

This plan addresses all critical, high, and medium priority issues identified in the audit. Items are organized by phase, with clear dependencies and acceptance criteria.

---

## Phase 1: Critical Fixes (Week 1) - BLOCKING PRODUCTION

**Goal:** Fix all P0/P1 issues that prevent production deployment

### 1.1 Fix Test Suite Failures
**Priority:** CRITICAL | **Effort:** 2-3 hours | **Owner:** Backend Dev

#### Tasks:

**A. Fix smoke_test.go nil logger panic**

**File:** `internal/api/smoke_test.go`

```go
// Line 50 - Change from:
router := NewRouter(database, prov, stripeSvc, stripeWebhook, cfg, nil)

// To:
logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
router := NewRouter(database, prov, stripeSvc, stripeWebhook, cfg, logger)
```

**B. Fix service_test.go compilation errors**

**File:** `internal/provisioner/service_test.go`

```go
// Line 4 - Remove unused import:
// "context"  <- DELETE THIS LINE

// Line 208 - Remove unused variable:
// svc := NewService(...)  <- DELETE THIS LINE
// OR use it:
require.NotNil(t, svc)
```

**C. Fix stripe_test.go type mismatch**

**File:** `internal/stripe/stripe_test.go`

**Problem:** `mockProvisioner` doesn't satisfy `*provisioner.Service` type

**Solution 1:** Create an interface
```go
// Create a provisioner interface
type Provisioner interface {
    Provision(ctx context.Context, customerID string) error
    Suspend(ctx context.Context, customerID string) error
    Resume(ctx context.Context, customerID string) error
    Terminate(ctx context.Context, customerID string) error
    ValidateBotToken(token string) (*telegram.BotInfo, error)
}

// Update WebhookHandler to use interface
type WebhookHandler struct {
    db            *db.DB
    provisioner   Provisioner  // Change from *provisioner.Service
    webhookSecret string
}
```

**Solution 2:** Use testify/mock (recommended)
```go
import "github.com/stretchr/testify/mock"

type MockProvisioner struct {
    mock.Mock
}

func (m *MockProvisioner) Provision(ctx context.Context, customerID string) error {
    args := m.Called(ctx, customerID)
    return args.Error(0)
}
// ... implement other methods
```

**Acceptance Criteria:**
- [ ] `go test ./...` compiles without errors
- [ ] All tests pass or have valid skip conditions
- [ ] `go vet ./...` shows no issues

---

### 1.2 Implement Docker Secrets for API Keys
**Priority:** CRITICAL | **Effort:** 4-6 hours | **Owner:** Backend Dev

**Problem:** OpenAI API key exposed in docker-compose.yml

**File Changes:**
- `internal/provisioner/compose.go`
- `internal/provisioner/service.go`
- `internal/workspace/config.go`

#### Implementation:

**A. Modify compose generation to use env_file**

```go
// internal/provisioner/compose.go
func (cg *ComposeGenerator) Generate(customerID string, port int, openAIKey string) error {
    compose := fmt.Sprintf(`version: '3.8'
services:
  openclaw:
    image: node:22-alpine
    container_name: blytz-%s
    working_dir: /app
    command: sh -c "npm install -g openclaw@latest && openclaw gateway --port 18789"
    ports:
      - "%d:18789"
    volumes:
      - ./.openclaw:/root/.openclaw
    env_file:
      - .env.secret  # Use env file instead of inline env vars
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '0.5'
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:18789/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
`, customerID, port)

    // ... rest of function
}
```

**B. Create .env.secret file generator**

```go
// internal/provisioner/compose.go - Add new method

func (cg *ComposeGenerator) GenerateEnvFile(customerID string, openAIKey string) error {
    envContent := fmt.Sprintf("OPENAI_API_KEY=%s\n", openAIKey)
    
    customerDir := filepath.Join(cg.baseDir, customerID)
    envPath := filepath.Join(customerDir, ".env.secret")
    
    // Write with restricted permissions (owner read/write only)
    if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
        return fmt.Errorf("write env file: %w", err)
    }
    
    return nil
}
```

**C. Update Service.Provision to generate env file**

```go
// internal/provisioner/service.go - in Provision method

if err := s.compose.GenerateEnvFile(customerID, s.openAIKey); err != nil {
    s.cleanup(customerID, port)
    s.db.UpdateCustomerStatus(ctx, customerID, "pending")
    return fmt.Errorf("generate env file: %w", err)
}

if err := s.compose.Generate(customerID, port, s.openAIKey); err != nil {
    // ...
}
```

**D. Update cleanup to remove env file**

```go
// internal/provisioner/service.go

func (s *Service) cleanup(customerID string, port int) {
    s.docker.Remove(context.Background(), customerID)
    s.db.ReleasePort(context.Background(), port)
    s.ports.ReleasePort(port)
    
    // Remove env file
    envPath := filepath.Join(s.baseDir, customerID, ".env.secret")
    os.Remove(envPath)  // Ignore errors
}
```

**Acceptance Criteria:**
- [ ] API keys not visible in docker-compose.yml
- [ ] .env.secret files created with 0600 permissions
- [ ] Containers can still access OpenAI API key
- [ ] Cleanup removes env files

---

### 1.3 Add Mutex to PortAllocator
**Priority:** CRITICAL | **Effort:** 1-2 hours | **Owner:** Backend Dev

**File:** `internal/provisioner/ports.go`

```go
package provisioner

import (
    "context"
    "fmt"
    "sync"  // ADD THIS IMPORT
)

type PortAllocator struct {
    mu        sync.Mutex  // ADD THIS FIELD
    startPort int
    endPort   int
    allocated map[int]bool
}

func (pa *PortAllocator) AllocatePort() (int, error) {
    pa.mu.Lock()         // ADD LOCK
    defer pa.mu.Unlock() // ADD UNLOCK
    
    for port := pa.startPort; port <= pa.endPort; port++ {
        if !pa.allocated[port] {
            pa.allocated[port] = true
            return port, nil
        }
    }
    return 0, fmt.Errorf("no available ports in range %d-%d", pa.startPort, pa.endPort)
}

func (pa *PortAllocator) ReleasePort(port int) {
    pa.mu.Lock()         // ADD LOCK
    defer pa.mu.Unlock() // ADD UNLOCK
    
    delete(pa.allocated, port)
}

func (pa *PortAllocator) LoadAllocatedPorts(ctx context.Context, db interface {
    GetAllocatedPorts(context.Context) ([]int, error)
}) error {
    pa.mu.Lock()         // ADD LOCK
    defer pa.mu.Unlock() // ADD UNLOCK
    
    ports, err := db.GetAllocatedPorts(ctx)
    if err != nil {
        return fmt.Errorf("load allocated ports: %w", err)
    }

    for _, port := range ports {
        pa.allocated[port] = true
    }

    return nil
}
```

**Acceptance Criteria:**
- [ ] PortAllocator is thread-safe
- [ ] Concurrent tests pass without race conditions
- [ ] Run `go test -race ./internal/provisioner` shows no issues

---

### 1.4 Fix Container Port Persistence
**Priority:** CRITICAL | **Effort:** 2-3 hours | **Owner:** Backend Dev

**Files:**
- `internal/db/db.go` - Add new method
- `internal/provisioner/service.go` - Use new method

**A. Add UpdateCustomerPort method to DB**

```go
// internal/db/db.go - Add after line 306

func (db *DB) UpdateCustomerPort(ctx context.Context, id string, port int) error {
    query := `UPDATE customers SET container_port = ?, updated_at = ? WHERE id = ?`
    _, err := db.conn.ExecContext(ctx, query, port, time.Now(), id)
    if err != nil {
        return fmt.Errorf("update customer port: %w", err)
    }
    return nil
}
```

**B. Update Service.Provision to persist port**

```go
// internal/provisioner/service.go - After AllocatePort section

if err := s.db.AllocatePort(ctx, customerID, port); err != nil {
    s.ports.ReleasePort(port)
    s.db.UpdateCustomerStatus(ctx, customerID, "pending")
    return fmt.Errorf("record port allocation: %w", err)
}

// ADD THIS:
if err := s.db.UpdateCustomerPort(ctx, customerID, port); err != nil {
    s.cleanup(customerID, port)
    s.db.UpdateCustomerStatus(ctx, customerID, "pending")
    return fmt.Errorf("update customer port: %w", err)
}
```

**Acceptance Criteria:**
- [ ] Customer records have container_port populated after provisioning
- [ ] Terminate properly releases ports from both DB tables
- [ ] GetCustomerByID returns correct container_port

---

### 1.5 Sanitize Customer ID Generation
**Priority:** CRITICAL | **Effort:** 2-3 hours | **Owner:** Backend Dev

**File:** `internal/db/db.go` (lines 91-96)

```go
func generateCustomerID(email string) string {
    // Validate email first
    if !isValidEmail(email) {
        return ""
    }
    
    id := strings.ToLower(email)
    
    // Replace @ and . with safe separators
    id = strings.ReplaceAll(id, "@", "-")
    id = strings.ReplaceAll(id, ".", "-")
    
    // Remove any remaining non-alphanumeric characters except hyphen
    var result strings.Builder
    for _, r := range id {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
            result.WriteRune(r)
        }
    }
    id = result.String()
    
    // Remove leading/trailing hyphens and consecutive hyphens
    id = strings.Trim(id, "-")
    for strings.Contains(id, "--") {
        id = strings.ReplaceAll(id, "--", "-")
    }
    
    // Enforce max length (63 chars for Docker container names)
    if len(id) > 63 {
        id = id[:63]
    }
    
    return id
}

func isValidEmail(email string) bool {
    // Simple validation - consider using regex for production
    return strings.Contains(email, "@") && 
           len(email) > 3 && 
           len(email) < 254
}
```

**Add comprehensive tests:**

```go
// internal/db/db_test.go

func TestGenerateCustomerID_Sanitization(t *testing.T) {
    tests := []struct {
        email    string
        expected string
    }{
        {"user@example.com", "user-example-com"},
        {"User.Name@Example.COM", "user-name-example-com"},
        {"test+tag@domain.co.uk", "test-tag-domain-co-uk"},
        {"../../etc/passwd@example.com", "etc-passwd-example-com"},
        {"test<script>@example.com", "testscript-example-com"},
        {"a@b.co", "a-b-co"},
        {"very-long-email-that-exceeds-sixty-three-characters-limit@example.com", 
         "very-long-email-that-exceeds-sixty-three-characters"},
    }

    for _, tt := range tests {
        t.Run(tt.email, func(t *testing.T) {
            got := generateCustomerID(tt.email)
            if got != tt.expected {
                t.Errorf("generateCustomerID(%q) = %q, want %q", tt.email, got, tt.expected)
            }
        })
    }
}
```

**Acceptance Criteria:**
- [ ] Directory traversal attempts sanitized
- [ ] Special characters removed
- [ ] Length limited to 63 characters
- [ ] Invalid emails return empty string
- [ ] All test cases pass

---

### 1.6 Create Missing Template Files
**Priority:** CRITICAL | **Effort:** 3-4 hours | **Owner:** Backend Dev

**Directory:** `internal/workspace/templates/personal-assistant/`

**Files to create:**

**A. AGENTS.md.tmpl**
```markdown
# AGENTS.md

## Identity
You are {{.AssistantName}}, a personal AI assistant.

## User Profile
{{.UserDescription}}

## Core Responsibilities
{{.ResponsibilitiesList}}

## Communication Style
- Be concise but thorough
- Ask clarifying questions when needed
- Remember context from previous conversations
- Be proactive in suggesting helpful actions

## Boundaries
- Do not share personal information
- Do not make commitments on behalf of the user
- Escalate sensitive decisions to the user
```

**B. USER.md.tmpl**
```markdown
# USER.md

## Background
{{.CustomInstructions}}

## Preferences
- Tone: Professional yet friendly
- Detail level: Moderate (provide key points, offer details on request)
- Response time: Within 24 hours for non-urgent matters

## Contact Information
- Primary communication: Telegram
- Assistant name preference: {{.AssistantName}}

## Important Notes
This file contains personalized instructions for {{.AssistantName}} to provide context-aware assistance.
```

**C. SOUL.md.tmpl**
```markdown
# SOUL.md

## Persona
You are {{.AssistantName}}, dedicated to helping your user achieve their goals efficiently.

## Core Values
1. Reliability - Always follow through on commitments
2. Proactivity - Anticipate needs before they're expressed
3. Adaptability - Learn and adjust to user preferences over time
4. Discretion - Maintain confidentiality and privacy

## Behavioral Guidelines
- Greet users warmly but professionally
- Summarize long conversations when returning after absence
- Offer to help with recurring tasks proactively
- Keep responses focused on the user's stated needs

## Learning Priorities
- User's work style and preferences
- Common tasks and workflows
- Important contacts and relationships
- Preferred communication patterns
```

**Acceptance Criteria:**
- [ ] All three template files exist
- [ ] Files use Go template syntax correctly
- [ ] `go test ./internal/workspace` passes
- [ ] Generator successfully creates workspace files

---

## Phase 2: High Priority (Week 2) - PRODUCTION HARDENING

### 2.1 Implement Caddy RemoveSubdomain
**Priority:** HIGH | **Effort:** 4-5 hours | **Owner:** Backend Dev

**File:** `internal/caddy/caddy.go`

```go
func (c *Client) RemoveSubdomain(subdomain string) error {
    // First, get all routes
    url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes", c.adminURL)
    
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("get routes: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("get routes failed: %s", resp.Status)
    }

    // Parse routes
    var routes []Route
    if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
        return fmt.Errorf("decode routes: %w", err)
    }

    // Find route index by subdomain
    var routeIndex *int
    for i, route := range routes {
        for _, match := range route.Match {
            for _, host := range match.Host {
                if host == subdomain {
                    idx := i
                    routeIndex = &idx
                    break
                }
            }
            if routeIndex != nil {
                break
            }
        }
        if routeIndex != nil {
            break
        }
    }

    if routeIndex == nil {
        return fmt.Errorf("route not found for subdomain: %s", subdomain)
    }

    // Delete the specific route using index
    deleteURL := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes/%d", c.adminURL, *routeIndex)
    
    req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
    if err != nil {
        return fmt.Errorf("create delete request: %w", err)
    }

    resp, err = http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("delete route: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("delete route failed: %s", resp.Status)
    }

    return nil
}
```

**Acceptance Criteria:**
- [ ] Can successfully add and remove subdomains
- [ ] Proper error handling when subdomain doesn't exist
- [ ] Integration test with Caddy admin API

---

### 2.2 Add Rate Limiting
**Priority:** HIGH | **Effort:** 3-4 hours | **Owner:** Backend Dev

**Implementation:**

```bash
# Add dependency
go get github.com/ulule/limiter/v3
go get github.com/ulule/limiter/v3/drivers/store/memory
```

**File:** `internal/api/ratelimit.go` (new file)

```go
package api

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/ulule/limiter/v3"
    "github.com/ulule/limiter/v3/drivers/store/memory"
)

func rateLimitMiddleware(rate limiter.Rate) gin.HandlerFunc {
    store := memory.NewStore()
    instance := limiter.New(store, rate)

    return func(c *gin.Context) {
        context, err := instance.Get(c, c.ClientIP())
        if err != nil {
            c.AbortWithStatus(http.StatusInternalServerError)
            return
        }

        c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
        c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
        c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

        if context.Reached {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "rate_limit_exceeded",
                "message": "Too many requests. Please try again later.",
            })
            return
        }

        c.Next()
    }
}

// Specific rate limits
func signupRateLimit() gin.HandlerFunc {
    // 5 requests per minute per IP
    return rateLimitMiddleware(limiter.Rate{
        Period: 1 * time.Minute,
        Limit:  5,
    })
}

func webhookRateLimit() gin.HandlerFunc {
    // 100 requests per minute (Stripe can send bursts)
    return rateLimitMiddleware(limiter.Rate{
        Period: 1 * time.Minute,
        Limit:  100,
    })
}
```

**Update router.go:**

```go
// router.go - Apply rate limiting
router.POST("/api/signup", signupRateLimit(), handler.CreateCustomer)
router.POST("/api/webhook/stripe", webhookRateLimit(), stripeWebhook.HandleWebhook)
```

**Acceptance Criteria:**
- [ ] Signup endpoint limited to 5 req/min per IP
- [ ] Webhook endpoint limited to 100 req/min
- [ ] Proper 429 responses with retry headers
- [ ] Rate limit headers in responses

---

### 2.3 Migrate to Structured Logging (Zap)
**Priority:** HIGH | **Effort:** 6-8 hours | **Owner:** Backend Dev

This is a cross-cutting change affecting all packages.

```bash
go get go.uber.org/zap
```

**Implementation Strategy:**

1. **Update main.go:**
```go
import "go.uber.org/zap"

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // Pass logger to all components
}
```

2. **Update config to use zap:**
```go
type Config struct {
    Logger *zap.Logger  // Inject logger
}
```

3. **Update each package gradually:**

Example in handler.go:
```go
// Before:
h.logger.Printf("Failed to create customer: %v", err)

// After:
h.logger.Error("failed to create customer",
    zap.Error(err),
    zap.String("email", req.Email),
    zap.String("operation", "CreateCustomer"),
)
```

**Acceptance Criteria:**
- [ ] All log.Printf replaced with structured logging
- [ ] Logs include operation context
- [ ] Error logs include stack traces
- [ ] No sensitive data in logs

---

### 2.4 Enhance Health Check Endpoint
**Priority:** HIGH | **Effort:** 3-4 hours | **Owner:** Backend Dev

**File:** `internal/api/handler.go`

```go
func (h *Handler) HealthCheck(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Check database
    dbHealthy := h.checkDatabase(ctx)
    
    // Check Docker (if provisioner available)
    dockerHealthy := h.checkDocker(ctx)
    
    allHealthy := dbHealthy && dockerHealthy
    
    status := http.StatusOK
    if !allHealthy {
        status = http.StatusServiceUnavailable
    }
    
    c.JSON(status, gin.H{
        "status":  map[bool]string{true: "healthy", false: "unhealthy"}[allHealthy],
        "version": "1.0.0",
        "checks": gin.H{
            "database": map[string]interface{}{
                "status": map[bool]string{true: "pass", false: "fail"}[dbHealthy],
            },
            "docker": map[string]interface{}{
                "status": map[bool]string{true: "pass", false: "fail"}[dockerHealthy],
            },
        },
        "timestamp": time.Now().UTC(),
    })
}

func (h *Handler) checkDatabase(ctx context.Context) bool {
    if h.db == nil {
        return false
    }
    // Try a simple query
    _, err := h.db.CountActiveCustomers(ctx)
    return err == nil
}

func (h *Handler) checkDocker(ctx context.Context) bool {
    if h.provisioner == nil {
        return false
    }
    // Implementation depends on provisioner interface
    return true
}
```

---

## Phase 3: Medium Priority (Week 3) - OPTIMIZATION

### 3.1 Add Database Indexes
**File:** `internal/db/db.go`

Add to Migrate():
```go
`CREATE INDEX IF NOT EXISTS idx_customers_stripe_subscription ON customers(stripe_subscription_id)`,
`CREATE INDEX IF NOT EXISTS idx_customers_stripe_session ON customers(stripe_checkout_session_id)`,
`CREATE INDEX IF NOT EXISTS idx_customers_container_port ON customers(container_port)`,
```

### 3.2 Add Circuit Breakers
**New Package:** `internal/circuitbreaker/`

Use library: `github.com/sony/gobreaker`

### 3.3 Add Request Validation Middleware
Extract validation logic from handler.go into reusable middleware.

---

## Phase 4: Polish (Week 4) - COVERAGE & REFACTORING

### 4.1 Increase Test Coverage
- Target: provisioner 70%, api 80%
- Add integration tests for webhook flows
- Add unit tests for edge cases

### 4.2 Dependency Injection Refactoring
Refactor to use interfaces:
```go
type Provisioner interface {
    Provision(ctx context.Context, customerID string) error
    Suspend(ctx context.Context, customerID string) error
    Resume(ctx context.Context, customerID string) error
    Terminate(ctx context.Context, customerID string) error
}
```

### 4.3 Documentation
- Add comprehensive README.md
- Document API endpoints
- Create deployment guide

---

## Dependencies & Prerequisites

### External Dependencies to Add:
```bash
go get github.com/ulule/limiter/v3
go get go.uber.org/zap
go get github.com/sony/gobreaker  # For circuit breakers
```

### Required Files to Create:
1. `internal/workspace/templates/personal-assistant/AGENTS.md.tmpl`
2. `internal/workspace/templates/personal-assistant/USER.md.tmpl`
3. `internal/workspace/templates/personal-assistant/SOUL.md.tmpl`
4. `internal/api/ratelimit.go`
5. `internal/circuitbreaker/` (new package)

---

## Testing Strategy

### Unit Tests
- Run after each Phase 1 task: `go test ./...`
- Coverage check: `go test -cover ./...`

### Integration Tests
- Run after Phase 2: `go test -tags=integration ./internal/e2e/`

### Load Tests
- Run after Phase 3: Test concurrent signup requests

### Security Tests
- Run after Phase 1: Verify API keys not in compose files
- Run after Phase 2: Verify rate limiting works

---

## Acceptance Criteria by Phase

### Phase 1 (Production Blockers)
- [ ] All tests pass
- [ ] No compilation errors
- [ ] API keys not exposed in files
- [ ] Port allocation thread-safe
- [ ] Customer IDs sanitized
- [ ] Template files exist

### Phase 2 (Hardening)
- [ ] Caddy routes removable
- [ ] Rate limiting active
- [ ] Structured logging implemented
- [ ] Health checks comprehensive

### Phase 3 (Optimization)
- [ ] Database performance optimized
- [ ] External services resilient
- [ ] Input validation comprehensive

### Phase 4 (Polish)
- [ ] Test coverage meets targets
- [ ] Documentation complete
- [ ] Code refactored to interfaces

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking changes | Medium | High | Comprehensive tests before each phase |
| Performance regression | Low | Medium | Benchmark tests |
| Security vulnerability | Low | Critical | Security review after Phase 1 |
| Scope creep | High | Medium | Strict phase boundaries |

---

## Success Metrics

1. **Test Success Rate:** 100% tests passing
2. **Code Coverage:** provisioner 70%+, api 80%+
3. **Security Scan:** Zero critical vulnerabilities
4. **Performance:** p95 response time < 200ms
5. **Reliability:** Zero data integrity issues

---

**Next Steps:**
1. Review and approve this plan
2. Create tickets for Phase 1 tasks
3. Begin with test suite fixes (1.1)
4. Schedule daily standups during Phase 1
5. Deploy to staging after Phase 1 complete

**Estimated Total Effort:** 3-4 weeks (1-2 developers)  
**Go-Live Target:** End of Week 2 (after Phase 2) with Phase 3+ as fast-follow

---

*Plan Version: 1.0*  
*Last Updated: 2026-02-19*
