---
type: plan
title: Blytz Cloud Multi-Agent Integration Roadmap
resource: blytz-cloud
description: "**Status:** Architecture complete ➜ Integration pending"
tags: [go, planning, react, stripe]
updated: 2026-06-18
---

# Blytz Cloud Multi-Agent Integration Roadmap

## 📋 Summary

This document provides the integration plan for adding multi-agent orchestration capabilities to Blytz Cloud. The architecture files and core implementation have been created - this guide explains how to integrate them into the existing codebase.

**Status:** Architecture complete ➜ Integration pending

---

## 📁 What Was Created

### New Packages (Ready to Use)

```
internal/
├── orchestrator/
│   └── orchestrator.go      # Request router & agent manager
├── agents/
│   └── base.go              # Base agent + specialized agents
├── costs/
│   └── tracker.go           # Budget tracking & circuit breakers
└── messagebus/
    └── (to be created)      # Inter-agent communication
```

### Documentation

```
docs/
├── MULTI_AGENT_ARCHITECTURE.md    # Full system architecture
├── IMPLEMENTATION_SUMMARY.md      # Quick integration guide
└── INTEGRATION_ROADMAP.md         # This file
```

### Example Code

```
cmd/example/main.go              # Working demonstration
```

---

## 🎯 Integration Plan

### Phase 1: Foundation (Week 1)

#### Step 1: Install Dependencies

```bash
cd /root/blytz-cloud

# Core dependencies
go get github.com/google/uuid
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/trace

# If not already present
go get github.com/gin-gonic/gin
go get go.uber.org/zap
```

#### Step 2: Create LLM Client Interface

Create file: `internal/llm/client.go`

```go
package llm

import "context"

// Client interface for LLM operations
type Client interface {
    SimpleChat(ctx context.Context, systemPrompt, userMessage string) (string, error)
}

// Implement this interface using your existing LLM provider code
// Wrap your current provider implementation
```

**Action Required:** 
- Create the interface
- Wrap your existing LLM provider to implement this interface
- Update `internal/config/config.go` to support LLM provider configuration

#### Step 3: Update Database Schema

Run these migrations:

```sql
-- migrations/003_multi_agent.up.sql

-- Agent instances table
CREATE TABLE agent_instances (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    agent_type TEXT NOT NULL, -- 'screening', 'matching', 'pricing', 'support'
    status TEXT NOT NULL DEFAULT 'active', -- 'active', 'paused', 'error'
    config JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);

-- Agent conversations table
CREATE TABLE agent_conversations (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    agent_instance_id TEXT NOT NULL,
    request_content TEXT NOT NULL,
    response_content TEXT,
    cost_usd DECIMAL(10,4) DEFAULT 0,
    tokens_used INTEGER DEFAULT 0,
    latency_ms INTEGER,
    agent_name TEXT,
    confidence DECIMAL(3,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_instance_id) REFERENCES agent_instances(id)
);

-- Cost tracking table
CREATE TABLE daily_costs (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    date DATE NOT NULL,
    total_cost_usd DECIMAL(10,4) DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    request_count INTEGER DEFAULT 0,
    UNIQUE(customer_id, date)
);

-- Human approval requests table
CREATE TABLE approval_requests (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    agent_instance_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    action_details JSON,
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'rejected', 'expired'
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP,
    responded_by TEXT,
    FOREIGN KEY (agent_instance_id) REFERENCES agent_instances(id)
);

-- Indexes for performance
CREATE INDEX idx_agent_instances_customer ON agent_instances(customer_id);
CREATE INDEX idx_agent_conversations_customer ON agent_conversations(customer_id);
CREATE INDEX idx_agent_conversations_agent ON agent_conversations(agent_instance_id);
CREATE INDEX idx_daily_costs_customer_date ON daily_costs(customer_id, date);
```

**Action Required:**
- Create migration file
- Add migration execution to database initialization
- Update `internal/db/db.go` with new CRUD operations

