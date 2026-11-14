package monitor

import (
	"testing"

	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
)

func TestEvilTwinDetection(t *testing.T) {
	engine := NewMonitorEngine()

	// Initial AP
	engine.ProcessAP(marauder.APScanEventData{
		SSID:    "Corporate_WiFi",
		BSSID:   "00:11:22:33:44:55",
		Channel: 6,
		RSSI:    -50,
	})

	// Same SSID from different BSSID -> Evil Twin
	engine.ProcessAP(marauder.APScanEventData{
		SSID:    "Corporate_WiFi",
		BSSID:   "de:ad:be:ef:00:01",
		Channel: 6,
		RSSI:    -40,
	})

	alerts := engine.GetAlerts()
	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Category != "evil_twin" {
		t.Errorf("Expected evil_twin category, got %s", alerts[0].Category)
	}
}

func TestDeauthBurstDetection(t *testing.T) {
	engine := NewMonitorEngine()

	for i := 0; i < 11; i++ {
		engine.ProcessDeauth(marauder.SniffDeauthEventData{
			SourceMAC: "00:11:22:33:44:55",
			DestMAC:   "ff:ff:ff:ff:ff:ff",
			BSSID:     "00:11:22:33:44:55",
			Reason:    "Deauth",
		})
	}

	alerts := engine.GetAlerts()
	if len(alerts) == 0 {
		t.Fatalf("Expected deauth burst alert, got none")
	}
	if alerts[0].Category != "deauth_flood" {
		t.Errorf("Expected deauth_flood alert, got %s", alerts[0].Category)
	}
}
