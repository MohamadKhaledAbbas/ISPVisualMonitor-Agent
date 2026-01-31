package privacy

import (
	"crypto/sha256"
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Redactor provides data redaction utilities
type Redactor struct {
	redactUsernames   bool
	redactIPAddresses bool
}

// NewRedactor creates a new redactor with the given settings
func NewRedactor(redactUsernames, redactIPAddresses bool) *Redactor {
	return &Redactor{
		redactUsernames:   redactUsernames,
		redactIPAddresses: redactIPAddresses,
	}
}

// RedactUsername redacts a username if enabled
func (r *Redactor) RedactUsername(username string) string {
	if !r.redactUsernames || username == "" {
		return username
	}
	return r.hashString(username)
}

// RedactIPAddress redacts an IP address if enabled
func (r *Redactor) RedactIPAddress(ip string) string {
	if !r.redactIPAddresses || ip == "" {
		return ip
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ip
	}

	// For IPv4, keep network portion, redact host portion
	if ipv4 := parsedIP.To4(); ipv4 != nil {
		return fmt.Sprintf("%d.%d.xxx.xxx", ipv4[0], ipv4[1])
	}

	// For IPv6, keep prefix, redact interface ID
	return r.hashString(ip)[:8] + "::xxxx"
}

// RedactMACAddress redacts a MAC address by keeping only the OUI (first 3 octets)
func (r *Redactor) RedactMACAddress(mac string) string {
	if mac == "" {
		return mac
	}
	
	parts := regexp.MustCompile(`[:-]`).Split(mac, -1)
	if len(parts) < 3 {
		return mac
	}
	
	// Keep OUI (vendor identifier), redact device-specific part
	return strings.Join(parts[:3], ":") + ":xx:xx:xx"
}

// hashString creates a deterministic hash of a string
func (r *Redactor) hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash[:8])
}

// ShouldRedactUsernames returns whether username redaction is enabled
func (r *Redactor) ShouldRedactUsernames() bool {
	return r.redactUsernames
}

// ShouldRedactIPAddresses returns whether IP address redaction is enabled
func (r *Redactor) ShouldRedactIPAddresses() bool {
	return r.redactIPAddresses
}
