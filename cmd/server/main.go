package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blytz/internal/api"
	"blytz/internal/caddy"
	"blytz/internal/config"
	"blytz/internal/db"
	"blytz/internal/provisioner"
	"blytz/internal/stripe"
)

func main() {
	logger := log.New(os.Stdout, "[BLYTZ] ", log.LstdFlags)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	database, err := db.New(cfg.DatabasePath)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	ctx := context.Background()
	if err := loadAllocatedPorts(ctx, database); err != nil {
		logger.Fatalf("Failed to load allocated ports: %v", err)
	}

	var caddyClient *caddy.Client
	if cfg.CaddyAdminURL != "" {
		caddyClient = caddy.NewClient(cfg.CaddyAdminURL)
	}

	prov := provisioner.NewService(
		database,
		cfg.TemplatesDir,
		cfg.CustomersDir,
		cfg.OpenAIAPIKey,
		cfg.PortRangeStart,
		cfg.PortRangeEnd,
		caddyClient,
		cfg.BaseDomain,
		logger,
	)

	stripeSvc := stripe.NewService(cfg.StripeSecretKey, cfg.StripePriceID)
	stripeWebhook := stripe.NewWebhookHandler(database, prov, cfg.StripeWebhookSecret)

	router := api.NewRouter(database, prov, stripeSvc, stripeWebhook, cfg, logger)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exited")
}

func loadAllocatedPorts(ctx context.Context, database *db.DB) error {
	ports, err := database.GetAllocatedPorts(ctx)
	if err != nil {
		return err
	}

	for _, port := range ports {
		_ = port
	}

	return nil
}
