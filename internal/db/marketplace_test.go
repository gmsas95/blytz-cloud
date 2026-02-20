package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAgentTypes(t *testing.T) {
	database, err := New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	// Should return seeded agents
	agents, err := database.GetAgentTypes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents)

	// Verify OpenClaw exists
	var foundOpenClaw bool
	for _, agent := range agents {
		if agent.ID == "openclaw" {
			foundOpenClaw = true
			assert.Equal(t, "OpenClaw", agent.Name)
			assert.Equal(t, "nodejs", agent.Language)
			assert.Equal(t, 18789, agent.InternalPort)
			assert.Equal(t, 18790, agent.InternalPortBridge)
			assert.Equal(t, "/health", agent.HealthEndpoint)
			assert.Equal(t, "512M", agent.MinMemory)
			assert.Equal(t, "0.25", agent.MinCPU)
			assert.True(t, agent.IsActive)
		}
	}
	assert.True(t, foundOpenClaw, "OpenClaw agent should be seeded")

	// Verify Myrai exists
	var foundMyrai bool
	for _, agent := range agents {
		if agent.ID == "myrai" {
			foundMyrai = true
			assert.Equal(t, "Myrai", agent.Name)
			assert.Equal(t, "go", agent.Language)
			assert.Equal(t, 8080, agent.InternalPort)
			assert.Equal(t, 0, agent.InternalPortBridge) // Myrai doesn't use bridge
			assert.Equal(t, "/api/health", agent.HealthEndpoint)
		}
	}
	assert.True(t, foundMyrai, "Myrai agent should be seeded")
}

func TestGetAgentType(t *testing.T) {
	database, err := New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	// Test getting OpenClaw
	agent, err := database.GetAgentType(ctx, "openclaw")
	require.NoError(t, err)
	assert.Equal(t, "openclaw", agent.ID)
	assert.Equal(t, "OpenClaw", agent.Name)

	// Test getting Myrai
	agent, err = database.GetAgentType(ctx, "myrai")
	require.NoError(t, err)
	assert.Equal(t, "myrai", agent.ID)
	assert.Equal(t, "Myrai", agent.Name)

	// Test non-existent agent
	_, err = database.GetAgentType(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetLLMProviders(t *testing.T) {
	database, err := New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	// Should return seeded providers
	providers, err := database.GetLLMProviders(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, providers)

	// Verify all expected providers exist
	expectedProviders := map[string]string{
		"openai":    "OPENAI_API_KEY",
		"anthropic": "ANTHROPIC_API_KEY",
		"groq":      "GROQ_API_KEY",
		"ollama":    "OLLAMA_HOST",
	}

	foundProviders := make(map[string]bool)
	for _, provider := range providers {
		foundProviders[provider.ID] = true
		if expectedEnv, ok := expectedProviders[provider.ID]; ok {
			assert.Equal(t, expectedEnv, provider.EnvKey)
			assert.True(t, provider.IsActive)
		}
	}

	for id := range expectedProviders {
		assert.True(t, foundProviders[id], "Provider %s should be seeded", id)
	}
}

func TestGetLLMProvider(t *testing.T) {
	database, err := New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	// Test getting OpenAI
	provider, err := database.GetLLMProvider(ctx, "openai")
	require.NoError(t, err)
	assert.Equal(t, "openai", provider.ID)
	assert.Equal(t, "OpenAI", provider.Name)
	assert.Equal(t, "OPENAI_API_KEY", provider.EnvKey)

	// Test getting Anthropic
	provider, err = database.GetLLMProvider(ctx, "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "anthropic", provider.ID)
	assert.Equal(t, "Anthropic", provider.Name)

	// Test non-existent provider
	_, err = database.GetLLMProvider(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCreateCustomerWithMarketplaceFields(t *testing.T) {
	database, err := New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	// Create customer with marketplace selections
	customer, err := database.CreateCustomer(ctx, &CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "Test Assistant",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
		AgentTypeID:        "openclaw",
		LLMProviderID:      "anthropic",
		LLMAPIKey:          "sk-ant-test",
		CustomConfig:       `{"key": "value"}`,
	})
	require.NoError(t, err)

	// Verify marketplace fields stored
	assert.Equal(t, "openclaw", customer.AgentTypeID)
	assert.Equal(t, "anthropic", customer.LLMProviderID)

	// Retrieve and verify
	retrieved, err := database.GetCustomerByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, "openclaw", retrieved.AgentTypeID)
	assert.Equal(t, "anthropic", retrieved.LLMProviderID)
	assert.Equal(t, `{"key": "value"}`, retrieved.CustomConfig)
}

func TestCreateCustomerDefaults(t *testing.T) {
	database, err := New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	// Create customer without marketplace fields (should use defaults)
	customer, err := database.CreateCustomer(ctx, &CreateCustomerRequest{
		Email:              "test2@example.com",
		AssistantName:      "Test Assistant",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
		// AgentTypeID and LLMProviderID not provided
	})
	require.NoError(t, err)

	// Should have defaults
	assert.Equal(t, "openclaw", customer.AgentTypeID)
	assert.Equal(t, "openai", customer.LLMProviderID)
}
