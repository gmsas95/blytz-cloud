package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCustomerID(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"user@example.com", "user-example-com"},
		{"User.Name@Example.COM", "user-name-example-com"},
		{"test+tag@domain.co.uk", "testtag-domain-co-uk"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := generateCustomerID(tt.email)
			if got != tt.expected {
				t.Errorf("generateCustomerID(%q) = %q, want %q", tt.email, got, tt.expected)
			}
		})
	}
}

func TestGenerateCustomerID_Sanitization(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"user@example.com", "user-example-com"},
		{"User.Name@Example.COM", "user-name-example-com"},
		{"test+tag@domain.co.uk", "testtag-domain-co-uk"},
		{"../../etc/passwd@example.com", "etc-passwd-example-com"},
		{"test<script>@example.com", "testscript-example-com"},
		{"a@b.co", "a-b-co"},
		{"very-long-email-that-exceeds-sixty-three-characters-limit@example.com",
			"very-long-email-that-exceeds-sixty-three-characters-limit-examp"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := generateCustomerID(tt.email)
			if got != tt.expected {
				t.Errorf("generateCustomerID(%q) = %q, want %q", tt.email, got, tt.expected)
			}
		})
	}
}

func TestDBCustomerCRUD(t *testing.T) {
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	req := &CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "TestBot",
		CustomInstructions: "Test instructions",
		TelegramBotToken:   "123456:ABCdef",
	}

	customer, err := database.CreateCustomer(ctx, req)
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}

	if customer.ID != "test-example-com" {
		t.Errorf("Expected ID 'test-example-com', got %q", customer.ID)
	}

	if customer.Status != "pending" {
		t.Errorf("Expected status 'pending', got %q", customer.Status)
	}

	retrieved, err := database.GetCustomerByID(ctx, customer.ID)
	if err != nil {
		t.Fatalf("GetCustomerByID failed: %v", err)
	}

	if retrieved.Email != req.Email {
		t.Errorf("Expected email %q, got %q", req.Email, retrieved.Email)
	}

	err = database.UpdateCustomerStatus(ctx, customer.ID, "active")
	if err != nil {
		t.Fatalf("UpdateCustomerStatus failed: %v", err)
	}

	updated, err := database.GetCustomerByID(ctx, customer.ID)
	if err != nil {
		t.Fatalf("GetCustomerByID after update failed: %v", err)
	}

	if updated.Status != "active" {
		t.Errorf("Expected status 'active', got %q", updated.Status)
	}
}

func TestPortAllocation(t *testing.T) {
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	err = database.AllocatePort(ctx, "customer-1", 30001)
	if err != nil {
		t.Fatalf("AllocatePort failed: %v", err)
	}

	err = database.AllocatePort(ctx, "customer-2", 30002)
	if err != nil {
		t.Fatalf("AllocatePort failed: %v", err)
	}

	ports, err := database.GetAllocatedPorts(ctx)
	if err != nil {
		t.Fatalf("GetAllocatedPorts failed: %v", err)
	}

	if len(ports) != 2 {
		t.Errorf("Expected 2 ports, got %d", len(ports))
	}

	err = database.ReleasePort(ctx, 30001)
	if err != nil {
		t.Fatalf("ReleasePort failed: %v", err)
	}

	ports, err = database.GetAllocatedPorts(ctx)
	if err != nil {
		t.Fatalf("GetAllocatedPorts after release failed: %v", err)
	}

	if len(ports) != 1 {
		t.Errorf("Expected 1 port after release, got %d", len(ports))
	}
}

func TestCountActiveCustomers(t *testing.T) {
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	count, err := database.CountActiveCustomers(ctx)
	if err != nil {
		t.Fatalf("CountActiveCustomers failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 active customers, got %d", count)
	}

	_, err = database.CreateCustomer(ctx, &CreateCustomerRequest{
		Email:              "test1@example.com",
		AssistantName:      "Bot1",
		CustomInstructions: "Instructions",
		TelegramBotToken:   "token1",
	})
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}

	count, err = database.CountActiveCustomers(ctx)
	if err != nil {
		t.Fatalf("CountActiveCustomers failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 active customer, got %d", count)
	}
}

func TestGetCustomerByEmail(t *testing.T) {
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	// Test non-existent customer
	customer, err := database.GetCustomerByEmail(ctx, "nonexistent@example.com")
	if err != nil {
		t.Fatalf("GetCustomerByEmail failed: %v", err)
	}
	if customer != nil {
		t.Error("Expected nil for non-existent customer")
	}

	// Create a customer
	_, err = database.CreateCustomer(ctx, &CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "TestBot",
		CustomInstructions: "Instructions",
		TelegramBotToken:   "123:abc",
	})
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}

	// Get by email
	customer, err = database.GetCustomerByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetCustomerByEmail failed: %v", err)
	}
	if customer == nil {
		t.Fatal("Expected customer, got nil")
	}
	if customer.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %q", customer.Email)
	}
}

func TestUpdateCustomerTelegramUsername(t *testing.T) {
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	customer, err := database.CreateCustomer(ctx, &CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "TestBot",
		CustomInstructions: "Instructions",
		TelegramBotToken:   "123:abc",
	})
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}

	err = database.UpdateCustomerTelegramUsername(ctx, customer.ID, "my_test_bot")
	if err != nil {
		t.Fatalf("UpdateCustomerTelegramUsername failed: %v", err)
	}

	retrieved, err := database.GetCustomerByID(ctx, customer.ID)
	if err != nil {
		t.Fatalf("GetCustomerByID failed: %v", err)
	}
	if retrieved.TelegramBotUsername == nil || *retrieved.TelegramBotUsername != "my_test_bot" {
		t.Errorf("Expected username 'my_test_bot', got %v", retrieved.TelegramBotUsername)
	}
}

