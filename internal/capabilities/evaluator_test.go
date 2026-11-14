package capabilities

import (
	"testing"

	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
)

func TestEvaluatorModernV120(t *testing.T) {
	evaluator := NewEvaluator()
	profile := &marauder.FirmwareProfile{
		RawVersion:   "v1.20.0",
		Major:        1,
		Minor:        20,
		Patch:        0,
		HardwareName: "Marauder v6",
		HasSD:        true,
		HasGPS:       true,
		SupportedCommands: map[string]bool{
			"scanall":     true,
			"sniffbeacon": true,
			"sniffpmkid":  true,
			"sniffraw":    true,
			"sniffbt":     true,
			"attack":      true,
		},
	}

	matrix := evaluator.Evaluate("dev-01", profile)
	if !matrix.IsAvailable("wifi_survey") {
		t.Errorf("Expected wifi_survey to be available")
	}
	if !matrix.IsAvailable("sniff_raw_sd") {
		t.Errorf("Expected sniff_raw_sd to be available")
	}
	if !matrix.IsAvailable("ble_observation") {
		t.Errorf("Expected ble_observation to be available")
	}
	if !matrix.IsAvailable("gps_telemetry") {
		t.Errorf("Expected gps_telemetry to be available")
	}
	if !matrix.IsAvailable("authorized_lab") {
		t.Errorf("Expected authorized_lab to be available")
	}
}

func TestEvaluatorFlipperS2NoBLE(t *testing.T) {
	evaluator := NewEvaluator()
	profile := &marauder.FirmwareProfile{
		RawVersion:   "v1.20.0",
		HardwareName: "Flipper Zero WiFi S2",
		HasSD:        false,
		HasGPS:       false,
		SupportedCommands: map[string]bool{
			"scanall": true,
		},
	}

	matrix := evaluator.Evaluate("flipper-01", profile)

	bleCap, _ := matrix.Get("ble_observation")
	if bleCap.Status != StatusUnsupported {
		t.Errorf("Expected BLE on S2 to be Unsupported, got %s", bleCap.Status)
	}

	sdCap, _ := matrix.Get("sniff_raw_sd")
	if sdCap.Status != StatusRequiresHardware {
		t.Errorf("Expected sniff_raw_sd without SD to be RequiresHardware, got %s", sdCap.Status)
	}

	gpsCap, _ := matrix.Get("gps_telemetry")
	if gpsCap.Status != StatusRequiresHardware {
		t.Errorf("Expected gps_telemetry without GPS to be RequiresHardware, got %s", gpsCap.Status)
	}
}
