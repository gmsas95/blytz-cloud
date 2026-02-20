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
		// Agent marketplace tables
		`CREATE TABLE IF NOT EXISTS agent_types (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			language TEXT,
			base_image TEXT,
			internal_port INTEGER,
			internal_port_bridge INTEGER,
			health_endpoint TEXT,
			min_memory TEXT,
			min_cpu TEXT,
			config_template TEXT,
			env_vars TEXT,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS llm_providers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			env_key TEXT NOT NULL,
			base_url TEXT,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// Add agent_type_id to customers (migration for existing DBs)
		`ALTER TABLE customers ADD COLUMN agent_type_id TEXT DEFAULT 'openclaw'`,
		`ALTER TABLE customers ADD COLUMN llm_provider_id TEXT DEFAULT 'openai'`,
		`ALTER TABLE customers ADD COLUMN custom_config TEXT`,
		// Indexes for new columns
		`CREATE INDEX IF NOT EXISTS idx_customers_agent_type ON customers(agent_type_id)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_types_active ON agent_types(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_llm_providers_active ON llm_providers(is_active)`,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			// Ignore "duplicate column" errors for ALTER TABLE
			if strings.Contains(err.Error(), "duplicate column") {
				continue
			}
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Seed default agents and LLM providers
	if err := db.seedMarketplaceData(); err != nil {
		return fmt.Errorf("seed marketplace data: %w", err)
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
	// Marketplace fields
	AgentTypeID   string `json:"agent_type_id" db:"agent_type_id"`
	LLMProviderID string `json:"llm_provider_id" db:"llm_provider_id"`
	CustomConfig  string `json:"custom_config" db:"custom_config"`
}

type AgentType struct {
	ID                 string    `json:"id" db:"id"`
	Name               string    `json:"name" db:"name"`
	Description        string    `json:"description" db:"description"`
	Language           string    `json:"language" db:"language"`
	BaseImage          string    `json:"base_image" db:"base_image"`
	InternalPort       int       `json:"internal_port" db:"internal_port"`
	InternalPortBridge int       `json:"internal_port_bridge" db:"internal_port_bridge"`
	HealthEndpoint     string    `json:"health_endpoint" db:"health_endpoint"`
	MinMemory          string    `json:"min_memory" db:"min_memory"`
	MinCPU             string    `json:"min_cpu" db:"min_cpu"`
	ConfigTemplate     string    `json:"config_template" db:"config_template"`
	EnvVars            string    `json:"env_vars" db:"env_vars"`
	IsActive           bool      `json:"is_active" db:"is_active"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

type LLMProvider struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	EnvKey      string    `json:"env_key" db:"env_key"`
	BaseURL     string    `json:"base_url" db:"base_url"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type CreateCustomerRequest struct {
	Email              string `json:"email"`
	AssistantName      string `json:"assistant_name"`
	CustomInstructions string `json:"custom_instructions"`
	TelegramBotToken   string `json:"telegram_bot_token"`
	// Marketplace selections
	AgentTypeID   string `json:"agent_type_id"`
	LLMProviderID string `json:"llm_provider_id"`
	LLMAPIKey     string `json:"llm_api_key"`
	CustomConfig  string `json:"custom_config"`
}

func (db *DB) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*Customer, error) {
	id := generateCustomerID(req.Email)
	if id == "" {
		return nil, fmt.Errorf("invalid email format")
	}

	// Set defaults if not provided
	agentTypeID := req.AgentTypeID
	if agentTypeID == "" {
		agentTypeID = "openclaw"
	}
	llmProviderID := req.LLMProviderID
	if llmProviderID == "" {
		llmProviderID = "openai"
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
		AgentTypeID:        agentTypeID,
		LLMProviderID:      llmProviderID,
		CustomConfig:       req.CustomConfig,
	}

	query := `INSERT INTO customers (id, email, assistant_name, custom_instructions, telegram_bot_token, status, created_at, updated_at, agent_type_id, llm_provider_id, custom_config) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.conn.ExecContext(ctx, query,
		customer.ID, customer.Email, customer.AssistantName,
		customer.CustomInstructions, customer.TelegramBotToken,
		customer.Status, customer.CreatedAt, customer.UpdatedAt,
		customer.AgentTypeID, customer.LLMProviderID, customer.CustomConfig)

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
			  paid_at, suspended_at, cancelled_at, agent_type_id, llm_provider_id, custom_config
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
		&customer.SuspendedAt, &customer.CancelledAt, &customer.AgentTypeID,
		&customer.LLMProviderID, &customer.CustomConfig,
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

// GetAgentTypes returns all active agent types
func (db *DB) GetAgentTypes(ctx context.Context) ([]AgentType, error) {
	query := `SELECT id, name, description, language, base_image, internal_port, internal_port_bridge, 
			health_endpoint, min_memory, min_cpu, config_template, env_vars, is_active, created_at 
			FROM agent_types WHERE is_active = true ORDER BY name`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query agent types: %w", err)
	}
	defer rows.Close()

	var agents []AgentType
	for rows.Next() {
		var agent AgentType
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Description, &agent.Language, &agent.BaseImage,
			&agent.InternalPort, &agent.InternalPortBridge, &agent.HealthEndpoint,
			&agent.MinMemory, &agent.MinCPU, &agent.ConfigTemplate, &agent.EnvVars,
			&agent.IsActive, &agent.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan agent type: %w", err)
		}
		agents = append(agents, agent)
	}

	return agents, rows.Err()
}

