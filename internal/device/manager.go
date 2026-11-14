package device

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/capabilities"
	"github.com/MJKhaani/ESP32-Marauder/internal/serialtransport"
)

var (
	ErrDeviceNotFound  = errors.New("device not found")
	ErrDeviceAlreadyConnected = errors.New("device is already connected")
)

// Manager coordinates device detection, connection, and sessions.
type Manager struct {
	mu               sync.RWMutex
	enumerator       serialtransport.PortEnumerator
	evaluator        *capabilities.Evaluator
	devices          map[string]*Device
	sessions         map[string]*Session
	listeners        []DeviceEventListener
	listenersMu      sync.RWMutex
	sessionListeners []SessionListener
}

// NewManager creates a new Device Manager.
func NewManager(enumerator serialtransport.PortEnumerator) *Manager {
	return &Manager{
		enumerator:       enumerator,
		evaluator:        capabilities.NewEvaluator(),
		devices:          make(map[string]*Device),
		sessions:         make(map[string]*Session),
		listeners:        make([]DeviceEventListener, 0),
		sessionListeners: make([]SessionListener, 0),
	}
}

// AddSessionListener registers a listener to be attached to all current and future sessions.
func (m *Manager) AddSessionListener(l SessionListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessionListeners = append(m.sessionListeners, l)
	for _, s := range m.sessions {
		s.AddListener(l)
	}
}

// AddListener registers an event listener for device state changes.
func (m *Manager) AddListener(l DeviceEventListener) {
	m.listenersMu.Lock()
	defer m.listenersMu.Unlock()
	m.listeners = append(m.listeners, l)
}

func (m *Manager) notify(event string, d *Device, data interface{}) {
	m.listenersMu.RLock()
	listeners := append([]DeviceEventListener(nil), m.listeners...)
	m.listenersMu.RUnlock()

	for _, l := range listeners {
		l(event, d, data)
	}
}

// ScanPorts enumerates available serial ports and registers detected candidates.
func (m *Manager) ScanPorts() ([]*Device, error) {
	ports, err := m.enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate ports: %w", err)
	}

	m.mu.Lock()
	now := time.Now()
	for _, p := range ports {
		id := p.Path
		if d, exists := m.devices[id]; exists {
			d.LastSeen = now
			d.Port = p
		} else {
			d = &Device{
				ID:   id,
				Port: p,
				Identity: Identity{
					StablePortID: id,
					Path:         p.Path,
					VID:          p.VID,
					PID:          p.PID,
					Confidence:   "unverified",
				},
				State:    StateDisconnected,
				LastSeen: now,
			}
			m.devices[id] = d
		}
	}

	result := make([]*Device, 0, len(m.devices))
	for _, d := range m.devices {
		result = append(result, d)
	}
	m.mu.Unlock()

	return result, nil
}

// Connect establishes a serial connection to the specified device path.
func (m *Manager) Connect(path string) (*Device, error) {
	m.mu.Lock()
	d, exists := m.devices[path]
	if !exists {
		d = &Device{
			ID:   path,
			Port: serialtransport.PortInfo{Path: path, IsCandidate: true},
			Identity: Identity{
				StablePortID: path,
				Path:         path,
				Confidence:   "unverified",
			},
			State:    StateDisconnected,
			LastSeen: time.Now(),
		}
		m.devices[path] = d
	}

	if d.State == StateConnected {
		m.mu.Unlock()
		return d, ErrDeviceAlreadyConnected
	}

	d.State = StateConnecting
	m.mu.Unlock()
	m.notify("connecting", d, nil)

	port, err := m.enumerator.OpenPort(path, 115200)
	if err != nil {
		m.mu.Lock()
		d.State = StateError
		d.LastError = err.Error()
		m.mu.Unlock()
		m.notify("error", d, err.Error())
		return d, fmt.Errorf("failed to open port %s: %w", path, err)
	}

	session := NewSession(d, port)
	for _, sl := range m.sessionListeners {
		session.AddListener(sl)
	}

	m.mu.Lock()
	m.sessions[d.ID] = session
	d.State = StateConnected
	d.LastError = ""
	m.mu.Unlock()

	// Interrogate device with `info` and `help`
	_ = session.InitializeInterrogation(m.evaluator)

	m.notify("connected", d, nil)
	return d, nil
}

// Disconnect gracefully terminates a device session.
func (m *Manager) Disconnect(deviceID string) error {
	m.mu.Lock()
	session, exists := m.sessions[deviceID]
	if !exists {
		m.mu.Unlock()
		return ErrDeviceNotFound
	}

	delete(m.sessions, deviceID)
	d := m.devices[deviceID]
	if d != nil {
		d.State = StateDisconnected
	}
	m.mu.Unlock()

	_ = session.Close()
	if d != nil {
		m.notify("disconnected", d, nil)
	}
	return nil
}

// GetSession retrieves the active session for a device.
func (m *Manager) GetSession(deviceID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[deviceID]
	return s, ok
}

// GetDevice retrieves device metadata by ID.
func (m *Manager) GetDevice(deviceID string) (*Device, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.devices[deviceID]
	return d, ok
}

// ListDevices returns all known devices.
func (m *Manager) ListDevices() []*Device {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Device, 0, len(m.devices))
	for _, d := range m.devices {
		list = append(list, d)
	}
	return list
}
