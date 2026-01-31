package mikrotik

import (
	"context"
	"testing"

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

func TestCollector_Collect(t *testing.T) {
	c := NewCollector()
	ctx := context.Background()

	router := &models.RouterConfig{
		ID:      "test-router",
		Name:    "Test Router",
		Type:    "mikrotik",
		Address: "192.168.1.1",
	}

	metrics, err := c.Collect(ctx, router)
	if err != nil {
		t.Fatalf("Collect() failed: %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	if metrics.RouterID != router.ID {
		t.Errorf("Expected router ID '%s', got '%s'", router.ID, metrics.RouterID)
	}

	if len(metrics.Interfaces) == 0 {
		t.Error("Expected at least one interface in stub data")
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
			wantErr: false,
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
			err := c.HealthCheck(ctx, tt.router)
			if (err != nil) != tt.wantErr {
				t.Errorf("HealthCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
