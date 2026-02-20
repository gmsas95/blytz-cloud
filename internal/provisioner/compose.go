package provisioner

import (
	"fmt"
	"os"
	"path/filepath"
)

type ComposeGenerator struct {
	baseDir string
}

func NewComposeGenerator(baseDir string) *ComposeGenerator {
	return &ComposeGenerator{baseDir: baseDir}
}

func (cg *ComposeGenerator) Generate(customerID string, port int, openAIKey string) error {
	compose := fmt.Sprintf(`version: '3.8'
services:
  openclaw:
    image: node:22-bookworm
    container_name: blytz-%s
    working_dir: /app
    user: "1000:1000"
    command: >
      sh -c "npm install -g openclaw@latest &&
             mkdir -p /home/node/.openclaw &&
             openclaw gateway --port 18789 --bind lan"
    ports:
      - "%d:18789"
      - "%d:18790"
    volumes:
      - ./.openclaw:/home/node/.openclaw
    env_file:
      - .env.secret
    environment:
      - HOME=/home/node
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.25'
        reservations:
          memory: 128M
          cpus: '0.1'
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:18789/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
`, customerID, port, port+1)

	customerDir := filepath.Join(cg.baseDir, customerID)
	if err := os.MkdirAll(customerDir, 0755); err != nil {
		return fmt.Errorf("create customer directory: %w", err)
	}

	composePath := filepath.Join(customerDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(compose), 0644); err != nil {
		return fmt.Errorf("write compose file: %w", err)
	}

	return nil
}

func (cg *ComposeGenerator) GenerateEnvFile(customerID string, openAIKey string) error {
	envContent := fmt.Sprintf("OPENAI_API_KEY=%s\n", openAIKey)

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
