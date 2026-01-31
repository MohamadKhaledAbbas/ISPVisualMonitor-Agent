package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/api/proto/agentpb"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a gRPC client connection to the monitoring server
type Client struct {
	config     *config.ServerConfig
	conn       *grpc.ClientConn
	agentClient agentpb.AgentServiceClient
}

// NewClient creates a new gRPC client
func NewClient(cfg *config.ServerConfig) (*Client, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("server address is required")
	}

	return &Client{
		config: cfg,
	}, nil
}

// Connect establishes a connection to the server
func (c *Client) Connect(ctx context.Context) error {
	var opts []grpc.DialOption

	// Configure TLS if enabled
	if c.config.TLS.Enabled {
		tlsConfig, err := c.loadTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to load TLS config: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add interceptors for authentication and logging
	opts = append(opts, grpc.WithUnaryInterceptor(authUnaryInterceptor()))
	opts = append(opts, grpc.WithStreamInterceptor(authStreamInterceptor()))

	conn, err := grpc.DialContext(ctx, c.config.Address, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	c.agentClient = agentpb.NewAgentServiceClient(conn)

	return nil
}

// GetAgentClient returns the agent service client
func (c *Client) GetAgentClient() agentpb.AgentServiceClient {
	return c.agentClient
}

// Close closes the connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// loadTLSConfig loads TLS configuration from files
func (c *Client) loadTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load CA certificate if provided
	if c.config.TLS.CACert != "" {
		caCert, err := os.ReadFile(c.config.TLS.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate if provided
	if c.config.TLS.ClientCert != "" && c.config.TLS.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(c.config.TLS.ClientCert, c.config.TLS.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
