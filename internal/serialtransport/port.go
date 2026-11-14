package serialtransport

import (
	"errors"
	"io"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// Common USB Vendor IDs for ESP32 and serial bridges.
const (
	VID_ESPRESSIF = "303a" // Native ESP32-S2/S3/C3/C5/C6 CDC
	VID_SILABS    = "10c4" // CP210x USB to UART Bridge
	VID_WCH       = "1a86" // CH340 / CH910x USB to UART
	VID_FTDI      = "0403" // FTDI FT232R
)

// ErrPortClosed is returned when an operation is performed on a closed port.
var ErrPortClosed = errors.New("serial port is closed")

// SerialPort abstracts a physical or virtual serial communication channel.
type SerialPort interface {
	io.ReadWriteCloser
	SetReadTimeout(timeout time.Duration) error
	ResetInputBuffer() error
	ResetOutputBuffer() error
	SetDTR(dtr bool) error
	SetRTS(rts bool) error
}

// RealPort wraps go.bug.st/serial.Port to fulfill SerialPort.
type RealPort struct {
	port serial.Port
}

func (r *RealPort) Read(p []byte) (n int, err error) {
	return r.port.Read(p)
}

func (r *RealPort) Write(p []byte) (n int, err error) {
	return r.port.Write(p)
}

func (r *RealPort) Close() error {
	return r.port.Close()
}

func (r *RealPort) SetReadTimeout(timeout time.Duration) error {
	return r.port.SetReadTimeout(timeout)
}

func (r *RealPort) ResetInputBuffer() error {
	return r.port.ResetInputBuffer()
}

func (r *RealPort) ResetOutputBuffer() error {
	return r.port.ResetOutputBuffer()
}

func (r *RealPort) SetDTR(dtr bool) error {
	return r.port.SetDTR(dtr)
}

func (r *RealPort) SetRTS(rts bool) error {
	return r.port.SetRTS(rts)
}

// PortInfo describes a detected serial port.
type PortInfo struct {
	Path         string `json:"path"`
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	SerialNumber string `json:"serial_number"`
	Product      string `json:"product"`
	Manufacturer string `json:"manufacturer"`
	IsCandidate  bool   `json:"is_candidate"`
}

// PortEnumerator provides methods to discover available serial ports.
type PortEnumerator interface {
	GetDetailedPortsList() ([]PortInfo, error)
	OpenPort(path string, baudRate int) (SerialPort, error)
}

// SystemPortEnumerator wraps go.bug.st/serial/enumerator for host hardware.
type SystemPortEnumerator struct{}

// NewSystemPortEnumerator returns a new SystemPortEnumerator.
func NewSystemPortEnumerator() *SystemPortEnumerator {
	return &SystemPortEnumerator{}
}

// GetDetailedPortsList returns all available host serial ports with candidate classification.
func (e *SystemPortEnumerator) GetDetailedPortsList() ([]PortInfo, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	result := make([]PortInfo, 0, len(ports))
	for _, p := range ports {
		vid := strings.ToLower(p.VID)
		pid := strings.ToLower(p.PID)
		isCandidate := isEspCandidate(vid, pid, p.Product, p.Manufacturer)

		result = append(result, PortInfo{
			Path:         p.Name,
			VID:          vid,
			PID:          pid,
			SerialNumber: p.SerialNumber,
			Product:      p.Product,
			Manufacturer: p.Manufacturer,
			IsCandidate:  isCandidate,
		})
	}
	return result, nil
}

// OpenPort opens the specified serial port with standard Marauder parameters (115200 baud, 8N1).
func (e *SystemPortEnumerator) OpenPort(path string, baudRate int) (SerialPort, error) {
	if baudRate <= 0 {
		baudRate = 115200
	}
	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	p, err := serial.Open(path, mode)
	if err != nil {
		return nil, err
	}
	return &RealPort{port: p}, nil
}

func isEspCandidate(vid, pid, product, manufacturer string) bool {
	switch vid {
	case VID_ESPRESSIF, VID_SILABS, VID_WCH, VID_FTDI:
		return true
	}
	prodLower := strings.ToLower(product)
	mfgLower := strings.ToLower(manufacturer)
	if strings.Contains(prodLower, "esp32") || strings.Contains(prodLower, "cp210") ||
		strings.Contains(prodLower, "ch340") || strings.Contains(prodLower, "ch910") ||
		strings.Contains(mfgLower, "espressif") {
		return true
	}
	return false
}
