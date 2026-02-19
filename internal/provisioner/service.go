package provisioner

import (
	"context"
	"fmt"
	"log"

	"blytz/internal/caddy"
	"blytz/internal/db"
	"blytz/internal/telegram"
	"blytz/internal/workspace"

	"github.com/google/uuid"
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
	logger     *log.Logger
	baseDomain string
	openAIKey  string
	baseDir    string
	portStart  int
	portEnd    int
}

func NewService(database *db.DB, templatesDir, baseDir, openAIKey string, portStart, portEnd int, caddyClient *caddy.Client, baseDomain string, logger *log.Logger) *Service {
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

	if err := s.compose.Generate(customerID, port, s.openAIKey); err != nil {
		s.db.ReleasePort(ctx, port)
		s.ports.ReleasePort(port)
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("generate compose: %w", err)
	}

	if err := workspace.GenerateOpenClawConfig(s.baseDir, customerID, customer.TelegramBotToken, generateGatewayToken(), port); err != nil {
		s.cleanup(customerID, port)
		s.db.UpdateCustomerStatus(ctx, customerID, "pending")
		return fmt.Errorf("generate openclaw config: %w", err)
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
			s.logger.Printf("Failed to add Caddy subdomain (non-fatal): %v", err)
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
}
