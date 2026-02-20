package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"blytz/internal/db"
)

// MarketplaceHandler handles marketplace-related API endpoints
type MarketplaceHandler struct {
	db     *db.DB
	logger *zap.Logger
}

// NewMarketplaceHandler creates a new marketplace handler
func NewMarketplaceHandler(database *db.DB, logger *zap.Logger) *MarketplaceHandler {
	return &MarketplaceHandler{
		db:     database,
		logger: logger,
	}
}

// ListAgents returns all available agent types
func (h *MarketplaceHandler) ListAgents(c *gin.Context) {
	agents, err := h.db.GetAgentTypes(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get agent types", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve agent types",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
	})
}

// GetAgent returns details for a specific agent type
func (h *MarketplaceHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, err := h.db.GetAgentType(c.Request.Context(), agentID)
	if err != nil {
		h.logger.Error("Failed to get agent type", zap.String("agent_id", agentID), zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Agent type not found",
		})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// ListLLMProviders returns all available LLM providers
func (h *MarketplaceHandler) ListLLMProviders(c *gin.Context) {
	providers, err := h.db.GetLLMProviders(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get LLM providers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve LLM providers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}

// GetLLMProvider returns details for a specific LLM provider
func (h *MarketplaceHandler) GetLLMProvider(c *gin.Context) {
	providerID := c.Param("id")

	provider, err := h.db.GetLLMProvider(c.Request.Context(), providerID)
	if err != nil {
		h.logger.Error("Failed to get LLM provider", zap.String("provider_id", providerID), zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "LLM provider not found",
		})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// GetStacks returns popular agent + LLM combinations
func (h *MarketplaceHandler) GetStacks(c *gin.Context) {
	// TODO: Implement based on actual usage analytics
	// For now, return hardcoded popular combinations
	stacks := []gin.H{
		{
			"name":        "OpenClaw + Claude",
			"agent_id":    "openclaw",
			"llm_id":      "anthropic",
			"description": "Best for multi-channel communication with excellent reasoning",
			"recommended": true,
		},
		{
			"name":        "OpenClaw + GPT-4",
			"agent_id":    "openclaw",
			"llm_id":      "openai",
			"description": "Most popular combination, reliable for general use",
			"recommended": false,
		},
		{
			"name":        "Myrai + Groq",
			"agent_id":    "myrai",
			"llm_id":      "groq",
			"description": "Fast and affordable, great for high-volume usage",
			"recommended": false,
		},
		{
			"name":        "Myrai + Local (Ollama)",
			"agent_id":    "myrai",
			"llm_id":      "ollama",
			"description": "100% free, runs locally on your hardware",
			"recommended": false,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"stacks": stacks,
	})
}
