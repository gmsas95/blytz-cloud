package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"blytz/internal/db"
)

func setupMarketplaceTest(t *testing.T) (*gin.Engine, *db.DB) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(":memory:")
	require.NoError(t, err)
	require.NoError(t, database.Migrate())

	logger, _ := zap.NewDevelopment()
	handler := NewMarketplaceHandler(database, logger)

	router := gin.New()
	router.GET("/api/marketplace/agents", handler.ListAgents)
	router.GET("/api/marketplace/agents/:id", handler.GetAgent)
	router.GET("/api/marketplace/llm-providers", handler.ListLLMProviders)
	router.GET("/api/marketplace/llm-providers/:id", handler.GetLLMProvider)
	router.GET("/api/marketplace/stacks", handler.GetStacks)

	return router, database
}

func TestListAgents(t *testing.T) {
	router, _ := setupMarketplaceTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/marketplace/agents", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	agents, ok := response["agents"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, agents)

	// Verify OpenClaw is in the list
	var foundOpenClaw bool
	for _, agent := range agents {
		agentMap := agent.(map[string]interface{})
		if agentMap["id"] == "openclaw" {
			foundOpenClaw = true
			assert.Equal(t, "OpenClaw", agentMap["name"])
			assert.Equal(t, "nodejs", agentMap["language"])
			assert.Equal(t, float64(18789), agentMap["internal_port"])
		}
	}
	assert.True(t, foundOpenClaw, "OpenClaw should be in agents list")
}

func TestGetAgent(t *testing.T) {
	router, _ := setupMarketplaceTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/marketplace/agents/openclaw", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "openclaw", response["id"])
	assert.Equal(t, "OpenClaw", response["name"])
	assert.Equal(t, "nodejs", response["language"])
	assert.Equal(t, float64(18789), response["internal_port"])
}

func TestGetAgentNotFound(t *testing.T) {
	router, _ := setupMarketplaceTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/marketplace/agents/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "not_found", response.Error)
}

func TestListLLMProviders(t *testing.T) {
	router, _ := setupMarketplaceTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/marketplace/llm-providers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	providers, ok := response["providers"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, providers)

	// Verify expected providers exist
	expectedProviders := []string{"openai", "anthropic", "groq", "ollama"}
	foundProviders := make(map[string]bool)
	for _, provider := range providers {
		providerMap := provider.(map[string]interface{})
		foundProviders[providerMap["id"].(string)] = true
	}

	for _, expected := range expectedProviders {
		assert.True(t, foundProviders[expected], "Provider %s should be listed", expected)
	}
}

func TestGetLLMProvider(t *testing.T) {
	router, _ := setupMarketplaceTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/marketplace/llm-providers/openai", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "openai", response["id"])
	assert.Equal(t, "OpenAI", response["name"])
	assert.Equal(t, "OPENAI_API_KEY", response["env_key"])
}

func TestGetLLMProviderNotFound(t *testing.T) {
	router, _ := setupMarketplaceTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/marketplace/llm-providers/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "not_found", response.Error)
}

func TestGetStacks(t *testing.T) {
	router, _ := setupMarketplaceTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/marketplace/stacks", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	stacks, ok := response["stacks"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, stacks)

	// Verify at least one recommended stack
	var foundRecommended bool
	for _, stack := range stacks {
		stackMap := stack.(map[string]interface{})
		if stackMap["recommended"].(bool) {
			foundRecommended = true
			break
		}
	}
	assert.True(t, foundRecommended, "Should have at least one recommended stack")
}
