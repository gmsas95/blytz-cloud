---
type: architecture
title: Multi-Agent Orchestration Architecture for Blytz Cloud
resource: blytz-cloud
description: "**Current Architecture:** Single agent per container (Layer 2)"
tags: [architecture, go, react]
updated: 2026-06-18
---

# Multi-Agent Orchestration Architecture for Blytz Cloud

## Current State Analysis

**Current Architecture:** Single agent per container (Layer 2)
- User gets 1 container slot
- Container runs 1 agent instance (OpenClaw/Myrai)
- Agent handles all tasks

**Target Architecture:** Multi-agent orchestration within container (Layer 3)
- User gets 1 container slot
- Container runs orchestrator + N specialized agents
- Agents collaborate on complex tasks

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    USER CONTAINER                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐   │
│  │           AGENT ORCHESTRATOR (Go)                  │   │
│  │  - Route requests to appropriate agent             │   │
│  │  - Manage agent lifecycle                          │   │
│  │  - Handle inter-agent communication                │   │
│  │  - Track costs & enforce limits                    │   │
│  └────────────────┬────────────────────────────────────┘   │
│                   │                                          │
│       ┌───────────┼───────────┬──────────────┐              │
│       ▼           ▼           ▼              ▼              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐  ┌─────────────┐       │
│  │ Screening│ │ Matching│ │ Pricing │  │  Support    │       │
│  │  Agent   │ │  Agent  │ │  Agent  │  │   Agent     │       │
│  └─────────┘ └─────────┘ └─────────┘  └─────────────┘       │
│       │           │           │              │              │
│       └───────────┴───────────┴──────────────┘              │
│                   │                                         │
│                   ▼                                         │
│         ┌─────────────────┐                                 │
│         │  Shared State   │                                 │
│         │   (BadgerDB)    │                                 │
│         └─────────────────┘                                 │
└─────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. Agent Orchestrator

**Responsibilities:**
- Request classification (which agent should handle this?)
- Agent lifecycle management (start/stop agents on demand)
- Cost tracking per agent
- Human approval gates
- Failure handling & retries

**Key Interfaces:**

```go
// Agent interface that all specialized agents implement
type Agent interface {
    Name() string
    CanHandle(ctx context.Context, request Request) (float64, error) // confidence score
    Execute(ctx context.Context, request Request) (Response, error)
    GetCapabilities() []Capability
    GetCostEstimate(request Request) (CostEstimate, error)
}

// Orchestrator manages multiple agents
type Orchestrator struct {
    agents        map[string]Agent
    stateStore    StateStore
    costTracker   CostTracker
    messageBus    MessageBus
    logger        *zap.Logger
    mu            sync.RWMutex
}
```

### 2. Specialized Agents

Based on your talent marketplace use case:

```go
// ScreeningAgent - vets talent profiles
type ScreeningAgent struct {
    llmClient     *llm.Client
    criteria      []ScreeningCriteria
    minConfidence float64
}

// MatchingAgent - pairs talent with clients
type MatchingAgent struct {
    llmClient     *llm.Client
    vectorDB      VectorDB
    matchThreshold float64
}

// PricingAgent - suggests rates
type PricingAgent struct {
    llmClient     *llm.Client
    marketData    MarketDataProvider
}

// SupportAgent - handles customer inquiries
type SupportAgent struct {
    llmClient     *llm.Client
    kbSearch      KnowledgeBase
}
```

### 3. Message Bus for Inter-Agent Communication

```go
type MessageBus interface {
    Publish(ctx context.Context, msg Message) error
    Subscribe(ctx context.Context, agentID string, handler MessageHandler) error
    Request(ctx context.Context, targetAgent string, req Request) (Response, error)
}

type Message struct {
    ID          string
    From        string
    To          string
    Type        MessageType // request, response, event
    Content     interface{}
    Timestamp   time.Time
    CorrelationID string // for tracking request chains
}
```

### 4. Cost Tracking & Circuit Breakers

