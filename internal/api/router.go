package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

func NewRouter(database *db.DB, prov *provisioner.Service, stripeSvc *stripe.Service, stripeWebhook *stripe.WebhookHandler, cfg *config.Config, logger *log.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(logger))

	handler := NewHandler(database, prov, stripeSvc, cfg, logger)

	// Health check
	router.GET("/api/health", handler.HealthCheck)

	// API endpoints
	router.POST("/api/signup", handler.CreateCustomer)
	router.GET("/api/status/:id", handler.GetCustomerStatus)
	router.POST("/api/webhook/stripe", stripeWebhook.HandleWebhook)

	// HTML pages
	router.GET("/", serveIndex)
	router.GET("/configure", serveConfigure)
	router.GET("/success", serveSuccess)

	return router
}

func loggingMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Printf("%s %s %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
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
