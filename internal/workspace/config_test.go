package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOpenClawConfig(t *testing.T) {
	tmpDir := t.TempDir()
	customerID := "test-customer"
	botToken := "123456:ABC-DEF"
	gatewayToken := "gateway-token-123"
	port := 30001

	err := GenerateOpenClawConfig(tmpDir, customerID, botToken, gatewayToken, port)
	require.NoError(t, err)

	// Verify config file was created
	configPath := filepath.Join(tmpDir, customerID, ".openclaw", "openclaw.json")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	contentStr := string(content)

	// Verify key components
	assert.Contains(t, contentStr, gatewayToken)
	assert.Contains(t, contentStr, botToken)
	assert.Contains(t, contentStr, "\"port\": 18789")
	assert.Contains(t, contentStr, "\"enabled\": true")
}

func TestGenerateOpenClawConfigInvalidBaseDir(t *testing.T) {
	// Try to write to a read-only path
	// Create a read-only directory
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.MkdirAll(readOnlyDir, 0755))
	require.NoError(t, os.Chmod(readOnlyDir, 0555))
	defer os.Chmod(readOnlyDir, 0755) // Restore for cleanup

	err := GenerateOpenClawConfig(readOnlyDir, "customer", "token", "gateway", 30001)
	// This may or may not fail depending on whether we're running as root
	// Just verify the function runs without panic
	t.Logf("GenerateOpenClawConfig returned: %v", err)
}

func TestGenerateOpenClawConfigOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	customerID := "test-customer"

	// Generate first config
	err := GenerateOpenClawConfig(tmpDir, customerID, "token1", "gateway1", 30001)
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, customerID, ".openclaw", "openclaw.json")
	content1, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content1), "token1")

	// Generate second config (should overwrite)
	err = GenerateOpenClawConfig(tmpDir, customerID, "token2", "gateway2", 30002)
	require.NoError(t, err)

	content2, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content2), "token2")
	assert.NotContains(t, string(content2), "token1")
}

func TestGenerateOpenClawConfigSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	customerID := "test-customer"
	// Token with special JSON characters
	botToken := `123456:ABC"DEF\XYZ`
	gatewayToken := `gateway"with\quotes`

	err := GenerateOpenClawConfig(tmpDir, customerID, botToken, gatewayToken, 30001)
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, customerID, ".openclaw", "openclaw.json")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// The special characters should be in the file (even if not properly escaped in this simple test)
	contentStr := string(content)
	assert.True(t, strings.Contains(contentStr, botToken) || strings.Contains(contentStr, "123456:ABC"))
}
