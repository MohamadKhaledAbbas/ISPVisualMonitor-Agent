package models

import "time"

// RouterConfig represents a router to monitor
type RouterConfig struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Type        string                 `yaml:"type"`
	Address     string                 `yaml:"address"`
	Credentials RouterCredentials      `yaml:"credentials"`
	Collect     CollectorFlags         `yaml:"collect"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

// RouterCredentials holds authentication information
type RouterCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSHKey   string `yaml:"ssh_key,omitempty"`
}

// CollectorFlags controls what data to collect
type CollectorFlags struct {
	System        bool `yaml:"system"`
	Interfaces    bool `yaml:"interfaces"`
	PPPoESessions bool `yaml:"pppoe_sessions"`
	NATSessions   bool `yaml:"nat_sessions"`
	DHCPLeases    bool `yaml:"dhcp_leases"`
}

// MetricsData represents collected metrics from a router
type MetricsData struct {
	RouterID  string
	Timestamp time.Time
	System    SystemMetrics
	Interfaces []InterfaceMetrics
}

// SystemMetrics represents router system metrics
type SystemMetrics struct {
	CPUPercent        float64
	MemoryPercent     float64
	MemoryTotalBytes  int64
	MemoryUsedBytes   int64
	UptimeSeconds     int64
	TemperatureCelsius float64
	FirmwareVersion   string
	BoardName         string
}

// InterfaceMetrics represents network interface metrics
type InterfaceMetrics struct {
	Name        string
	Description string
	IsUp        bool
	SpeedMbps   int64
	RxBytes     int64
	TxBytes     int64
	RxPackets   int64
	TxPackets   int64
	RxErrors    int64
	TxErrors    int64
	RxDrops     int64
	TxDrops     int64
}
