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
    image: node:22-alpine
    container_name: blytz-%s
    working_dir: /app
    command: sh -c "npm install -g openclaw@latest && openclaw gateway --port 18789"
    ports:
      - "%d:18789"
    volumes:
      - ./.openclaw:/root/.openclaw
    env_file:
      - .env.secret
    environment:
      - OPENCLAW_STATE_DIR=/root/.openclaw
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '0.5'
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:18789/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
`, customerID, port)

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

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		return fmt.Errorf("write env file: %w", err)
	}

	return nil
}
