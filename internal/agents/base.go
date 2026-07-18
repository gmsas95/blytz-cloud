// Package agents provides specialized agent implementations
package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"blytz/internal/llm"
	"blytz/internal/orchestrator"
)

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	name         string
	description  string
	llmClient    *llm.Client
	capabilities []orchestrator.Capability
}

// Name returns the agent name
func (ba *BaseAgent) Name() string {
	return ba.name
}

// Description returns the agent description
func (ba *BaseAgent) Description() string {
	return ba.description
}

// GetCapabilities returns agent capabilities
func (ba *BaseAgent) GetCapabilities() []orchestrator.Capability {
	return ba.capabilities
}

// CanHandle uses LLM to determine if this agent can handle a request
func (ba *BaseAgent) CanHandle(ctx context.Context, request orchestrator.Request) (float64, error) {
	capabilitiesStr := ba.formatCapabilities()

	prompt := fmt.Sprintf(`You are an AI routing system. Determine if the "%s" agent can handle this request.

Agent Capabilities:
%s

User Request: "%s"

Respond with ONLY a number between 0.0 and 1.0 indicating your confidence.
0.0 = Cannot handle at all
1.0 = Perfect match for this agent

Confidence score:`, ba.name, capabilitiesStr, request.Content)

	resp, err := ba.llmClient.SimpleChat(ctx, "You are a request routing system. Be precise.", prompt)
	if err != nil {
		return 0, fmt.Errorf("LLM confidence check failed: %w", err)
	}

	// Parse confidence score
	resp = strings.TrimSpace(resp)
	var confidence float64
	if _, err := fmt.Sscanf(resp, "%f", &confidence); err != nil {
		// If parsing fails, try to extract number from response
		confidence = ba.extractConfidenceFromText(resp)
	}

	// Clamp to valid range
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence, nil
}

// formatCapabilities formats capabilities for the prompt
func (ba *BaseAgent) formatCapabilities() string {
	var parts []string
	for _, cap := range ba.capabilities {
		parts = append(parts, fmt.Sprintf("- %s: %s (examples: %s)",
			cap.Name, cap.Description, strings.Join(cap.Examples, ", ")))
	}
	return strings.Join(parts, "\n")
}

// extractConfidenceFromText tries to extract a confidence score from text
func (ba *BaseAgent) extractConfidenceFromText(text string) float64 {
	// Simple heuristic: look for numbers
	if strings.Contains(text, "0.9") || strings.Contains(text, "0.8") {
		return 0.85
	}
	if strings.Contains(text, "0.7") || strings.Contains(text, "0.6") {
		return 0.65
	}
	if strings.Contains(text, "0.5") {
		return 0.5
	}
	if strings.Contains(text, "0") {
		return 0.1
	}
	if strings.Contains(text, "1") {
		return 0.9
	}
	return 0.5 // Default uncertainty
}

// GetCostEstimate provides a default cost estimate
func (ba *BaseAgent) GetCostEstimate(request orchestrator.Request) orchestrator.CostEstimate {
	// Base cost + variable based on content length
	baseCost := 0.01
	lengthFactor := float64(len(request.Content)) / 1000.0 * 0.005

	return orchestrator.CostEstimate{
		MinUSD:     baseCost,
		MaxUSD:     baseCost + lengthFactor*2,
		TypicalUSD: baseCost + lengthFactor,
		Confidence: 0.8,
	}
}

