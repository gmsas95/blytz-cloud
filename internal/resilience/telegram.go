// Package resilience provides circuit breaker wrappers for external services
package resilience

import (
	"context"
	"fmt"

	"blytz/internal/circuitbreaker"
	"blytz/internal/telegram"
)

// TelegramClient wraps telegram validation with circuit breaker
type TelegramClient struct {
	breaker *circuitbreaker.CircuitBreaker
}

// NewTelegramClient creates a new circuit breaker protected telegram client
func NewTelegramClient() *TelegramClient {
	return &TelegramClient{
		breaker: circuitbreaker.NewWithDefaults(),
	}
}

// ValidateToken validates a bot token with circuit breaker protection
func (c *TelegramClient) ValidateToken(ctx context.Context, token string) (*telegram.BotInfo, error) {
	result, err := c.breaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return telegram.ValidateToken(token)
	})

	if err != nil {
		return nil, fmt.Errorf("telegram validation failed: %w", err)
	}

	return result.(*telegram.BotInfo), nil
}

// Stats returns circuit breaker statistics
func (c *TelegramClient) Stats() map[string]interface{} {
	return c.breaker.Stats()
}

// Reset resets the circuit breaker
func (c *TelegramClient) Reset() {
	c.breaker.Reset()
}
