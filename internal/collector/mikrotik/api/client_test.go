package api

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.Timeout)
	}
	
	if cfg.RetryAttempts != 3 {
		t.Errorf("expected 3 retry attempts, got %d", cfg.RetryAttempts)
	}
	
	if cfg.RetryDelay != time.Second {
		t.Errorf("expected 1s retry delay, got %v", cfg.RetryDelay)
	}
}

func TestNewClient(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		cfg := &ClientConfig{
			Address:  "192.168.1.1:8728",
			Username: "admin",
			Password: "password",
			Timeout:  5 * time.Second,
		}
		
		client := NewClient(cfg)
		
		if client == nil {
			t.Error("expected non-nil client")
		}
		
		if client.config.Address != "192.168.1.1:8728" {
			t.Errorf("expected address '192.168.1.1:8728', got %q", client.config.Address)
		}
	})
	
	t.Run("with nil config", func(t *testing.T) {
		client := NewClient(nil)
		
		if client == nil {
			t.Error("expected non-nil client")
		}
		
		if client.config.Timeout != 10*time.Second {
			t.Errorf("expected default timeout, got %v", client.config.Timeout)
		}
	})
}

func TestClientIsConnected(t *testing.T) {
	client := NewClient(&ClientConfig{
		Address: "192.168.1.1:8728",
	})
	
	if client.IsConnected() {
		t.Error("new client should not be connected")
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := newCircuitBreaker(3, 100*time.Millisecond)

	// Initially should allow attempts
	if !cb.canAttempt() {
		t.Error("circuit breaker should allow first attempt")
	}

	// Record failures
	cb.recordFailure()
	cb.recordFailure()
	if !cb.canAttempt() {
		t.Error("circuit breaker should still allow attempt with 2 failures")
	}

	// Third failure should open the circuit
	cb.recordFailure()
	if cb.canAttempt() {
		t.Error("circuit breaker should be open after 3 failures")
	}

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)
	if !cb.canAttempt() {
		t.Error("circuit breaker should be half-open after timeout")
	}

	// Success should close the circuit
	cb.recordSuccess()
	if !cb.canAttempt() {
		t.Error("circuit breaker should be closed after success")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := newCircuitBreaker(2, time.Second)

	// Open the circuit
	cb.recordFailure()
	cb.recordFailure()
	if cb.canAttempt() {
		t.Error("circuit should be open")
	}

	// Reset should close it
	cb.reset()
	if !cb.canAttempt() {
		t.Error("circuit should be closed after reset")
	}
}

func TestClientConnectNoServer(t *testing.T) {
	client := NewClient(&ClientConfig{
		Address:       "127.0.0.1:19999", // Non-existent port
		Username:      "admin",
		Password:      "password",
		Timeout:       100 * time.Millisecond,
		RetryAttempts: 0, // No retries for faster test
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		t.Error("expected connection error")
		client.Close()
	}

	if !IsConnectionError(err) {
		t.Logf("error type: %T, value: %v", err, err)
		// Note: context deadline exceeded is also acceptable
	}
}

func TestClientExecuteNotConnected(t *testing.T) {
	client := NewClient(&ClientConfig{
		Address: "192.168.1.1:8728",
	})

	ctx := context.Background()
	_, err := client.Execute(ctx, NewSentence("/test"))

	if err != ErrNotConnected {
		t.Errorf("expected ErrNotConnected, got %v", err)
	}
}

func TestClientResetCircuitBreaker(t *testing.T) {
	client := NewClient(&ClientConfig{
		Address: "192.168.1.1:8728",
	})

	// Force open the circuit breaker
	for i := 0; i < 5; i++ {
		client.circuitBreaker.recordFailure()
	}

	if client.circuitBreaker.canAttempt() {
		t.Error("circuit should be open")
	}

	client.ResetCircuitBreaker()

	if !client.circuitBreaker.canAttempt() {
		t.Error("circuit should be closed after reset")
	}
}
