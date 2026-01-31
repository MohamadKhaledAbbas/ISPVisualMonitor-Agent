package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/collector/mikrotik"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/config"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/license"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/privacy"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/transport/grpc"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/models"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/version"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "/etc/ispagent/agent.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ISP Visual Monitor Agent v%s\n", version.GetVersion())
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Starting ISP Visual Monitor Agent v%s", version.GetVersion())
	log.Printf("Agent ID: %s", cfg.Agent.ID)

	// Initialize audit logger if enabled
	var auditLogger *privacy.AuditLogger
	if cfg.Privacy.AuditLogging {
		auditLogger, err = privacy.NewAuditLogger(cfg.Privacy.AuditLogPath)
		if err != nil {
			log.Fatalf("Failed to initialize audit logger: %v", err)
		}
		defer auditLogger.Close()
		log.Printf("Audit logging enabled: %s", cfg.Privacy.AuditLogPath)
	}

	// Initialize license client
	ctx := context.Background()
	licenseClient, err := license.NewClient(cfg.License.ValidationURL, cfg.License.Key, cfg.Agent.ID)
	if err != nil {
		log.Fatalf("Failed to create license client: %v", err)
	}

	// Validate license (with timeout)
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := licenseClient.Connect(validateCtx); err != nil {
		log.Printf("Warning: Failed to connect to license server: %v", err)
		log.Printf("Running in offline mode with grace period")
	} else {
		result, err := licenseClient.Validate(validateCtx)
		if err != nil {
			log.Fatalf("License validation failed: %v", err)
		}

		if !result.Valid {
			log.Fatalf("Invalid license: %s", result.Message)
		}

		log.Printf("License validated: Tier=%s, MaxRouters=%d", result.Tier, result.MaxRouters)
		licenseClient.Close()
	}

	// Initialize collector registry
	registry := collector.NewRegistry()
	if err := registry.Register(mikrotik.NewCollector()); err != nil {
		log.Fatalf("Failed to register MikroTik collector: %v", err)
	}
	log.Printf("Registered collectors: %v", registry.List())

	// Initialize gRPC client
	grpcClient, err := grpc.NewClient(&cfg.Server)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}

	// Connect to server
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := grpcClient.Connect(connectCtx); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer grpcClient.Close()

	log.Printf("Connected to server: %s", cfg.Server.Address)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start collection loop
	ticker := time.NewTicker(time.Duration(cfg.Collection.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	log.Printf("Starting collection loop (interval: %ds)", cfg.Collection.IntervalSeconds)

	for {
		select {
		case <-ticker.C:
			// Collect from all configured routers
			for _, router := range cfg.Routers {
				go collectFromRouter(ctx, registry, &router, auditLogger)
			}

		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down gracefully...", sig)
			return
		}
	}
}

func collectFromRouter(ctx context.Context, registry *collector.Registry, router *models.RouterConfig, auditLogger *privacy.AuditLogger) {
	log.Printf("Collecting from router: %s (%s)", router.Name, router.ID)

	// Get the appropriate collector
	coll, err := registry.Get(router.Type)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Collect metrics
	metrics, err := coll.Collect(ctx, router)
	if err != nil {
		log.Printf("Error collecting from %s: %v", router.Name, err)
		return
	}

	log.Printf("Collected metrics from %s: CPU=%.1f%%, Memory=%.1f%%, Interfaces=%d",
		router.Name, metrics.System.CPUPercent, metrics.System.MemoryPercent, len(metrics.Interfaces))

	// Log to audit if enabled
	if auditLogger != nil {
		details := map[string]interface{}{
			"router_name": router.Name,
			"cpu_percent": metrics.System.CPUPercent,
			"interfaces":  len(metrics.Interfaces),
		}
		if err := auditLogger.LogCollection(router.ID, "metrics", 1, details); err != nil {
			log.Printf("Warning: Failed to log audit entry: %v", err)
		}
	}

	// TODO: Send metrics to server via transport
}
