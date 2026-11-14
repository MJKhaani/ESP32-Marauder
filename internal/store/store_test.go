package store

import (
	"testing"
)

func TestStoreMigrationsAndOperations(t *testing.T) {
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore(:memory:) failed: %v", err)
	}
	defer s.Close()

	// 1. Save Device
	err = s.SaveDevice("dev-01", "Lab Node 1", "Marauder v6", "ESP32", "v1.20.0")
	if err != nil {
		t.Fatalf("SaveDevice failed: %v", err)
	}

	// Upsert update
	err = s.SaveDevice("dev-01", "Lab Node 1 (Renamed)", "Marauder v6", "ESP32", "v1.20.1")
	if err != nil {
		t.Fatalf("SaveDevice update failed: %v", err)
	}

	// 2. Save Audit Log
	err = s.SaveAuditLog("aud-01", "Bob", "Lab 4", "attack -t deauth", `["00:11:22:33:44:55"]`, 30, true)
	if err != nil {
		t.Fatalf("SaveAuditLog failed: %v", err)
	}
}
