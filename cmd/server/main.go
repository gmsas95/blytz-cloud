package main

import (
	"context"
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
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	database, err := db.New(cfg.DatabasePath)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	ctx := context.Background()
	if err := loadAllocatedPorts(ctx, database); err != nil {
		logger.Fatal("Failed to load allocated ports", zap.Error(err))
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
		logger.Info("Server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
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