```go
type CostTracker struct {
    dailyBudget   float64
    sessionBudget float64
    spentToday    float64
    mu            sync.RWMutex
    alerts        chan CostAlert
}

type CostAlert struct {
    Level     AlertLevel // warning, critical
    Amount    float64
    Limit     float64
    Action    ActionType // notify, throttle, halt
}

func (ct *CostTracker) CheckBudget(cost float64) error {
    ct.mu.Lock()
    defer ct.mu.Unlock()
    
    ct.spentToday += cost
    
    if ct.spentToday > ct.dailyBudget {
        return fmt.Errorf("daily budget exceeded: $%.2f / $%.2f", 
            ct.spentToday, ct.dailyBudget)
    }
    
    if ct.spentToday > ct.dailyBudget*0.8 {
        ct.alerts <- CostAlert{
            Level:  Warning,
            Amount: ct.spentToday,
            Limit:  ct.dailyBudget,
            Action: Notify,
        }
    }
    
    return nil
}
```

### 5. Human-in-the-Loop Gates

```go
type HumanGate struct {
    approvalRequired bool
    pendingApprovals map[string]ApprovalRequest
    mu               sync.RWMutex
}

type ApprovalRequest struct {
    ID          string
    AgentName   string
    Action      string
    Details     interface{}
    RequestedAt time.Time
    Status      ApprovalStatus // pending, approved, rejected
    ApprovedBy  string
    ApprovedAt  *time.Time
}

func (hg *HumanGate) RequestApproval(ctx context.Context, req ApprovalRequest) (bool, error) {
    // Store approval request
    // Send notification (email, Telegram, etc.)
    // Wait for human response (with timeout)
    // Return result
}
```

## Request Flow Example

```
User: "Find me a React developer"

1. Request arrives at Orchestrator
2. Orchestrator asks all agents "Can you handle this?"
   - ScreeningAgent: 0.3 (not screening-specific)
   - MatchingAgent: 0.95 (this is my job!)
   - PricingAgent: 0.1 (not relevant)
   - SupportAgent: 0.2 (not a support question)
3. Orchestrator routes to MatchingAgent
4. MatchingAgent breaks task into steps:
   a. Search talent database
   b. Filter by React skill
   c. Check availability
   d. Return top 3 matches
5. Each step may invoke other agents via MessageBus
6. Final response returned to user
7. Cost tracked, logged
```

## Implementation Phases

### Phase 1: Foundation (Week 1-2)

**Goals:**
- Implement Agent interface
- Create basic Orchestrator
- Add message bus (in-memory first)

**Files to Create:**
```
internal/
├── agent/
│   ├── orchestrator.go      # Main orchestrator
│   ├── registry.go          # Agent registration
│   └── router.go            # Request routing
├── agents/
│   ├── base.go              # Base agent implementation
│   ├── screening.go         # ScreeningAgent
│   ├── matching.go          # MatchingAgent
│   └── support.go           # SupportAgent
├── messagebus/
│   ├── bus.go               # MessageBus interface
│   └── memory.go            # In-memory implementation
└── costs/
    ├── tracker.go           # Cost tracking
    └── limits.go            # Budget limits
```

### Phase 2: Specialized Agents (Week 3-4)

**Goals:**
- Implement 2-3 specialized agents
- Add inter-agent communication
- Basic cost tracking

**Example Screening Agent:**

```go
package agents

type ScreeningAgent struct {
    BaseAgent
    llmClient *llm.Client
    minScore  float64
}

func NewScreeningAgent(llm *llm.Client) *ScreeningAgent {
    return &ScreeningAgent{
        BaseAgent: BaseAgent{name: "screening"},
        llmClient: llm,
        minScore:  0.7,
    }
}

func (sa *ScreeningAgent) CanHandle(ctx context.Context, req agent.Request) (float64, error) {
    // Use LLM to determine if this is a screening request
    prompt := fmt.Sprintf(`
        Is this a talent screening request? 
        Request: "%s"
        Respond with just a number 0-1 indicating confidence.
    `, req.Content)
    
    resp, err := sa.llmClient.SimpleChat(ctx, "", prompt)
    if err != nil {
        return 0, err
    }
    
    score, _ := strconv.ParseFloat(strings.TrimSpace(resp), 64)
    return score, nil
}

func (sa *ScreeningAgent) Execute(ctx context.Context, req agent.Request) (agent.Response, error) {
    // 1. Parse screening criteria from request
    // 2. Query talent database
    // 3. Score each candidate
    // 4. Return results
}
```

### Phase 3: Production Features (Week 5-6)

