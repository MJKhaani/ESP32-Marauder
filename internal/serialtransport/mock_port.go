package serialtransport

import (
	"bytes"
	"io"
	"sync"
	"time"
)

// MockPort provides an in-memory implementation of SerialPort for deterministic testing.
type MockPort struct {
	mu          sync.Mutex
	cond        *sync.Cond
	readBuf     *bytes.Buffer
	writtenBuf  *bytes.Buffer
	handlers    map[string][]byte
	closed      bool
	readTimeout time.Duration
	failOnWrite bool
	failOnRead  bool
	chunkSize   int
	dtr         bool
	rts         bool
}

// NewMockPort creates a new MockPort instance.
func NewMockPort() *MockPort {
	mp := &MockPort{
		readBuf:    bytes.NewBuffer(nil),
		writtenBuf: bytes.NewBuffer(nil),
		handlers:   make(map[string][]byte),
		chunkSize:  0, // no chunking by default
	}
	mp.cond = sync.NewCond(&mp.mu)
	return mp
}

// RegisterResponse associates an input command string with a canned response output.
func (m *MockPort) RegisterResponse(command string, response []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[command] = response
}

// FeedBytes directly injects bytes into the read buffer.
func (m *MockPort) FeedBytes(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readBuf.Write(data)
	m.cond.Broadcast()
}

func (m *MockPort) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for m.readBuf.Len() == 0 {
		if m.closed {
			return 0, io.EOF
		}
		if m.failOnRead {
			return 0, io.ErrUnexpectedEOF
		}
		m.cond.Wait()
	}

	if m.closed {
		return 0, io.EOF
	}

	toRead := len(p)
	if m.chunkSize > 0 && toRead > m.chunkSize {
		toRead = m.chunkSize
	}
	return m.readBuf.Read(p[:toRead])
}

func (m *MockPort) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, ErrPortClosed
	}
	if m.failOnWrite {
		return 0, io.ErrClosedPipe
	}

	m.writtenBuf.Write(p)
	cmd := string(bytes.TrimSpace(p))

	// Match handler
	if resp, ok := m.handlers[cmd]; ok {
		m.readBuf.Write(resp)
		m.cond.Broadcast()
	}

	return len(p), nil
}

func (m *MockPort) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.cond.Broadcast()
	return nil
}

func (m *MockPort) SetReadTimeout(timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readTimeout = timeout
	return nil
}

func (m *MockPort) ResetInputBuffer() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readBuf.Reset()
	return nil
}

func (m *MockPort) ResetOutputBuffer() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writtenBuf.Reset()
	return nil
}

func (m *MockPort) SetDTR(dtr bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dtr = dtr
	return nil
}

func (m *MockPort) SetRTS(rts bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rts = rts
	return nil
}

func (m *MockPort) GetWrittenBytes() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	b := make([]byte, m.writtenBuf.Len())
	copy(b, m.writtenBuf.Bytes())
	return b
}

func (m *MockPort) SetFailOnWrite(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnWrite = fail
}

func (m *MockPort) SetFailOnRead(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnRead = fail
	m.cond.Broadcast()
}

// MockPortEnumerator provides a test enumerator that yields mock ports.
type MockPortEnumerator struct {
	mu        sync.Mutex
	ports     []PortInfo
	mockPorts map[string]*MockPort
}

// NewMockPortEnumerator returns a new MockPortEnumerator with mock devices.
func NewMockPortEnumerator() *MockPortEnumerator {
	return &MockPortEnumerator{
		ports: []PortInfo{
			{
				Path:         "/dev/tty.usbserial-0001",
				VID:          VID_SILABS,
				PID:          "ea60",
				SerialNumber: "0001",
				Product:      "CP2102 USB to UART Bridge Controller",
				Manufacturer: "Silicon Labs",
				IsCandidate:  true,
			},
			{
				Path:         "/dev/tty.usbmodem1101",
				VID:          VID_ESPRESSIF,
				PID:          "1001",
				SerialNumber: "ESP32S3_01",
				Product:      "ESP32-S3 USB JTAG/serial debug unit",
				Manufacturer: "Espressif",
				IsCandidate:  true,
			},
		},
		mockPorts: make(map[string]*MockPort),
	}
}

func (m *MockPortEnumerator) SetPorts(ports []PortInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ports = ports
}

func (m *MockPortEnumerator) RegisterPort(path string, port *MockPort) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockPorts[path] = port
}

func (m *MockPortEnumerator) GetDetailedPortsList() ([]PortInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]PortInfo, len(m.ports))
	copy(res, m.ports)
	return res, nil
}

func (m *MockPortEnumerator) OpenPort(path string, baudRate int) (SerialPort, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.mockPorts[path]; ok {
		return p, nil
	}
	mp := NewMockPort()
	m.mockPorts[path] = mp
	return mp, nil
}
