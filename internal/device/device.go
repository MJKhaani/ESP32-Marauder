package device

import (
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/capabilities"
	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
	"github.com/MJKhaani/ESP32-Marauder/internal/serialtransport"
)

// State describes the connection and operational state of a device.
type State string

const (
	StateDisconnected State = "disconnected"
	StateConnecting   State = "connecting"
	StateConnected    State = "connected"
	StateScanning     State = "scanning"
	StateError        State = "error"
)

// Identity represents the verified hardware and firmware signature of a device.
type Identity struct {
	StablePortID string `json:"stable_port_id"`
	Path         string `json:"path"`
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	ChipFamily   string `json:"chip_family"`
	BoardModel   string `json:"board_model"`
	FirmwareVer  string `json:"firmware_version"`
	Confidence   string `json:"confidence"` // "verified", "presumed", "unknown"
}

// Device encapsulates a detected or connected hardware unit.
type Device struct {
	ID           string                         `json:"id"`
	Port         serialtransport.PortInfo       `json:"port"`
	Identity     Identity                       `json:"identity"`
	State        State                          `json:"state"`
	Profile      *marauder.FirmwareProfile      `json:"profile,omitempty"`
	Capabilities *capabilities.CapabilityMatrix `json:"capabilities,omitempty"`
	LastSeen     time.Time                      `json:"last_seen"`
	LastError    string                         `json:"last_error,omitempty"`
}

// DeviceEventListener handles asynchronous notifications about device changes.
type DeviceEventListener func(event string, d *Device, data interface{})
