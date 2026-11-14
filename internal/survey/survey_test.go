package survey

import (
	"bytes"
	"strings"
	"testing"

	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
)

func TestSurveyStoreRecordAndFilter(t *testing.T) {
	store := NewSurveyStore()

	// Record 2 APs
	store.RecordAP(marauder.APScanEventData{
		Index:      0,
		SSID:       "Lab_5G",
		BSSID:      "00:11:22:33:44:55",
		RSSI:       -45,
		Channel:    36,
		Encryption: "WPA2",
	})
	store.RecordAP(marauder.APScanEventData{
		Index:      1,
		SSID:       "Public_Open",
		BSSID:      "aa:bb:cc:dd:ee:ff",
		RSSI:       -75,
		Channel:    6,
		Encryption: "OPEN",
	})

	aps := store.GetAPs(false)
	if len(aps) != 2 {
		t.Fatalf("Expected 2 APs, got %d", len(aps))
	}
	if aps[0].Band != "5 GHz" {
		t.Errorf("Expected 5 GHz band for channel 36, got %s", aps[0].Band)
	}
	if aps[0].Vendor != "Cisco Systems" {
		t.Errorf("Expected Cisco Systems vendor for 00:11:22, got %s", aps[0].Vendor)
	}

	// Set allowlist to only lab AP
	store.SetAllowlist([]string{"00:11:22:33:44:55"})
	filtered := store.GetAPs(true)
	if len(filtered) != 1 {
		t.Fatalf("Expected 1 allowlisted AP, got %d", len(filtered))
	}
	if filtered[0].SSID != "Lab_5G" {
		t.Errorf("Expected Lab_5G, got %s", filtered[0].SSID)
	}

	// Test Station record
	store.RecordStation(marauder.StationScanEventData{
		Index:           0,
		MAC:             "12:34:56:78:9a:bc",
		RSSI:            -60,
		AssociatedBSSID: "00:11:22:33:44:55",
	})

	stas := store.GetStations()
	if len(stas) != 1 {
		t.Fatalf("Expected 1 station, got %d", len(stas))
	}
	if stas[0].AssociatedBSSID != "00:11:22:33:44:55" {
		t.Errorf("Expected associated BSSID match, got %s", stas[0].AssociatedBSSID)
	}

	// Test CSV export
	var buf bytes.Buffer
	err := store.ExportCSV(&buf)
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Lab_5G") {
		t.Errorf("CSV missing Lab_5G")
	}
}
