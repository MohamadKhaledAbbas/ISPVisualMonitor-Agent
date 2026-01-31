package mikrotik

import (
	"testing"
	"time"
)

func TestParseUptime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"empty", "", 0},
		{"seconds only", "45s", 45},
		{"minutes and seconds", "5m30s", 330},
		{"hours minutes seconds", "2h30m15s", 9015},
		{"days", "1d", 86400},
		{"days and hours", "1d12h", 129600},
		{"weeks", "1w", 604800},
		{"full format", "1w2d3h4m5s", 788645},
		{"with spaces", " 1h 30m ", 5400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseUptime(tt.input)
			if result != tt.expected {
				t.Errorf("ParseUptime(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSpeed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"empty", "", 0},
		{"auto", "auto", 0},
		{"unknown", "unknown", 0},
		{"gbps", "1Gbps", 1000},
		{"mbps", "100Mbps", 100},
		{"10gbps", "10Gbps", 10000},
		{"lowercase", "100mbps", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSpeed(tt.input)
			if result != tt.expected {
				t.Errorf("ParseSpeed(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"false", false},
		{"no", false},
		{"", false},
		{"maybe", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseBool(tt.input)
			if result != tt.expected {
				t.Errorf("ParseBool(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"", 0},
		{"0", 0},
		{"123", 123},
		{"-456", -456},
		{"  789  ", 789},
		{"invalid", 0},
		{"12.34", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseInt64(tt.input)
			if result != tt.expected {
				t.Errorf("ParseInt64(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"", 0},
		{"0", 0},
		{"123.45", 123.45},
		{"-67.89", -67.89},
		{"  12.5  ", 12.5},
		{"75%", 75},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseFloat64(tt.input)
			if result != tt.expected {
				t.Errorf("ParseFloat64(%q) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseMemory(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"empty", "", 0},
		{"bytes", "1024", 1024},
		{"kib", "1KiB", 1024},
		{"mib", "100MiB", 104857600},
		{"gib", "1GiB", 1073741824},
		{"lowercase", "512mib", 536870912},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMemory(tt.input)
			if result != tt.expected {
				t.Errorf("ParseMemory(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{"empty", "", 0},
		{"seconds", "30s", 30 * time.Second},
		{"minutes", "5m", 5 * time.Minute},
		{"hours", "2h", 2 * time.Hour},
		{"complex", "1h30m45s", time.Hour + 30*time.Minute + 45*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseDuration(tt.input)
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMatchFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		include  []string
		exclude  []string
		expected bool
	}{
		{
			name:     "no filters",
			input:    "ether1",
			include:  nil,
			exclude:  nil,
			expected: true,
		},
		{
			name:     "include match",
			input:    "ether1",
			include:  []string{"ether*"},
			exclude:  nil,
			expected: true,
		},
		{
			name:     "include no match",
			input:    "bridge1",
			include:  []string{"ether*"},
			exclude:  nil,
			expected: false,
		},
		{
			name:     "exclude match",
			input:    "bridge-local",
			include:  nil,
			exclude:  []string{"bridge-local"},
			expected: false,
		},
		{
			name:     "include then exclude",
			input:    "ether1-wan",
			include:  []string{"ether*"},
			exclude:  []string{"*-wan"},
			expected: false,
		},
		{
			name:     "exact match",
			input:    "ether1",
			include:  []string{"ether1"},
			exclude:  nil,
			expected: true,
		},
		{
			name:     "multiple patterns",
			input:    "sfp1",
			include:  []string{"ether*", "sfp*"},
			exclude:  nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchFilter(tt.input, tt.include, tt.exclude)
			if result != tt.expected {
				t.Errorf("MatchFilter(%q, %v, %v) = %v, want %v",
					tt.input, tt.include, tt.exclude, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{1099511627776, "1.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int64
		expected string
	}{
		{0, "0s"},
		{45, "45s"},
		{60, "1m"},
		{90, "1m30s"},
		{3600, "1h"},
		{3661, "1h1m1s"},
		{86400, "1d"},
		{90061, "1d1h1m1s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatDuration(tt.seconds)
			if result != tt.expected {
				t.Errorf("FormatDuration(%d) = %q, want %q", tt.seconds, result, tt.expected)
			}
		})
	}
}

func TestSafeString(t *testing.T) {
	if SafeString("value", "default") != "value" {
		t.Error("SafeString should return value when not empty")
	}
	if SafeString("", "default") != "default" {
		t.Error("SafeString should return default when empty")
	}
}

func TestHandle32BitCounterWrap(t *testing.T) {
	tests := []struct {
		name     string
		current  uint64
		previous uint64
		expected uint64
	}{
		{"no wrap", 1000, 500, 500},
		{"wrap", 100, 0xFFFFFF00, 356},
		{"zero delta", 500, 500, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Handle32BitCounterWrap(tt.current, tt.previous)
			if result != tt.expected {
				t.Errorf("Handle32BitCounterWrap(%d, %d) = %d, want %d",
					tt.current, tt.previous, result, tt.expected)
			}
		})
	}
}

func TestHandle64BitCounterWrap(t *testing.T) {
	tests := []struct {
		name     string
		current  uint64
		previous uint64
		expected uint64
	}{
		{"no wrap", 1000000, 500000, 500000},
		{"zero delta", 12345, 12345, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Handle64BitCounterWrap(tt.current, tt.previous)
			if result != tt.expected {
				t.Errorf("Handle64BitCounterWrap(%d, %d) = %d, want %d",
					tt.current, tt.previous, result, tt.expected)
			}
		})
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern  string
		s        string
		expected bool
	}{
		{"ether*", "ether1", true},
		{"ether*", "ether10", true},
		{"ether*", "bridge1", false},
		{"*-wan", "ether1-wan", true},
		{"*-wan", "ether1-lan", false},
		{"ether1", "ether1", true},
		{"ether1", "ether2", false},
		{"", "", true},
		{"", "something", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.s, func(t *testing.T) {
			result := matchGlob(tt.pattern, tt.s)
			if result != tt.expected {
				t.Errorf("matchGlob(%q, %q) = %v, want %v",
					tt.pattern, tt.s, result, tt.expected)
			}
		})
	}
}
