//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

func TestE2EProvisionCustomer(t *testing.T) {
	if os.Getenv("RUN_E2E") != "true" {
		t.Skip("Skipping E2E test. Set RUN_E2E=true to run.")
	}

	database, err := db.New("./test_e2e.db")
	require.NoError(t, err)
	defer database.Close()
	defer os.Remove("./test_e2e.db")
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		TemplatesDir:   "../internal/workspace/templates",
		CustomersDir:   tmpDir,
		PortRangeStart: 30000,
		PortRangeEnd:   30005,
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		BaseDomain:     "localhost",
	}

	prov := provisioner.NewService(
		database,
		cfg.TemplatesDir,
		cfg.CustomersDir,
		cfg.OpenAIAPIKey,
		cfg.PortRangeStart,
		cfg.PortRangeEnd,
		nil,
		cfg.BaseDomain,
		nil,
	)

	stripeSvc := stripe.NewService(
		os.Getenv("STRIPE_SECRET_KEY"),
		os.Getenv("STRIPE_PRICE_ID"),
	)
	stripeWebhook := stripe.NewWebhookHandler(database, prov, os.Getenv("STRIPE_WEBHOOK_SECRET"))

	// Step 1: Create customer
	customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "e2e@example.com",
		AssistantName:      "E2EBot",
		CustomInstructions: "I'm a freelance developer needing help with proposals",
		TelegramBotToken:   "123456:TEST_TOKEN_E2E",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, customer.ID)

	// Step 2: Create Stripe checkout session
	checkoutURL, err := stripeSvc.CreateCheckoutSession(customer.ID, customer.Email)
	// Skip if no Stripe credentials
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		t.Skip("Skipping Stripe tests - no credentials")
	}
	require.NoError(t, err)
	assert.NotEmpty(t, checkoutURL)

	// Step 3: Simulate webhook - update customer with Stripe info
	err = database.UpdateStripeInfo(ctx, customer.ID, "cus_e2e_123", "sub_e2e_456")
	require.NoError(t, err)

	// Step 4: Provision customer (requires Docker)
	if os.Getenv("DOCKER_TEST") != "true" {
		t.Skip("Skipping Docker tests - set DOCKER_TEST=true")
	}

	err = prov.Provision(ctx, customer.ID)
	require.NoError(t, err)

	// Verify workspace files created
	workspaceDir := filepath.Join(cfg.CustomersDir, customer.ID, ".openclaw", "workspace")
	_, err = os.Stat(workspaceDir)
	require.NoError(t, err)

	// Verify compose file
	composePath := filepath.Join(cfg.CustomersDir, customer.ID, "docker-compose.yml")
	_, err = os.Stat(composePath)
	require.NoError(t, err)

	// Verify customer status
	updated, err := database.GetCustomerByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", updated.Status)

	// Step 5: Test suspend/resume
	err = prov.Suspend(ctx, customer.ID)
	require.NoError(t, err)

	err = prov.Resume(ctx, customer.ID)
	require.NoError(t, err)

	// Step 6: Test terminate
	err = prov.Terminate(ctx, customer.ID)
	require.NoError(t, err)

	updated, err = database.GetCustomerByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", updated.Status)

	// Verify port released
	ports, err := database.GetAllocatedPorts(ctx)
	require.NoError(t, err)
	assert.Empty(t, ports)
}

func TestE2EStripeWebhookFlow(t *testing.T) {
	if os.Getenv("RUN_E2E") != "true" || os.Getenv("STRIPE_WEBHOOK_SECRET") == "" {
		t.Skip("Skipping E2E Stripe tests")
	}

	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	// Create customer
	customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "webhook@example.com",
		AssistantName:      "WebhookBot",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
	})
	require.NoError(t, err)

	prov := provisioner.NewService(database, "", "", "", 30000, 30005, nil, "localhost", nil)
	webhookHandler := stripe.NewWebhookHandler(database, prov, os.Getenv("STRIPE_WEBHOOK_SECRET"))

	// Simulate checkout.completed webhook
	event := map[string]interface{}{
		"type": "checkout.session.completed",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"metadata": map[string]string{
					"customer_id": customer.ID,
				},
				"customer":     "cus_webhook_123",
				"subscription": "sub_webhook_456",
			},
		},
	}

	eventJSON, _ := json.Marshal(event)

	// In real test, would send HTTP request with proper signature
	_ = webhookHandler
	_ = eventJSON

	// Verify customer updated
	updated, err := database.GetCustomerByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, "cus_webhook_123", updated.StripeCustomerID)
}

func TestE2EMultipleCustomers(t *testing.T) {
	if os.Getenv("RUN_E2E") != "true" {
		t.Skip("Skipping E2E test")
	}

	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	// Create 5 customers
	for i := 0; i < 5; i++ {
		_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              string(rune('a'+i)) + "@e2e.com",
			AssistantName:      "Bot" + string(rune('0'+i)),
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc" + string(rune('0'+i)),
		})
		require.NoError(t, err)
	}

	// Verify count
	count, err := database.CountActiveCustomers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	// Verify all have unique IDs
	customers := make(map[string]bool)
	for i := 0; i < 5; i++ {
		id := string(rune('a'+i)) + "-e2e-com"
		customer, err := database.GetCustomerByID(ctx, id)
		require.NoError(t, err)
		customers[customer.ID] = true
	}
	assert.Len(t, customers, 5)
}

func TestE2EDockerAvailability(t *testing.T) {
	if os.Getenv("RUN_E2E") != "true" {
		t.Skip("Skipping E2E test")
	}

	// Check if Docker is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "version")
	err := cmd.Run()

	if err != nil {
		t.Skip("Docker not available")
	}

	t.Log("Docker is available")
}
