// Package orchestrator manages multiple AI agents and routes requests to the appropriate agent
package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("blytz/orchestrator")

// Agent is the interface that all specialized agents must implement
type Agent interface {
	// Name returns the unique name of the agent
	Name() string

	// CanHandle returns a confidence score (0-1) indicating how well this agent can handle the request
	CanHandle(ctx context.Context, request Request) (float64, error)

	// Execute processes the request and returns a response
	Execute(ctx context.Context, request Request) (Response, error)

	// GetCapabilities returns the list of capabilities this agent has
	GetCapabilities() []Capability

	// GetCostEstimate returns an estimated cost for handling the request
	GetCostEstimate(request Request) CostEstimate
}

// Capability represents what an agent can do
type Capability struct {
	Name        string
	Description string
	Examples    []string
}

// Request represents a user request to the orchestrator
type Request struct {
	ID              string
	CustomerID      string
	Content         string
	Context         map[string]interface{}
	Priority        int     // 1-5, higher = more important
	MaxCostUSD      float64 // Maximum cost willing to spend
	RequireApproval bool    // Whether human approval is required
	Timeout         time.Duration
}

// Response represents the orchestrator's response
type Response struct {
	ID            string
	RequestID     string
	AgentName     string
	Content       string
	Confidence    float64
	CostUSD       float64
	TokensUsed    int
	LatencyMs     int64
	ToolsUsed     []string
	HumanApproved bool
	CreatedAt     time.Time
}

// CostEstimate represents the estimated cost to handle a request
type CostEstimate struct {
	MinUSD     float64
	MaxUSD     float64
	TypicalUSD float64
	Confidence float64 // How confident we are in this estimate
}

// RoutingStrategy determines how requests are routed to agents
type RoutingStrategy string

const (
	StrategyConfidence    RoutingStrategy = "confidence"  // Route to highest confidence agent
	StrategyCostOptimized RoutingStrategy = "cost"        // Route to cheapest capable agent
	StrategyRoundRobin    RoutingStrategy = "round_robin" // Distribute evenly
)

// Orchestrator manages multiple agents and routes requests
type Orchestrator struct {
	agents        map[string]Agent
	strategy      RoutingStrategy
	minConfidence float64
	costTracker   CostTracker
	messageBus    MessageBus
	stateStore    StateStore
	humanGate     HumanGate
	logger        *zap.Logger
	mu            sync.RWMutex
}

// Config holds orchestrator configuration
type Config struct {
	Strategy      RoutingStrategy
	MinConfidence float64
	MaxCostUSD    float64
	Logger        *zap.Logger
}

// New creates a new orchestrator
func New(cfg Config) *Orchestrator {
	if cfg.MinConfidence == 0 {
		cfg.MinConfidence = 0.6
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	return &Orchestrator{
		agents:        make(map[string]Agent),
		strategy:      cfg.Strategy,
		minConfidence: cfg.MinConfidence,
		logger:        cfg.Logger,
	}
}

// RegisterAgent adds an agent to the orchestrator
func (o *Orchestrator) RegisterAgent(agent Agent) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	name := agent.Name()
	if _, exists := o.agents[name]; exists {
		return fmt.Errorf("agent %s already registered", name)
	}

	o.agents[name] = agent
	o.logger.Info("Agent registered", zap.String("agent", name))
	return nil
}

// UnregisterAgent removes an agent from the orchestrator
func (o *Orchestrator) UnregisterAgent(name string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.agents, name)
	o.logger.Info("Agent unregistered", zap.String("agent", name))
}

