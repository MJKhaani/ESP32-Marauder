package marauder

import (
	"os"
	"strings"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		maj   int
		min   int
		patch int
		err   bool
	}{
		{"v1.20.0", 1, 20, 0, false},
		{"v0.13.4", 0, 13, 4, false},
		{"1.0", 1, 0, 0, false},
		{"invalid", 0, 0, 0, true},
	}

	for _, tc := range tests {
		maj, min, patch, err := ParseVersion(tc.input)
		if tc.err && err == nil {
			t.Errorf("ParseVersion(%s) expected error, got nil", tc.input)
		}
		if !tc.err && err != nil {
			t.Errorf("ParseVersion(%s) unexpected error: %v", tc.input, err)
		}
		if maj != tc.maj || min != tc.min || patch != tc.patch {
			t.Errorf("ParseVersion(%s) = (%d,%d,%d); want (%d,%d,%d)",
				tc.input, maj, min, patch, tc.maj, tc.min, tc.patch)
		}
	}
}

func TestWireParserInfoV120(t *testing.T) {
	data, err := os.ReadFile("../../testdata/info_v120.txt")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	parser := NewWireParser()
	var infoFound bool
	var infoData InfoEventData

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		ev := parser.ParseLine(line)
		if ev != nil && ev.Type == EventInfo {
			infoFound = true
			infoData = ev.Data.(InfoEventData)
		}
	}

	if !infoFound {
		t.Fatalf("Expected InfoEvent, none found")
	}
	if infoData.Version != "v1.20.0" {
		t.Errorf("Expected version v1.20.0, got %s", infoData.Version)
	}
	if infoData.Hardware != "Marauder v6" {
		t.Errorf("Expected Hardware 'Marauder v6', got %s", infoData.Hardware)
	}
	if !infoData.HasSD {
		t.Errorf("Expected HasSD to be true")
	}
	if !infoData.HasGPS {
		t.Errorf("Expected HasGPS to be true")
	}
}

func TestWireParserScanAll(t *testing.T) {
	data, err := os.ReadFile("../../testdata/scanall_output.txt")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	parser := NewWireParser()
	apCount := 0
	staCount := 0

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		ev := parser.ParseLine(line)
		if ev == nil {
			continue
		}
		switch ev.Type {
		case EventAP:
			apCount++
			ap := ev.Data.(APScanEventData)
			if ap.BSSID == "" {
				t.Errorf("Empty BSSID in parsed AP")
			}
		case EventSTA:
			staCount++
			sta := ev.Data.(StationScanEventData)
			if sta.MAC == "" {
				t.Errorf("Empty MAC in parsed STA")
			}
		}
	}

	if apCount != 3 {
		t.Errorf("Expected 3 AP events, got %d", apCount)
	}
	if staCount != 2 {
		t.Errorf("Expected 2 STA events, got %d", staCount)
	}
}

func TestCommandBuilder(t *testing.T) {
	cb := NewCommandBuilder()
	if cb.ScanAll() != "scanall\r\n" {
		t.Errorf("Unexpected ScanAll output: %s", cb.ScanAll())
	}
	if cb.StopScan() != "stopscan\r\n" {
		t.Errorf("Unexpected StopScan output: %s", cb.StopScan())
	}

	chCmd, err := cb.Channel(6)
	if err != nil || chCmd != "channel -s 6\r\n" {
		t.Errorf("Unexpected Channel output: %s (err: %v)", chCmd, err)
	}

	_, err = cb.Channel(200)
	if err == nil {
		t.Errorf("Expected error for channel 200, got nil")
	}

	_, err = cb.FormatCustom("bad\ncmd")
	if err == nil {
		t.Errorf("Expected error for newline in custom command, got nil")
	}
}
