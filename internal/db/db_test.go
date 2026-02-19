package db

import (
	"context"
	"testing"
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
