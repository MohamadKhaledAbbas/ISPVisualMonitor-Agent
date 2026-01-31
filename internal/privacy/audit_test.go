package privacy

import (
	"os"
	"testing"
)

func TestNewAuditLogger(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	logger, err := NewAuditLogger(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Error("Expected non-nil logger")
	}
}

func TestAuditLogger_Log(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	logger, err := NewAuditLogger(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	entry := AuditEntry{
		EventType:   "test_event",
		RouterID:    "router-01",
		DataType:    "test_data",
		RecordCount: 5,
	}

	err = logger.Log(entry)
	if err != nil {
		t.Errorf("Failed to log audit entry: %v", err)
	}
}

func TestAuditLogger_LogCollection(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	logger, err := NewAuditLogger(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	err = logger.LogCollection("router-01", "metrics", 10, map[string]interface{}{
		"cpu": 25.5,
	})
	if err != nil {
		t.Errorf("Failed to log collection: %v", err)
	}
}

func TestAuditLogger_LogTransmission(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	logger, err := NewAuditLogger(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	err = logger.LogTransmission("router-01", "metrics", 10, "server.example.com:443")
	if err != nil {
		t.Errorf("Failed to log transmission: %v", err)
	}
}
