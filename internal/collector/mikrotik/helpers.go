package mikrotik

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseUptime parses RouterOS uptime string (e.g., "1w2d3h4m5s") to seconds.
func ParseUptime(uptime string) int64 {
	if uptime == "" {
		return 0
	}

	var total int64
	uptime = strings.TrimSpace(uptime)

	// Regular expression to match time components
	re := regexp.MustCompile(`(\d+)([wdhms])`)
	matches := re.FindAllStringSubmatch(uptime, -1)

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		value, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			continue
		}

		switch match[2] {
		case "w":
			total += value * 7 * 24 * 60 * 60
		case "d":
			total += value * 24 * 60 * 60
		case "h":
			total += value * 60 * 60
		case "m":
			total += value * 60
		case "s":
			total += value
		}
	}

	return total
}

// ParseSpeed parses RouterOS speed string (e.g., "1Gbps", "100Mbps") to Mbps.
func ParseSpeed(speed string) int64 {
	if speed == "" {
		return 0
	}

	speed = strings.TrimSpace(strings.ToLower(speed))

	// Handle "auto" or unknown values
	if speed == "auto" || speed == "unknown" {
		return 0
	}

	// Extract number and unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(gbps|mbps|kbps|bps)?$`)
	matches := re.FindStringSubmatch(speed)

	if len(matches) < 2 {
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	unit := "mbps"
	if len(matches) >= 3 && matches[2] != "" {
		unit = matches[2]
	}

	switch unit {
	case "gbps":
		return int64(value * 1000)
	case "mbps":
		return int64(value)
	case "kbps":
		return int64(value / 1000)
	case "bps":
		return int64(value / 1000000)
	default:
		return int64(value)
	}
}

// ParseBool parses RouterOS boolean string (true/false/yes/no).
func ParseBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "true" || s == "yes"
}

// ParseInt64 parses a string to int64, returning 0 on error.
func ParseInt64(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// ParseUint64 parses a string to uint64, returning 0 on error.
func ParseUint64(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// ParseFloat64 parses a string to float64, returning 0 on error.
func ParseFloat64(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// Handle percentage suffix
	s = strings.TrimSuffix(s, "%")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// ParseMemory parses memory size string (e.g., "1024MiB", "1GiB") to bytes.
func ParseMemory(mem string) int64 {
	if mem == "" {
		return 0
	}

	mem = strings.TrimSpace(strings.ToLower(mem))

	// Extract number and unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(gib|mib|kib|gb|mb|kb|b)?$`)
	matches := re.FindStringSubmatch(mem)

	if len(matches) < 2 {
		// Try parsing as plain number (bytes)
		return ParseInt64(mem)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	unit := "b"
	if len(matches) >= 3 && matches[2] != "" {
		unit = matches[2]
	}

	switch unit {
	case "gib", "gb":
		return int64(value * 1024 * 1024 * 1024)
	case "mib", "mb":
		return int64(value * 1024 * 1024)
	case "kib", "kb":
		return int64(value * 1024)
	default:
		return int64(value)
	}
}

// ParseDuration parses RouterOS duration string to time.Duration.
func ParseDuration(dur string) time.Duration {
	seconds := ParseUptime(dur)
	return time.Duration(seconds) * time.Second
}

// ParseTimestamp parses RouterOS timestamp string.
func ParseTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}

	// RouterOS formats: "jan/01/2024 12:00:00", "2024-01-01 12:00:00"
	layouts := []string{
		"jan/02/2006 15:04:05",
		"2006-01-02 15:04:05",
		"Jan/02/2006 15:04:05",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, ts); err == nil {
			return t
		}
	}

	return time.Time{}
}

// MatchFilter checks if a name matches include/exclude patterns.
func MatchFilter(name string, include, exclude []string) bool {
	// If no include patterns, everything is included by default
	included := len(include) == 0

	// Check include patterns
	for _, pattern := range include {
		if matchGlob(pattern, name) {
			included = true
			break
		}
	}

	if !included {
		return false
	}

	// Check exclude patterns
	for _, pattern := range exclude {
		if matchGlob(pattern, name) {
			return false
		}
	}

	return true
}

// matchGlob performs simple glob matching with * wildcard.
func matchGlob(pattern, s string) bool {
	if pattern == "" {
		return s == ""
	}

	// Simple * matching
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			return strings.HasPrefix(s, prefix) && strings.HasSuffix(s, suffix)
		}
	}

	return pattern == s
}

// FormatBytes formats bytes to human-readable string.
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return strconv.FormatFloat(float64(bytes)/TB, 'f', 2, 64) + " TB"
	case bytes >= GB:
		return strconv.FormatFloat(float64(bytes)/GB, 'f', 2, 64) + " GB"
	case bytes >= MB:
		return strconv.FormatFloat(float64(bytes)/MB, 'f', 2, 64) + " MB"
	case bytes >= KB:
		return strconv.FormatFloat(float64(bytes)/KB, 'f', 2, 64) + " KB"
	default:
		return strconv.FormatInt(bytes, 10) + " B"
	}
}

// FormatDuration formats seconds to human-readable duration.
func FormatDuration(seconds int64) string {
	d := time.Duration(seconds) * time.Second

	days := int64(d.Hours() / 24)
	hours := int64(d.Hours()) % 24
	minutes := int64(d.Minutes()) % 60
	secs := int64(d.Seconds()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, strconv.FormatInt(days, 10)+"d")
	}
	if hours > 0 {
		parts = append(parts, strconv.FormatInt(hours, 10)+"h")
	}
	if minutes > 0 {
		parts = append(parts, strconv.FormatInt(minutes, 10)+"m")
	}
	if secs > 0 || len(parts) == 0 {
		parts = append(parts, strconv.FormatInt(secs, 10)+"s")
	}

	return strings.Join(parts, "")
}

// SafeString returns a default value if the string is empty.
func SafeString(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}

// Handle32BitCounterWrap handles 32-bit counter wraparound.
// Returns the delta between current and previous, handling wrap.
func Handle32BitCounterWrap(current, previous uint64) uint64 {
	if current >= previous {
		return current - previous
	}
	// Counter wrapped around (32-bit max is 4294967295)
	return (0xFFFFFFFF - previous) + current + 1
}

// Handle64BitCounterWrap handles 64-bit counter wraparound.
// Returns the delta between current and previous, handling wrap.
func Handle64BitCounterWrap(current, previous uint64) uint64 {
	if current >= previous {
		return current - previous
	}
	// Counter wrapped around
	return (0xFFFFFFFFFFFFFFFF - previous) + current + 1
}
