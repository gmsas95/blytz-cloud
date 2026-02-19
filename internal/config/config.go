package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port                  string
	DatabasePath          string
	CustomersDir          string
	TemplatesDir          string
	MaxCustomers          int
	PortRangeStart        int
	PortRangeEnd          int
	BaseDomain            string
	CaddyAdminURL         string
	OpenAIAPIKey          string
	StripeSecretKey       string
	StripeWebhookSecret   string
	StripePriceID         string
	OpenClawGatewayPrefix string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		DatabasePath:          getEnv("DATABASE_PATH", "./tmp/platform/database.sqlite"),
		CustomersDir:          getEnv("CUSTOMERS_DIR", "./tmp/customers"),
		TemplatesDir:          getEnv("TEMPLATES_DIR", "./internal/workspace/templates"),
		MaxCustomers:          getEnvInt("MAX_CUSTOMERS", 20),
		PortRangeStart:        getEnvInt("PORT_RANGE_START", 30000),
		PortRangeEnd:          getEnvInt("PORT_RANGE_END", 30999),
		BaseDomain:            getEnv("BASE_DOMAIN", "localhost"),
		CaddyAdminURL:         getEnv("CADDY_ADMIN_URL", ""),
		OpenAIAPIKey:          os.Getenv("OPENAI_API_KEY"),
		StripeSecretKey:       os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:   os.Getenv("STRIPE_WEBHOOK_SECRET"),
		StripePriceID:         os.Getenv("STRIPE_PRICE_ID"),
		OpenClawGatewayPrefix: getEnv("OPENCLAW_GATEWAY_TOKEN_PREFIX", "blytz_"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.MaxCustomers <= 0 {
		return fmt.Errorf("MAX_CUSTOMERS must be positive")
	}
	if c.PortRangeEnd <= c.PortRangeStart {
		return fmt.Errorf("PORT_RANGE_END must be greater than PORT_RANGE_START")
	}
	if c.PortRangeEnd-c.PortRangeStart < c.MaxCustomers {
		return fmt.Errorf("port range must accommodate MAX_CUSTOMERS")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