---

### Phase 2: Core Integration (Week 2)

#### Step 4: Update Handler

Modify: `internal/api/handler.go`

```go
type Handler struct {
    db           *db.DB
    provisioner  provisioner.Provisioner
    stripe       *stripe.Service
    orchestrator *orchestrator.Orchestrator  // ADD THIS
    costTracker  *costs.Tracker              // ADD THIS
    cfg          *config.Config
    logger       *zap.Logger
}

// Update constructor
func NewHandler(
    database *db.DB, 
    prov provisioner.Provisioner, 
    stripeSvc *stripe.Service,
    orch *orchestrator.Orchestrator,      // ADD THIS
    costTracker *costs.Tracker,           // ADD THIS
    cfg *config.Config, 
    logger *zap.Logger,
) *Handler {
    return &Handler{
        db:           database,
        provisioner:  prov,
        stripe:       stripeSvc,
        orchestrator: orch,                 // ADD THIS
        costTracker:  costTracker,          // ADD THIS
        cfg:          cfg,
        logger:       logger,
    }
}
```

**Action Required:**
- Add orchestrator and cost tracker to Handler struct
- Update NewHandler constructor
- Update all places where NewHandler is called

#### Step 5: Add New API Endpoints

Add to: `internal/api/handler.go` (or create new file: `internal/api/agent_handlers.go`)

```go
// ProcessAgentRequest handles requests through the multi-agent system
func (h *Handler) ProcessAgentRequest(c *gin.Context) {
    customerID := c.GetString("customer_id")
    if customerID == "" {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }
    
    var req struct {
        Content       string                 `json:"content" binding:"required"`
        MaxCostUSD    float64               `json:"max_cost_usd"`
        Context       map[string]interface{} `json:"context"`
        RequireApproval bool                `json:"require_approval"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request: " + err.Error()})
        return
    }
    
    // Create request
    request := orchestrator.Request{
        ID:              generateID(),
        CustomerID:      customerID,
        Content:         req.Content,
        Context:         req.Context,
        MaxCostUSD:      req.MaxCostUSD,
        RequireApproval: req.RequireApproval,
        Timeout:         30 * time.Second,
    }
    
    // Process through orchestrator
    resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), request)
    if err != nil {
        h.logger.Error("Agent request failed", zap.Error(err))
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "id":              resp.ID,
        "content":         resp.Content,
        "agent_name":      resp.AgentName,
        "confidence":      resp.Confidence,
        "cost_usd":        resp.CostUSD,
        "tokens_used":     resp.TokensUsed,
        "latency_ms":      resp.LatencyMs,
        "human_approved":  resp.HumanApproved,
    })
}

// GetAgentStatus returns status of all agents
func (h *Handler) GetAgentStatus(c *gin.Context) {
    agents := h.orchestrator.GetAgents()
    
    var status []gin.H
    for _, agent := range agents {
        caps := agent.GetCapabilities()
        var capNames []string
        for _, cap := range caps {
            capNames = append(capNames, cap.Name)
        }
        
        status = append(status, gin.H{
            "name":         agent.Name(),
            "capabilities": capNames,
        })
    }
    
    c.JSON(200, gin.H{"agents": status})
}

// GetCostStatus returns current cost tracking status
func (h *Handler) GetCostStatus(c *gin.Context) {
    customerID := c.GetString("customer_id")
    if customerID == "" {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }
    
    spend, err := h.costTracker.GetCustomerSpend(customerID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "daily_spend":     spend.SpentToday,
        "monthly_spend":   spend.SpentThisMonth,
        "session_spend":   spend.SessionSpend,
        "request_count":   spend.RequestCount,
        "token_count":     spend.TokenCount,
    })
}

// GetApprovalRequests returns pending approval requests
func (h *Handler) GetApprovalRequests(c *gin.Context) {
    customerID := c.GetString("customer_id")
    // Query database for pending approvals
    // Return list
}

