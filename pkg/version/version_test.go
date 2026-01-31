package version

import "testing"

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("Version should not be empty")
	}
}

func TestVersionDefault(t *testing.T) {
	// Default version should be "dev"
	if Version != "dev" && Version == "" {
		t.Error("Default version should be 'dev' or set by ldflags")
	}
}
