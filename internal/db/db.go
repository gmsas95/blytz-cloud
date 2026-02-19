package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	// Open with WAL mode for better concurrency
	conn, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// WAL mode allows concurrent reads during writes
	// busy_timeout = wait up to 5 seconds if DB is locked
	// synchronous = NORMAL (good balance of safety/performance)
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(time.Hour)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS customers (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			assistant_name TEXT NOT NULL,
			custom_instructions TEXT NOT NULL,
			telegram_bot_token TEXT NOT NULL,
			telegram_bot_username TEXT,
			container_port INTEGER,
			container_id TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			stripe_customer_id TEXT,
			stripe_subscription_id TEXT,
			stripe_checkout_session_id TEXT,
			subscription_status TEXT,
			current_period_end TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			paid_at TIMESTAMP,
			suspended_at TIMESTAMP,
			cancelled_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_stripe_customer ON customers(stripe_customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_stripe_subscription ON customers(stripe_subscription_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_stripe_session ON customers(stripe_checkout_session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_container_port ON customers(container_port)`,
		`CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id TEXT NOT NULL,
			action TEXT NOT NULL,
			details TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_customer ON audit_log(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at)`,
		`CREATE TABLE IF NOT EXISTS port_allocations (
			port INTEGER PRIMARY KEY,
			customer_id TEXT NOT NULL,
			allocated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id)
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

func generateCustomerID(email string) string {
	// Validate email format first
	if !isValidEmail(email) {
		return ""
	}

	id := strings.ToLower(email)

	// Replace special characters with safe separators
	id = strings.ReplaceAll(id, "@", "-")
	id = strings.ReplaceAll(id, ".", "-")
	id = strings.ReplaceAll(id, "/", "-")

	// Remove any remaining non-alphanumeric characters except hyphen
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	id = result.String()

	// Remove leading/trailing hyphens and consecutive hyphens
	id = strings.Trim(id, "-")
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}

	// Enforce max length (63 chars for Docker container names)
	if len(id) > 63 {
		id = id[:63]
	}

	return id
}

func isValidEmail(email string) bool {
	// Simple validation - consider using regex for production
	return strings.Contains(email, "@") &&
		len(email) > 3 &&
		len(email) < 254
}

type Customer struct {
	ID                      string     `json:"id" db:"id"`
	Email                   string     `json:"email" db:"email"`
	AssistantName           string     `json:"assistant_name" db:"assistant_name"`
	CustomInstructions      string     `json:"custom_instructions" db:"custom_instructions"`
	TelegramBotToken        string     `json:"telegram_bot_token" db:"telegram_bot_token"`
	TelegramBotUsername     *string    `json:"telegram_bot_username" db:"telegram_bot_username"`
	ContainerPort           *int       `json:"container_port" db:"container_port"`
	ContainerID             *string    `json:"container_id" db:"container_id"`
	Status                  string     `json:"status" db:"status"`
	StripeCustomerID        *string    `json:"stripe_customer_id" db:"stripe_customer_id"`
	StripeSubscriptionID    *string    `json:"stripe_subscription_id" db:"stripe_subscription_id"`
	StripeCheckoutSessionID *string    `json:"stripe_checkout_session_id" db:"stripe_checkout_session_id"`
	SubscriptionStatus      *string    `json:"subscription_status" db:"subscription_status"`
	CurrentPeriodEnd        *time.Time `json:"current_period_end" db:"current_period_end"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
	PaidAt                  *time.Time `json:"paid_at" db:"paid_at"`
	SuspendedAt             *time.Time `json:"suspended_at" db:"suspended_at"`
	CancelledAt             *time.Time `json:"cancelled_at" db:"cancelled_at"`
}

type CreateCustomerRequest struct {
	Email              string `json:"email"`
	AssistantName      string `json:"assistant_name"`
	CustomInstructions string `json:"custom_instructions"`
	TelegramBotToken   string `json:"telegram_bot_token"`
}

func (db *DB) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*Customer, error) {
	id := generateCustomerID(req.Email)
	if id == "" {
		return nil, fmt.Errorf("invalid email format")
	}

	customer := &Customer{
		ID:                 id,
		Email:              req.Email,
		AssistantName:      req.AssistantName,
		CustomInstructions: req.CustomInstructions,
		TelegramBotToken:   req.TelegramBotToken,
		Status:             "pending",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	query := `INSERT INTO customers (id, email, assistant_name, custom_instructions, telegram_bot_token, status, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.conn.ExecContext(ctx, query,
		customer.ID, customer.Email, customer.AssistantName,
		customer.CustomInstructions, customer.TelegramBotToken,
		customer.Status, customer.CreatedAt, customer.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("insert customer: %w", err)
	}

	if err := db.logAudit(ctx, customer.ID, "created", nil); err != nil {
		return nil, fmt.Errorf("log audit: %w", err)
	}

	return customer, nil
}

func (db *DB) GetCustomerByID(ctx context.Context, id string) (*Customer, error) {
	query := `SELECT id, email, assistant_name, custom_instructions, telegram_bot_token, 
			  telegram_bot_username, container_port, container_id, status, 
			  stripe_customer_id, stripe_subscription_id, stripe_checkout_session_id,
			  subscription_status, current_period_end, created_at, updated_at, 
			  paid_at, suspended_at, cancelled_at
			  FROM customers WHERE id = ?`

	row := db.conn.QueryRowContext(ctx, query, id)

	customer := &Customer{}
	err := row.Scan(
		&customer.ID, &customer.Email, &customer.AssistantName,
		&customer.CustomInstructions, &customer.TelegramBotToken,
		&customer.TelegramBotUsername, &customer.ContainerPort, &customer.ContainerID,
		&customer.Status, &customer.StripeCustomerID, &customer.StripeSubscriptionID,
		&customer.StripeCheckoutSessionID, &customer.SubscriptionStatus, &customer.CurrentPeriodEnd,
		&customer.CreatedAt, &customer.UpdatedAt, &customer.PaidAt,
		&customer.SuspendedAt, &customer.CancelledAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("customer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query customer: %w", err)
	}

	return customer, nil
}

func (db *DB) GetCustomerByEmail(ctx context.Context, email string) (*Customer, error) {
	query := `SELECT id FROM customers WHERE email = ?`
	row := db.conn.QueryRowContext(ctx, query, email)

	var id string
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query customer by email: %w", err)
	}

	return db.GetCustomerByID(ctx, id)
}

func (db *DB) UpdateCustomerStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE customers SET status = ?, updated_at = ? WHERE id = ?`
	_, err := db.conn.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update customer status: %w", err)
	}
	return nil
}

func (db *DB) CountActiveCustomers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM customers WHERE status IN ('pending', 'provisioning', 'active')`
	row := db.conn.QueryRowContext(ctx, query)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count customers: %w", err)
	}

	return count, nil
}

func (db *DB) AllocatePort(ctx context.Context, customerID string, port int) error {
	query := `INSERT INTO port_allocations (port, customer_id) VALUES (?, ?)`
	_, err := db.conn.ExecContext(ctx, query, port, customerID)
	if err != nil {
		return fmt.Errorf("allocate port: %w", err)
	}
	return nil
}

func (db *DB) ReleasePort(ctx context.Context, port int) error {
	query := `DELETE FROM port_allocations WHERE port = ?`
	_, err := db.conn.ExecContext(ctx, query, port)
	if err != nil {
		return fmt.Errorf("release port: %w", err)
	}
	return nil
}

func (db *DB) GetAllocatedPorts(ctx context.Context) ([]int, error) {
	query := `SELECT port FROM port_allocations`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query allocated ports: %w", err)
	}
	defer rows.Close()

	var ports []int
	for rows.Next() {
		var port int
		if err := rows.Scan(&port); err != nil {
			return nil, fmt.Errorf("scan port: %w", err)
		}
		ports = append(ports, port)
	}

	return ports, rows.Err()
}

