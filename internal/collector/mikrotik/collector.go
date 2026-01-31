package mikrotik

import (
	"context"
	"fmt"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
)

// Collector implements the collector interface for MikroTik RouterOS
type Collector struct {
	name string
}

// NewCollector creates a new MikroTik collector
func NewCollector() collector.Collector {
	return &Collector{
		name: "mikrotik",
	}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return c.name
}

// Type returns the router type
func (c *Collector) Type() string {
	return "mikrotik"
}

// Collect collects metrics from a MikroTik router
func (c *Collector) Collect(ctx context.Context, router *models.RouterConfig) (*models.MetricsData, error) {
	// TODO: Implement actual MikroTik API collection
	// This is a stub implementation that returns sample data
	
	metrics := &models.MetricsData{
		RouterID:  router.ID,
		Timestamp: time.Now(),
		System: models.SystemMetrics{
			CPUPercent:         25.5,
			MemoryPercent:      45.2,
			MemoryTotalBytes:   1024 * 1024 * 1024, // 1GB
			MemoryUsedBytes:    461 * 1024 * 1024,  // ~450MB
			UptimeSeconds:      86400,              // 1 day
			TemperatureCelsius: 42.0,
			FirmwareVersion:    "7.12",
			BoardName:          "RB4011iGS+",
		},
		Interfaces: []models.InterfaceMetrics{
			{
				Name:        "ether1",
				Description: "WAN",
				IsUp:        true,
				SpeedMbps:   1000,
				RxBytes:     1024 * 1024 * 100, // 100MB
				TxBytes:     1024 * 1024 * 50,  // 50MB
				RxPackets:   10000,
				TxPackets:   8000,
				RxErrors:    0,
				TxErrors:    0,
				RxDrops:     0,
				TxDrops:     0,
			},
			{
				Name:        "ether2",
				Description: "LAN",
				IsUp:        true,
				SpeedMbps:   1000,
				RxBytes:     1024 * 1024 * 50,
				TxBytes:     1024 * 1024 * 100,
				RxPackets:   8000,
				TxPackets:   10000,
				RxErrors:    0,
				TxErrors:    0,
				RxDrops:     0,
				TxDrops:     0,
			},
		},
	}

	return metrics, nil
}

// HealthCheck verifies connectivity to the MikroTik router
func (c *Collector) HealthCheck(ctx context.Context, router *models.RouterConfig) error {
	// TODO: Implement actual health check using RouterOS API
	// For now, just verify configuration is present
	
	if router.Address == "" {
		return fmt.Errorf("router address is required")
	}
	
	if router.Credentials.Username == "" {
		return fmt.Errorf("router username is required")
	}
	
	if router.Credentials.Password == "" {
		return fmt.Errorf("router password is required")
	}

	// Stub: Return success for now
	return nil
}
