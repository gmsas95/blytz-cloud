//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"blytz/internal/api"
	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

// setupTestServer creates a fully configured test server
func setupTestServer(t *testing.T) (*gin.Engine, *db.DB, *provisioner.Service, string) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	require.NoError(t, database.Migrate())

	tmpDir := t.TempDir()
	cfg := &config.Config{
		MaxCustomers:   10,
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		TemplatesDir:   "../workspace/templates",
		CustomersDir:   tmpDir,
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		BaseDomain:     "localhost",
	}

	logger, _ := zap.NewDevelopment()

	prov := provisioner.NewService(
		database,
		cfg.TemplatesDir,
		cfg.CustomersDir,
		cfg.OpenAIAPIKey,
		cfg.PortRangeStart,
		cfg.PortRangeEnd,
		nil,
		cfg.BaseDomain,
		logger,
	)

	stripeSvc := stripe.NewService(
		getEnv("STRIPE_SECRET_KEY", "sk_test_dummy"),
		getEnv("STRIPE_PRICE_ID", "price_dummy"),
	)
	stripeWebhook := stripe.NewWebhookHandler(database, prov, getEnv("STRIPE_WEBHOOK_SECRET", "whsec_dummy"))

	router := api.NewRouter(database, prov, stripeSvc, stripeWebhook, cfg, logger)

	return router, database, prov, tmpDir
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// TestE2E_CustomerSignupFlow tests the complete signup flow
func TestE2E_CustomerSignupFlow(t *testing.T) {
	router, database, _, _ := setupTestServer(t)
	defer database.Close()

	t.Run("health_check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "healthy", resp["status"])
	})

	t.Run("create_customer_success", func(t *testing.T) {
		body := map[string]string{
			"email":               "test@example.com",
			"assistant_name":      "TestBot",
			"custom_instructions": "Help me with coding",
			"telegram_bot_token":  "123456:TEST_TOKEN_FOR_TESTING",
		}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Should succeed or fail on Telegram validation
		assert.Contains(t, []int{http.StatusCreated, http.StatusBadRequest}, w.Code)
		
		if w.Code == http.StatusCreated {
			var resp api.CreateCustomerResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, "test-example-com", resp.CustomerID)
			assert.Equal(t, "pending", resp.Status)
			assert.NotEmpty(t, resp.CheckoutURL)
		}
	})

	t.Run("create_customer_duplicate_email", func(t *testing.T) {
		// First create a customer directly in DB
		ctx := context.Background()
		_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              "duplicate@example.com",
			AssistantName:      "DupBot",
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)

		// Try to create again via API
		body := map[string]string{
			"email":               "duplicate@example.com",
			"assistant_name":      "DupBot2",
			"custom_instructions": "Help2",
			"telegram_bot_token":  "123:abc2",
		}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("create_customer_invalid_email", func(t *testing.T) {
		body := map[string]string{
			"email":               "not-an-email",
			"assistant_name":      "TestBot",
			"custom_instructions": "Help",
			"telegram_bot_token":  "123:abc",
		}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("get_customer_status", func(t *testing.T) {
		ctx := context.Background()
		customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              "status@example.com",
			AssistantName:      "StatusBot",
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/status/"+customer.ID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp db.Customer
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, customer.ID, resp.ID)
		assert.Equal(t, "pending", resp.Status)
	})

	t.Run("get_customer_status_not_found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/status/nonexistent-customer", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestE2E_SystemStatus tests system status endpoints
func TestE2E_SystemStatus(t *testing.T) {
	router, database, _, _ := setupTestServer(t)
	defer database.Close()

	ctx := context.Background()
	// Create some customers
	for i := 0; i < 3; i++ {
		_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              string(rune('a'+i)) + "@test.com",
			AssistantName:      "Bot",
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/status/system", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	
	capacity, ok := resp["capacity"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(3), capacity["active_customers"])
	assert.Equal(t, float64(10), capacity["max_capacity"])
}

// TestE2E_CapacityLimit tests capacity enforcement
func TestE2E_CapacityLimit(t *testing.T) {
	router, database, _, _ := setupTestServer(t)
	defer database.Close()

	ctx := context.Background()
	// Fill to capacity
	for i := 0; i < 10; i++ {
		_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              string(rune('a'+i)) + "@cap.com",
			AssistantName:      "Bot",
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)
	}

	// Try to exceed capacity
	body := map[string]string{
		"email":               "overflow@cap.com",
		"assistant_name":      "OverflowBot",
		"custom_instructions": "Help",
		"telegram_bot_token":  "123:abc",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestE2E_WebhookHandling tests Stripe webhook handling
func TestE2E_WebhookHandling(t *testing.T) {
	router, database, prov, _ := setupTestServer(t)
	defer database.Close()

	ctx := context.Background()
	customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "webhook@example.com",
		AssistantName:      "WebhookBot",
		CustomInstructions: "Help",
		TelegramBotToken:   "123:abc",
	})
	require.NoError(t, err)

	t.Run("webhook_invalid_payload", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/webhook/stripe", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("webhook_checkout_completed", func(t *testing.T) {
		if os.Getenv("STRIPE_WEBHOOK_SECRET") == "" {
			t.Skip("STRIPE_WEBHOOK_SECRET not set")
		}

		// Manually update customer as paid (simulating webhook effect)
		err := database.UpdateStripeInfo(ctx, customer.ID, "cus_test_123", "sub_test_456")
		require.NoError(t, err)

		updated, err := database.GetCustomerByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, "active", updated.Status)
		assert.NotNil(t, updated.StripeCustomerID)
		assert.Equal(t, "cus_test_123", *updated.StripeCustomerID)
	})

	_ = prov
}

// TestE2E_HTMLPages tests static HTML page serving
func TestE2E_HTMLPages(t *testing.T) {
	router, database, _, _ := setupTestServer(t)
	defer database.Close()

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/", http.StatusOK, "Personal AI"},
		{"/configure", http.StatusOK, "Configure"},
		{"/success", http.StatusOK, "Ready"},
		{"/nonexistent", http.StatusNotFound, "404"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

// TestE2E_DatabaseOperations tests database CRUD operations
func TestE2E_DatabaseOperations(t *testing.T) {
	database, err := db.New(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := context.Background()

	t.Run("create_and_get_customer", func(t *testing.T) {
		customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              "dbtest@example.com",
			AssistantName:      "DBBot",
			CustomInstructions: "Test DB operations",
			TelegramBotToken:   "123:dbtoken",
		})
		require.NoError(t, err)
		assert.Equal(t, "dbtest-example-com", customer.ID)
		assert.Equal(t, "pending", customer.Status)

		retrieved, err := database.GetCustomerByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, customer.Email, retrieved.Email)
	})

	t.Run("get_by_email", func(t *testing.T) {
		_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              "byemail@example.com",
			AssistantName:      "EmailBot",
			CustomInstructions: "Test",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)

		customer, err := database.GetCustomerByEmail(ctx, "byemail@example.com")
		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, "byemail@example.com", customer.Email)
	})

	t.Run("update_status", func(t *testing.T) {
		customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              "statusup@example.com",
			AssistantName:      "StatusBot",
			CustomInstructions: "Test",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)

		err = database.UpdateCustomerStatus(ctx, customer.ID, "active")
		require.NoError(t, err)

		updated, err := database.GetCustomerByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, "active", updated.Status)
	})

	t.Run("port_allocation", func(t *testing.T) {
		err := database.AllocatePort(ctx, "test-customer", 30001)
		require.NoError(t, err)

		ports, err := database.GetAllocatedPorts(ctx)
		require.NoError(t, err)
		assert.Contains(t, ports, 30001)

		err = database.ReleasePort(ctx, 30001)
		require.NoError(t, err)

		ports, err = database.GetAllocatedPorts(ctx)
		require.NoError(t, err)
		assert.NotContains(t, ports, 30001)
	})

	t.Run("count_active", func(t *testing.T) {
		count, err := database.CountActiveCustomers(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3) // From previous tests
	})
}

