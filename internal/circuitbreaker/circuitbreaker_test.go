package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	cb := NewWithDefaults()

	ctx := context.Background()
	err := cb.Execute(ctx, func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_Execute_Failure(t *testing.T) {
	cb := New(Config{
		MaxFailures:      2,
		Timeout:          100 * time.Millisecond,
		HalfOpenMax:      1,
		SuccessThreshold: 1,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// First two failures should not open circuit
	err := cb.Execute(ctx, func() error { return testErr })
	assert.ErrorIs(t, err, testErr)
	assert.Equal(t, StateClosed, cb.State())

	err = cb.Execute(ctx, func() error { return testErr })
	assert.ErrorIs(t, err, testErr)
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_CircuitOpen(t *testing.T) {
	cb := New(Config{
		MaxFailures:      1,
		Timeout:          1 * time.Second,
		HalfOpenMax:      1,
		SuccessThreshold: 1,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// Trigger circuit open
	cb.Execute(ctx, func() error { return testErr })
	assert.Equal(t, StateOpen, cb.State())

	// Next execution should fail immediately with ErrCircuitOpen
	err := cb.Execute(ctx, func() error { return nil })
	assert.ErrorIs(t, err, ErrCircuitOpen)
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := New(Config{
		MaxFailures:      1,
		Timeout:          50 * time.Millisecond,
		HalfOpenMax:      1,
		SuccessThreshold: 1,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open circuit
	cb.Execute(ctx, func() error { return testErr })
	assert.Equal(t, StateOpen, cb.State())

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Should be half-open now
	err := cb.Execute(ctx, func() error { return nil })
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_HalfOpen_FailureReopens(t *testing.T) {
	cb := New(Config{
		MaxFailures:      1,
		Timeout:          50 * time.Millisecond,
		HalfOpenMax:      1,
		SuccessThreshold: 1,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open circuit
	cb.Execute(ctx, func() error { return testErr })
	assert.Equal(t, StateOpen, cb.State())

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Failure in half-open should reopen
	err := cb.Execute(ctx, func() error { return testErr })
	assert.ErrorIs(t, err, testErr)
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_ExecuteWithResult(t *testing.T) {
	cb := NewWithDefaults()

	ctx := context.Background()
	result, err := cb.ExecuteWithResult(ctx, func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := NewWithDefaults()

	ctx := context.Background()
	cb.Execute(ctx, func() error { return errors.New("fail") })

	stats := cb.Stats()
	assert.Equal(t, "closed", stats["state"])
	assert.Equal(t, 1, stats["failures"])
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := New(Config{
		MaxFailures: 1,
	})

	ctx := context.Background()
	cb.Execute(ctx, func() error { return errors.New("fail") })
	assert.Equal(t, StateOpen, cb.State())

	cb.Reset()
	assert.Equal(t, StateClosed, cb.State())
	assert.Equal(t, 0, cb.failures)
}

func TestState_String(t *testing.T) {
	assert.Equal(t, "closed", StateClosed.String())
	assert.Equal(t, "open", StateOpen.String())
	assert.Equal(t, "half-open", StateHalfOpen.String())
	assert.Equal(t, "unknown", State(999).String())
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 5, cfg.MaxFailures)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.HalfOpenMax)
	assert.Equal(t, 2, cfg.SuccessThreshold)
}