// ProcessRequest routes a request to the appropriate agent and executes it
func (o *Orchestrator) ProcessRequest(ctx context.Context, req Request) (Response, error) {
	ctx, span := tracer.Start(ctx, "orchestrator.process_request")
	defer span.End()

	// Generate request ID if not provided
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	o.logger.Info("Processing request",
		zap.String("request_id", req.ID),
		zap.String("customer_id", req.CustomerID),
		zap.String("content", req.Content),
	)

	// Check budget
	if o.costTracker != nil {
		if err := o.costTracker.CheckBudget(req.CustomerID, req.MaxCostUSD); err != nil {
			o.logger.Warn("Budget check failed",
				zap.String("request_id", req.ID),
				zap.Error(err),
			)
			return Response{}, fmt.Errorf("budget exceeded: %w", err)
		}
	}

	// Select the best agent for this request
	agent, confidence, err := o.selectAgent(ctx, req)
	if err != nil {
		span.RecordError(err)
		return Response{}, fmt.Errorf("failed to select agent: %w", err)
	}

	if confidence < o.minConfidence {
		return Response{}, fmt.Errorf("no agent confident enough to handle request (max confidence: %.2f)", confidence)
	}

	o.logger.Info("Agent selected",
		zap.String("request_id", req.ID),
		zap.String("agent", agent.Name()),
		zap.Float64("confidence", confidence),
	)

	// Check if human approval is required
	if o.humanGate != nil && (req.RequireApproval || o.needsHumanApproval(agent, req)) {
		approved, err := o.humanGate.RequestApproval(ctx, ApprovalRequest{
			RequestID:   req.ID,
			CustomerID:  req.CustomerID,
			AgentName:   agent.Name(),
			Action:      "execute",
			Details:     req,
			RequestedAt: time.Now(),
		})
		if err != nil {
			return Response{}, fmt.Errorf("approval request failed: %w", err)
		}
		if !approved {
			return Response{}, fmt.Errorf("action rejected by human")
		}
	}

	// Execute the request
	start := time.Now()
	resp, err := agent.Execute(ctx, req)
	latency := time.Since(start)

	if err != nil {
		span.RecordError(err)
		o.logger.Error("Agent execution failed",
			zap.String("request_id", req.ID),
			zap.String("agent", agent.Name()),
			zap.Error(err),
		)
		return Response{}, fmt.Errorf("agent execution failed: %w", err)
	}

	// Fill in response metadata
	resp.ID = uuid.New().String()
	resp.RequestID = req.ID
	resp.AgentName = agent.Name()
	resp.Confidence = confidence
	resp.LatencyMs = latency.Milliseconds()
	resp.CreatedAt = time.Now()

	// Track costs
	if o.costTracker != nil {
		o.costTracker.RecordUsage(req.CustomerID, resp.CostUSD, resp.TokensUsed)
	}

	// Store state
	if o.stateStore != nil {
		if err := o.stateStore.SaveInteraction(ctx, req, resp); err != nil {
			o.logger.Warn("Failed to save interaction", zap.Error(err))
		}
	}

	o.logger.Info("Request processed successfully",
		zap.String("request_id", req.ID),
		zap.String("agent", agent.Name()),
		zap.Float64("cost_usd", resp.CostUSD),
		zap.Int("tokens", resp.TokensUsed),
		zap.Duration("latency", latency),
	)

	return resp, nil
}

// agentScore holds scoring information for agent selection
type agentScore struct {
	agent      Agent
	confidence float64
	cost       CostEstimate
}

// selectAgent chooses the best agent for a request based on the routing strategy
func (o *Orchestrator) selectAgent(ctx context.Context, req Request) (Agent, float64, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if len(o.agents) == 0 {
		return nil, 0, fmt.Errorf("no agents registered")
	}

	var scores []agentScore

	// Get confidence scores from all agents
	for _, agent := range o.agents {
		confidence, err := agent.CanHandle(ctx, req)
		if err != nil {
			o.logger.Warn("Agent confidence check failed",
				zap.String("agent", agent.Name()),
				zap.Error(err),
			)
			continue
		}

		cost := agent.GetCostEstimate(req)
		scores = append(scores, agentScore{
			agent:      agent,
			confidence: confidence,
			cost:       cost,
		})
	}

	if len(scores) == 0 {
		return nil, 0, fmt.Errorf("no agents available")
	}

	// Apply routing strategy
	switch o.strategy {
	case StrategyCostOptimized:
		return o.selectByCost(scores)
	case StrategyRoundRobin:
		return o.selectRoundRobin(scores)
	default: // StrategyConfidence
		return o.selectByConfidence(scores)
	}
}

