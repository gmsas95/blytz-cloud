// Package circuitbreaker provides circuit breaker functionality for external service calls
package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Config holds circuit breaker configuration
type Config struct {
	MaxFailures      int           // Number of failures before opening circuit
	Timeout          time.Duration // Duration to wait before trying again (half-open)
	HalfOpenMax      int           // Max requests allowed in half-open state
	SuccessThreshold int           // Number of successes needed to close circuit
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		MaxFailures:      5,
		Timeout:          30 * time.Second,
		HalfOpenMax:      3,
		SuccessThreshold: 2,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config        Config
	state         State
	failures      int
	successes     int
	halfOpenCount int
	lastFailure   time.Time
	mu            sync.RWMutex
}

// ErrCircuitOpen is returned when the circuit is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// New creates a new circuit breaker with the given configuration
func New(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// NewWithDefaults creates a circuit breaker with default configuration
func NewWithDefaults() *CircuitBreaker {
	return New(DefaultConfig())
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Execute runs the given function if the circuit allows it
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := cb.canExecute(); err != nil {
		return err
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// ExecuteWithResult runs the given function and returns its result if the circuit allows it
func (cb *CircuitBreaker) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	if err := cb.canExecute(); err != nil {
		return nil, err
	}

	result, err := fn()
	cb.recordResult(err)
	return result, err
}

func (cb *CircuitBreaker) canExecute() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			cb.successes = 0
			return nil
		}
		return fmt.Errorf("%w: retry after %v", ErrCircuitOpen, cb.config.Timeout-time.Since(cb.lastFailure))
	case StateHalfOpen:
		if cb.halfOpenCount >= cb.config.HalfOpenMax {
			return fmt.Errorf("%w: half-open limit reached", ErrCircuitOpen)
		}
		cb.halfOpenCount++
		return nil
	}

	return nil
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.halfOpenCount = 0
		}
	case StateClosed:
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateHalfOpen:
		cb.state = StateOpen
		cb.halfOpenCount = 0
		cb.successes = 0
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
		}
	}
}

// Stats returns current circuit breaker statistics
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":           cb.state.String(),
		"failures":        cb.failures,
		"successes":       cb.successes,
		"half_open_count": cb.halfOpenCount,
		"last_failure":    cb.lastFailure,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenCount = 0
}
