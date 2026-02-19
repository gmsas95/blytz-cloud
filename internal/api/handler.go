package api

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

type Handler struct {
	db          *db.DB
	provisioner *provisioner.Service
	stripe      *stripe.Service
	cfg         *config.Config
	logger      *log.Logger
}

func NewHandler(database *db.DB, prov *provisioner.Service, stripeSvc *stripe.Service, cfg *config.Config, logger *log.Logger) *Handler {
	return &Handler{
		db:          database,
		provisioner: prov,
		stripe:      stripeSvc,
		cfg:         cfg,
		logger:      logger,
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_failed",
			Message: "Invalid request body",
		})
		return
	}

	if err := h.validateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	count, err := h.db.CountActiveCustomers(ctx)
	if err != nil {
		h.logger.Printf("Failed to count customers: %v", err)
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
		h.logger.Printf("Failed to check existing customer: %v", err)
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
		h.logger.Printf("Failed to create customer: %v", err)
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
		h.logger.Printf("Failed to create checkout session: %v", err)
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

func (h *Handler) validateRequest(req *CreateCustomerRequest) error {
	if len(req.CustomInstructions) > 5000 {
		return errors.New("custom_instructions exceeds maximum length of 5000 characters")
	}

	if len(req.AssistantName) > 50 {
		return errors.New("assistant_name exceeds maximum length of 50 characters")
	}

	if !strings.Contains(req.TelegramBotToken, ":") {
		return errors.New("telegram_bot_token format should be: <numbers>:<alphanumeric>")
	}

	return nil
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
