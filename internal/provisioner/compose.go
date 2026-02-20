package provisioner

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ComposeGenerator creates docker-compose files for different agent types
type ComposeGenerator struct {
	baseDir string
}

// AgentConfig contains configuration for generating compose files
type AgentConfig struct {
	CustomerID         string
	AgentType          string // "openclaw", "myrai", etc.
	ExternalPort       int
	ExternalPortBridge int
	InternalPort       int
	InternalPortBridge int
	BaseImage          string
	LLMEnvKey          string
	LLMKey             string
	GatewayToken       string
	HealthEndpoint     string
	MinMemory          string
	MinCPU             string
}

// AgentTemplates contains docker-compose templates for each agent type
var AgentTemplates = map[string]string{
	"openclaw": `version: '3.8'
services:
  agent:
    image: {{.BaseImage}}
    container_name: blytz-{{.CustomerID}}
    working_dir: /app
    user: "1000:1000"
    command: >
      sh -c "npm install -g openclaw@latest &&
             mkdir -p /home/node/.openclaw &&
             openclaw gateway --port {{.InternalPort}} --bind lan"
    ports:
      - "{{.ExternalPort}}:{{.InternalPort}}"
      - "{{.ExternalPortBridge}}:{{.InternalPortBridge}}"
    volumes:
      - ./config:/home/node/.openclaw
    env_file:
      - .env.secret
    environment:
      - HOME=/home/node
      - {{.LLMEnvKey}}=${{.LLMEnvKey}}
    deploy:
      resources:
        limits:
          memory: {{.MinMemory}}
          cpus: '{{.MinCPU}}'
        reservations:
          memory: 128M
          cpus: '0.1'
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:{{.InternalPort}}{{.HealthEndpoint}}"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
`,
	"myrai": `version: '3.8'
services:
  agent:
    image: {{.BaseImage}}
    container_name: blytz-{{.CustomerID}}
    command: ["myrai", "server", "--port", "{{.InternalPort}}"]
    ports:
      - "{{.ExternalPort}}:{{.InternalPort}}"
    volumes:
      - ./data:/app/data
    env_file:
      - .env.secret
    environment:
      - {{.LLMEnvKey}}=${{.LLMEnvKey}}
      - MYRAI_GATEWAY_TOKEN={{.GatewayToken}}
      - MYRAI_SERVER_PORT={{.InternalPort}}
      - MYRAI_SERVER_ADDRESS=0.0.0.0
    deploy:
      resources:
        limits:
          memory: {{.MinMemory}}
          cpus: '{{.MinCPU}}'
        reservations:
          memory: 128M
          cpus: '0.1'
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:{{.InternalPort}}{{.HealthEndpoint}}"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
`,
}

// NewComposeGenerator creates a new compose generator
func NewComposeGenerator(baseDir string) *ComposeGenerator {
	return &ComposeGenerator{baseDir: baseDir}
}

// Generate creates a docker-compose.yml for the specified agent type
func (cg *ComposeGenerator) Generate(config AgentConfig) error {
	templateStr, ok := AgentTemplates[config.AgentType]
	if !ok {
		return fmt.Errorf("unknown agent type: %s", config.AgentType)
	}

	tmpl, err := template.New("compose").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	customerDir := filepath.Join(cg.baseDir, config.CustomerID)
	if err := os.MkdirAll(customerDir, 0755); err != nil {
		return fmt.Errorf("create customer directory: %w", err)
	}

	composePath := filepath.Join(customerDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write compose file: %w", err)
	}

	return nil
}

// GenerateEnvFile creates the .env.secret file with API keys
func (cg *ComposeGenerator) GenerateEnvFile(customerID string, envVars map[string]string) error {
	var envContent string
	for key, value := range envVars {
		envContent += fmt.Sprintf("%s=%s\n", key, value)
	}

	customerDir := filepath.Join(cg.baseDir, customerID)
	envPath := filepath.Join(customerDir, ".env.secret")

	// Create customer directory if it doesn't exist
	if err := os.MkdirAll(customerDir, 0755); err != nil {
		return fmt.Errorf("create customer directory: %w", err)
	}

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		return fmt.Errorf("write env file: %w", err)
	}

	return nil
}
