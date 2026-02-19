package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

func GenerateOpenClawConfig(baseDir, customerID string, botToken string, gatewayToken string, port int) error {
	config := fmt.Sprintf(`{
  "gateway": {
    "port": 18789,
    "auth": {
      "token": "%s"
    }
  },
  "agents": {
    "defaults": {
      "workspace": "/root/.openclaw/workspace"
    }
  },
  "channels": {
    "telegram": {
      "enabled": true,
      "botToken": "%s",
      "dmPolicy": "open",
      "allowFrom": ["*"]
    }
  }
}`, gatewayToken, botToken)

	configDir := filepath.Join(baseDir, customerID, ".openclaw")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "openclaw.json")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
