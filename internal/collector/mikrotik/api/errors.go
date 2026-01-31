// Package api implements the MikroTik RouterOS API protocol.
package api

import (
	"errors"
	"fmt"
)

// ErrType represents the type of API error.
type ErrType int

const (
	// ErrTypeConnection indicates a connection error.
	ErrTypeConnection ErrType = iota
	// ErrTypeAuth indicates an authentication error.
	ErrTypeAuth
	// ErrTypeTrap indicates a RouterOS trap error.
	ErrTypeTrap
	// ErrTypeFatal indicates a fatal RouterOS error.
	ErrTypeFatal
	// ErrTypeTimeout indicates a timeout error.
	ErrTypeTimeout
	// ErrTypeCircuitOpen indicates the circuit breaker is open.
	ErrTypeCircuitOpen
	// ErrTypeProtocol indicates a protocol error.
	ErrTypeProtocol
)

// APIError represents an error from the RouterOS API.
type APIError struct {
	Type      ErrType
	Message   string
	Category  string // RouterOS error category
	Wrapped   error
	Temporary bool
}

// Error returns the error message.
func (e *APIError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Wrapped)
	}
	return e.Message
}

// Unwrap returns the wrapped error.
func (e *APIError) Unwrap() error {
	return e.Wrapped
}

// IsTemporary returns true if the error is temporary and can be retried.
func (e *APIError) IsTemporary() bool {
	return e.Temporary
}

// Common errors
var (
	// ErrNotConnected indicates the client is not connected.
	ErrNotConnected = &APIError{
		Type:      ErrTypeConnection,
		Message:   "not connected to router",
		Temporary: false,
	}

	// ErrConnectionClosed indicates the connection was closed.
	ErrConnectionClosed = &APIError{
		Type:      ErrTypeConnection,
		Message:   "connection closed",
		Temporary: true,
	}

	// ErrAuthFailed indicates authentication failed.
	ErrAuthFailed = &APIError{
		Type:      ErrTypeAuth,
		Message:   "authentication failed",
		Temporary: false,
	}

	// ErrCircuitOpen indicates the circuit breaker is open.
	ErrCircuitOpen = &APIError{
		Type:      ErrTypeCircuitOpen,
		Message:   "circuit breaker is open",
		Temporary: true,
	}

	// ErrTimeout indicates a timeout occurred.
	ErrTimeout = &APIError{
		Type:      ErrTypeTimeout,
		Message:   "operation timed out",
		Temporary: true,
	}
)

// NewConnectionError creates a new connection error.
func NewConnectionError(msg string, err error) *APIError {
	return &APIError{
		Type:      ErrTypeConnection,
		Message:   msg,
		Wrapped:   err,
		Temporary: true,
	}
}

// NewAuthError creates a new authentication error.
func NewAuthError(msg string) *APIError {
	return &APIError{
		Type:      ErrTypeAuth,
		Message:   msg,
		Temporary: false,
	}
}

// NewTrapError creates an error from a RouterOS trap.
func NewTrapError(reply *Reply) *APIError {
	msg := reply.GetMessage()
	if msg == "" {
		msg = "unknown error"
	}
	return &APIError{
		Type:      ErrTypeTrap,
		Message:   msg,
		Category:  reply.Data["category"],
		Temporary: isTrapTemporary(reply),
	}
}

// NewFatalError creates an error from a RouterOS fatal.
func NewFatalError(reply *Reply) *APIError {
	msg := reply.GetMessage()
	if msg == "" {
		msg = "fatal error"
	}
	return &APIError{
		Type:      ErrTypeFatal,
		Message:   msg,
		Temporary: false,
	}
}

// NewProtocolError creates a new protocol error.
func NewProtocolError(msg string, err error) *APIError {
	return &APIError{
		Type:      ErrTypeProtocol,
		Message:   msg,
		Wrapped:   err,
		Temporary: false,
	}
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(msg string) *APIError {
	return &APIError{
		Type:      ErrTypeTimeout,
		Message:   msg,
		Temporary: true,
	}
}

// isTrapTemporary determines if a trap error is temporary.
func isTrapTemporary(reply *Reply) bool {
	// Some traps are temporary (resource busy, etc.)
	category := reply.Data["category"]
	switch category {
	case "1": // Generic
		return true
	case "2": // Resource unavailable
		return true
	case "3": // Resource busy
		return true
	default:
		return false
	}
}

// IsConnectionError checks if the error is a connection error.
func IsConnectionError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrTypeConnection
	}
	return false
}

// IsAuthError checks if the error is an authentication error.
func IsAuthError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrTypeAuth
	}
	return false
}

// IsTemporaryError checks if the error is temporary.
func IsTemporaryError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Temporary
	}
	return false
}

// IsCircuitOpenError checks if the error is a circuit breaker open error.
func IsCircuitOpenError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrTypeCircuitOpen
	}
	return false
}
