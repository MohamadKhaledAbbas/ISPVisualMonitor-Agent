package privacy

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditLogger logs all data collection events for transparency
type AuditLogger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	EventType  string                 `json:"event_type"`
	RouterID   string                 `json:"router_id"`
	DataType   string                 `json:"data_type"`
	RecordCount int                   `json:"record_count"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(filePath string) (*AuditLogger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &AuditLogger{
		filePath: filePath,
		file:     file,
	}, nil
}

// Log writes an audit entry to the log
func (a *AuditLogger) Log(entry AuditEntry) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry.Timestamp = time.Now()
	
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	if _, err := a.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	return nil
}

// LogCollection logs a data collection event
func (a *AuditLogger) LogCollection(routerID, dataType string, recordCount int, details map[string]interface{}) error {
	return a.Log(AuditEntry{
		EventType:   "data_collection",
		RouterID:    routerID,
		DataType:    dataType,
		RecordCount: recordCount,
		Details:     details,
	})
}

// LogTransmission logs a data transmission event
func (a *AuditLogger) LogTransmission(routerID, dataType string, recordCount int, destination string) error {
	return a.Log(AuditEntry{
		EventType:   "data_transmission",
		RouterID:    routerID,
		DataType:    dataType,
		RecordCount: recordCount,
		Details: map[string]interface{}{
			"destination": destination,
		},
	})
}

// Close closes the audit log file
func (a *AuditLogger) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.file != nil {
		return a.file.Close()
	}
	return nil
}
