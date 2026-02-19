package provisioner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"blytz/internal/db"
)

func TestNewService(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	tmpDir := t.TempDir()
	svc := NewService(
		database,
		tmpDir,
		tmpDir,
		"sk-test",
		30000,
		30005,
		nil,
		"localhost",
		nil,
	)

	require.NotNil(t, svc)
	assert.NotNil(t, svc.db)
	assert.NotNil(t, svc.ports)
	assert.NotNil(t, svc.workspace)
	assert.NotNil(t, svc.docker)
	assert.NotNil(t, svc.compose)
}

func TestServiceProvisionRollback(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

	// Create customer
	customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "Test",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()
	svc := NewService(
		database,
		tmpDir,
		tmpDir,
		"sk-test",
		30000,
		30005,
		nil,
		"localhost",
		nil,
	)

	// This will fail because Docker isn't available, testing rollback
	err = svc.Provision(ctx, customer.ID)
	require.Error(t, err)

	// Verify customer status rolled back to pending
	updated, err := database.GetCustomerByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, "pending", updated.Status)

	// Verify port was released
	ports, err := database.GetAllocatedPorts(ctx)
	require.NoError(t, err)
	assert.Empty(t, ports)
}

func TestServiceSuspendResume(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

	// Create customer
	customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "Test",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
	})
	require.NoError(t, err)

	// Set to active first
	err = database.UpdateCustomerStatus(ctx, customer.ID, "active")
	require.NoError(t, err)

	tmpDir := t.TempDir()
	svc := NewService(
		database,
		tmpDir,
		tmpDir,
		"sk-test",
		30000,
		30005,
		nil,
		"localhost",
		nil,
	)

	// Test suspend - will fail without Docker but tests DB update path
	err = svc.Suspend(ctx, customer.ID)
	// Expect error due to Docker not available
	require.Error(t, err)

	// Test resume - will fail without Docker
	err = svc.Resume(ctx, customer.ID)
	require.Error(t, err)
}

func TestServiceTerminate(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

	// Create customer with port
	customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "Test",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
	})
	require.NoError(t, err)

	// Allocate a port and update customer
	err = database.AllocatePort(ctx, customer.ID, 30001)
	require.NoError(t, err)
	err = database.UpdateCustomerPort(ctx, customer.ID, 30001)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	svc := NewService(
		database,
		tmpDir,
		tmpDir,
		"sk-test",
		30000,
		30005,
		nil,
		"localhost",
		nil,
	)

	// Terminate - will fail without Docker but tests DB path
	err = svc.Terminate(ctx, customer.ID)
	require.Error(t, err) // Docker not available

	// Verify port released
	ports, err := database.GetAllocatedPorts(ctx)
	require.NoError(t, err)
	assert.Empty(t, ports)
}

func TestServiceValidateBotToken(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()

	tmpDir := t.TempDir()
	svc := NewService(
		database,
		tmpDir,
		tmpDir,
		"sk-test",
		30000,
		30005,
		nil,
		"localhost",
		nil,
	)

	// Test with invalid token
	botInfo, err := svc.ValidateBotToken("123:invalid")
	require.Error(t, err)
	assert.Nil(t, botInfo)
	assert.Contains(t, err.Error(), "telegram")
}

func TestPortExhaustion(t *testing.T) {
	// Test that port allocator correctly handles exhaustion
	allocator := NewPortAllocator(30000, 30002)

	// Allocate all 3 ports
	for i := 0; i < 3; i++ {
		port, err := allocator.AllocatePort()
		require.NoError(t, err, "Should allocate port %d", i)
		assert.Equal(t, 30000+i, port)
	}

	// 4th allocation should fail
	_, err := allocator.AllocatePort()
	assert.Error(t, err, "Should fail when no ports available")
	assert.Contains(t, err.Error(), "no available ports")

	// Release one port and try again
	allocator.ReleasePort(30001)
	port, err := allocator.AllocatePort()
	require.NoError(t, err, "Should allocate after release")
	assert.Equal(t, 30001, port)
}

func TestComposeFileGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	customerID := "test-customer"
	port := 30001
	openAIKey := "sk-test-key"

	err := gen.Generate(customerID, port, openAIKey)
	require.NoError(t, err)

	// Verify file was created
	composePath := filepath.Join(tmpDir, customerID, "docker-compose.yml")
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify key components
	assert.Contains(t, contentStr, "blytz-test-customer")
	assert.Contains(t, contentStr, "30001:18789")
	assert.Contains(t, contentStr, "env_file:") // Check for env_file directive
	assert.Contains(t, contentStr, ".env.secret")
	assert.NotContains(t, contentStr, openAIKey) // API key should NOT be in compose file
	assert.Contains(t, contentStr, "memory: 1G")
	assert.Contains(t, contentStr, "cpus: '0.5'")
	assert.Contains(t, contentStr, "restart: unless-stopped")

	// Verify env file was created
	err = gen.GenerateEnvFile(customerID, openAIKey)
	require.NoError(t, err)

	envPath := filepath.Join(tmpDir, customerID, ".env.secret")
	envContent, err := os.ReadFile(envPath)
	require.NoError(t, err)
	assert.Contains(t, string(envContent), openAIKey)
}

func TestWorkspaceGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	err := gen.Generate("test-customer", 30001, "sk-test")
	require.NoError(t, err)

	// Verify directory was created
	dirPath := filepath.Join(tmpDir, "test-customer")
	_, err = os.Stat(dirPath)
	require.NoError(t, err)

	// Verify compose file exists
	composePath := filepath.Join(dirPath, "docker-compose.yml")
	_, err = os.Stat(composePath)
	require.NoError(t, err)
}
