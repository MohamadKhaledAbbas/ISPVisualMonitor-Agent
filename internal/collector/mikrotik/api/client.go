// Package api implements the MikroTik RouterOS API protocol.
package api

import (
	"bufio"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"io"
	"net"
	"sync"
	"time"
)

// ClientConfig contains configuration for the RouterOS API client.
type ClientConfig struct {
	Address            string        // Router address (host:port)
	Username           string        // API username
	Password           string        // API password
	UseTLS             bool          // Use TLS connection (port 8729)
	TLSConfig          *tls.Config   // Custom TLS configuration
	InsecureSkipVerify bool          // Skip TLS certificate verification (not recommended for production)
	Timeout            time.Duration // Connection and read/write timeout
	RetryAttempts      int           // Number of retry attempts
	RetryDelay         time.Duration // Delay between retries
}

// DefaultConfig returns a ClientConfig with default values.
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:            10 * time.Second,
		RetryAttempts:      3,
		RetryDelay:         time.Second,
		InsecureSkipVerify: false, // Secure by default
	}
}

// Client represents a connection to a MikroTik RouterOS API.
type Client struct {
	config     *ClientConfig
	conn       net.Conn
	reader     *bufio.Reader
	mu         sync.Mutex
	connected  bool
	apiVersion string // Detected API version for login method selection

	// Circuit breaker
	circuitBreaker *circuitBreaker
}

// circuitBreaker implements the circuit breaker pattern.
type circuitBreaker struct {
	mu               sync.Mutex
	failures         int
	maxFailures      int
	state            circuitState
	lastFailure      time.Time
	resetTimeout     time.Duration
	halfOpenAttempts int
}

type circuitState int

const (
	circuitClosed circuitState = iota
	circuitOpen
	circuitHalfOpen
)

func newCircuitBreaker(maxFailures int, resetTimeout time.Duration) *circuitBreaker {
	return &circuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        circuitClosed,
	}
}

func (cb *circuitBreaker) canAttempt() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case circuitClosed:
		return true
	case circuitOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = circuitHalfOpen
			cb.halfOpenAttempts = 0
			return true
		}
		return false
	case circuitHalfOpen:
		return cb.halfOpenAttempts < 1 // Allow one attempt in half-open state
	}
	return false
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = circuitClosed
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.state == circuitHalfOpen {
		cb.state = circuitOpen
		return
	}

	if cb.failures >= cb.maxFailures {
		cb.state = circuitOpen
	}
}

func (cb *circuitBreaker) reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = circuitClosed
}

// NewClient creates a new RouterOS API client.
func NewClient(config *ClientConfig) *Client {
	if config == nil {
		config = DefaultConfig()
	}
	return &Client{
		config:         config,
		circuitBreaker: newCircuitBreaker(5, 30*time.Second),
	}
}

// Connect establishes a connection to the router and authenticates.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	if !c.circuitBreaker.canAttempt() {
		return ErrCircuitOpen
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		if err := c.connect(ctx); err != nil {
			lastErr = err
			continue
		}

		if err := c.authenticate(ctx); err != nil {
			c.closeConn()
			lastErr = err
			// Don't retry auth errors
			if IsAuthError(err) {
				c.circuitBreaker.recordFailure()
				return err
			}
			continue
		}

		c.connected = true
		c.circuitBreaker.recordSuccess()
		return nil
	}

	c.circuitBreaker.recordFailure()
	return lastErr
}

