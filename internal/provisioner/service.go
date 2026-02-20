package provisioner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"blytz/internal/caddy"
	"blytz/internal/db"
	"blytz/internal/telegram"
	"blytz/internal/workspace"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func generateGatewayToken() string {
	return uuid.New().String()
}

type Service struct {
	db         *db.DB
	workspace  *workspace.Generator
	docker     *DockerProvisioner
	compose    *ComposeGenerator
	ports      *PortAllocator
	caddy      *caddy.Client
	logger     *zap.Logger
	baseDomain string
	openAIKey  string
	baseDir    string
	portStart  int
	portEnd    int
}

func NewService(database *db.DB, templatesDir, baseDir, openAIKey string, portStart, portEnd int, caddyClient *caddy.Client, baseDomain string, logger *zap.Logger) *Service {
	return &Service{
		db:         database,
		workspace:  workspace.NewWithBaseDir(templatesDir, baseDir),
		docker:     NewDockerProvisioner(baseDir),
		compose:    NewComposeGenerator(baseDir),
		ports:      NewPortAllocator(portStart, portEnd),
		caddy:      caddyClient,
		logger:     logger,
		baseDomain: baseDomain,
		openAIKey:  openAIKey,
		baseDir:    baseDir,
		portStart:  portStart,
		portEnd:    portEnd,
	}
}

func (s *Service) Provision(ctx context.Context, customerID string) error {
	customer, err := s.db.GetCustomerByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}

	if err := s.db.UpdateCustomerStatus(ctx, customerID, "provisioning"); err != nil {
		return fmt.Errorf("update status to provisioning: %w", err)
	}

	// Get agent type configuration
	agentType, err := s.db.GetAgentType(ctx, customer.AgentTypeID)
	if err != nil {
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("get agent type: %w", err)
	}

	// Get LLM provider configuration
	llmProvider, err := s.db.GetLLMProvider(ctx, customer.LLMProviderID)
	if err != nil {
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("get llm provider: %w", err)
	}

	if err := s.workspace.Generate(customerID, customer.AssistantName, customer.CustomInstructions); err != nil {
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("generate workspace: %w", err)
	}

	port, err := s.ports.AllocatePort()
	if err != nil {
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("allocate port: %w", err)
	}

	if err := s.db.AllocatePort(ctx, customerID, port); err != nil {
		s.ports.ReleasePort(port)
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("record port allocation: %w", err)
	}

	// Also update the customer's container_port field
	if err := s.db.UpdateCustomerPort(ctx, customerID, port); err != nil {
		s.cleanup(customerID, port)
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("update customer port: %w", err)
	}

	// Build agent configuration
	gatewayToken := generateGatewayToken()
	agentConfig := AgentConfig{
		CustomerID:         customerID,
		AgentType:          customer.AgentTypeID,
		ExternalPort:       port,
		ExternalPortBridge: port + 1,
		InternalPort:       agentType.InternalPort,
		InternalPortBridge: agentType.InternalPortBridge,
		BaseImage:          agentType.BaseImage,
		LLMEnvKey:          llmProvider.EnvKey,
		LLMKey:             s.openAIKey,
		GatewayToken:       gatewayToken,
		HealthEndpoint:     agentType.HealthEndpoint,
		MinMemory:          agentType.MinMemory,
		MinCPU:             agentType.MinCPU,
	}

	// Generate environment variables map
	envVars := map[string]string{
		llmProvider.EnvKey: s.openAIKey,
	}
	if customer.AgentTypeID == "myrai" {
		envVars["MYRAI_GATEWAY_TOKEN"] = gatewayToken
	}

	if err := s.compose.GenerateEnvFile(customerID, envVars); err != nil {
		s.cleanup(customerID, port)
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("generate env file: %w", err)
	}

	if err := s.compose.Generate(agentConfig); err != nil {
		s.cleanup(customerID, port)
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("generate compose: %w", err)
	}

	// Generate agent-specific config
	if customer.AgentTypeID == "openclaw" {
		if err := workspace.GenerateOpenClawConfig(s.baseDir, customerID, customer.TelegramBotToken, gatewayToken, port); err != nil {
			s.cleanup(customerID, port)
			s.db.UpdateCustomerStatus(ctx, customerID, "pending")
			return fmt.Errorf("generate openclaw config: %w", err)
		}
	}

	if err := s.docker.Create(ctx, customerID); err != nil {
		s.cleanup(customerID, port)
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("create container: %w", err)
	}

	if err := s.docker.Start(ctx, customerID); err != nil {
		s.docker.Remove(ctx, customerID)
		s.cleanup(customerID, port)
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("start container: %w", err)
	}

	if err := s.db.UpdateCustomerStatus(ctx, customerID, "active"); err != nil {
		return fmt.Errorf("update status to active: %w", err)
	}

	if s.caddy != nil {
		subdomain := fmt.Sprintf("%s.%s", customerID, s.baseDomain)
		target := fmt.Sprintf("localhost:%d", port)
		if err := s.caddy.AddSubdomain(subdomain, target); err != nil {
			s.logger.Warn("Failed to add Caddy subdomain (non-fatal)", zap.Error(err))
		}
	}

	return nil
}

func (s *Service) Suspend(ctx context.Context, customerID string) error {
	if err := s.docker.Stop(ctx, customerID); err != nil {
		return fmt.Errorf("stop container: %w", err)
	}

	if err := s.db.UpdateCustomerStatus(ctx, customerID, "suspended"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

func (s *Service) Resume(ctx context.Context, customerID string) error {
	if err := s.docker.Start(ctx, customerID); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	if err := s.db.UpdateCustomerStatus(ctx, customerID, "active"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

func (s *Service) Terminate(ctx context.Context, customerID string) error {
	customer, err := s.db.GetCustomerByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}

	if customer.ContainerPort != nil {
		s.db.ReleasePort(ctx, *customer.ContainerPort)
		s.ports.ReleasePort(*customer.ContainerPort)
		// Clear the container_port from customer record
		if err := s.db.ClearCustomerPort(ctx, customerID); err != nil {
			return fmt.Errorf("clear customer port: %w", err)
		}
	}

	if err := s.docker.Remove(ctx, customerID); err != nil {
		return fmt.Errorf("remove container: %w", err)
	}

	if err := s.db.UpdateCustomerStatus(ctx, customerID, "cancelled"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

func (s *Service) ValidateBotToken(token string) (*telegram.BotInfo, error) {
	return telegram.ValidateToken(token)
}

func (s *Service) cleanup(customerID string, port int) {
	s.docker.Remove(context.Background(), customerID)
	s.db.ReleasePort(context.Background(), port)
	s.ports.ReleasePort(port)

	// Remove env file
	envPath := filepath.Join(s.baseDir, customerID, ".env.secret")
	os.Remove(envPath) // Ignore errors
}
