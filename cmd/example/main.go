// Example: How to integrate the multi-agent orchestrator into Blytz Cloud
// This file demonstrates the main integration points

package main

import (
	"context"
	"log"
	"time"

	"blytz/internal/agents"
	"blytz/internal/costs"
	"blytz/internal/orchestrator"
	"go.uber.org/zap"
)

func main() {
	// 1. Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 2. Create cost tracker with budgets
	costTracker := costs.NewTracker()

	// Set budget for a customer
	costTracker.SetBudget("customer-123", costs.BudgetConfig{
		DailyBudgetUSD:   10.00,
		SessionBudgetUSD: 2.00,
		RequestBudgetUSD: 0.50,
		WarningThreshold: 0.8,
	})

	// 3. Create orchestrator
	orch := orchestrator.New(orchestrator.Config{
		Strategy:      orchestrator.StrategyConfidence,
		MinConfidence: 0.6,
		Logger:        logger,
	})

	// 4. Set cost tracker
	orch.SetCostTracker(costTracker)

	// 5. Create and register specialized agents
	// In production, you'd inject proper LLM clients
	llmClient := &MockLLMClient{}

	screeningAgent := agents.NewScreeningAgent(llmClient)
	matchingAgent := agents.NewMatchingAgent(llmClient)
	supportAgent := agents.NewSupportAgent(llmClient)

	if err := orch.RegisterAgent(screeningAgent); err != nil {
		log.Fatal(err)
	}
	if err := orch.RegisterAgent(matchingAgent); err != nil {
		log.Fatal(err)
	}
	if err := orch.RegisterAgent(supportAgent); err != nil {
		log.Fatal(err)
	}

	logger.Info("Orchestrator initialized",
		zap.Int("agents", len(orch.GetAgents())))

	// 6. Process requests
	ctx := context.Background()

	// Example 1: Screening request
	req1 := orchestrator.Request{
		ID:         "req-001",
		CustomerID: "customer-123",
		Content:    "Screen this React developer profile for senior level",
		Context: map[string]interface{}{
			"profile_id": "talent-456",
		},
		MaxCostUSD: 0.50,
		Timeout:    30 * time.Second,
	}

	resp1, err := orch.ProcessRequest(ctx, req1)
	if err != nil {
		logger.Error("Request failed", zap.Error(err))
	} else {
		logger.Info("Request processed",
			zap.String("agent", resp1.AgentName),
			zap.Float64("cost", resp1.CostUSD),
			zap.Int("tokens", resp1.TokensUsed),
		)
	}

	// Example 2: Matching request
	req2 := orchestrator.Request{
		ID:         "req-002",
		CustomerID: "customer-123",
		Content:    "Find me a Python developer with 3+ years experience",
		MaxCostUSD: 1.00,
	}

	resp2, err := orch.ProcessRequest(ctx, req2)
	if err != nil {
		logger.Error("Request failed", zap.Error(err))
	} else {
		logger.Info("Request processed",
			zap.String("agent", resp2.AgentName),
			zap.Float64("cost", resp2.CostUSD),
		)
	}

	// Example 3: Support request
	req3 := orchestrator.Request{
		ID:         "req-003",
		CustomerID: "customer-123",
		Content:    "How do I reset my password?",
		MaxCostUSD: 0.10,
	}

	resp3, err := orch.ProcessRequest(ctx, req3)
	if err != nil {
		logger.Error("Request failed", zap.Error(err))
	} else {
		logger.Info("Request processed",
			zap.String("agent", resp3.AgentName),
		)
	}

	// 7. Check spending
	spend, _ := costTracker.GetDailySpend("customer-123")
	logger.Info("Daily spend", zap.Float64("amount", spend))

	// 8. Monitor cost alerts
	go func() {
		for alert := range costTracker.GetAlertsChannel() {
			logger.Warn("Cost alert",
				zap.String("customer", alert.CustomerID),
				zap.String("level", string(alert.Level)),
				zap.Float64("amount", alert.Amount),
				zap.String("message", alert.Message),
			)

			// Send notification to customer/admin
			// email.SendAlert(alert)
			// slack.SendNotification(alert)
		}
	}()
}