func (c *Client) connect(ctx context.Context) error {
	dialer := &net.Dialer{
		Timeout: c.config.Timeout,
	}

	var conn net.Conn
	var err error

	if c.config.UseTLS {
		tlsConfig := c.config.TLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
			// Only skip verification if explicitly configured
			// This is sometimes needed for RouterOS self-signed certs
			if c.config.InsecureSkipVerify {
				tlsConfig.InsecureSkipVerify = true // #nosec G402 - User explicitly configured insecure mode
			}
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", c.config.Address, tlsConfig)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", c.config.Address)
	}

	if err != nil {
		return NewConnectionError("failed to connect", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	return nil
}

func (c *Client) authenticate(ctx context.Context) error {
	// Try new login method first (RouterOS 6.43+)
	if err := c.tryNewLogin(); err == nil {
		return nil
	}

	// Fall back to legacy challenge-response login
	return c.tryLegacyLogin()
}

func (c *Client) tryNewLogin() error {
	sentence := NewSentence("/login")
	sentence.AddAttribute("name", c.config.Username)
	sentence.AddAttribute("password", c.config.Password)

	if err := c.writeSentence(sentence); err != nil {
		return err
	}

	reply, err := c.readReply()
	if err != nil {
		return err
	}

	if reply.IsTrap() {
		return NewAuthError(reply.GetMessage())
	}

	if !reply.IsDone() {
		// Check if we got a challenge (old method)
		if ret, ok := reply.Data["ret"]; ok {
			return c.handleLegacyChallenge(ret)
		}
		return NewProtocolError("unexpected login response", nil)
	}

	return nil
}

func (c *Client) tryLegacyLogin() error {
	sentence := NewSentence("/login")

	if err := c.writeSentence(sentence); err != nil {
		return err
	}

	reply, err := c.readReply()
	if err != nil {
		return err
	}

	if reply.IsTrap() {
		return NewAuthError(reply.GetMessage())
	}

	if !reply.IsDone() {
		return NewProtocolError("unexpected login response", nil)
	}

	// Get challenge
	challenge, ok := reply.Data["ret"]
	if !ok {
		return NewProtocolError("no challenge received", nil)
	}

	return c.handleLegacyChallenge(challenge)
}

func (c *Client) handleLegacyChallenge(challenge string) error {
	// Decode hex challenge
	challengeBytes, err := hex.DecodeString(challenge)
	if err != nil {
		return NewProtocolError("invalid challenge format", err)
	}

	// Create MD5 hash: 0x00 + password + challenge
	hash := md5.New()
	hash.Write([]byte{0})
	hash.Write([]byte(c.config.Password))
	hash.Write(challengeBytes)
	response := hex.EncodeToString(hash.Sum(nil))

	// Send response
	sentence := NewSentence("/login")
	sentence.AddAttribute("name", c.config.Username)
	sentence.AddAttribute("response", "00"+response)

	if err := c.writeSentence(sentence); err != nil {
		return err
	}

	reply, err := c.readReply()
	if err != nil {
		return err
	}

	if reply.IsTrap() {
		return NewAuthError(reply.GetMessage())
	}

	if !reply.IsDone() {
		return NewProtocolError("unexpected login response", nil)
	}

	return nil
}

// Close closes the connection to the router.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closeConn()
}

func (c *Client) closeConn() error {
	c.connected = false
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.reader = nil
		return err
	}
	return nil
}

// IsConnected returns true if the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// Execute sends a command and returns all replies.
func (c *Client) Execute(ctx context.Context, sentence *Sentence) ([]*Reply, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, ErrNotConnected
	}

	// Set deadline
	if c.config.Timeout > 0 {
		deadline := time.Now().Add(c.config.Timeout)
		if err := c.conn.SetDeadline(deadline); err != nil {
			return nil, NewConnectionError("failed to set deadline", err)
		}
	}

	if err := c.writeSentence(sentence); err != nil {
		return nil, err
	}

	return c.readAllReplies()
}

// Run executes a command and returns the data replies (filtering out !done).
func (c *Client) Run(ctx context.Context, command string, args map[string]string) ([]map[string]string, error) {
	sentence := NewSentence(command)
	for k, v := range args {
		sentence.AddAttribute(k, v)
	}

	replies, err := c.Execute(ctx, sentence)
	if err != nil {
		return nil, err
	}

	var result []map[string]string
	for _, reply := range replies {
		if reply.IsTrap() {
			return nil, NewTrapError(reply)
		}
		if reply.IsFatal() {
			c.closeConn()
			return nil, NewFatalError(reply)
		}
		if reply.IsData() {
			result = append(result, reply.Data)
		}
	}

	return result, nil
}

// RunOne executes a command and returns the first data reply.
func (c *Client) RunOne(ctx context.Context, command string, args map[string]string) (map[string]string, error) {
	results, err := c.Run(ctx, command, args)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (c *Client) writeSentence(sentence *Sentence) error {
	data := EncodeSentence(sentence)
	_, err := c.conn.Write(data)
	if err != nil {
		return NewConnectionError("failed to write sentence", err)
	}
	return nil
}

func (c *Client) readReply() (*Reply, error) {
	reply, err := DecodeSentence(c.reader)
	if err != nil {
		if err == io.EOF {
			c.closeConn()
			return nil, ErrConnectionClosed
		}
		return nil, NewProtocolError("failed to decode reply", err)
	}
	return reply, nil
}

func (c *Client) readAllReplies() ([]*Reply, error) {
	var replies []*Reply

	for {
		reply, err := c.readReply()
		if err != nil {
			return replies, err
		}

		replies = append(replies, reply)

		if reply.IsDone() || reply.IsFatal() {
			break
		}
	}

	return replies, nil
}

// Ping sends a test command to verify the connection is alive.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.RunOne(ctx, "/system/resource/print", nil)
	return err
}

// GetRouterIdentity returns the router's identity.
func (c *Client) GetRouterIdentity(ctx context.Context) (string, error) {
	result, err := c.RunOne(ctx, "/system/identity/print", nil)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}
	return result["name"], nil
}

// ResetCircuitBreaker resets the circuit breaker to allow new connections.
func (c *Client) ResetCircuitBreaker() {
	c.circuitBreaker.reset()
}
