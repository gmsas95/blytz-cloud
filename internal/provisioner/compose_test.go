package provisioner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOpenClawCompose(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	config := AgentConfig{
		CustomerID:         "customer-123",
		AgentType:          "openclaw",
		ExternalPort:       30001,
		ExternalPortBridge: 30002,
		InternalPort:       18789,
		InternalPortBridge: 18790,
		BaseImage:          "node:22-bookworm",
		LLMEnvKey:          "OPENAI_API_KEY",
		LLMKey:             "sk-test",
		GatewayToken:       "test-token",
		HealthEndpoint:     "/health",
		MinMemory:          "512M",
		MinCPU:             "0.25",
	}

	err := gen.Generate(config)
	require.NoError(t, err)

	// Verify file was created
	composePath := filepath.Join(tmpDir, "customer-123", "docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify OpenClaw-specific content
	assert.Contains(t, contentStr, "image: node:22-bookworm")
	assert.Contains(t, contentStr, "npm install -g openclaw@latest")
	assert.Contains(t, contentStr, "30001:18789")
	assert.Contains(t, contentStr, "30002:18790")          // Bridge port
	assert.Contains(t, contentStr, "/home/node/.openclaw") // Volume mount
	assert.Contains(t, contentStr, "user: \"1000:1000\"")
	assert.Contains(t, contentStr, "http://localhost:18789/health")
}

func TestGenerateMyraiCompose(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	config := AgentConfig{
		CustomerID:         "customer-456",
		AgentType:          "myrai",
		ExternalPort:       30100,
		ExternalPortBridge: 30101,
		InternalPort:       8080,
		InternalPortBridge: 0,
		BaseImage:          "ghcr.io/gmsas95/myrai:latest",
		LLMEnvKey:          "ANTHROPIC_API_KEY",
		LLMKey:             "sk-ant-test",
		GatewayToken:       "myrai-token",
		HealthEndpoint:     "/api/health",
		MinMemory:          "512M",
		MinCPU:             "0.25",
	}

	err := gen.Generate(config)
	require.NoError(t, err)

	// Verify file was created
	composePath := filepath.Join(tmpDir, "customer-456", "docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify Myrai-specific content
	assert.Contains(t, contentStr, "image: ghcr.io/gmsas95/myrai:latest")
	assert.Contains(t, contentStr, `["myrai", "server", "--port", "8080"]`)
	assert.Contains(t, contentStr, "30100:8080")
	assert.Contains(t, contentStr, "MYRAI_GATEWAY_TOKEN=myrai-token")
	assert.Contains(t, contentStr, "MYRAI_SERVER_PORT=8080")
	assert.Contains(t, contentStr, "http://localhost:8080/api/health")

	// Myrai doesn't have bridge port
	assert.NotContains(t, contentStr, "30101:")
}

func TestGenerateEnvFileMultipleVars(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	customerID := "test-customer"
	envVars := map[string]string{
		"OPENAI_API_KEY":     "sk-test-openai",
		"ANTHROPIC_API_KEY":  "sk-ant-test",
		"TELEGRAM_BOT_TOKEN": "123:abc",
		"CUSTOM_VAR":         "custom_value",
	}

	err := gen.GenerateEnvFile(customerID, envVars)
	require.NoError(t, err)

	// Verify file was created
	envPath := filepath.Join(tmpDir, customerID, ".env.secret")
	content, err := os.ReadFile(envPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify all environment variables
	assert.Contains(t, contentStr, "OPENAI_API_KEY=sk-test-openai")
	assert.Contains(t, contentStr, "ANTHROPIC_API_KEY=sk-ant-test")
	assert.Contains(t, contentStr, "TELEGRAM_BOT_TOKEN=123:abc")
	assert.Contains(t, contentStr, "CUSTOM_VAR=custom_value")

	// Verify file permissions (0600)
	info, err := os.Stat(envPath)
	require.NoError(t, err)
	mode := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), mode, "Env file should have 0600 permissions")
}

func TestGenerateUnknownAgentType(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	config := AgentConfig{
		CustomerID: "customer-789",
		AgentType:  "unknown-agent",
	}

	err := gen.Generate(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown agent type")
}

func TestAgentTemplateCompleteness(t *testing.T) {
	// Verify all templates have required fields
	requiredPlaceholders := []string{
		"{{.CustomerID}}",
		"{{.ExternalPort}}",
		"{{.InternalPort}}",
		"{{.BaseImage}}",
		"{{.LLMEnvKey}}",
		"{{.MinMemory}}",
		"{{.MinCPU}}",
	}

	for agentType, template := range AgentTemplates {
		for _, placeholder := range requiredPlaceholders {
			assert.Contains(t, template, placeholder,
				"Template for %s missing placeholder %s", agentType, placeholder)
		}
	}
}

func TestComposeGenerationIdempotency(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	config := AgentConfig{
		CustomerID:     "customer-test",
		AgentType:      "openclaw",
		ExternalPort:   30001,
		InternalPort:   18789,
		BaseImage:      "node:22-bookworm",
		LLMEnvKey:      "OPENAI_API_KEY",
		HealthEndpoint: "/health",
		MinMemory:      "512M",
		MinCPU:         "0.25",
	}

	// Generate first time
	err := gen.Generate(config)
	require.NoError(t, err)

	content1, err := os.ReadFile(filepath.Join(tmpDir, "customer-test", "docker-compose.yml"))
	require.NoError(t, err)

	// Generate second time (should overwrite)
	err = gen.Generate(config)
	require.NoError(t, err)

	content2, err := os.ReadFile(filepath.Join(tmpDir, "customer-test", "docker-compose.yml"))
	require.NoError(t, err)

	// Should be identical
	assert.Equal(t, string(content1), string(content2))
}

func TestDockerComposeValidity(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	testCases := []struct {
		name   string
		config AgentConfig
	}{
		{
			name: "OpenClaw",
			config: AgentConfig{
				CustomerID:         "openclaw-test",
				AgentType:          "openclaw",
				ExternalPort:       30001,
				ExternalPortBridge: 30002,
				InternalPort:       18789,
				InternalPortBridge: 18790,
				BaseImage:          "node:22-bookworm",
				LLMEnvKey:          "OPENAI_API_KEY",
				GatewayToken:       "token",
				HealthEndpoint:     "/health",
				MinMemory:          "512M",
				MinCPU:             "0.25",
			},
		},
		{
			name: "Myrai",
			config: AgentConfig{
				CustomerID:     "myrai-test",
				AgentType:      "myrai",
				ExternalPort:   30100,
				InternalPort:   8080,
				BaseImage:      "ghcr.io/gmsas95/myrai:latest",
				LLMEnvKey:      "ANTHROPIC_API_KEY",
				GatewayToken:   "token",
				HealthEndpoint: "/api/health",
				MinMemory:      "512M",
				MinCPU:         "0.25",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := gen.Generate(tc.config)
			require.NoError(t, err)

			composePath := filepath.Join(tmpDir, tc.config.CustomerID, "docker-compose.yml")
			content, err := os.ReadFile(composePath)
			require.NoError(t, err)
			contentStr := string(content)

			// Basic YAML validity checks
			assert.True(t, strings.HasPrefix(contentStr, "version: '3.8'"),
				"Should have version declaration")
			assert.Contains(t, contentStr, "services:",
				"Should have services section")
			assert.Contains(t, contentStr, "container_name: blytz-"+tc.config.CustomerID,
				"Should have correct container name")
		})
	}
}