func (db *DB) UpdateCustomerTelegramUsername(ctx context.Context, id string, username string) error {
	query := `UPDATE customers SET telegram_bot_username = ?, updated_at = ? WHERE id = ?`
	_, err := db.conn.ExecContext(ctx, query, username, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update telegram username: %w", err)
	}
	return nil
}

func (db *DB) ClearCustomerPort(ctx context.Context, id string) error {
	query := `UPDATE customers SET container_port = NULL, updated_at = ? WHERE id = ?`
	_, err := db.conn.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("clear customer port: %w", err)
	}
	return nil
}

func (db *DB) UpdateStripeInfo(ctx context.Context, id string, stripeCustomerID, stripeSubscriptionID string) error {
	query := `UPDATE customers SET 
		stripe_customer_id = ?, 
		stripe_subscription_id = ?, 
		status = 'active',
		paid_at = ?,
		updated_at = ?
		WHERE id = ?`
	now := time.Now()
	_, err := db.conn.ExecContext(ctx, query, stripeCustomerID, stripeSubscriptionID, now, now, id)
	if err != nil {
		return fmt.Errorf("update stripe info: %w", err)
	}
	return nil
}

func (db *DB) GetCustomerByStripeID(ctx context.Context, stripeCustomerID string) (*Customer, error) {
	query := `SELECT id FROM customers WHERE stripe_customer_id = ?`
	row := db.conn.QueryRowContext(ctx, query, stripeCustomerID)

	var id string
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("customer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query customer by stripe id: %w", err)
	}

	return db.GetCustomerByID(ctx, id)
}

func (db *DB) UpdateCustomerPort(ctx context.Context, id string, port int) error {
	query := `UPDATE customers SET container_port = ?, updated_at = ? WHERE id = ?`
	_, err := db.conn.ExecContext(ctx, query, port, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update customer port: %w", err)
	}
	return nil
}

func (db *DB) logAudit(ctx context.Context, customerID, action string, details interface{}) error {
	query := `INSERT INTO audit_log (customer_id, action, details) VALUES (?, ?, ?)`
	_, err := db.conn.ExecContext(ctx, query, customerID, action, details)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
