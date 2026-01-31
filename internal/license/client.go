package license

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/api/proto/licensepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Client handles license validation with the server
type Client struct {
	validationURL string
	licenseKey    string
	agentID       string
	grpcClient    licensepb.LicenseServiceClient
	conn          *grpc.ClientConn
}

// NewClient creates a new license validation client
func NewClient(validationURL, licenseKey, agentID string) (*Client, error) {
	if validationURL == "" {
		return nil, fmt.Errorf("validation URL is required")
	}
	if licenseKey == "" {
		return nil, fmt.Errorf("license key is required")
	}
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	return &Client{
		validationURL: validationURL,
		licenseKey:    licenseKey,
		agentID:       agentID,
	}, nil
}

// Connect establishes a connection to the license server
func (c *Client) Connect(ctx context.Context) error {
	var opts []grpc.DialOption

	// Use TLS for production, insecure for development
	if os.Getenv("ISPAGENT_DEV_MODE") == "true" {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	conn, err := grpc.DialContext(ctx, c.validationURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to license server: %w", err)
	}

	c.conn = conn
	c.grpcClient = licensepb.NewLicenseServiceClient(conn)

	return nil
}

// Validate validates the license with the server
func (c *Client) Validate(ctx context.Context) (*ValidationResult, error) {
	if c.grpcClient == nil {
		return nil, fmt.Errorf("client not connected, call Connect first")
	}

	req := &licensepb.ValidateRequest{
		LicenseKey:          c.licenseKey,
		AgentId:             c.agentID,
		HardwareFingerprint: generateHardwareFingerprint(),
	}

	resp, err := c.grpcClient.ValidateLicense(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("license validation failed: %w", err)
	}

	result := &ValidationResult{
		Valid:           resp.Valid,
		Tier:            resp.Tier,
		MaxRouters:      int(resp.MaxRouters),
		EnabledFeatures: resp.EnabledFeatures,
		Message:         resp.Message,
	}

	if resp.ExpiresAt != nil {
		result.ExpiresAt = resp.ExpiresAt.AsTime()
	}

	return result, nil
}

// Refresh refreshes the license status
func (c *Client) Refresh(ctx context.Context) (*RefreshResult, error) {
	if c.grpcClient == nil {
		return nil, fmt.Errorf("client not connected, call Connect first")
	}

	req := &licensepb.RefreshRequest{
		LicenseKey: c.licenseKey,
		AgentId:    c.agentID,
	}

	resp, err := c.grpcClient.RefreshLicense(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("license refresh failed: %w", err)
	}

	result := &RefreshResult{
		Success: resp.Success,
	}

	if resp.NextRefresh != nil {
		result.NextRefresh = resp.NextRefresh.AsTime()
	}

	return result, nil
}

// Close closes the connection to the license server
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ValidationResult contains the result of a license validation
type ValidationResult struct {
	Valid           bool
	Tier            string
	ExpiresAt       time.Time
	MaxRouters      int
	EnabledFeatures []string
	Message         string
}

// RefreshResult contains the result of a license refresh
type RefreshResult struct {
	Success     bool
	NextRefresh time.Time
}

// generateHardwareFingerprint generates a hardware fingerprint for the agent
func generateHardwareFingerprint() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("hw-%s", hostname)
}
