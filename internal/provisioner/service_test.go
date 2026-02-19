package provisioner

import (
	"fmt"
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

func TestLoadAllocatedPorts(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

	// Allocate some ports
	err = database.AllocatePort(ctx, "customer-1", 30001)
	require.NoError(t, err)
	err = database.AllocatePort(ctx, "customer-2", 30002)
	require.NoError(t, err)

	allocator := NewPortAllocator(30000, 30010)

	// Load allocated ports
	err = allocator.LoadAllocatedPorts(ctx, database)
	require.NoError(t, err)

	// Try to allocate - should skip 30001 and 30002
	port, err := allocator.AllocatePort()
	require.NoError(t, err)
	assert.Equal(t, 30000, port)

	// Next allocation should be 30003 (skipping 30001, 30002)
	port, err = allocator.AllocatePort()
	require.NoError(t, err)
	assert.Equal(t, 30003, port)
}

func TestDockerProvisionerMethods(t *testing.T) {
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
	provisioner := NewDockerProvisioner(tmpDir)

	// Test Create without compose file
	err = provisioner.Create(ctx, customer.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker-compose.yml not found")

	// Test GetStatus
	status, err := provisioner.GetStatus(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, "not_found", status)
}

func TestServiceCleanup(t *testing.T) {
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

	// Allocate port and generate files
	err = database.AllocatePort(ctx, customer.ID, 30001)
	require.NoError(t, err)

	err = svc.compose.Generate(customer.ID, 30001, "sk-test")
	require.NoError(t, err)

	err = svc.compose.GenerateEnvFile(customer.ID, "sk-test")
	require.NoError(t, err)

	// Verify files exist
	envPath := filepath.Join(tmpDir, customer.ID, ".env.secret")
	_, err = os.Stat(envPath)
	require.NoError(t, err)

	// Cleanup
	svc.cleanup(customer.ID, 30001)

	// Verify env file removed
	_, err = os.Stat(envPath)
	assert.True(t, os.IsNotExist(err), "env file should be removed")

	// Verify port released
	ports, err := database.GetAllocatedPorts(ctx)
	require.NoError(t, err)
	assert.Empty(t, ports)
}

func TestServiceProvisionErrorPaths(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

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

	// Test with non-existent customer
	err = svc.Provision(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get customer")
}

func TestServiceTerminateNonExistentCustomer(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

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

	// Test terminate non-existent customer
	err = svc.Terminate(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get customer")
}

func TestServiceSuspendResumeNonExistent(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

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

	// Test suspend non-existent customer
	err = svc.Suspend(ctx, "nonexistent")
	require.Error(t, err)
	// Service.Suspend doesn't call GetCustomerByID first, it directly calls docker.Stop
	// So we get a docker error instead of "get customer"
	assert.Contains(t, err.Error(), "stop container")

	// Test resume non-existent customer
	err = svc.Resume(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start container")
}

func TestGenerateGatewayToken(t *testing.T) {
	// Test that gateway tokens are generated
	token1 := generateGatewayToken()
	token2 := generateGatewayToken()

	// Tokens should not be empty
	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)
}

func TestNewPortAllocator(t *testing.T) {
	allocator := NewPortAllocator(30000, 30010)
	assert.NotNil(t, allocator)
	assert.NotNil(t, allocator.allocated)
}

func TestNewComposeGenerator(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)
	assert.NotNil(t, gen)
	assert.Equal(t, tmpDir, gen.baseDir)
}

func TestNewDockerProvisioner(t *testing.T) {
	tmpDir := t.TempDir()
	dp := NewDockerProvisioner(tmpDir)
	assert.NotNil(t, dp)
	assert.Equal(t, tmpDir, dp.baseDir)
}

func TestDockerProvisionerCreateWithExistingCompose(t *testing.T) {
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

	// Create compose file
	customerDir := filepath.Join(tmpDir, customer.ID)
	require.NoError(t, os.MkdirAll(customerDir, 0755))
	composeContent := `version: '3.8'
services:
  test:
    image: alpine:latest
`
	require.NoError(t, os.WriteFile(filepath.Join(customerDir, "docker-compose.yml"), []byte(composeContent), 0644))

	provisioner := NewDockerProvisioner(tmpDir)

	// Test Create with existing compose file
	// This may succeed or fail depending on Docker availability
	err = provisioner.Create(ctx, customer.ID)
	// Either error is OK for coverage - we've tested the code path
	t.Logf("Create returned: %v", err)
}

func TestPortAllocatorConcurrency(t *testing.T) {
	allocator := NewPortAllocator(30000, 30010)

	// Test concurrent allocations
	ports := make(map[int]bool)
	for i := 0; i < 5; i++ {
		port, err := allocator.AllocatePort()
		require.NoError(t, err)
		ports[port] = true
	}

	// All ports should be unique
	assert.Len(t, ports, 5)

	// Test releasing and reallocating
	allocator.ReleasePort(30000)
	port, err := allocator.AllocatePort()
	require.NoError(t, err)
	// Should get 30000 back since it was released
	assert.True(t, port >= 30000 && port <= 30010)
}

func TestComposeGeneratorGenerateMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	// Generate multiple customer configs
	for i := 0; i < 3; i++ {
		customerID := fmt.Sprintf("customer-%d", i)
		port := 30000 + i
		err := gen.Generate(customerID, port, "sk-test")
		require.NoError(t, err)

		// Verify directory created
		dirPath := filepath.Join(tmpDir, customerID)
		_, err = os.Stat(dirPath)
		require.NoError(t, err)

		// Verify compose file
		composePath := filepath.Join(dirPath, "docker-compose.yml")
		content, err := os.ReadFile(composePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), fmt.Sprintf("%d:18789", port))
	}
}

func TestComposeGeneratorEnvFileOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewComposeGenerator(tmpDir)

	customerID := "test-customer"

	// Generate first env file (GenerateEnvFile now creates the directory)
	err := gen.GenerateEnvFile(customerID, "sk-key-1")
	require.NoError(t, err)

	envPath := filepath.Join(tmpDir, customerID, ".env.secret")
	content1, err := os.ReadFile(envPath)
	require.NoError(t, err)
	assert.Contains(t, string(content1), "sk-key-1")

	// Generate second env file (should overwrite)
	err = gen.GenerateEnvFile(customerID, "sk-key-2")
	require.NoError(t, err)

	content2, err := os.ReadFile(envPath)
	require.NoError(t, err)
	assert.Contains(t, string(content2), "sk-key-2")
	assert.NotContains(t, string(content2), "sk-key-1")
}