// MockLLMClient is a mock LLM client for demonstration
type MockLLMClient struct{}

func (m *MockLLMClient) SimpleChat(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	// Simulate LLM response
	return "0.9", nil // High confidence
}

func (m *MockLLMClient) ChatCompletion(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

// Additional agent implementations

// MatchingAgent matches talent with opportunities
type MatchingAgent struct {
	*agents.BaseAgent
}

func NewMatchingAgent(llmClient interface{}) *MatchingAgent {
	ba := agents.NewBaseAgent(
		"matching",
		"Matches talent profiles with suitable opportunities",
		llmClient,
	)

	ba.AddCapability(orchestrator.Capability{
		Name:        "talent_matching",
		Description: "Find best matching talent for a role",
		Examples:    []string{"Find React developers", "Match designers to projects", "Search for Python engineers"},
	})

	return &MatchingAgent{BaseAgent: &ba}
}

func (ma *MatchingAgent) CanHandle(ctx context.Context, req orchestrator.Request) (float64, error) {
	// Check for matching keywords
	content := req.Content
	if containsAny(content, []string{"find", "match", "search", "looking for", "need"}) {
		return 0.9, nil
	}
	return 0.3, nil
}

func (ma *MatchingAgent) Execute(ctx context.Context, req orchestrator.Request) (orchestrator.Response, error) {
	// Perform matching logic
	// 1. Parse requirements
	// 2. Search talent database
	// 3. Score matches
	// 4. Return top results

	return orchestrator.Response{
		Content:    "Found 3 matching candidates for your React developer search",
		CostUSD:    0.05,
		TokensUsed: 150,
	}, nil
}

func (ma *MatchingAgent) GetCostEstimate(req orchestrator.Request) orchestrator.CostEstimate {
	return orchestrator.CostEstimate{
		MinUSD:     0.02,
		MaxUSD:     0.10,
		TypicalUSD: 0.05,
		Confidence: 0.9,
	}
}

// SupportAgent handles customer support
type SupportAgent struct {
	*agents.BaseAgent
}

func NewSupportAgent(llmClient interface{}) *SupportAgent {
	ba := agents.NewBaseAgent(
		"support",
		"Answers customer support questions",
		llmClient,
	)

	ba.AddCapability(orchestrator.Capability{
		Name:        "customer_support",
		Description: "Answer questions and provide assistance",
		Examples:    []string{"How do I...?", "What is...?", "Help with..."},
	})

	return &SupportAgent{BaseAgent: &ba}
}

func (sa *SupportAgent) CanHandle(ctx context.Context, req orchestrator.Request) (float64, error) {
	// Check for support keywords
	content := req.Content
	if containsAny(content, []string{"how", "what", "help", "question", "?", "support"}) {
		return 0.8, nil
	}
	return 0.2, nil
}

func (sa *SupportAgent) Execute(ctx context.Context, req orchestrator.Request) (orchestrator.Response, error) {
	// Look up answer in knowledge base
	// Or use LLM to generate response

	return orchestrator.Response{
		Content:    "To reset your password, go to Settings > Security > Reset Password",
		CostUSD:    0.01,
		TokensUsed: 50,
	}, nil
}

func (sa *SupportAgent) GetCostEstimate(req orchestrator.Request) orchestrator.CostEstimate {
	return orchestrator.CostEstimate{
		MinUSD:     0.005,
		MaxUSD:     0.02,
		TypicalUSD: 0.01,
		Confidence: 0.95,
	}
}

// Helper function
func containsAny(s string, substrs []string) bool {
	lower := strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(lower, substr) {
			return true
		}
	}
	return false
}
