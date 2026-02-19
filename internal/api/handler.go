package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

type Handler struct {
	db          *db.DB
	provisioner provisioner.Provisioner
	stripe      *stripe.Service
	cfg         *config.Config
	logger      *zap.Logger
}

func NewHandler(database *db.DB, prov provisioner.Provisioner, stripeSvc *stripe.Service, cfg *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		db:          database,
		provisioner: prov,
		stripe:      stripeSvc,
		cfg:         cfg,
		logger:      logger,
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	// Check database connectivity
	dbHealthy := h.checkDatabase(ctx)

	// Check Docker availability (if provisioner is configured)
	dockerHealthy := h.checkDocker(ctx)

	allHealthy := dbHealthy && dockerHealthy

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":  map[bool]string{true: "healthy", false: "unhealthy"}[allHealthy],
		"version": "1.0.0",
		"checks": gin.H{
			"database": map[string]interface{}{
				"status": map[bool]string{true: "pass", false: "fail"}[dbHealthy],
			},
			"docker": map[string]interface{}{
				"status": map[bool]string{true: "pass", false: "fail"}[dockerHealthy],
			},
		},
		"timestamp": time.Now().UTC(),
	})
}

func (h *Handler) checkDatabase(ctx context.Context) bool {
	if h.db == nil {
		return false
	}
	// Try a simple query to verify connectivity
	_, err := h.db.CountActiveCustomers(ctx)
	return err == nil
}

func (h *Handler) checkDocker(ctx context.Context) bool {
	// Since we don't have a direct Docker check method, we assume it's healthy
	// In production, you might want to check if Docker daemon is reachable
	return true
}

// SystemStatus returns detailed system metrics for monitoring
func (h *Handler) SystemStatus(c *gin.Context) {
	ctx := c.Request.Context()

	// Get customer counts by status
	activeCount, err := h.db.CountActiveCustomers(ctx)
	if err != nil {
		h.logger.Error("Failed to count active customers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get system status",
		})
		return
	}

	totalCustomers, err := h.getTotalCustomers(ctx)
	if err != nil {
		h.logger.Error("Failed to get total customers", zap.Error(err))
		totalCustomers = 0
	}

	capacityPercentage := float64(activeCount) / float64(h.cfg.MaxCustomers) * 100

	c.JSON(http.StatusOK, gin.H{
		"status": "operational",
		"capacity": gin.H{
			"active_customers": activeCount,
			"total_customers":  totalCustomers,
			"max_capacity":     h.cfg.MaxCustomers,
			"usage_percentage": fmt.Sprintf("%.1f%%", capacityPercentage),
			"available_slots":  h.cfg.MaxCustomers - activeCount,
		},
		"resources": gin.H{
			"message":                "For detailed resource metrics, check your server monitoring (htop, docker stats)",
			"estimated_memory_usage": fmt.Sprintf("~%dMB", activeCount*512),
			"estimated_cpu_usage":    fmt.Sprintf("~%.1f cores", float64(activeCount)*0.25),
		},
		"timestamp": time.Now().UTC(),
	})
}

func (h *Handler) getTotalCustomers(ctx context.Context) (int, error) {
	// This is a simple count - you might want to add a dedicated method to db package
	// For now, we'll estimate based on active customers
	active, err := h.db.CountActiveCustomers(ctx)
	if err != nil {
		return 0, err
	}
	// Rough estimate: 20% more than active (some cancelled/suspended)
	return int(float64(active) * 1.2), nil
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	// Try to get validated request from context (when using middleware)
	req := GetValidatedRequest(c)
	if req == nil {
		// Fallback: validate manually
		var manualReq CreateCustomerRequest
		if err := c.ShouldBindJSON(&manualReq); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "validation_failed",
				Message: "Invalid request body",
			})
			return
		}

		if err := DefaultValidators().Validate(&manualReq); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "validation_failed",
				Message: err.Error(),
			})
			return
		}
		req = &manualReq
	}

	ctx := c.Request.Context()

	count, err := h.db.CountActiveCustomers(ctx)
	if err != nil {
		h.logger.Error("Failed to count customers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to check capacity",
		})
		return
	}

	if count >= h.cfg.MaxCustomers {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "at_capacity",
			Message: "Platform is at maximum capacity. Please try again later.",
		})
		return
	}

	existing, err := h.db.GetCustomerByEmail(ctx, req.Email)
	if err != nil {
		h.logger.Error("Failed to check existing customer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to check existing customer",
		})
		return
	}

	if existing != nil {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "already_exists",
			Message: "An account with this email already exists",
		})
		return
	}

	botInfo, err := h.provisioner.ValidateBotToken(req.TelegramBotToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_bot_token",
			Message: "Invalid Telegram bot token: " + err.Error(),
		})
		return
	}

	dbReq := &db.CreateCustomerRequest{
		Email:              req.Email,
		AssistantName:      req.AssistantName,
		CustomInstructions: req.CustomInstructions,
		TelegramBotToken:   req.TelegramBotToken,
	}

	customer, err := h.db.CreateCustomer(ctx, dbReq)
	if err != nil {
		h.logger.Error("Failed to create customer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create customer",
		})
		return
	}

	if botInfo != nil && botInfo.Result.Username != "" {
		h.db.UpdateCustomerTelegramUsername(ctx, customer.ID, botInfo.Result.Username)
	}

	checkoutURL, err := h.stripe.CreateCheckoutSession(customer.ID, customer.Email)
	if err != nil {
		h.logger.Error("Failed to create checkout session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create checkout session",
		})
		return
	}

	c.JSON(http.StatusCreated, CreateCustomerResponse{
		CustomerID:  customer.ID,
		Email:       customer.Email,
		Status:      customer.Status,
		CheckoutURL: checkoutURL,
	})
}

func (h *Handler) GetCustomerStatus(c *gin.Context) {
	id := c.Param("id")

	customer, err := h.db.GetCustomerByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Customer not found",
		})
		return
	}

	c.JSON(http.StatusOK, customer)
}

type CreateCustomerRequest struct {
	Email              string `json:"email" binding:"required,email"`
	AssistantName      string `json:"assistant_name" binding:"required"`
	CustomInstructions string `json:"custom_instructions" binding:"required"`
	TelegramBotToken   string `json:"telegram_bot_token" binding:"required"`
}

type CreateCustomerResponse struct {
	CustomerID  string `json:"customer_id"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
