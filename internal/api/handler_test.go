package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

func setupTestServer(t *testing.T) (*gin.Engine, *db.DB) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	cfg := &config.Config{
		MaxCustomers:   20,
		PortRangeStart: 30000,
		PortRangeEnd:   30999,
	}

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

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

	return router, database
}

func TestHealthCheck(t *testing.T) {
	router, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %s", response["status"])
	}
}

func TestCreateCustomerValidation(t *testing.T) {
	router, _ := setupTestServer(t)

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name: "missing email",
			body: map[string]string{
				"assistant_name":      "Test",
				"custom_instructions": "Help me",
				"telegram_bot_token":  "123:abc",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			body: map[string]string{
				"email":               "not-an-email",
				"assistant_name":      "Test",
				"custom_instructions": "Help me",
				"telegram_bot_token":  "123:abc",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "instructions too long",
			body: map[string]string{
				"email":               "test@example.com",
				"assistant_name":      "Test",
				"custom_instructions": string(make([]byte, 5001)),
				"telegram_bot_token":  "123:abc",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestGetCustomerStatus(t *testing.T) {
	router, database := setupTestServer(t)

	// Create a customer first
	ctx := t.Context()
	_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "Test",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
	})
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/status/test-example-com", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetCustomerStatusNotFound(t *testing.T) {
	router, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/status/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}