// ApproveRequest approves a pending request
func (h *Handler) ApproveRequest(c *gin.Context) {
    // Handle approval/rejection
}
```

#### Step 6: Update Router

Modify: `internal/api/router.go`

```go
func SetupRoutes(r *gin.Engine, h *Handler, logger *zap.Logger) {
    // Existing routes...
    
    // Health check (no auth required)
    r.GET("/api/health", h.HealthCheck)
    
    // Public routes
    r.POST("/api/signup", h.RateLimitMiddleware(5, time.Minute), h.Signup)
    r.GET("/api/status/:id", h.GetStatus)
    r.POST("/api/webhook/stripe", h.RateLimitMiddleware(100, time.Minute), h.StripeWebhook)
    
    // Authenticated routes
    api := r.Group("/api")
    api.Use(AuthMiddleware())
    {
        // Existing routes...
        
        // NEW: Multi-agent routes
        api.POST("/agent/request", h.ProcessAgentRequest)
        api.GET("/agent/status", h.GetAgentStatus)
        api.GET("/agent/costs", h.GetCostStatus)
        api.GET("/agent/approvals", h.GetApprovalRequests)
        api.POST("/agent/approvals/:id", h.ApproveRequest)
    }
}
```

---

### Phase 3: Main Integration (Week 2-3)

#### Step 7: Initialize Orchestrator in Main

Modify: `cmd/server/main.go`

```go
func main() {
    // ... existing setup ...
    
    // Initialize database
    database, err := db.New(cfg.DatabasePath)
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }
    
    // NEW: Initialize orchestrator
    orchestrator := setupOrchestrator(cfg, logger)
    
    // NEW: Initialize cost tracker
    costTracker := setupCostTracker(cfg)
    
    // Initialize handler WITH orchestrator
    handler := api.NewHandler(
        database, 
        prov, 
        stripeSvc,
        orchestrator,    // NEW
        costTracker,     // NEW
        cfg, 
        logger,
    )
    
    // Setup routes
    api.SetupRoutes(router, handler, logger)
    
    // ... rest of server setup ...
}

func setupOrchestrator(cfg *config.Config, logger *zap.Logger) *orchestrator.Orchestrator {
    // Create orchestrator
    orch := orchestrator.New(orchestrator.Config{
        Strategy:      orchestrator.StrategyConfidence,
        MinConfidence: 0.6,
        MaxCostUSD:    10.0,
        Logger:        logger,
    })
    
    // Create LLM client
    llmClient := createLLMClient(cfg)
    
    // Register agents
    orch.RegisterAgent(agents.NewScreeningAgent(llmClient))
    orch.RegisterAgent(agents.NewMatchingAgent(llmClient))
    orch.RegisterAgent(agents.NewSupportAgent(llmClient))
    // Add more agents as needed...
    
    logger.Info("Orchestrator initialized", 
        zap.Int("agents", len(orch.GetAgents())))
    
    return orch
}

func setupCostTracker(cfg *config.Config) *costs.Tracker {
    tracker := costs.NewTracker()
    
    // Set default budget
    tracker.SetBudget("default", costs.BudgetConfig{
        DailyBudgetUSD:   10.00,
        SessionBudgetUSD: 2.00,
        RequestBudgetUSD: 0.50,
        WarningThreshold: 0.8,
    })
    
    return tracker
}

