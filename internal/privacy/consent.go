package privacy

import (
	"fmt"
	"time"
)

// ConsentManager manages user consent for data collection
type ConsentManager struct {
	consentGiven bool
	consentDate  time.Time
}

// NewConsentManager creates a new consent manager
func NewConsentManager() *ConsentManager {
	return &ConsentManager{
		consentGiven: false,
	}
}

// GrantConsent records that consent has been given
func (c *ConsentManager) GrantConsent() {
	c.consentGiven = true
	c.consentDate = time.Now()
}

// RevokeConsent revokes previously given consent
func (c *ConsentManager) RevokeConsent() {
	c.consentGiven = false
}

// HasConsent returns whether consent has been given
func (c *ConsentManager) HasConsent() bool {
	return c.consentGiven
}

// GetConsentDate returns when consent was granted
func (c *ConsentManager) GetConsentDate() time.Time {
	return c.consentDate
}

// ValidateConsent checks if consent is valid for the given operation
func (c *ConsentManager) ValidateConsent(operation string) error {
	if !c.consentGiven {
		return fmt.Errorf("consent not granted for operation: %s", operation)
	}
	return nil
}
