package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
	"gopkg.in/yaml.v3"
)

// Config represents the agent configuration
type Config struct {
	Agent      AgentConfig      `yaml:"agent"`
	Server     ServerConfig     `yaml:"server"`
	License    LicenseConfig    `yaml:"license"`
	Collection CollectionConfig `yaml:"collection"`
	Routers    []models.RouterConfig `yaml:"routers"`
	Privacy    PrivacyConfig    `yaml:"privacy"`
	Logging    LoggingConfig    `yaml:"logging"`
}

// AgentConfig contains agent identification
type AgentConfig struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

// ServerConfig contains gRPC server connection details
type ServerConfig struct {
	Address string    `yaml:"address"`
	TLS     TLSConfig `yaml:"tls"`
}

// TLSConfig contains TLS certificate configuration
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CACert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"client_cert"`
	ClientKey  string `yaml:"client_key"`
}

// LicenseConfig contains license validation settings
type LicenseConfig struct {
	Key               string `yaml:"key"`
	ValidationURL     string `yaml:"validation_url"`
	OfflineGraceHours int    `yaml:"offline_grace_hours"`
}

// CollectionConfig contains data collection settings
type CollectionConfig struct {
	IntervalSeconds int `yaml:"interval_seconds"`
}

// PrivacyConfig contains privacy and audit settings
type PrivacyConfig struct {
	AuditLogging      bool   `yaml:"audit_logging"`
	AuditLogPath      string `yaml:"audit_log_path"`
	RedactUsernames   bool   `yaml:"redact_usernames"`
	RedactIPAddresses bool   `yaml:"redact_ip_addresses"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Load loads configuration from a YAML file and expands environment variables
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in the config
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if cfg.Agent.ID == "" {
		cfg.Agent.ID = generateAgentID()
	}
	if cfg.Collection.IntervalSeconds == 0 {
		cfg.Collection.IntervalSeconds = 60
	}
	if cfg.License.OfflineGraceHours == 0 {
		cfg.License.OfflineGraceHours = 72
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}

	return &cfg, nil
}

// generateAgentID creates a unique agent ID based on hostname
func generateAgentID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("agent-%s", strings.ToLower(hostname))
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server.address is required")
	}
	if c.License.Key == "" {
		return fmt.Errorf("license.key is required")
	}
	if len(c.Routers) == 0 {
		return fmt.Errorf("at least one router must be configured")
	}
	for i, router := range c.Routers {
		if router.ID == "" {
			return fmt.Errorf("router[%d].id is required", i)
		}
		if router.Type == "" {
			return fmt.Errorf("router[%d].type is required", i)
		}
		if router.Address == "" {
			return fmt.Errorf("router[%d].address is required", i)
		}
	}
	return nil
}
