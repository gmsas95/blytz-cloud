package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v84/webhook"

	"blytz/internal/db"
	"blytz/internal/provisioner"
)

type WebhookHandler struct {
	db            *db.DB
	provisioner   *provisioner.Service
	webhookSecret string
}

func NewWebhookHandler(database *db.DB, prov *provisioner.Service, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		db:            database,
		provisioner:   prov,
		webhookSecret: webhookSecret,
	}
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), h.webhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	ctx := c.Request.Context()

	switch event.Type {
	case "checkout.session.completed":
		if err := h.handleCheckoutCompleted(ctx, event.Data.Raw); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "customer.subscription.deleted":
		if err := h.handleSubscriptionDeleted(ctx, event.Data.Raw); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "invoice.payment_failed":
		if err := h.handlePaymentFailed(ctx, event.Data.Raw); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *WebhookHandler) handleCheckoutCompleted(ctx context.Context, data json.RawMessage) error {
	var session struct {
		Metadata struct {
			CustomerID string `json:"customer_id"`
		} `json:"metadata"`
		Customer     string `json:"customer"`
		Subscription string `json:"subscription"`
	}

	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("unmarshal session: %w", err)
	}

	if err := h.db.UpdateStripeInfo(ctx, session.Metadata.CustomerID, session.Customer, session.Subscription); err != nil {
		return fmt.Errorf("update stripe info: %w", err)
	}

	if err := h.provisioner.Provision(ctx, session.Metadata.CustomerID); err != nil {
		return fmt.Errorf("provision customer: %w", err)
	}

	return nil
}

func (h *WebhookHandler) handleSubscriptionDeleted(ctx context.Context, data json.RawMessage) error {
	var subscription struct {
		Customer string `json:"customer"`
	}

	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("unmarshal subscription: %w", err)
	}

	customer, err := h.db.GetCustomerByStripeID(ctx, subscription.Customer)
	if err != nil {
		return fmt.Errorf("get customer by stripe id: %w", err)
	}

	if err := h.provisioner.Terminate(ctx, customer.ID); err != nil {
		return fmt.Errorf("terminate customer: %w", err)
	}

	return nil
}

func (h *WebhookHandler) handlePaymentFailed(ctx context.Context, data json.RawMessage) error {
	var invoice struct {
		Customer string `json:"customer"`
	}

	if err := json.Unmarshal(data, &invoice); err != nil {
		return fmt.Errorf("unmarshal invoice: %w", err)
	}

	customer, err := h.db.GetCustomerByStripeID(ctx, invoice.Customer)
	if err != nil {
		return fmt.Errorf("get customer by stripe id: %w", err)
	}

	if err := h.db.UpdateCustomerStatus(ctx, customer.ID, "suspended"); err != nil {
		return fmt.Errorf("suspend customer: %w", err)
	}

	if err := h.provisioner.Suspend(ctx, customer.ID); err != nil {
		return fmt.Errorf("suspend container: %w", err)
	}

	return nil
}