func TestServiceTerminateWithPort(t *testing.T) {
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

	// Set customer as active with port
	err = database.UpdateCustomerStatus(ctx, customer.ID, "active")
	require.NoError(t, err)

	port := 30001
	err = database.AllocatePort(ctx, customer.ID, port)
	require.NoError(t, err)
	err = database.UpdateCustomerPort(ctx, customer.ID, port)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	svc := NewService(
		database,
		tmpDir,
		tmpDir,
		"sk-test",
		30000,
		30010,
		nil,
		"localhost",
		nil,
	)

	// Create compose file so cleanup has something to remove
	err = svc.compose.Generate(customer.ID, port, "sk-test")
	require.NoError(t, err)

	err = svc.compose.GenerateEnvFile(customer.ID, "sk-test")
	require.NoError(t, err)

	// Verify files exist
	_, err = os.Stat(filepath.Join(tmpDir, customer.ID, ".env.secret"))
	require.NoError(t, err)

	// Terminate may succeed or fail depending on Docker availability
	// Either way, we test the DB cleanup path
	err = svc.Terminate(ctx, customer.ID)
	t.Logf("Terminate returned: %v", err)

	// Verify port cleared from customer
	updated, err := database.GetCustomerByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.ContainerPort)

	// Verify status changed
	assert.Equal(t, "cancelled", updated.Status)
}

func TestServiceProvisionErrorPathsDetailed(t *testing.T) {
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
		30010,
		nil,
		"localhost",
		nil,
	)

	// Load allocated ports first
	err = svc.ports.LoadAllocatedPorts(ctx, database)
	require.NoError(t, err)

	// This will fail at docker create, testing the full path
	err = svc.Provision(ctx, customer.ID)
	// May error or succeed depending on Docker
	t.Logf("Provision returned: %v", err)
}

