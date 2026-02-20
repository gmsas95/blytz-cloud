package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

func NewRouter(database *db.DB, prov provisioner.Provisioner, stripeSvc *stripe.Service, stripeWebhook *stripe.WebhookHandler, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(logger))

	handler := NewHandler(database, prov, stripeSvc, cfg, logger)
	marketplaceHandler := NewMarketplaceHandler(database, logger)

	// Health and status checks
	router.GET("/api/health", handler.HealthCheck)
	router.GET("/api/status/system", handler.SystemStatus)

	// Marketplace endpoints
	router.GET("/api/marketplace/agents", marketplaceHandler.ListAgents)
	router.GET("/api/marketplace/agents/:id", marketplaceHandler.GetAgent)
	router.GET("/api/marketplace/llm-providers", marketplaceHandler.ListLLMProviders)
	router.GET("/api/marketplace/llm-providers/:id", marketplaceHandler.GetLLMProvider)
	router.GET("/api/marketplace/stacks", marketplaceHandler.GetStacks)

	// API endpoints with rate limiting
	router.POST("/api/signup", signupRateLimit(), handler.CreateCustomer)
	router.GET("/api/status/:id", handler.GetCustomerStatus)
	router.POST("/api/webhook/stripe", webhookRateLimit(), stripeWebhook.HandleWebhook)

	// HTML pages
	router.GET("/", serveIndex)
	router.GET("/configure", serveConfigure)
	router.GET("/success", serveSuccess)

	return router
}

func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
		)
	}
}

func serveIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, indexHTML)
}

func serveConfigure(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, configureHTML)
}

func serveSuccess(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, successHTML)
}