// selectByConfidence picks the agent with highest confidence score
func (o *Orchestrator) selectByConfidence(scores []agentScore) (Agent, float64, error) {
	var best agentScore
	for _, s := range scores {
		if s.confidence > best.confidence {
			best = s
		}
	}
	return best.agent, best.confidence, nil
}

// selectByCost picks the cheapest capable agent
func (o *Orchestrator) selectByCost(scores []agentScore) (Agent, float64, error) {
	var best agentScore
	for _, s := range scores {
		if s.confidence >= o.minConfidence {
			if best.agent == nil || s.cost.TypicalUSD < best.cost.TypicalUSD {
				best = s
			}
		}
	}
	if best.agent == nil {
		return o.selectByConfidence(scores)
	}
	return best.agent, best.confidence, nil
}

// selectRoundRobin distributes requests evenly (simplified)
func (o *Orchestrator) selectRoundRobin(scores []agentScore) (Agent, float64, error) {
	// In a real implementation, track last used time per agent
	// For now, just select by confidence
	return o.selectByConfidence(scores)
}

// needsHumanApproval determines if a request needs human approval
func (o *Orchestrator) needsHumanApproval(agent Agent, req Request) bool {
	// Check agent-specific rules
	// Check request type
	// Check cost threshold
	// etc.
	return false
}

// GetAgents returns all registered agents
func (o *Orchestrator) GetAgents() []Agent {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agents := make([]Agent, 0, len(o.agents))
	for _, agent := range o.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetAgent returns a specific agent by name
func (o *Orchestrator) GetAgent(name string) (Agent, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agent, exists := o.agents[name]
	return agent, exists
}

// SetCostTracker sets the cost tracker
func (o *Orchestrator) SetCostTracker(tracker CostTracker) {
	o.costTracker = tracker
}

// SetMessageBus sets the message bus
func (o *Orchestrator) SetMessageBus(bus MessageBus) {
	o.messageBus = bus
}

// SetStateStore sets the state store
func (o *Orchestrator) SetStateStore(store StateStore) {
	o.stateStore = store
}

// SetHumanGate sets the human approval gate
func (o *Orchestrator) SetHumanGate(gate HumanGate) {
	o.humanGate = gate
}

// CostTracker interface for tracking costs
type CostTracker interface {
	CheckBudget(customerID string, amount float64) error
	RecordUsage(customerID string, costUSD float64, tokens int)
	GetDailySpend(customerID string) (float64, error)
}

// MessageBus interface for inter-agent communication
type MessageBus interface {
	Publish(ctx context.Context, msg Message) error
	Subscribe(ctx context.Context, agentID string, handler MessageHandler) error
	Request(ctx context.Context, targetAgent string, req Request) (Response, error)
}

// Message represents a message in the message bus
type Message struct {
	ID            string
	From          string
	To            string
	Type          MessageType
	Content       interface{}
	Timestamp     time.Time
	CorrelationID string
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeRequest  MessageType = "request"
	MessageTypeResponse MessageType = "response"
	MessageTypeEvent    MessageType = "event"
)

// MessageHandler is a function that handles messages
type MessageHandler func(ctx context.Context, msg Message) error

// StateStore interface for persisting state
type StateStore interface {
	SaveInteraction(ctx context.Context, req Request, resp Response) error
	GetInteractionHistory(ctx context.Context, customerID string, limit int) ([]Interaction, error)
}

// Interaction represents a saved request-response pair
type Interaction struct {
	Request  Request
	Response Response
}

// HumanGate interface for human-in-the-loop approvals
type HumanGate interface {
	RequestApproval(ctx context.Context, req ApprovalRequest) (bool, error)
}

// ApprovalRequest represents a request for human approval
type ApprovalRequest struct {
	ID          string
	RequestID   string
	CustomerID  string
	AgentName   string
	Action      string
	Details     interface{}
	RequestedAt time.Time
}
