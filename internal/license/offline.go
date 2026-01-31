package license

import (
	"fmt"
	"sync"
	"time"
)

// OfflineManager manages offline grace period for license validation
type OfflineManager struct {
	graceHours       int
	lastValidation   time.Time
	lastValidationOK bool
	mu               sync.RWMutex
}

// NewOfflineManager creates a new offline manager
func NewOfflineManager(graceHours int) *OfflineManager {
	return &OfflineManager{
		graceHours: graceHours,
	}
}

// RecordValidation records a successful license validation
func (o *OfflineManager) RecordValidation(success bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.lastValidation = time.Now()
	o.lastValidationOK = success
}

// CanOperate checks if the agent can operate during offline period
func (o *OfflineManager) CanOperate() (bool, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.lastValidation.IsZero() {
		return false, fmt.Errorf("no previous validation found, must validate online first")
	}

	if !o.lastValidationOK {
		return false, fmt.Errorf("last validation was not successful")
	}

	gracePeriod := time.Duration(o.graceHours) * time.Hour
	expiresAt := o.lastValidation.Add(gracePeriod)

	if time.Now().After(expiresAt) {
		return false, fmt.Errorf("offline grace period expired at %s", expiresAt.Format(time.RFC3339))
	}

	return true, nil
}

// GetLastValidation returns information about the last validation
func (o *OfflineManager) GetLastValidation() (time.Time, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.lastValidation, o.lastValidationOK
}

// GetGraceRemaining returns the remaining grace period
func (o *OfflineManager) GetGraceRemaining() time.Duration {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.lastValidation.IsZero() {
		return 0
	}

	gracePeriod := time.Duration(o.graceHours) * time.Hour
	expiresAt := o.lastValidation.Add(gracePeriod)
	remaining := time.Until(expiresAt)

	if remaining < 0 {
		return 0
	}

	return remaining
}
