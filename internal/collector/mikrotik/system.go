package mikrotik

import (
	"context"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector/mikrotik/api"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
)

// SystemMetrics contains extended system metrics.
type SystemMetrics struct {
	models.SystemMetrics
	
	// Extended fields
	Architecture    string  `json:"architecture,omitempty"`
	LicenseLevel    int     `json:"license_level,omitempty"`
	DiskTotalBytes  int64   `json:"disk_total_bytes,omitempty"`
	DiskUsedBytes   int64   `json:"disk_used_bytes,omitempty"`
	DiskFreeBytes   int64   `json:"disk_free_bytes,omitempty"`
	DiskPercent     float64 `json:"disk_percent,omitempty"`
	VoltageMV       int64   `json:"voltage_mv,omitempty"`
	FanSpeedRPM     int64   `json:"fan_speed_rpm,omitempty"`
	RouterIdentity  string  `json:"router_identity,omitempty"`
}

// collectSystem collects system resource metrics from the router.
func (c *Collector) collectSystem(ctx context.Context, client *api.Client) (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// Get system resource info
	resource, err := client.RunOne(ctx, "/system/resource/print", nil)
	if err != nil {
		return nil, err
	}

	if resource != nil {
		metrics.CPUPercent = ParseFloat64(resource["cpu-load"])
		metrics.UptimeSeconds = ParseUptime(resource["uptime"])
		metrics.FirmwareVersion = resource["version"]
		metrics.BoardName = resource["board-name"]
		metrics.Architecture = resource["architecture-name"]
		
		// Memory
		metrics.MemoryTotalBytes = ParseInt64(resource["total-memory"])
		metrics.MemoryUsedBytes = metrics.MemoryTotalBytes - ParseInt64(resource["free-memory"])
		if metrics.MemoryTotalBytes > 0 {
			metrics.MemoryPercent = float64(metrics.MemoryUsedBytes) / float64(metrics.MemoryTotalBytes) * 100
		}
		
		// Disk
		metrics.DiskTotalBytes = ParseInt64(resource["total-hdd-space"])
		metrics.DiskFreeBytes = ParseInt64(resource["free-hdd-space"])
		metrics.DiskUsedBytes = metrics.DiskTotalBytes - metrics.DiskFreeBytes
		if metrics.DiskTotalBytes > 0 {
			metrics.DiskPercent = float64(metrics.DiskUsedBytes) / float64(metrics.DiskTotalBytes) * 100
		}
	}

	// Get router identity
	identity, err := client.RunOne(ctx, "/system/identity/print", nil)
	if err == nil && identity != nil {
		metrics.RouterIdentity = identity["name"]
	}

	// Get license level (optional, may fail on some models)
	license, err := client.RunOne(ctx, "/system/license/print", nil)
	if err == nil && license != nil {
		metrics.LicenseLevel = int(ParseInt64(license["level"]))
	}

	// Get health info (optional, hardware dependent)
	health, err := client.Run(ctx, "/system/health/print", nil)
	if err == nil && len(health) > 0 {
		for _, h := range health {
			name := h["name"]
			value := h["value"]
			
			switch name {
			case "temperature", "cpu-temperature":
				metrics.TemperatureCelsius = ParseFloat64(value)
			case "voltage":
				metrics.VoltageMV = int64(ParseFloat64(value) * 1000)
			case "fan-speed", "fan1-speed":
				metrics.FanSpeedRPM = ParseInt64(value)
			}
		}
	}

	// RouterOS 6.x health format (single record with multiple fields)
	if len(health) == 1 {
		h := health[0]
		if temp, ok := h["temperature"]; ok {
			metrics.TemperatureCelsius = ParseFloat64(temp)
		}
		if voltage, ok := h["voltage"]; ok {
			metrics.VoltageMV = int64(ParseFloat64(voltage) * 1000)
		}
		if fan, ok := h["fan-speed"]; ok {
			metrics.FanSpeedRPM = ParseInt64(fan)
		}
	}

	return metrics, nil
}
