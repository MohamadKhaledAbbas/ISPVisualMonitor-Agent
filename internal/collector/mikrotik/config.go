package mikrotik

import (
	"time"
)

// Config contains configuration for the MikroTik collector.
type Config struct {
	// API connection settings
	API APIConfig `yaml:"api"`

	// What data to collect
	Collect CollectConfig `yaml:"collect"`

	// Interface filtering
	InterfaceInclude []string `yaml:"interface_include,omitempty"`
	InterfaceExclude []string `yaml:"interface_exclude,omitempty"`

	// NAT collection settings
	NAT NATConfig `yaml:"nat,omitempty"`
}

// APIConfig contains API connection settings.
type APIConfig struct {
	Port          int           `yaml:"port"`
	UseTLS        bool          `yaml:"use_tls"`
	Timeout       time.Duration `yaml:"timeout"`
	RetryAttempts int           `yaml:"retry_attempts"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
}

// CollectConfig controls which metrics to collect.
type CollectConfig struct {
	System     bool `yaml:"system"`
	Interfaces bool `yaml:"interfaces"`
	PPPoE      bool `yaml:"pppoe"`
	NAT        bool `yaml:"nat"`
	DHCP       bool `yaml:"dhcp"`
}

// NATConfig contains NAT-specific collection settings.
type NATConfig struct {
	// SamplingEnabled enables connection sampling for high-traffic routers
	SamplingEnabled bool `yaml:"sampling_enabled"`
	// SampleRate is the fraction of connections to sample (0.0-1.0)
	SampleRate float64 `yaml:"sample_rate"`
	// MaxConnections limits the number of connections to collect
	MaxConnections int `yaml:"max_connections"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			Port:          8728,
			UseTLS:        false,
			Timeout:       10 * time.Second,
			RetryAttempts: 3,
			RetryDelay:    time.Second,
		},
		Collect: CollectConfig{
			System:     true,
			Interfaces: true,
			PPPoE:      true,
			NAT:        false, // Disabled by default due to performance impact
			DHCP:       true,
		},
		NAT: NATConfig{
			SamplingEnabled: false,
			SampleRate:      1.0,
			MaxConnections:  10000,
		},
	}
}

// WithTLS returns a config with TLS enabled.
func (c *Config) WithTLS() *Config {
	c.API.UseTLS = true
	c.API.Port = 8729
	return c
}

// WithTimeout returns a config with custom timeout.
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.API.Timeout = timeout
	return c
}

// WithRetry returns a config with custom retry settings.
func (c *Config) WithRetry(attempts int, delay time.Duration) *Config {
	c.API.RetryAttempts = attempts
	c.API.RetryDelay = delay
	return c
}

// WithInterfaceFilter returns a config with interface filtering.
func (c *Config) WithInterfaceFilter(include, exclude []string) *Config {
	c.InterfaceInclude = include
	c.InterfaceExclude = exclude
	return c
}

// WithNATSampling returns a config with NAT sampling enabled.
func (c *Config) WithNATSampling(rate float64, maxConn int) *Config {
	c.NAT.SamplingEnabled = true
	c.NAT.SampleRate = rate
	c.NAT.MaxConnections = maxConn
	return c
}

// EnableAll enables collection of all metric types.
func (c *Config) EnableAll() *Config {
	c.Collect.System = true
	c.Collect.Interfaces = true
	c.Collect.PPPoE = true
	c.Collect.NAT = true
	c.Collect.DHCP = true
	return c
}

// DisableAll disables collection of all metric types.
func (c *Config) DisableAll() *Config {
	c.Collect.System = false
	c.Collect.Interfaces = false
	c.Collect.PPPoE = false
	c.Collect.NAT = false
	c.Collect.DHCP = false
	return c
}
