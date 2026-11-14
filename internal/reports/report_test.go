package reports

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/monitor"
	"github.com/MJKhaani/ESP32-Marauder/internal/survey"
)

func TestSanitizeReport(t *testing.T) {
	aps := []*survey.AccessPoint{
		{
			BSSID:      "00:11:22:33:44:55",
			SSID:       "Secret_Lab",
			Channel:    6,
			Band:       "2.4 GHz",
			RSSI:       -50,
			Encryption: "WPA3",
			Vendor:     "Cisco",
		},
	}
	alerts := []monitor.AnomalyAlert{
		{
			ID:        "1",
			Timestamp: time.Now(),
			Severity:  monitor.SeverityWarning,
			Category:  "evil_twin",
			Detail:    "Rogue AP detected",
		},
	}

	report := GenerateSanitizedReport("Lab Survey", "Marauder v6", "v1.20.0", aps, nil, alerts, nil)
	if report.AccessPoints[0].BSSID != "00:11:22:xx:xx:xx" {
		t.Errorf("BSSID not sanitized: %s", report.AccessPoints[0].BSSID)
	}

	var htmlBuf bytes.Buffer
	err := report.ExportHTML(&htmlBuf)
	if err != nil {
		t.Fatalf("ExportHTML failed: %v", err)
	}

	if !strings.Contains(htmlBuf.String(), "00:11:22:xx:xx:xx") {
		t.Errorf("Sanitized BSSID missing from HTML")
	}
}
