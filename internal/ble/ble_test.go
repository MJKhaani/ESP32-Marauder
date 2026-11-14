package ble

import (
	"testing"
)

func TestBLEStoreTracking(t *testing.T) {
	store := NewBLEStore()

	// Apple AirTag advertisement
	d1 := store.RecordAdvertisement("aa:bb:cc:11:22:33", "AirTag", -65, "4c00 1219 findmy")
	if !d1.IsTracker {
		t.Errorf("Expected device to be classified as tracker")
	}
	if d1.TrackerType != "Apple AirTag / FindMy" {
		t.Errorf("Expected Apple AirTag / FindMy, got %s", d1.TrackerType)
	}

	// Update advertisement
	d1Update := store.RecordAdvertisement("aa:bb:cc:11:22:33", "", -60, "")
	if d1Update.Count != 2 {
		t.Errorf("Expected count 2, got %d", d1Update.Count)
	}
	if d1Update.RSSI != -60 {
		t.Errorf("Expected updated RSSI -60, got %d", d1Update.RSSI)
	}

	devices := store.GetDevices()
	if len(devices) != 1 {
		t.Fatalf("Expected 1 device, got %d", len(devices))
	}
}