func TestUpdateCustomerPort(t *testing.T) {
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	customer, err := database.CreateCustomer(ctx, &CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "TestBot",
		CustomInstructions: "Instructions",
		TelegramBotToken:   "123:abc",
	})
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}

	err = database.UpdateCustomerPort(ctx, customer.ID, 30001)
	if err != nil {
		t.Fatalf("UpdateCustomerPort failed: %v", err)
	}

	retrieved, err := database.GetCustomerByID(ctx, customer.ID)
	if err != nil {
		t.Fatalf("GetCustomerByID failed: %v", err)
	}
	if retrieved.ContainerPort == nil || *retrieved.ContainerPort != 30001 {
		t.Errorf("Expected port 30001, got %v", retrieved.ContainerPort)
	}

	// Test ClearCustomerPort
	err = database.ClearCustomerPort(ctx, customer.ID)
	if err != nil {
		t.Fatalf("ClearCustomerPort failed: %v", err)
	}

	retrieved, err = database.GetCustomerByID(ctx, customer.ID)
	if err != nil {
		t.Fatalf("GetCustomerByID failed: %v", err)
	}
	if retrieved.ContainerPort != nil {
		t.Errorf("Expected nil port, got %v", *retrieved.ContainerPort)
	}
}

func TestUpdateStripeInfo(t *testing.T) {
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	customer, err := database.CreateCustomer(ctx, &CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "TestBot",
		CustomInstructions: "Instructions",
		TelegramBotToken:   "123:abc",
	})
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}

	err = database.UpdateStripeInfo(ctx, customer.ID, "cus_test123", "sub_test456")
	if err != nil {
		t.Fatalf("UpdateStripeInfo failed: %v", err)
	}

	retrieved, err := database.GetCustomerByID(ctx, customer.ID)
	if err != nil {
		t.Fatalf("GetCustomerByID failed: %v", err)
	}
	if retrieved.StripeCustomerID == nil || *retrieved.StripeCustomerID != "cus_test123" {
		t.Errorf("Expected stripe customer ID 'cus_test123', got %v", retrieved.StripeCustomerID)
	}
	if retrieved.StripeSubscriptionID == nil || *retrieved.StripeSubscriptionID != "sub_test456" {
		t.Errorf("Expected stripe subscription ID 'sub_test456', got %v", retrieved.StripeSubscriptionID)
	}
	if retrieved.Status != "active" {
		t.Errorf("Expected status 'active', got %q", retrieved.Status)
	}
}

func TestGetCustomerByStripeID(t *testing.T) {
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	ctx := context.Background()

	customer, err := database.CreateCustomer(ctx, &CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "TestBot",
		CustomInstructions: "Instructions",
		TelegramBotToken:   "123:abc",
	})
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}

	err = database.UpdateStripeInfo(ctx, customer.ID, "cus_test123", "sub_test456")
	if err != nil {
		t.Fatalf("UpdateStripeInfo failed: %v", err)
	}

	// Get by stripe ID
	retrieved, err := database.GetCustomerByStripeID(ctx, "cus_test123")
	if err != nil {
		t.Fatalf("GetCustomerByStripeID failed: %v", err)
	}
	if retrieved.ID != customer.ID {
		t.Errorf("Expected customer ID %q, got %q", customer.ID, retrieved.ID)
	}

	// Test non-existent stripe ID
	_, err = database.GetCustomerByStripeID(ctx, "cus_nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent stripe ID")
	}
}

func TestDBErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("GetCustomerByID with non-existent ID", func(t *testing.T) {
		database, err := New(":memory:")
		require.NoError(t, err)
		defer database.Close()
		require.NoError(t, database.Migrate())

		_, err = database.GetCustomerByID(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "customer not found")
	})

	t.Run("CreateCustomer with invalid email", func(t *testing.T) {
		database, err := New(":memory:")
		require.NoError(t, err)
		defer database.Close()
		require.NoError(t, database.Migrate())

		_, err = database.CreateCustomer(ctx, &CreateCustomerRequest{
			Email:              "not-an-email",
			AssistantName:      "Test",
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("CreateCustomer duplicate email", func(t *testing.T) {
		database, err := New(":memory:")
		require.NoError(t, err)
		defer database.Close()
		require.NoError(t, database.Migrate())

		_, err = database.CreateCustomer(ctx, &CreateCustomerRequest{
			Email:              "test@example.com",
			AssistantName:      "Test",
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)

		// Try to create duplicate
		_, err = database.CreateCustomer(ctx, &CreateCustomerRequest{
			Email:              "test@example.com",
			AssistantName:      "Test2",
			CustomInstructions: "Help2",
			TelegramBotToken:   "456:def",
		})
		assert.Error(t, err)
	})

	t.Run("UpdateCustomerStatus non-existent customer", func(t *testing.T) {
		database, err := New(":memory:")
		require.NoError(t, err)
		defer database.Close()
		require.NoError(t, database.Migrate())

		err = database.UpdateCustomerStatus(ctx, "nonexistent", "active")
		// SQLite doesn't error on UPDATE for non-existent rows, just returns 0 rows affected
		assert.NoError(t, err)
	})
}

func TestDBClose(t *testing.T) {
	database, err := New(":memory:")
	require.NoError(t, err)

	err = database.Close()
	require.NoError(t, err)
}