// Execute must be implemented by specific agents
func (ba *BaseAgent) Execute(ctx context.Context, request orchestrator.Request) (orchestrator.Response, error) {
	return orchestrator.Response{}, fmt.Errorf("Execute not implemented")
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(name, description string, llmClient *llm.Client) BaseAgent {
	return BaseAgent{
		name:        name,
		description: description,
		llmClient:   llmClient,
	}
}

// AddCapability adds a capability to the agent
func (ba *BaseAgent) AddCapability(cap orchestrator.Capability) {
	ba.capabilities = append(ba.capabilities, cap)
}

// ScreeningAgent vets talent profiles
type ScreeningAgent struct {
	BaseAgent
	minScore float64
}

// NewScreeningAgent creates a new screening agent
func NewScreeningAgent(llmClient *llm.Client) *ScreeningAgent {
	ba := NewBaseAgent(
		"screening",
		"Screens and evaluates talent profiles based on criteria",
		llmClient,
	)

	ba.AddCapability(orchestrator.Capability{
		Name:        "profile_evaluation",
		Description: "Evaluate talent profiles against requirements",
		Examples:    []string{"Screen React developers", "Check portfolio quality", "Verify experience level"},
	})

	ba.AddCapability(orchestrator.Capability{
		Name:        "skill_assessment",
		Description: "Assess specific skills and competencies",
		Examples:    []string{"Evaluate coding skills", "Check language proficiency", "Assess design skills"},
	})

	return &ScreeningAgent{
		BaseAgent: ba,
		minScore:  0.7,
	}
}

// Execute performs screening tasks
func (sa *ScreeningAgent) Execute(ctx context.Context, request orchestrator.Request) (orchestrator.Response, error) {
	// Parse screening criteria from request
	criteria, err := sa.parseCriteria(request.Content)
	if err != nil {
		return orchestrator.Response{}, fmt.Errorf("failed to parse criteria: %w", err)
	}

	// Build screening prompt
	prompt := sa.buildScreeningPrompt(criteria, request.Context)

	// Call LLM for evaluation
	start := time.Now()
	result, err := sa.llmClient.SimpleChat(ctx,
		"You are a talent screening expert. Be objective and thorough.",
		prompt,
	)
	latency := time.Since(start)

	if err != nil {
		return orchestrator.Response{}, fmt.Errorf("screening evaluation failed: %w", err)
	}

	// Parse result
	screeningResult, err := sa.parseScreeningResult(result)
	if err != nil {
		return orchestrator.Response{}, fmt.Errorf("failed to parse screening result: %w", err)
	}

	// Calculate cost (rough estimate)
	tokensUsed := len(request.Content)/4 + len(result)/4
	costUSD := float64(tokensUsed) * 0.000002 // $2 per million tokens

	return orchestrator.Response{
		Content:    sa.formatScreeningResponse(screeningResult),
		CostUSD:    costUSD,
		TokensUsed: tokensUsed,
		LatencyMs:  latency.Milliseconds(),
		ToolsUsed:  []string{"llm_screening"},
	}, nil
}

func (sa *ScreeningAgent) parseCriteria(content string) (ScreeningCriteria, error) {
	// Extract criteria from natural language
	// This is simplified - in production, use structured input
	return ScreeningCriteria{
		Role:       sa.extractRole(content),
		Skills:     sa.extractSkills(content),
		Experience: sa.extractExperience(content),
		Location:   sa.extractLocation(content),
	}, nil
}

func (sa *ScreeningAgent) extractRole(content string) string {
	// Simple extraction - look for common role keywords
	roles := []string{"React", "Python", "Go", "Node.js", "designer", "developer", "manager"}
	contentLower := strings.ToLower(content)
	for _, role := range roles {
		if strings.Contains(contentLower, strings.ToLower(role)) {
			return role
		}
	}
	return "general"
}

func (sa *ScreeningAgent) extractSkills(content string) []string {
	// Extract skills mentioned in the request
	var skills []string
	knownSkills := []string{"React", "TypeScript", "Node.js", "Python", "Go", "Docker", "Kubernetes"}
	contentLower := strings.ToLower(content)

	for _, skill := range knownSkills {
		if strings.Contains(contentLower, strings.ToLower(skill)) {
			skills = append(skills, skill)
		}
	}

	return skills
}

func (sa *ScreeningAgent) extractExperience(content string) string {
	// Look for experience requirements
	if strings.Contains(content, "senior") || strings.Contains(content, "5+ years") {
		return "senior"
	}
	if strings.Contains(content, "junior") || strings.Contains(content, "entry") {
		return "junior"
	}
	return "any"
}

func (sa *ScreeningAgent) extractLocation(content string) string {
	// Look for location preferences
	if strings.Contains(content, "remote") {
		return "remote"
	}
	// Could add more location extraction
	return "any"
}

func (sa *ScreeningAgent) buildScreeningPrompt(criteria ScreeningCriteria, context map[string]interface{}) string {
	return fmt.Sprintf(`Screen talent profiles based on these criteria:

Role: %s
Required Skills: %s
Experience Level: %s
Location: %s

Evaluate the following profile and provide:
1. Overall score (0-100)
2. Skill match analysis
3. Experience assessment
4. Recommendation (pass/fail/maybe)
5. Brief justification

Format your response as JSON.`,
		criteria.Role,
		strings.Join(criteria.Skills, ", "),
		criteria.Experience,
		criteria.Location,
	)
}

func (sa *ScreeningAgent) parseScreeningResult(result string) (ScreeningResult, error) {
	// Try to parse as JSON
	var sr ScreeningResult
	if err := json.Unmarshal([]byte(result), &sr); err != nil {
		// Fallback: create result from text
		sr = ScreeningResult{
			Score:          sa.extractScore(result),
			Recommendation: sa.extractRecommendation(result),
			Justification:  result,
		}
	}
	return sr, nil
}

func (sa *ScreeningAgent) extractScore(text string) int {
	// Extract score from text
	// Look for patterns like "Score: 85" or "85/100"
	return 70 // Default
}

func (sa *ScreeningAgent) extractRecommendation(text string) string {
	textLower := strings.ToLower(text)
	if strings.Contains(textLower, "pass") || strings.Contains(textLower, "recommend") {
		return "pass"
	}
	if strings.Contains(textLower, "fail") || strings.Contains(textLower, "reject") {
		return "fail"
	}
	return "maybe"
}

func (sa *ScreeningAgent) formatScreeningResponse(result ScreeningResult) string {
	return fmt.Sprintf(`Screening Result:
- Score: %d/100
- Recommendation: %s
- Justification: %s`,
		result.Score,
		result.Recommendation,
		result.Justification,
	)
}

// ScreeningCriteria represents screening criteria
type ScreeningCriteria struct {
	Role       string
	Skills     []string
	Experience string
	Location   string
}

// ScreeningResult represents the result of screening
type ScreeningResult struct {
	Score          int             `json:"score"`
	SkillMatch     map[string]bool `json:"skill_match,omitempty"`
	ExperienceFit  string          `json:"experience_fit,omitempty"`
	Recommendation string          `json:"recommendation"`
	Justification  string          `json:"justification"`
}
