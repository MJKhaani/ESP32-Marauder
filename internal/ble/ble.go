package ble

import (
	"strings"
	"sync"
	"time"
)

// BLEDevice models an observed Bluetooth Low Energy advertisement source.
type BLEDevice struct {
	MAC          string    `json:"mac"`
	Name         string    `json:"name,omitempty"`
	RSSI         int       `json:"rssi"`
	CompanyID    string    `json:"company_id,omitempty"`
	Manufacturer string    `json:"manufacturer,omitempty"`
	IsTracker    bool      `json:"is_tracker"`
	TrackerType  string    `json:"tracker_type,omitempty"` // "Apple AirTag", "Tile", "SmartTag"
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	Count        int       `json:"count"`
}

// BLEStore tracks BLE advertisements.
type BLEStore struct {
	mu      sync.RWMutex
	devices map[string]*BLEDevice
}

// NewBLEStore creates a new BLEStore.
func NewBLEStore() *BLEStore {
	return &BLEStore{
		devices: make(map[string]*BLEDevice),
	}
}

// RecordAdvertisement registers or updates an observed BLE device.
func (s *BLEStore) RecordAdvertisement(mac, name string, rssi int, rawAdv string) *BLEDevice {
	s.mu.Lock()
	defer s.mu.Unlock()

	normMAC := strings.ToLower(strings.TrimSpace(mac))
	now := time.Now()

	dev, exists := s.devices[normMAC]
	if !exists {
		dev = &BLEDevice{
			MAC:       normMAC,
			Name:      name,
			RSSI:      rssi,
			FirstSeen: now,
			LastSeen:  now,
			Count:     1,
		}
		s.devices[normMAC] = dev
	} else {
		dev.LastSeen = now
		dev.RSSI = rssi
		dev.Count++
		if name != "" && dev.Name == "" {
			dev.Name = name
		}
	}

	// Classify tracker or manufacturer if identifiable
	lowerAdv := strings.ToLower(rawAdv)
	if strings.Contains(lowerAdv, "4c00") || strings.Contains(lowerAdv, "apple") {
		dev.Manufacturer = "Apple, Inc."
		if strings.Contains(lowerAdv, "1219") || strings.Contains(lowerAdv, "findmy") {
			dev.IsTracker = true
			dev.TrackerType = "Apple AirTag / FindMy"
		}
	} else if strings.Contains(lowerAdv, "tile") {
		dev.Manufacturer = "Tile, Inc."
		dev.IsTracker = true
		dev.TrackerType = "Tile Tracker"
	} else if strings.Contains(lowerAdv, "0075") || strings.Contains(lowerAdv, "samsung") {
		dev.Manufacturer = "Samsung Electronics"
		if strings.Contains(lowerAdv, "smarttag") {
			dev.IsTracker = true
			dev.TrackerType = "Samsung SmartTag"
		}
	} else if strings.Contains(lowerAdv, "0006") || strings.Contains(lowerAdv, "microsoft") {
		dev.Manufacturer = "Microsoft Corp."
	} else if dev.Manufacturer == "" {
		dev.Manufacturer = "Generic BLE"
	}

	return dev
}

// GetDevices returns all recorded BLE devices.
func (s *BLEStore) GetDevices() []*BLEDevice {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*BLEDevice, 0, len(s.devices))
	for _, d := range s.devices {
		res = append(res, d)
	}
	return res
}

// Clear flushes recorded BLE devices.
func (s *BLEStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = make(map[string]*BLEDevice)
}
