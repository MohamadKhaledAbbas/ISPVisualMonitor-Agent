package collector

import (
	"context"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
)

// Collector defines the interface for collecting data from routers
type Collector interface {
	// Name returns the collector name
	Name() string

	// Type returns the router type this collector supports
	Type() string

	// Collect collects metrics from the router
	Collect(ctx context.Context, router *models.RouterConfig) (*models.MetricsData, error)

	// HealthCheck verifies the collector can connect to the router
	HealthCheck(ctx context.Context, router *models.RouterConfig) error
}
