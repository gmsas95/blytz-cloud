package provisioner

import (
	"os"
	"testing"
)

func TestPortAllocator(t *testing.T) {
	allocator := NewPortAllocator(30000, 30005)

	// Test allocation
	port1, err := allocator.AllocatePort()
	if err != nil {
		t.Fatalf("First allocation failed: %v", err)
	}
	if port1 != 30000 {
		t.Errorf("Expected port 30000, got %d", port1)
	}

	port2, err := allocator.AllocatePort()
	if err != nil {
		t.Fatalf("Second allocation failed: %v", err)
	}
	if port2 != 30001 {
		t.Errorf("Expected port 30001, got %d", port2)
	}

	// Test release and reallocation
	allocator.ReleasePort(30000)
	port3, err := allocator.AllocatePort()
	if err != nil {
		t.Fatalf("Reallocation after release failed: %v", err)
	}
	if port3 != 30000 {
		t.Errorf("Expected port 30000 after release, got %d", port3)
	}

	// Test exhaustion
	allocator.AllocatePort() // 30002
	allocator.AllocatePort() // 30003
	allocator.AllocatePort() // 30004
	allocator.AllocatePort() // 30005

	_, err = allocator.AllocatePort() // Should fail
	if err == nil {
		t.Error("Expected error when no ports available, got nil")
	}
}

func TestComposeGenerator(t *testing.T) {
	tmpDir := t.TempDir()
	generator := NewComposeGenerator(tmpDir)

	config := AgentConfig{
		CustomerID:         "test-customer",
		AgentType:          "openclaw",
		ExternalPort:       30001,
		ExternalPortBridge: 30002,
		InternalPort:       18789,
		InternalPortBridge: 18790,
		BaseImage:          "node:22-bookworm",
		LLMEnvKey:          "OPENAI_API_KEY",
		LLMKey:             "sk-test-key",
		GatewayToken:       "test-token",
		HealthEndpoint:     "/health",
		MinMemory:          "512M",
		MinCPU:             "0.25",
	}

	err := generator.Generate(config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify the file was created
	expectedPath := tmpDir + "/test-customer/docker-compose.yml"
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected compose file at %s", expectedPath)
	}
}
