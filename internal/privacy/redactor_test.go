package privacy

import "testing"

func TestNewRedactor(t *testing.T) {
	r := NewRedactor(true, true)
	if r == nil {
		t.Error("Expected non-nil redactor")
	}
	if !r.ShouldRedactUsernames() {
		t.Error("Expected username redaction to be enabled")
	}
	if !r.ShouldRedactIPAddresses() {
		t.Error("Expected IP redaction to be enabled")
	}
}

func TestRedactor_RedactUsername(t *testing.T) {
	tests := []struct {
		name            string
		redactEnabled   bool
		username        string
		expectRedacted  bool
	}{
		{"enabled", true, "user@example.com", true},
		{"disabled", false, "user@example.com", false},
		{"empty", true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRedactor(tt.redactEnabled, false)
			result := r.RedactUsername(tt.username)

			if tt.expectRedacted && result == tt.username && tt.username != "" {
				t.Error("Expected username to be redacted but it wasn't")
			}
			if !tt.expectRedacted && result != tt.username {
				t.Errorf("Expected username to not be redacted, got %s", result)
			}
		})
	}
}

func TestRedactor_RedactIPAddress(t *testing.T) {
	tests := []struct {
		name           string
		redactEnabled  bool
		ip             string
		expectRedacted bool
	}{
		{"ipv4 enabled", true, "192.168.1.100", true},
		{"ipv4 disabled", false, "192.168.1.100", false},
		{"ipv6 enabled", true, "2001:db8::1", true},
		{"empty", true, "", false},
		{"invalid", true, "not-an-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRedactor(false, tt.redactEnabled)
			result := r.RedactIPAddress(tt.ip)

			if tt.expectRedacted && result == tt.ip && tt.ip != "" && tt.ip != "not-an-ip" {
				t.Errorf("Expected IP to be redacted but got original: %s", result)
			}
			if !tt.expectRedacted && result != tt.ip {
				t.Errorf("Expected IP to not be redacted, got %s", result)
			}
		})
	}
}

func TestRedactor_RedactMACAddress(t *testing.T) {
	r := NewRedactor(false, false)

	tests := []struct {
		mac      string
		expected string
	}{
		{"00:11:22:33:44:55", "00:11:22:xx:xx:xx"},
		{"00-11-22-33-44-55", "00:11:22:xx:xx:xx"},
		{"", ""},
		{"invalid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.mac, func(t *testing.T) {
			result := r.RedactMACAddress(tt.mac)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
