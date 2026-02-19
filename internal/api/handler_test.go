package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

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

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %v", response["status"])
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

func TestCreateCustomerDuplicate(t *testing.T) {
	router, database := setupTestServer(t)

	// Create first customer
	ctx := t.Context()
	_, err := database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "Test",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
	})
	if err != nil {
		t.Fatalf("Failed to create first customer: %v", err)
	}

	// Try to create duplicate
	body, _ := json.Marshal(map[string]string{
		"email":               "test@example.com",
		"assistant_name":      "Test2",
		"custom_instructions": "Help me too",
		"telegram_bot_token":  "456:def",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error != "already_exists" {
		t.Errorf("Expected error 'already_exists', got %q", response.Error)
	}
}

func TestCreateCustomerAtCapacity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	cfg := &config.Config{
		MaxCustomers:   1, // Set capacity to 1
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

	// Create first customer to reach capacity
	ctx := t.Context()
	_, err = database.CreateCustomer(ctx, &db.CreateCustomerRequest{
		Email:              "first@example.com",
		AssistantName:      "First",
		CustomInstructions: "Help me",
		TelegramBotToken:   "123:abc",
	})
	if err != nil {
		t.Fatalf("Failed to create first customer: %v", err)
	}

	// Try to create second customer
	body, _ := json.Marshal(map[string]string{
		"email":               "second@example.com",
		"assistant_name":      "Second",
		"custom_instructions": "Help me too",
		"telegram_bot_token":  "456:def",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error != "at_capacity" {
		t.Errorf("Expected error 'at_capacity', got %q", response.Error)
	}
}

func TestCheckDatabaseWithNilDB(t *testing.T) {
	handler := &Handler{
		db: nil,
	}

	if handler.checkDatabase(t.Context()) {
		t.Error("Expected checkDatabase to return false with nil DB")
	}
}

func TestCreateCustomerInvalidBotToken(t *testing.T) {
	router, _ := setupTestServer(t)

	body, _ := json.Marshal(map[string]string{
		"email":               "test@example.com",
		"assistant_name":      "Test",
		"custom_instructions": "Help me",
		"telegram_bot_token":  "invalid_token", // No colon
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCheckDatabaseWithError(t *testing.T) {
	// This test would require mocking the database to return an error
	// For now, we just verify the happy path works
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()
	database.Migrate()

	handler := &Handler{
		db: database,
	}

	if !handler.checkDatabase(t.Context()) {
		t.Error("Expected checkDatabase to return true with valid DB")
	}
}
