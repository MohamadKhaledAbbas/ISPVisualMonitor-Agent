package mikrotik

import (
	"context"
	"testing"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Error("Expected non-nil collector")
	}

	if c.Name() != "mikrotik" {
		t.Errorf("Expected name 'mikrotik', got '%s'", c.Name())
	}

	if c.Type() != "mikrotik" {
		t.Errorf("Expected type 'mikrotik', got '%s'", c.Type())
	}
}

func TestNewCollectorWithConfig(t *testing.T) {
	cfg := &Config{
		API: APIConfig{
			Port:    8729,
			UseTLS:  true,
			Timeout: 5 * time.Second,
		},
		Collect: CollectConfig{
			System:     true,
			Interfaces: true,
			PPPoE:      false,
			NAT:        false,
			DHCP:       false,
		},
	}

	c := NewCollectorWithConfig(cfg)
	if c == nil {
		t.Error("Expected non-nil collector")
	}

	gotCfg := c.GetConfig()
	if gotCfg.API.Port != 8729 {
		t.Errorf("Expected port 8729, got %d", gotCfg.API.Port)
	}
	if !gotCfg.API.UseTLS {
		t.Error("Expected TLS to be enabled")
	}
	if gotCfg.API.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", gotCfg.API.Timeout)
	}
}

func TestCollector_SetConfig(t *testing.T) {
	c := NewCollectorWithConfig(nil)

	newCfg := &Config{
		API: APIConfig{
			Port:    8728,
			UseTLS:  false,
			Timeout: 30 * time.Second,
		},
	}

	c.SetConfig(newCfg)

	gotCfg := c.GetConfig()
	if gotCfg.API.Port != 8728 {
		t.Errorf("Expected port 8728, got %d", gotCfg.API.Port)
	}
}