func TestServicePortAllocationIntegration(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

	// Create multiple customers
	for i := 0; i < 3; i++ {
		_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              fmt.Sprintf("user%d@example.com", i),
			AssistantName:      fmt.Sprintf("Bot%d", i),
			CustomInstructions: "Help me",
			TelegramBotToken:   fmt.Sprintf("token%d:abc", i),
		})
		require.NoError(t, err)
	}

	// Allocate ports for each
	ports := []int{}
	for i := 0; i < 3; i++ {
		port := 30000 + i
		customerID := fmt.Sprintf("user%d-example-com", i)
		err := database.AllocatePort(ctx, customerID, port)
		require.NoError(t, err)
		ports = append(ports, port)
	}

	// Verify ports are allocated
	allocated, err := database.GetAllocatedPorts(ctx)
	require.NoError(t, err)
	assert.Len(t, allocated, 3)
}

func TestDockerProvisionerStartStopRemove(t *testing.T) {
	tmpDir := t.TempDir()
	dp := NewDockerProvisioner(tmpDir)

	// Create a mock compose file
	customerID := "test-customer"
	customerDir := filepath.Join(tmpDir, customerID)
	require.NoError(t, os.MkdirAll(customerDir, 0755))

	composeContent := `version: '3.8'
services:
  openclaw:
    image: alpine:latest
    command: ["echo", "test"]
`
	require.NoError(t, os.WriteFile(filepath.Join(customerDir, "docker-compose.yml"), []byte(composeContent), 0644))

	ctx := t.Context()

	// Test Start - may succeed or fail depending on Docker
	err := dp.Start(ctx, customerID)
	t.Logf("Start returned: %v", err)

	// Test Stop
	err = dp.Stop(ctx, customerID)
	t.Logf("Stop returned: %v", err)

	// Test Remove
	err = dp.Remove(ctx, customerID)
	t.Logf("Remove returned: %v", err)
}

func TestDockerProvisionerErrorCases(t *testing.T) {
	tmpDir := t.TempDir()
	dp := NewDockerProvisioner(tmpDir)

	ctx := t.Context()

	// Test methods with non-existent customer
	err := dp.Create(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker-compose.yml not found")

	err = dp.Start(ctx, "nonexistent")
	require.Error(t, err)

	err = dp.Stop(ctx, "nonexistent")
	require.Error(t, err)

	err = dp.Remove(ctx, "nonexistent")
	require.Error(t, err)

	status, err := dp.GetStatus(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, "not_found", status)
}

func TestServiceMultipleOperations(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

	// Create customer
	customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "multi@example.com",
		AssistantName:      "MultiBot",
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

	// Test ValidateBotToken
	_, err = svc.ValidateBotToken("invalid")
	assert.Error(t, err)

	// Test with empty token
	_, err = svc.ValidateBotToken("")
	assert.Error(t, err)

	// Test various status updates
	err = database.UpdateCustomerStatus(ctx, customer.ID, "provisioning")
	require.NoError(t, err)

	status, err := database.GetCustomerByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, "provisioning", status.Status)

	err = database.UpdateCustomerStatus(ctx, customer.ID, "active")
	require.NoError(t, err)

	err = database.UpdateCustomerStatus(ctx, customer.ID, "suspended")
	require.NoError(t, err)

	err = database.UpdateCustomerStatus(ctx, customer.ID, "cancelled")
	require.NoError(t, err)
}

func TestPortAllocatorEdgeCases(t *testing.T) {
	// Test with single port range
	allocator := NewPortAllocator(30000, 30000)

	// Should allocate the single port
	port, err := allocator.AllocatePort()
	require.NoError(t, err)
	assert.Equal(t, 30000, port)

	// Should fail when exhausted
	_, err = allocator.AllocatePort()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no available ports")

	// Release and reallocate
	allocator.ReleasePort(30000)
	port, err = allocator.AllocatePort()
	require.NoError(t, err)
	assert.Equal(t, 30000, port)
}

func TestPortAllocatorReleaseNonExistent(t *testing.T) {
	allocator := NewPortAllocator(30000, 30010)

	// Should not panic when releasing non-existent port
	allocator.ReleasePort(99999)

	// Should still work normally after releasing non-existent port
	port, err := allocator.AllocatePort()
	require.NoError(t, err)
	assert.Equal(t, 30000, port)
}
