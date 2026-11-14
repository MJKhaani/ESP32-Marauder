package diagnostics

import (
	"encoding/json"
	"io"
	"runtime"
	"time"
)

// SupportBundle encapsulates non-sensitive environment diagnostics for issue reporting.
type SupportBundle struct {
	Timestamp      time.Time `json:"timestamp"`
	AppVersion     string    `json:"app_version"`
	OS             string    `json:"os"`
	Arch           string    `json:"arch"`
	GoVersion      string    `json:"go_version"`
	ActiveDevices  int       `json:"active_devices"`
	RecentErrors   []string  `json:"recent_errors"`
}

// GenerateSupportBundle creates a sanitized support report.
func GenerateSupportBundle(appVersion string, activeDevCount int, recentErrors []string) *SupportBundle {
	return &SupportBundle{
		Timestamp:     time.Now(),
		AppVersion:    appVersion,
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		GoVersion:     runtime.Version(),
		ActiveDevices: activeDevCount,
		RecentErrors:  recentErrors,
	}
}

// ExportJSON writes the bundle to an io.Writer.
func (b *SupportBundle) ExportJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(b)
}
