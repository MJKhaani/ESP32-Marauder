package reports

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/monitor"
	"github.com/MJKhaani/ESP32-Marauder/internal/policy"
	"github.com/MJKhaani/ESP32-Marauder/internal/survey"
)

// ReportSummary encapsulates findings from a research session.
type ReportSummary struct {
	Title        string                 `json:"title"`
	GeneratedAt  time.Time              `json:"generated_at"`
	DeviceModel  string                 `json:"device_model"`
	FirmwareVer  string                 `json:"firmware_version"`
	AccessPoints []*survey.AccessPoint  `json:"access_points"`
	Stations     []*survey.Station      `json:"stations"`
	Alerts       []monitor.AnomalyAlert `json:"alerts"`
	AuditTrail   []policy.AuditRecord   `json:"audit_trail,omitempty"`
}

// GenerateSanitizedReport creates a privacy-preserving summary of session data.
func GenerateSanitizedReport(title, devModel, fwVer string, aps []*survey.AccessPoint, stas []*survey.Station, alerts []monitor.AnomalyAlert, audit []policy.AuditRecord) *ReportSummary {
	sanitizedAPs := make([]*survey.AccessPoint, len(aps))
	for i, ap := range aps {
		cp := *ap
		cp.BSSID = SanitizeMAC(ap.BSSID)
		sanitizedAPs[i] = &cp
	}

	sanitizedSTAs := make([]*survey.Station, len(stas))
	for i, sta := range stas {
		cp := *sta
		cp.MAC = SanitizeMAC(sta.MAC)
		cp.AssociatedBSSID = SanitizeMAC(sta.AssociatedBSSID)
		sanitizedSTAs[i] = &cp
	}

	return &ReportSummary{
		Title:        title,
		GeneratedAt:  time.Now(),
		DeviceModel:  devModel,
		FirmwareVer:  fwVer,
		AccessPoints: sanitizedAPs,
		Stations:     sanitizedSTAs,
		Alerts:       alerts,
		AuditTrail:   audit,
	}
}

// SanitizeMAC masks the last 3 octets of a MAC address (e.g. 00:11:22:xx:xx:xx)
func SanitizeMAC(mac string) string {
	parts := strings.Split(mac, ":")
	if len(parts) == 6 {
		return fmt.Sprintf("%s:%s:%s:xx:xx:xx", parts[0], parts[1], parts[2])
	}
	return "xx:xx:xx:xx:xx:xx"
}

// ExportJSON formats the report summary as indented JSON.
func (r *ReportSummary) ExportJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// ExportHTML produces a standalone, styled HTML executive report.
func (r *ReportSummary) ExportHTML(w io.Writer) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>%s - Marauder Studio Report</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; padding: 2rem; }
    h1 { color: #38bdf8; border-bottom: 2px solid #1e293b; padding-bottom: 0.5rem; }
    .meta { color: #94a3b8; font-size: 0.875rem; margin-bottom: 2rem; }
    table { width: 100%%; border-collapse: collapse; margin-top: 1rem; margin-bottom: 2rem; background: #1e293b; border-radius: 6px; overflow: hidden; }
    th, td { padding: 0.75rem 1rem; text-align: left; border-bottom: 1px solid #334155; }
    th { background: #0f172a; color: #cbd5e1; }
    .badge { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600; }
    .badge-warn { background: #ca8a04; color: #000; }
    .badge-crit { background: #dc2626; color: #fff; }
  </style>
</head>
<body>
  <h1>%s</h1>
  <div class="meta">
    Device: %s | Firmware: %s | Generated: %s (Sanitized)
  </div>

  <h2>Discovered Access Points (%d)</h2>
  <table>
    <thead><tr><th>BSSID</th><th>SSID</th><th>Channel</th><th>Band</th><th>RSSI</th><th>Security</th><th>Vendor</th></tr></thead>
    <tbody>`,
		r.Title, r.Title, r.DeviceModel, r.FirmwareVer, r.GeneratedAt.Format(time.RFC3339), len(r.AccessPoints))

	for _, ap := range r.AccessPoints {
		html += fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%d</td><td>%s</td><td>%d dBm</td><td>%s</td><td>%s</td></tr>\n",
			ap.BSSID, ap.SSID, ap.Channel, ap.Band, ap.RSSI, ap.Encryption, ap.Vendor)
	}

	html += `</tbody></table>
  <h2>Defensive Anomaly Alerts (` + fmt.Sprintf("%d", len(r.Alerts)) + `)</h2>
  <table>
    <thead><tr><th>Time</th><th>Severity</th><th>Category</th><th>Description</th></tr></thead>
    <tbody>`

	for _, a := range r.Alerts {
		html += fmt.Sprintf("<tr><td>%s</td><td><span class=\"badge badge-warn\">%s</span></td><td>%s</td><td>%s</td></tr>\n",
			a.Timestamp.Format("15:04:05"), a.Severity, a.Category, a.Detail)
	}

	html += `</tbody></table>
</body></html>`

	_, err := io.WriteString(w, html)
	return err
}
