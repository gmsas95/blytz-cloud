package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"PORT":          "9090",
				"MAX_CUSTOMERS": "50",
			},
			wantErr: false,
		},
		{
			name: "invalid port range",
			envVars: map[string]string{
				"PORT_RANGE_START": "30000",
				"PORT_RANGE_END":   "30000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cfg == nil {
				t.Error("Load() returned nil config without error")
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				MaxCustomers:   20,
				PortRangeStart: 30000,
				PortRangeEnd:   30999,
			},
			wantErr: false,
		},
		{
			name: "negative max customers",
			cfg: &Config{
				MaxCustomers:   -1,
				PortRangeStart: 30000,
				PortRangeEnd:   30999,
			},
			wantErr: true,
		},
		{
			name: "insufficient port range",
			cfg: &Config{
				MaxCustomers:   1000,
				PortRangeStart: 30000,
				PortRangeEnd:   30050,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
