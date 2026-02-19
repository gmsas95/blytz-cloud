package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

func TestSmokeSignupFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	cfg := &config.Config{
		MaxCustomers:   20,
		PortRangeStart: 30000,
		PortRangeEnd:   30999,
	}

	logger, _ := zap.NewDevelopment()

	prov := provisioner.NewService(
		database,
		"./internal/workspace/templates",
		t.TempDir(),
		"sk-test",
		cfg.PortRangeStart,
		cfg.PortRangeEnd,
		nil,
		"localhost",
		logger,
	)

	stripeSvc := stripe.NewService("sk-test", "price-test")
	stripeWebhook := stripe.NewWebhookHandler(database, prov, "whsec-test")

	router := NewRouter(database, prov, stripeSvc, stripeWebhook, cfg, logger)

	// Step 1: Health check
	t.Run("health check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/health", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Step 2: Create customer with invalid Telegram token (will fail validation)
	t.Run("signup with invalid token", func(t *testing.T) {
		body := map[string]string{
			"email":               "test@example.com",
			"assistant_name":      "TestBot",
			"custom_instructions": "Help me with proposals",
			"telegram_bot_token":  "123:invalid_token",
		}
		bodyBytes, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Will fail because Telegram validation fails
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Step 3: Verify customer was NOT created
	t.Run("customer not created", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/status/test-example-com", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestSmokeCapacityLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	cfg := &config.Config{
		MaxCustomers:   2, // Low limit for testing
		PortRangeStart: 30000,
		PortRangeEnd:   30999,
	}

	logger, _ := zap.NewDevelopment()

	prov := provisioner.NewService(
		database,
		"./internal/workspace/templates",
		t.TempDir(),
		"sk-test",
		cfg.PortRangeStart,
		cfg.PortRangeEnd,
		nil,
		"localhost",
		logger,
	)

	stripeSvc := stripe.NewService("sk-test", "price-test")
	stripeWebhook := stripe.NewWebhookHandler(database, prov, "whsec-test")

	router := NewRouter(database, prov, stripeSvc, stripeWebhook, cfg, logger)

	// Fill up capacity
	ctx := t.Context()
	for i := 0; i < 2; i++ {
		_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              string(rune('a'+i)) + "@example.com",
			AssistantName:      "Test",
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)
	}

	// Try to create third customer
	body := map[string]string{
		"email":               "third@example.com",
		"assistant_name":      "Test",
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

func TestSmokeConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	cfg := &config.Config{
		MaxCustomers:   20,
		PortRangeStart: 30000,
		PortRangeEnd:   30999,
	}

	logger, _ := zap.NewDevelopment()

	prov := provisioner.NewService(
		database,
		"./internal/workspace/templates",
		t.TempDir(),
		"sk-test",
		cfg.PortRangeStart,
		cfg.PortRangeEnd,
		nil,
		"localhost",
		logger,
	)

	stripeSvc := stripe.NewService("sk-test", "price-test")
	stripeWebhook := stripe.NewWebhookHandler(database, prov, "whsec-test")

	router := NewRouter(database, prov, stripeSvc, stripeWebhook, cfg, logger)

	// Send multiple health check requests concurrently
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

	// Wait for all requests
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}

func TestSmokeDatabaseResilience(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

	// Create multiple customers rapidly
	for i := 0; i < 10; i++ {
		_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
			Email:              string(rune('a'+i)) + "@example.com",
			AssistantName:      "Test",
			CustomInstructions: "Help",
			TelegramBotToken:   "123:abc",
		})
		require.NoError(t, err)
	}

	// Verify all customers exist
	count, err := database.CountActiveCustomers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 10, count)

	// Update statuses in rapid succession
	for i := 0; i < 10; i++ {
		err := database.UpdateCustomerStatus(ctx, string(rune('a'+i))+"-example-com", "active")
		require.NoError(t, err)
	}
}

func TestSmokePortAllocation(t *testing.T) {
	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	ctx := t.Context()

	// Allocate multiple ports
	for i := 0; i < 5; i++ {
		customerID := string(rune('a'+i)) + "-customer"
		port := 30000 + i
		err := database.AllocatePort(ctx, customerID, port)
		require.NoError(t, err)
	}

	// Get allocated ports
	ports, err := database.GetAllocatedPorts(ctx)
	require.NoError(t, err)
	assert.Len(t, ports, 5)

	// Release one port
	err = database.ReleasePort(ctx, 30000)
	require.NoError(t, err)

	// Verify port was released
	ports, err = database.GetAllocatedPorts(ctx)
	require.NoError(t, err)
	assert.Len(t, ports, 4)
}

func TestSmokeWebhookInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	cfg := &config.Config{
		MaxCustomers:   20,
		PortRangeStart: 30000,
		PortRangeEnd:   30999,
	}

	prov := provisioner.NewService(
		database,
		"./internal/workspace/templates",
		t.TempDir(),
		"sk-test",
		cfg.PortRangeStart,
		cfg.PortRangeEnd,
		nil,
		"localhost",
		nil,
	)

	stripeWebhook := stripe.NewWebhookHandler(database, prov, "whsec_test")

	router := gin.New()
	router.POST("/webhook", stripeWebhook.HandleWebhook)

	// Test with empty body
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/webhook", bytes.NewReader([]byte{}))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test with malformed JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/webhook", bytes.NewReader([]byte("not json")))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSmokeHtmlPages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(":memory:")
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.Migrate())

	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()
	prov := provisioner.NewService(database, "", "", "", 30000, 30005, nil, "localhost", logger)
	stripeSvc := stripe.NewService("", "")
	stripeWebhook := stripe.NewWebhookHandler(database, prov, "")

	router := NewRouter(database, prov, stripeSvc, stripeWebhook, cfg, logger)

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/", http.StatusOK, "Your Personal AI Assistant"},
		{"/configure", http.StatusOK, "Configure Your Assistant"},
		{"/success", http.StatusOK, "Your Assistant is Ready"},
		{"/nonexistent", http.StatusNotFound, "404"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