// GetAgentType returns a specific agent type by ID
func (db *DB) GetAgentType(ctx context.Context, id string) (*AgentType, error) {
	query := `SELECT id, name, description, language, base_image, internal_port, internal_port_bridge, 
			health_endpoint, min_memory, min_cpu, config_template, env_vars, is_active, created_at 
			FROM agent_types WHERE id = ? AND is_active = true`

	row := db.conn.QueryRowContext(ctx, query, id)

	var agent AgentType
	err := row.Scan(
		&agent.ID, &agent.Name, &agent.Description, &agent.Language, &agent.BaseImage,
		&agent.InternalPort, &agent.InternalPortBridge, &agent.HealthEndpoint,
		&agent.MinMemory, &agent.MinCPU, &agent.ConfigTemplate, &agent.EnvVars,
		&agent.IsActive, &agent.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent type not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get agent type: %w", err)
	}

	return &agent, nil
}

// GetLLMProviders returns all active LLM providers
func (db *DB) GetLLMProviders(ctx context.Context) ([]LLMProvider, error) {
	query := `SELECT id, name, description, env_key, base_url, is_active, created_at 
			FROM llm_providers WHERE is_active = true ORDER BY name`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query llm providers: %w", err)
	}
	defer rows.Close()

	var providers []LLMProvider
	for rows.Next() {
		var provider LLMProvider
		err := rows.Scan(
			&provider.ID, &provider.Name, &provider.Description, &provider.EnvKey,
			&provider.BaseURL, &provider.IsActive, &provider.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan llm provider: %w", err)
		}
		providers = append(providers, provider)
	}

	return providers, rows.Err()
}

// GetLLMProvider returns a specific LLM provider by ID
func (db *DB) GetLLMProvider(ctx context.Context, id string) (*LLMProvider, error) {
	query := `SELECT id, name, description, env_key, base_url, is_active, created_at 
			FROM llm_providers WHERE id = ? AND is_active = true`

	row := db.conn.QueryRowContext(ctx, query, id)

	var provider LLMProvider
	err := row.Scan(
		&provider.ID, &provider.Name, &provider.Description, &provider.EnvKey,
		&provider.BaseURL, &provider.IsActive, &provider.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("llm provider not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get llm provider: %w", err)
	}

	return &provider, nil
}

// seedMarketplaceData populates default agent types and LLM providers
func (db *DB) seedMarketplaceData() error {
	ctx := context.Background()

	// Seed agent types
	agents := []struct {
		id          string
		name        string
		description string
		language    string
		baseImage   string
		port        int
		bridgePort  int
		healthPath  string
		memory      string
		cpu         string
		envVars     string
	}{
		{
			id:          "openclaw",
			name:        "OpenClaw",
			description: "Multi-channel AI assistant with voice, canvas, and 20+ LLM providers",
			language:    "nodejs",
			baseImage:   "node:22-bookworm",
			port:        18789,
			bridgePort:  18790,
			healthPath:  "/health",
			memory:      "512M",
			cpu:         "0.25",
			envVars:     `["OPENAI_API_KEY", "ANTHROPIC_API_KEY", "TELEGRAM_BOT_TOKEN"]`,
		},
		{
			id:          "myrai",
			name:        "Myrai",
			description: "Go-based AI assistant with persona system, memory, and 20+ LLM providers",
			language:    "go",
			baseImage:   "ghcr.io/gmsas95/myrai:latest",
			port:        8080,
			bridgePort:  0,
			healthPath:  "/api/health",
			memory:      "512M",
			cpu:         "0.25",
			envVars:     `["OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GROQ_API_KEY", "TELEGRAM_BOT_TOKEN", "MYRAI_GATEWAY_TOKEN"]`,
		},
	}

	for _, agent := range agents {
		query := `INSERT OR IGNORE INTO agent_types 
			(id, name, description, language, base_image, internal_port, internal_port_bridge, health_endpoint, min_memory, min_cpu, env_vars)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err := db.conn.ExecContext(ctx, query,
			agent.id, agent.name, agent.description, agent.language, agent.baseImage,
			agent.port, agent.bridgePort, agent.healthPath, agent.memory, agent.cpu, agent.envVars)
		if err != nil {
			return fmt.Errorf("seed agent %s: %w", agent.id, err)
		}
	}

	// Seed LLM providers
	providers := []struct {
		id      string
		name    string
		desc    string
		envKey  string
		baseURL string
	}{
		{
			id:     "openai",
			name:   "OpenAI",
			desc:   "GPT-4, GPT-3.5 - Most popular, reliable",
			envKey: "OPENAI_API_KEY",
		},
		{
			id:     "anthropic",
			name:   "Anthropic",
			desc:   "Claude - Great reasoning and long context",
			envKey: "ANTHROPIC_API_KEY",
		},
		{
			id:     "groq",
			name:   "Groq",
			desc:   "Fast inference at affordable prices",
			envKey: "GROQ_API_KEY",
		},
		{
			id:     "ollama",
			name:   "Ollama",
			desc:   "Free local inference - runs on your hardware",
			envKey: "OLLAMA_HOST",
		},
	}

	for _, provider := range providers {
		query := `INSERT OR IGNORE INTO llm_providers (id, name, description, env_key, base_url) VALUES (?, ?, ?, ?, ?)`
		_, err := db.conn.ExecContext(ctx, query, provider.id, provider.name, provider.desc, provider.envKey, provider.baseURL)
		if err != nil {
			return fmt.Errorf("seed provider %s: %w", provider.id, err)
		}
	}

	return nil
}
