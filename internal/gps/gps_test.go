package gps

import (
	"bytes"
	"strings"
	"testing"
)

func TestNMEAParserRMC(t *testing.T) {
	parser := NewNMEAParser()

	// 4807.038,N -> 48 deg + 7.038/60 = 48.1173
	// 01131.000,E -> 11 deg + 31.000/60 = 11.51666
	rmc := "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A"

	coord, err := parser.ParseSentence(rmc)
	if err != nil {
		t.Fatalf("ParseSentence failed: %v", err)
	}
	if coord.Fix != Fix3D {
		t.Errorf("Expected Fix3D, got %s", coord.Fix)
	}
	if coord.Latitude < 48.1 || coord.Latitude > 48.2 {
		t.Errorf("Unexpected latitude: %f", coord.Latitude)
	}
	if coord.Longitude < 11.5 || coord.Longitude > 11.6 {
		t.Errorf("Unexpected longitude: %f", coord.Longitude)
	}
}

func TestGPXExport(t *testing.T) {
	coords := []GPSCoordinate{
		{Latitude: 37.7749, Longitude: -122.4194, Altitude: 15.0},
		{Latitude: 37.7750, Longitude: -122.4195, Altitude: 16.0},
	}

	var buf bytes.Buffer
	err := ExportGPX(&buf, "Test Track", coords)
	if err != nil {
		t.Fatalf("ExportGPX failed: %v", err)
	}

	xmlStr := buf.String()
	if !strings.Contains(xmlStr, "<trk>") || !strings.Contains(xmlStr, "lat=\"37.7749\"") {
		t.Errorf("GPX output missing expected elements: %s", xmlStr)
	}
}

func TestWardriveCSVParser(t *testing.T) {
	csvData := `WigleWifi-1.4,appRelease=2.26,model=ESP32,release=1.20,device=Marauder
MAC,SSID,AuthMode,FirstSeen,Channel,RSSI,CurrentLatitude,CurrentLongitude,AltitudeMeters,AccuracyMeters,Type
00:11:22:33:44:55,Lab_WiFi,[WPA2-PSK-CCMP][ESS],2026-11-14 12:00:00,6,-50,37.7749,-122.4194,15.0,3.5,WIFI
`
	entries, err := ParseWardriveCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("ParseWardriveCSV failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].SSID != "Lab_WiFi" {
		t.Errorf("Expected Lab_WiFi, got %s", entries[0].SSID)
	}
	if entries[0].BSSID != "00:11:22:33:44:55" {
		t.Errorf("Expected normalized BSSID, got %s", entries[0].BSSID)
	}
}

func TestSanitizeCoord(t *testing.T) {
	lat, lon := SanitizeCoord(37.774929, -122.419416)
	if lat != 37.77 || lon != -122.41 {
		t.Errorf("Expected 37.77, -122.41, got %f, %f", lat, lon)
	}
}