**Goals:**
- Human approval gates
- Cost circuit breakers
- Observability (tracing, metrics)
- State persistence

**Observability Setup:**

```go
// OpenTelemetry tracing
import "go.opentelemetry.io/otel"

func (o *Orchestrator) ProcessRequest(ctx context.Context, req Request) (Response, error) {
    ctx, span := tracer.Start(ctx, "orchestrator.process_request")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("request.id", req.ID),
        attribute.String("request.content", req.Content),
    )
    
    // Route to agent
    agent, err := o.selectAgent(ctx, req)
    if err != nil {
        span.RecordError(err)
        return Response{}, err
    }
    
    span.SetAttributes(attribute.String("agent.selected", agent.Name()))
    
    // Execute
    resp, err := agent.Execute(ctx, req)
    if err != nil {
        span.RecordError(err)
        return Response{}, err
    }
    
    span.SetAttributes(
        attribute.Int("response.tokens", resp.TokensUsed),
        attribute.Float64("response.cost", resp.Cost),
    )
    
    return resp, nil
}
```

## Database Schema Updates

```sql
-- Agent instances table
CREATE TABLE agent_instances (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    agent_type TEXT NOT NULL, -- 'screening', 'matching', 'pricing', etc.
    status TEXT NOT NULL, -- 'active', 'paused', 'error'
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
    cost_usd DECIMAL(10,4),
    tokens_used INTEGER,
    latency_ms INTEGER,
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
    status TEXT NOT NULL, -- 'pending', 'approved', 'rejected', 'expired'
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP,
    responded_by TEXT,
    FOREIGN KEY (agent_instance_id) REFERENCES agent_instances(id)
);
```

## Configuration

```yaml
# agent-orchestrator.yaml
orchestrator:
  default_agent: "support"
  
  routing:
    strategy: "confidence_threshold" # or "round_robin", "cost_optimized"
    min_confidence: 0.6
    
  costs:
    daily_budget_usd: 10.00
    session_budget_usd: 2.00
    alert_threshold: 0.8
    
  human_gates:
    - action: "talent_hire"
      threshold: "always" # always require approval
    - action: "pricing_change"
      threshold: "cost > 5.00"
    - action: "data_export"
      threshold: "always"
      
  agents:
    screening:
      enabled: true
      min_confidence: 0.7
      max_cost_per_request: 0.50
      
    matching:
      enabled: true
      min_confidence: 0.8
      max_cost_per_request: 1.00
      
    pricing:
      enabled: true
      min_confidence: 0.6
      max_cost_per_request: 0.25
```

## Testing Strategy

```go
// Test agent routing
func TestOrchestrator_RouteToScreeningAgent(t *testing.T) {
    // Setup
    orchestrator := NewOrchestrator(...)
    screeningAgent := NewScreeningAgent(mockLLM)
    orchestrator.RegisterAgent(screeningAgent)
    
    // Test request
    req := Request{
        Content: "Screen this React developer profile",
    }
    
    // Execute
    resp, err := orchestrator.ProcessRequest(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "screening", resp.AgentName)
}

// Test cost limits
func TestCostTracker_BudgetExceeded(t *testing.T) {
    tracker := NewCostTracker(10.00, 2.00)
    tracker.spentToday = 9.50
    
    err := tracker.CheckBudget(1.00)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "budget exceeded")
}
```

## Migration Path

### From Single Agent to Multi-Agent

1. **Phase 1: Side-by-side**
   - Keep existing OpenClaw/Myrai as "legacy mode"
   - Add orchestrator as opt-in beta
   
2. **Phase 2: Gradual rollout**
   - New customers get multi-agent by default
   - Existing customers can migrate
   
3. **Phase 3: Full migration**
   - All customers on multi-agent
   - Legacy mode deprecated

## Success Metrics

- **Cost Efficiency:** 20% reduction in token usage through specialized agents
- **Response Quality:** 30% improvement in task completion accuracy
- **Human Oversight:** <5% of requests require human approval
- **System Reliability:** 99.9% uptime, graceful degradation when agents fail

## Next Steps

1. Review this architecture with your team
2. Prioritize which specialized agents to build first
3. Create detailed implementation tickets
4. Set up staging environment for testing
5. Begin Phase 1 implementation