func createLLMClient(cfg *config.Config) llm.Client {
    // Wrap your existing LLM provider
    // Return a client that implements llm.Client interface
}
```

#### Step 8: Update Configuration

Add to: `internal/config/config.go`

```go
type Config struct {
    // ... existing config ...
    
    // Agent Orchestrator
    AgentStrategy      string  `env:"AGENT_STRATEGY" envDefault:"confidence"`
    AgentMinConfidence float64 `env:"AGENT_MIN_CONFIDENCE" envDefault:"0.6"`
    AgentMaxCostUSD    float64 `env:"AGENT_MAX_COST_USD" envDefault:"10.0"`
    
    // Cost Tracking
    DailyBudgetUSD   float64 `env:"DAILY_BUDGET_USD" envDefault:"10.0"`
    SessionBudgetUSD float64 `env:"SESSION_BUDGET_USD" envDefault:"2.0"`
    RequestBudgetUSD float64 `env:"REQUEST_BUDGET_USD" envDefault:"0.5"`
}
```

---

### Phase 4: Frontend Integration (Week 3)

#### Step 9: Create Agent Dashboard UI

New file: `frontend/src/app/dashboard/agents/page.tsx`

```tsx
'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export default function AgentsPage() {
  const [agents, setAgents] = useState([])
  const [costs, setCosts] = useState(null)
  const [requestContent, setRequestContent] = useState('')
  const [response, setResponse] = useState(null)
  const [loading, setLoading] = useState(false)

  // Fetch agents and costs on mount
  useEffect(() => {
    fetchAgents()
    fetchCosts()
  }, [])

  const fetchAgents = async () => {
    const res = await fetch('/api/agent/status')
    const data = await res.json()
    setAgents(data.agents)
  }

  const fetchCosts = async () => {
    const res = await fetch('/api/agent/costs')
    const data = await res.json()
    setCosts(data)
  }

  const sendRequest = async () => {
    setLoading(true)
    const res = await fetch('/api/agent/request', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        content: requestContent,
        max_cost_usd: 1.00
      })
    })
    const data = await res.json()
    setResponse(data)
    setLoading(false)
    fetchCosts() // Refresh costs
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">AI Agents</h1>
      
      {/* Cost Overview */}
      {costs && (
        <div className="grid grid-cols-4 gap-4">
          <Card>
            <CardHeader>
              <CardTitle>Daily Spend</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold">${costs.daily_spend.toFixed(2)}</p>
            </CardContent>
          </Card>
          {/* More cost cards... */}
        </div>
      )}

      {/* Agent List */}
      <Card>
        <CardHeader>
          <CardTitle>Available Agents</CardTitle>
        </CardHeader>
        <CardContent>
          {agents.map(agent => (
            <div key={agent.name} className="flex items-center space-x-4 py-2">
              <div className="font-medium">{agent.name}</div>
              <div className="text-sm text-gray-500">
                {agent.capabilities.join(', ')}
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Request Form */}
      <Card>
        <CardHeader>
          <CardTitle>Send Request</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Input
              placeholder="What do you need help with?"
              value={requestContent}
              onChange={(e) => setRequestContent(e.target.value)}
            />
            <Button onClick={sendRequest} disabled={loading}>
              {loading ? 'Processing...' : 'Send Request'}
            </Button>
          </div>
          
          {response && (
            <div className="mt-4 p-4 bg-gray-50 rounded">
              <p><strong>Response:</strong> {response.content}</p>
              <p><strong>Agent:</strong> {response.agent_name}</p>
              <p><strong>Cost:</strong> ${response.cost_usd.toFixed(4)}</p>
              <p><strong>Confidence:</strong> {(response.confidence * 100).toFixed(0)}%</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
```

---

## 📊 Testing Checklist

### Unit Tests to Create

```go
// internal/orchestrator/orchestrator_test.go
func TestOrchestrator_RegisterAgent(t *testing.T)
func TestOrchestrator_ProcessRequest(t *testing.T)
func TestOrchestrator_SelectAgent(t *testing.T)
func TestOrchestrator_CostBudget(t *testing.T)

// internal/agents/base_test.go
func TestScreeningAgent_CanHandle(t *testing.T)
func TestScreeningAgent_Execute(t *testing.T)
func TestMatchingAgent_CanHandle(t *testing.T)

// internal/costs/tracker_test.go
func TestTracker_CheckBudget(t *testing.T)
func TestTracker_RecordUsage(t *testing.T)
func TestTracker_AlertThreshold(t *testing.T)
```

### Integration Tests

```go
// internal/api/agent_handlers_test.go
func TestProcessAgentRequest_Success(t *testing.T)
func TestProcessAgentRequest_BudgetExceeded(t *testing.T)
func TestGetAgentStatus(t *testing.T)
func TestGetCostStatus(t *testing.T)
```

### Manual Testing Steps

1. **Start server**: `go run ./cmd/server`
2. **Check health**: `curl http://localhost:8080/api/health`
3. **Signup**: Create a customer via `/api/signup`
4. **Test agent request**: 
   ```bash
   curl -X POST http://localhost:8080/api/agent/request \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <token>" \
     -d '{"content": "Find me a React developer"}'
   ```
5. **Check costs**: `curl http://localhost:8080/api/agent/costs`
6. **Check agents**: `curl http://localhost:8080/api/agent/status`

---

## 🚀 Deployment Considerations

### Environment Variables to Add

```bash
# Agent Orchestrator
AGENT_STRATEGY=confidence
AGENT_MIN_CONFIDENCE=0.6
AGENT_MAX_COST_USD=10.0

# Cost Tracking
DAILY_BUDGET_USD=10.0
SESSION_BUDGET_USD=2.0
REQUEST_BUDGET_USD=0.5
```

### Monitoring

Add these metrics:

```go
// Metrics to track
- agent_requests_total
- agent_request_duration_seconds
- agent_cost_usd_total
- agent_confidence_score
- agent_selection_count
```

---

## 📈 Success Metrics

Track these after integration:

| Metric | Target | Measurement |
|--------|--------|-------------|
| Cost Reduction | 20% | Compare token costs before/after |
| Response Quality | 30% improvement | User satisfaction scores |
| Routing Accuracy | >90% | Correct agent selection rate |
| Budget Compliance | 100% | No over-budget incidents |

---

## 🐛 Common Issues & Solutions

### Issue 1: LLM Client Interface Mismatch
**Problem:** Your existing LLM code doesn't match the interface
**Solution:** Create an adapter/wrapper

### Issue 2: Database Migration Failures
**Problem:** Migrations don't run or fail
**Solution:** Test migrations in isolation first

### Issue 3: Circular Dependencies
**Problem:** Packages importing each other
**Solution:** Extract interfaces to separate package

### Issue 4: High Latency
**Problem:** Agent selection adds latency
**Solution:** Cache confidence scores, use async LLM calls

---

## 📝 Next Steps (In Order)

### Immediate (Today)
- [ ] Review this roadmap
- [ ] Decide on integration approach (A, B, or C)
- [ ] Install dependencies

### This Week
- [ ] Create LLM client interface
- [ ] Run database migrations
- [ ] Update Handler struct
- [ ] Add new API endpoints

### Next Week
- [ ] Wire up orchestrator in main.go
- [ ] Test basic functionality
- [ ] Add frontend UI
- [ ] Write tests

### Following Week
- [ ] Production testing
- [ ] Monitor costs
- [ ] Add more specialized agents
- [ ] Optimize performance

---

## 📚 References

- **Architecture:** `docs/MULTI_AGENT_ARCHITECTURE.md`
- **Implementation:** `docs/IMPLEMENTATION_SUMMARY.md`
- **Example:** `cmd/example/main.go`
- **Core Code:** `internal/orchestrator/orchestrator.go`

---

## 💡 Tips

1. **Start Simple**: Get basic routing working first, then add features
2. **Test Early**: Run the example code to understand the flow
3. **Monitor Costs**: Watch token usage closely during testing
4. **Iterate**: Don't try to integrate everything at once
5. **Document**: Keep notes on what works and what doesn't

---

**Questions?** Refer to the architecture docs or the example code. The multi-agent system is ready - you just need to wire it into your existing infrastructure!

**Estimated Time to Complete:** 2-3 weeks (part-time)