func TestCollector_HealthCheck(t *testing.T) {
	c := NewCollector()
	ctx := context.Background()

	tests := []struct {
		name    string
		router  *models.RouterConfig
		wantErr bool
	}{
		{
			name: "valid config",
			router: &models.RouterConfig{
				ID:      "test",
				Address: "192.168.1.1",
				Credentials: models.RouterCredentials{
					Username: "admin",
					Password: "password",
				},
			},
			// Note: This will fail connection but pass validation
			wantErr: true, // Connection will fail since no router is available
		},
		{
			name: "missing address",
			router: &models.RouterConfig{
				ID: "test",
				Credentials: models.RouterCredentials{
					Username: "admin",
					Password: "password",
				},
			},
			wantErr: true,
		},
		{
			name: "missing username",
			router: &models.RouterConfig{
				ID:      "test",
				Address: "192.168.1.1",
				Credentials: models.RouterCredentials{
					Password: "password",
				},
			},
			wantErr: true,
		},
		{
			name: "missing password",
			router: &models.RouterConfig{
				ID:      "test",
				Address: "192.168.1.1",
				Credentials: models.RouterCredentials{
					Username: "admin",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a short timeout context for connection tests
			testCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cancel()

			err := c.HealthCheck(testCtx, tt.router)
			if (err != nil) != tt.wantErr {
				t.Errorf("HealthCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollector_HealthCheck_ValidatesConfig(t *testing.T) {
	c := NewCollector()
	ctx := context.Background()

	// Test missing address
	router := &models.RouterConfig{
		ID: "test",
		Credentials: models.RouterCredentials{
			Username: "admin",
			Password: "password",
		},
	}

	err := c.HealthCheck(ctx, router)
	if err == nil {
		t.Error("Expected error for missing address")
	}

	// Test missing username
	router.Address = "192.168.1.1"
	router.Credentials.Username = ""

	err = c.HealthCheck(ctx, router)
	if err == nil {
		t.Error("Expected error for missing username")
	}

	// Test missing password
	router.Credentials.Username = "admin"
	router.Credentials.Password = ""

	err = c.HealthCheck(ctx, router)
	if err == nil {
		t.Error("Expected error for missing password")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.API.Port != 8728 {
		t.Errorf("Expected default port 8728, got %d", cfg.API.Port)
	}
	if cfg.API.UseTLS {
		t.Error("Expected TLS to be disabled by default")
	}
	if cfg.API.Timeout != 10*time.Second {
		t.Errorf("Expected default timeout 10s, got %v", cfg.API.Timeout)
	}
	if cfg.API.RetryAttempts != 3 {
		t.Errorf("Expected 3 retry attempts, got %d", cfg.API.RetryAttempts)
	}
	if !cfg.Collect.System {
		t.Error("Expected system collection to be enabled by default")
	}
	if !cfg.Collect.Interfaces {
		t.Error("Expected interface collection to be enabled by default")
	}
	if cfg.Collect.NAT {
		t.Error("Expected NAT collection to be disabled by default")
	}
}

func TestConfig_WithTLS(t *testing.T) {
	cfg := DefaultConfig().WithTLS()

	if !cfg.API.UseTLS {
		t.Error("Expected TLS to be enabled")
	}
	if cfg.API.Port != 8729 {
		t.Errorf("Expected TLS port 8729, got %d", cfg.API.Port)
	}
}

func TestConfig_WithTimeout(t *testing.T) {
	cfg := DefaultConfig().WithTimeout(30 * time.Second)

	if cfg.API.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cfg.API.Timeout)
	}
}

func TestConfig_WithRetry(t *testing.T) {
	cfg := DefaultConfig().WithRetry(5, 2*time.Second)

	if cfg.API.RetryAttempts != 5 {
		t.Errorf("Expected 5 retry attempts, got %d", cfg.API.RetryAttempts)
	}
	if cfg.API.RetryDelay != 2*time.Second {
		t.Errorf("Expected 2s retry delay, got %v", cfg.API.RetryDelay)
	}
}

func TestConfig_WithInterfaceFilter(t *testing.T) {
	include := []string{"ether*", "sfp*"}
	exclude := []string{"bridge-local"}

	cfg := DefaultConfig().WithInterfaceFilter(include, exclude)

	if len(cfg.InterfaceInclude) != 2 {
		t.Errorf("Expected 2 include patterns, got %d", len(cfg.InterfaceInclude))
	}
	if len(cfg.InterfaceExclude) != 1 {
		t.Errorf("Expected 1 exclude pattern, got %d", len(cfg.InterfaceExclude))
	}
}

func TestConfig_WithNATSampling(t *testing.T) {
	cfg := DefaultConfig().WithNATSampling(0.5, 5000)

	if !cfg.NAT.SamplingEnabled {
		t.Error("Expected NAT sampling to be enabled")
	}
	if cfg.NAT.SampleRate != 0.5 {
		t.Errorf("Expected sample rate 0.5, got %f", cfg.NAT.SampleRate)
	}
	if cfg.NAT.MaxConnections != 5000 {
		t.Errorf("Expected max connections 5000, got %d", cfg.NAT.MaxConnections)
	}
}

func TestConfig_EnableDisableAll(t *testing.T) {
	cfg := DefaultConfig().DisableAll()

	if cfg.Collect.System {
		t.Error("Expected system collection to be disabled")
	}
	if cfg.Collect.Interfaces {
		t.Error("Expected interface collection to be disabled")
	}

	cfg.EnableAll()

	if !cfg.Collect.System {
		t.Error("Expected system collection to be enabled")
	}
	if !cfg.Collect.NAT {
		t.Error("Expected NAT collection to be enabled")
	}
}

func TestInterfaceTracker(t *testing.T) {
	tracker := newInterfaceTracker()

	// First call should not calculate rates (no previous data)
	rxBps, txBps, rxPps, txPps := tracker.updateAndCalculateRates("ether1", 1000, 500, 100, 50)
	if rxBps != 0 || txBps != 0 {
		t.Error("First call should return 0 rates")
	}
	if rxPps != 0 || txPps != 0 {
		t.Error("First call should return 0 packet rates")
	}

	// Second call with same values should return 0 rates (no change)
	rxBps, txBps, rxPps, txPps = tracker.updateAndCalculateRates("ether1", 1000, 500, 100, 50)
	if rxBps != 0 || txBps != 0 {
		t.Error("No change should return 0 rates")
	}
}

func TestCollectedData(t *testing.T) {
	data := &CollectedData{
		MetricsData: &models.MetricsData{
			RouterID:  "test-router",
			Timestamp: time.Now(),
		},
		CollectedAt: time.Now(),
	}

	if data.RouterID != "test-router" {
		t.Errorf("Expected router ID 'test-router', got %q", data.RouterID)
	}

	if data.CollectedAt.IsZero() {
		t.Error("Expected non-zero collection time")
	}
}
