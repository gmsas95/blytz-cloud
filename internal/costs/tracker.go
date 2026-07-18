// Package costs provides cost tracking and budget management
package costs

import (
	"fmt"
	"sync"
	"time"
)

// AlertLevel represents the severity of a cost alert
type AlertLevel string

const (
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// ActionType represents what action to take when a limit is reached
type ActionType string

const (
	ActionNotify   ActionType = "notify"
	ActionThrottle ActionType = "throttle"
	ActionHalt     ActionType = "halt"
)

// CostAlert represents a cost-related alert
type CostAlert struct {
	Level      AlertLevel
	CustomerID string
	Amount     float64
	Limit      float64
	Action     ActionType
	Message    string
	Timestamp  time.Time
}

// BudgetConfig holds budget configuration for a customer
type BudgetConfig struct {
	DailyBudgetUSD   float64
	SessionBudgetUSD float64
	RequestBudgetUSD float64
	WarningThreshold float64 // 0.0-1.0, e.g., 0.8 for 80%
	CustomerID       string
}

// CustomerSpend tracks spending for a customer
type CustomerSpend struct {
	CustomerID      string
	SpentToday      float64
	SpentThisMonth  float64
	SessionSpend    float64
	RequestCount    int
	TokenCount      int
	LastReset       time.Time
	LastRequestTime time.Time
}

// Tracker tracks costs and enforces budgets
type Tracker struct {
	budgets  map[string]BudgetConfig
	spending map[string]*CustomerSpend
	alerts   chan CostAlert
	mu       sync.RWMutex
}

// NewTracker creates a new cost tracker
func NewTracker() *Tracker {
	return &Tracker{
		budgets:  make(map[string]BudgetConfig),
		spending: make(map[string]*CustomerSpend),
		alerts:   make(chan CostAlert, 100),
	}
}

// SetBudget sets the budget for a customer
func (t *Tracker) SetBudget(customerID string, config BudgetConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()

	config.CustomerID = customerID
	t.budgets[customerID] = config

	// Initialize spend tracking if not exists
	if _, exists := t.spending[customerID]; !exists {
		t.spending[customerID] = &CustomerSpend{
			CustomerID: customerID,
			LastReset:  time.Now(),
		}
	}
}

// CheckBudget checks if a cost would exceed the budget
func (t *Tracker) CheckBudget(customerID string, estimatedCost float64) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	budget, hasBudget := t.budgets[customerID]
	spend, hasSpend := t.spending[customerID]

	if !hasBudget {
		// No budget set, allow but warn
		return nil
	}

	if !hasSpend {
		spend = &CustomerSpend{
			CustomerID: customerID,
			LastReset:  time.Now(),
		}
		t.spending[customerID] = spend
	}

	// Check daily budget
	newDailyTotal := spend.SpentToday + estimatedCost
	if budget.DailyBudgetUSD > 0 && newDailyTotal > budget.DailyBudgetUSD {
		return fmt.Errorf("daily budget exceeded: $%.2f spent of $%.2f limit (requesting $%.2f)",
			spend.SpentToday, budget.DailyBudgetUSD, estimatedCost)
	}

	// Check session budget
	newSessionTotal := spend.SessionSpend + estimatedCost
	if budget.SessionBudgetUSD > 0 && newSessionTotal > budget.SessionBudgetUSD {
		return fmt.Errorf("session budget exceeded: $%.2f spent of $%.2f limit (requesting $%.2f)",
			spend.SessionSpend, budget.SessionBudgetUSD, estimatedCost)
	}

	// Check request budget
	if budget.RequestBudgetUSD > 0 && estimatedCost > budget.RequestBudgetUSD {
		return fmt.Errorf("request cost $%.2f exceeds maximum allowed $%.2f",
			estimatedCost, budget.RequestBudgetUSD)
	}

	// Check warning thresholds
	t.checkThresholds(customerID, budget, spend, newDailyTotal)

	return nil
}

// RecordUsage records actual usage
func (t *Tracker) RecordUsage(customerID string, costUSD float64, tokens int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	spend, exists := t.spending[customerID]
	if !exists {
		spend = &CustomerSpend{
			CustomerID: customerID,
			LastReset:  time.Now(),
		}
		t.spending[customerID] = spend
	}

	// Update spend
	spend.SpentToday += costUSD
	spend.SpentThisMonth += costUSD
	spend.SessionSpend += costUSD
	spend.RequestCount++
	spend.TokenCount += tokens
	spend.LastRequestTime = time.Now()
}

// GetDailySpend returns the daily spend for a customer
func (t *Tracker) GetDailySpend(customerID string) (float64, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	spend, exists := t.spending[customerID]
	if !exists {
		return 0, fmt.Errorf("no spending data for customer %s", customerID)
	}

	// Check if we need to reset daily spend
	if time.Since(spend.LastReset) > 24*time.Hour {
		return 0, nil // Will be reset on next use
	}

	return spend.SpentToday, nil
}

// GetCustomerSpend returns full spending info for a customer
func (t *Tracker) GetCustomerSpend(customerID string) (*CustomerSpend, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	spend, exists := t.spending[customerID]
	if !exists {
		return nil, fmt.Errorf("no spending data for customer %s", customerID)
	}

	return spend, nil
}

// ResetDailySpend resets daily spending (call at midnight)
func (t *Tracker) ResetDailySpend(customerID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if spend, exists := t.spending[customerID]; exists {
		spend.SpentToday = 0
		spend.LastReset = time.Now()
	}
}

// ResetSessionSpend resets session spending
func (t *Tracker) ResetSessionSpend(customerID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if spend, exists := t.spending[customerID]; exists {
		spend.SessionSpend = 0
	}
}

// checkThresholds checks if we've crossed warning thresholds
func (t *Tracker) checkThresholds(customerID string, budget BudgetConfig, spend *CustomerSpend, newTotal float64) {
	if budget.WarningThreshold <= 0 || budget.WarningThreshold >= 1 {
		return
	}

	threshold := budget.DailyBudgetUSD * budget.WarningThreshold

	// Check if we're crossing the warning threshold
	if spend.SpentToday < threshold && newTotal >= threshold {
		select {
		case t.alerts <- CostAlert{
			Level:      AlertLevelWarning,
			CustomerID: customerID,
			Amount:     newTotal,
			Limit:      budget.DailyBudgetUSD,
			Action:     ActionNotify,
			Message:    fmt.Sprintf("Daily spend at %.0f%% of budget", budget.WarningThreshold*100),
			Timestamp:  time.Now(),
		}:
		default:
			// Channel full, skip
		}
	}
}

// GetAlertsChannel returns the alerts channel
func (t *Tracker) GetAlertsChannel() <-chan CostAlert {
	return t.alerts
}

// GetAllSpending returns spending for all customers
func (t *Tracker) GetAllSpending() map[string]*CustomerSpend {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]*CustomerSpend)
	for k, v := range t.spending {
		result[k] = v
	}
	return result
}

// Cleanup removes old spending data
func (t *Tracker) Cleanup(maxAge time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for id, spend := range t.spending {
		if now.Sub(spend.LastRequestTime) > maxAge {
			delete(t.spending, id)
		}
	}
}
