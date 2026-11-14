package serialtransport

import (
	"bytes"
	"testing"
	"time"
)

func TestMockPortReadWrite(t *testing.T) {
	mp := NewMockPort()
	mp.RegisterResponse("info", []byte("ESP32 Marauder v1.20.0\r\n"))

	// Write command
	n, err := mp.Write([]byte("info\r\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 6 {
		t.Fatalf("Expected 6 bytes written, got %d", n)
	}

	// Read response
	buf := make([]byte, 64)
	rn, err := mp.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if rn == 0 {
		t.Fatalf("Expected non-zero read")
	}

	got := string(buf[:rn])
	if !bytes.Contains([]byte(got), []byte("ESP32 Marauder v1.20.0")) {
		t.Fatalf("Unexpected read content: %s", got)
	}
}

func TestMockPortFailureModes(t *testing.T) {
	mp := NewMockPort()
	mp.SetFailOnWrite(true)

	_, err := mp.Write([]byte("test"))
	if err == nil {
		t.Fatalf("Expected error on write, got nil")
	}

	mp.SetFailOnWrite(false)
	mp.SetFailOnRead(true)

	buf := make([]byte, 10)
	_, err = mp.Read(buf)
	if err == nil {
		t.Fatalf("Expected error on read, got nil")
	}

	_ = mp.Close()
	_, err = mp.Write([]byte("after close"))
	if err != ErrPortClosed {
		t.Fatalf("Expected ErrPortClosed, got %v", err)
	}
}

func TestCandidateClassification(t *testing.T) {
	tests := []struct {
		vid      string
		pid      string
		product  string
		mfg      string
		expected bool
	}{
		{"303a", "1001", "ESP32-S3", "Espressif", true},
		{"10c4", "ea60", "CP2102", "Silicon Labs", true},
		{"1a86", "7523", "USB-SERIAL CH340", "wch.cn", true},
		{"0403", "6001", "FT232R USB UART", "FTDI", true},
		{"1234", "5678", "Generic Mouse", "Generic", false},
		{"ffff", "0000", "Custom ESP32 Board", "DIY", true},
	}

	for _, tc := range tests {
		got := isEspCandidate(tc.vid, tc.pid, tc.product, tc.mfg)
		if got != tc.expected {
			t.Errorf("isEspCandidate(%s, %s, %s, %s) = %v; want %v",
				tc.vid, tc.pid, tc.product, tc.mfg, got, tc.expected)
		}
	}
}

func TestMockPortEnumerator(t *testing.T) {
	enumerator := NewMockPortEnumerator()
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		t.Fatalf("GetDetailedPortsList error: %v", err)
	}
	if len(ports) < 2 {
		t.Fatalf("Expected at least 2 mock ports, got %d", len(ports))
	}

	port, err := enumerator.OpenPort(ports[0].Path, 115200)
	if err != nil {
		t.Fatalf("OpenPort failed: %v", err)
	}
	if port == nil {
		t.Fatalf("Expected non-nil port")
	}

	_ = port.SetReadTimeout(100 * time.Millisecond)
	_ = port.ResetInputBuffer()
	_ = port.ResetOutputBuffer()
	_ = port.SetDTR(true)
	_ = port.SetRTS(true)
	_ = port.Close()
}
