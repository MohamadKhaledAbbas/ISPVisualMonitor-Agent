package mikrotik

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector/mikrotik/api"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
)

// Collector implements the collector interface for MikroTik RouterOS.
type Collector struct {
	name        string
	config      *Config
	ifaceTracker *interfaceTracker
	mu          sync.RWMutex
}

// CollectedData contains all data collected from a MikroTik router.
type CollectedData struct {
	*models.MetricsData
	System       *SystemMetrics      `json:"system,omitempty"`
	Interfaces   []InterfaceMetrics  `json:"interfaces,omitempty"`
	PPPoE        []PPPoESession      `json:"pppoe_sessions,omitempty"`
	PPPoEServers []PPPoEServerStats  `json:"pppoe_servers,omitempty"`
	NAT          []NATConnection     `json:"nat_connections,omitempty"`
	NATStats     *NATStats           `json:"nat_stats,omitempty"`
	DHCPLeases   []DHCPLease         `json:"dhcp_leases,omitempty"`
	DHCPPools    []DHCPPoolStats     `json:"dhcp_pools,omitempty"`
	DHCPServers  []DHCPServerStats   `json:"dhcp_servers,omitempty"`
	CollectedAt  time.Time           `json:"collected_at"`
	Errors       []string            `json:"errors,omitempty"`
}

// NewCollector creates a new MikroTik collector with default configuration.
func NewCollector() collector.Collector {
	return NewCollectorWithConfig(DefaultConfig())
}

// NewCollectorWithConfig creates a new MikroTik collector with custom configuration.
func NewCollectorWithConfig(config *Config) *Collector {
	if config == nil {
		config = DefaultConfig()
	}
	return &Collector{
		name:         "mikrotik",
		config:       config,
		ifaceTracker: newInterfaceTracker(),
	}
}

// Name returns the collector name.
func (c *Collector) Name() string {
	return c.name
}

// Type returns the router type.
func (c *Collector) Type() string {
	return "mikrotik"
}

// SetConfig updates the collector configuration.
func (c *Collector) SetConfig(config *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config
}

// GetConfig returns a copy of the current configuration.
func (c *Collector) GetConfig() *Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.config == nil {
		return DefaultConfig()
	}
	
	// Return a shallow copy
	cfg := *c.config
	return &cfg
}

// Collect collects metrics from a MikroTik router.
func (c *Collector) Collect(ctx context.Context, router *models.RouterConfig) (*models.MetricsData, error) {
	data, err := c.CollectAll(ctx, router)
	if err != nil {
		return nil, err
	}
	return data.MetricsData, nil
}

// CollectAll collects all configured metrics from a MikroTik router.
func (c *Collector) CollectAll(ctx context.Context, router *models.RouterConfig) (*CollectedData, error) {
	c.mu.RLock()
	cfg := c.config
	c.mu.RUnlock()

	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Create API client
	client := c.createClient(router, cfg)

	// Connect to router
	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to router: %w", err)
	}
	defer client.Close()

	data := &CollectedData{
		MetricsData: &models.MetricsData{
			RouterID:  router.ID,
			Timestamp: time.Now(),
		},
		CollectedAt: time.Now(),
	}

	// Collect system metrics
	if cfg.Collect.System {
		sysMetrics, err := c.collectSystem(ctx, client)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Sprintf("system: %v", err))
		} else {
			data.System = sysMetrics
			data.MetricsData.System = sysMetrics.SystemMetrics
		}
	}

	// Collect interface metrics
	if cfg.Collect.Interfaces {
		ifaceMetrics, err := c.collectInterfaces(ctx, client)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Sprintf("interfaces: %v", err))
		} else {
			data.Interfaces = ifaceMetrics
			// Convert to base model
			for _, iface := range ifaceMetrics {
				data.MetricsData.Interfaces = append(data.MetricsData.Interfaces, iface.InterfaceMetrics)
			}
		}
	}

	// Collect PPPoE sessions
	if cfg.Collect.PPPoE {
		sessions, servers, err := c.collectPPPoE(ctx, client)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Sprintf("pppoe: %v", err))
		} else {
			data.PPPoE = sessions
			data.PPPoEServers = servers
		}
	}

	// Collect NAT connections
	if cfg.Collect.NAT {
		connections, stats, err := c.collectNAT(ctx, client)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Sprintf("nat: %v", err))
		} else {
			data.NAT = connections
			data.NATStats = stats
		}
	}

	// Collect DHCP leases
	if cfg.Collect.DHCP {
		leases, pools, servers, err := c.collectDHCP(ctx, client)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Sprintf("dhcp: %v", err))
		} else {
			data.DHCPLeases = leases
			data.DHCPPools = pools
			data.DHCPServers = servers
		}
	}

	return data, nil
}

// HealthCheck verifies connectivity to the MikroTik router.
func (c *Collector) HealthCheck(ctx context.Context, router *models.RouterConfig) error {
	if router.Address == "" {
		return fmt.Errorf("router address is required")
	}

	if router.Credentials.Username == "" {
		return fmt.Errorf("router username is required")
	}

	if router.Credentials.Password == "" {
		return fmt.Errorf("router password is required")
	}

	c.mu.RLock()
	cfg := c.config
	c.mu.RUnlock()

	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Create API client and test connection
	client := c.createClient(router, cfg)

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Ping to verify connection is working
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// createClient creates an API client for the given router configuration.
func (c *Collector) createClient(router *models.RouterConfig, cfg *Config) *api.Client {
	port := cfg.API.Port
	if port == 0 {
		if cfg.API.UseTLS {
			port = 8729
		} else {
			port = 8728
		}
	}

	clientConfig := &api.ClientConfig{
		Address:            fmt.Sprintf("%s:%d", router.Address, port),
		Username:           router.Credentials.Username,
		Password:           router.Credentials.Password,
		UseTLS:             cfg.API.UseTLS,
		InsecureSkipVerify: cfg.API.InsecureSkipVerify,
		Timeout:            cfg.API.Timeout,
		RetryAttempts:      cfg.API.RetryAttempts,
		RetryDelay:         cfg.API.RetryDelay,
	}

	if clientConfig.Timeout == 0 {
		clientConfig.Timeout = 10 * time.Second
	}
	if clientConfig.RetryAttempts == 0 {
		clientConfig.RetryAttempts = 3
	}
	if clientConfig.RetryDelay == 0 {
		clientConfig.RetryDelay = time.Second
	}

	return api.NewClient(clientConfig)
}
