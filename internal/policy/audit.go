package policy

import (
	"encoding/json"
	"io"
	"time"
)

// AuditRecord stores an unalterable local log of an authorized lab test.
type AuditRecord struct {
	ID              string    `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	Operator        string    `json:"operator"`
	Lab             string    `json:"lab"`
	DeviceID        string    `json:"device_id"`
	Command         string    `json:"command"`
	Targets         []string  `json:"targets"`
	DurationSeconds int       `json:"duration_seconds"`
	Approved        bool      `json:"approved"`
	StoppedAt       *time.Time `json:"stopped_at,omitempty"`
	StopConfirmed   bool      `json:"stop_confirmed"`
}

// ExportAuditTrail writes audit records to an io.Writer as formatted JSON.
func ExportAuditTrail(w io.Writer, records []AuditRecord) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}
