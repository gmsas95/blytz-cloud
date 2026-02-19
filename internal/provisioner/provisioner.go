package provisioner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type DockerProvisioner struct {
	baseDir string
}

func NewDockerProvisioner(baseDir string) *DockerProvisioner {
	return &DockerProvisioner{baseDir: baseDir}
}

func (dp *DockerProvisioner) Create(ctx context.Context, customerID string) error {
	customerDir := filepath.Join(dp.baseDir, customerID)
	composePath := filepath.Join(customerDir, "docker-compose.yml")

	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found for customer %s", customerID)
	}

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "create")
	cmd.Dir = customerDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create container: %w (output: %s)", err, string(output))
	}

	return nil
}

func (dp *DockerProvisioner) Start(ctx context.Context, customerID string) error {
	customerDir := filepath.Join(dp.baseDir, customerID)
	composePath := filepath.Join(customerDir, "docker-compose.yml")

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "up", "-d")
	cmd.Dir = customerDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("start container: %w (output: %s)", err, string(output))
	}

	return nil
}

func (dp *DockerProvisioner) Stop(ctx context.Context, customerID string) error {
	customerDir := filepath.Join(dp.baseDir, customerID)
	composePath := filepath.Join(customerDir, "docker-compose.yml")

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "stop")
	cmd.Dir = customerDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop container: %w (output: %s)", err, string(output))
	}

	return nil
}

func (dp *DockerProvisioner) Remove(ctx context.Context, customerID string) error {
	customerDir := filepath.Join(dp.baseDir, customerID)
	composePath := filepath.Join(customerDir, "docker-compose.yml")

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "down", "-v")
	cmd.Dir = customerDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("remove container: %w (output: %s)", err, string(output))
	}

	return nil
}

func (dp *DockerProvisioner) GetStatus(ctx context.Context, customerID string) (string, error) {
	containerName := fmt.Sprintf("blytz-%s", customerID)

	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Status}}", containerName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "not_found", nil
		}
		return "", fmt.Errorf("inspect container: %w", err)
	}

	return string(output), nil
}