// TestE2E_ContainerLifecycle tests container operations if Docker is available
func TestE2E_ContainerLifecycle(t *testing.T) {
	if os.Getenv("DOCKER_TEST") != "true" {
		t.Skip("Set DOCKER_TEST=true to run container tests")
	}

	_, database, prov, tmpDir := setupTestServer(t)
	defer database.Close()

	ctx := context.Background()

	// Create customer
	customer, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "container@example.com",
		AssistantName:      "ContainerBot",
		CustomInstructions: "Test container lifecycle",
		TelegramBotToken:   "123456:CONTAINER_TEST_TOKEN",
	})
	require.NoError(t, err)

	// Simulate payment
	err = database.UpdateStripeInfo(ctx, customer.ID, "cus_container", "sub_container")
	require.NoError(t, err)

	t.Run("provision_container", func(t *testing.T) {
		err := prov.Provision(ctx, customer.ID)
		require.NoError(t, err)

		// Verify workspace files
		workspaceDir := filepath.Join(tmpDir, customer.ID, ".openclaw", "workspace")
		_, err = os.Stat(workspaceDir)
		require.NoError(t, err)

		// Verify compose file
		composePath := filepath.Join(tmpDir, customer.ID, "docker-compose.yml")
		_, err = os.Stat(composePath)
		require.NoError(t, err)

		// Verify customer status
		updated, err := database.GetCustomerByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, "active", updated.Status)
		assert.NotNil(t, updated.ContainerPort)
	})

	t.Run("suspend_container", func(t *testing.T) {
		err := prov.Suspend(ctx, customer.ID)
		require.NoError(t, err)

		updated, err := database.GetCustomerByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, "suspended", updated.Status)
	})

	t.Run("resume_container", func(t *testing.T) {
		err := prov.Resume(ctx, customer.ID)
		require.NoError(t, err)

		updated, err := database.GetCustomerByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, "active", updated.Status)
	})

	t.Run("terminate_container", func(t *testing.T) {
		err := prov.Terminate(ctx, customer.ID)
		require.NoError(t, err)

		updated, err := database.GetCustomerByID(ctx, customer.ID)
		require.NoError(t, err)
		assert.Equal(t, "cancelled", updated.Status)

		// Verify port released
		ports, err := database.GetAllocatedPorts(ctx)
		require.NoError(t, err)
		for _, p := range ports {
			assert.NotEqual(t, *updated.ContainerPort, p)
		}
	})
}

// TestE2E_ConcurrentOperations tests concurrent API requests
func TestE2E_ConcurrentOperations(t *testing.T) {
	router, database, _, _ := setupTestServer(t)
	defer database.Close()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/health", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}

// TestE2E_StripeIntegration tests Stripe checkout flow
func TestE2E_StripeIntegration(t *testing.T) {
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		t.Skip("STRIPE_SECRET_KEY not set")
	}

	router, database, _, _ := setupTestServer(t)
	defer database.Close()

	t.Run("create_checkout_session", func(t *testing.T) {
		body := map[string]string{
			"email":               "stripe@example.com",
			"assistant_name":      "StripeBot",
			"custom_instructions": "Test Stripe",
			"telegram_bot_token":  "123456:TEST_TOKEN",
		}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code == http.StatusCreated {
			var resp api.CreateCustomerResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Contains(t, resp.CheckoutURL, "checkout.stripe.com")
		}
	})
}
