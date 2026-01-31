package config

import (
	"os"
	"testing"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := `
agent:
  id: "test-agent"
  name: "Test Agent"

server:
  address: "localhost:50051"
  tls:
    enabled: false

license:
  key: "test-license-key"
  validation_url: "localhost:50052"
  offline_grace_hours: 72

collection:
  interval_seconds: 60

routers:
  - id: "router-01"
    name: "Test Router"
    type: "mikrotik"
    address: "192.168.1.1"
    credentials:
      username: "admin"
      password: "admin"
    collect:
      system: true
      interfaces: true
      pppoe_sessions: true

privacy:
  audit_logging: true
  audit_log_path: "/tmp/audit.log"

logging:
  level: "info"
  format: "json"
  output: "stdout"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading config
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if cfg.Agent.ID != "test-agent" {
		t.Errorf("Expected agent.id 'test-agent', got '%s'", cfg.Agent.ID)
	}
	if cfg.Server.Address != "localhost:50051" {
		t.Errorf("Expected server.address 'localhost:50051', got '%s'", cfg.Server.Address)
	}
	if len(cfg.Routers) != 1 {
		t.Errorf("Expected 1 router, got %d", len(cfg.Routers))
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{Address: "localhost:50051"},
				License: LicenseConfig{Key: "test-key"},
				Routers: []models.RouterConfig{
					{ID: "r1", Type: "mikrotik", Address: "192.168.1.1"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing server address",
			config: &Config{
				License: LicenseConfig{Key: "test-key"},
				Routers: []models.RouterConfig{
					{ID: "r1", Type: "mikrotik", Address: "192.168.1.1"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing license key",
			config: &Config{
				Server: ServerConfig{Address: "localhost:50051"},
				Routers: []models.RouterConfig{
					{ID: "r1", Type: "mikrotik", Address: "192.168.1.1"},
				},
			},
			wantErr: true,
		},
		{
			name: "no routers",
			config: &Config{
				Server:  ServerConfig{Address: "localhost:50051"},
				License: LicenseConfig{Key: "test-key"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvVarExpansion(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_LICENSE_KEY", "my-secret-key")
	defer os.Unsetenv("TEST_LICENSE_KEY")

	content := `
agent:
  name: "Test"
server:
  address: "localhost:50051"
  tls:
    enabled: false
license:
  key: "${TEST_LICENSE_KEY}"
routers:
  - id: "r1"
    name: "Router"
    type: "mikrotik"
    address: "192.168.1.1"
    credentials:
      username: "admin"
      password: "admin"
    collect:
      system: true
`

	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.License.Key != "my-secret-key" {
		t.Errorf("Environment variable not expanded, got '%s'", cfg.License.Key)
	}
}
