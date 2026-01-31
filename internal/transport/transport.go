package transport

import (
	"context"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
)

// Transport defines the interface for sending data to the server
type Transport interface {
	// Connect establishes a connection to the server
	Connect(ctx context.Context) error

	// SendMetrics sends metrics data to the server
	SendMetrics(ctx context.Context, data *models.MetricsData) error

	// SendHeartbeat sends a heartbeat to the server
	SendHeartbeat(ctx context.Context) error

	// Close closes the connection
	Close() error
}
