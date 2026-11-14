package capture

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func buildSyntheticPCAP() []byte {
	var buf bytes.Buffer

	// Global header (24 bytes)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(PCAPMagicSameEndian))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(2))     // major
	_ = binary.Write(&buf, binary.LittleEndian, uint16(4))     // minor
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))      // zone
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))     // sigfigs
	_ = binary.Write(&buf, binary.LittleEndian, uint32(65535)) // snaplen
	_ = binary.Write(&buf, binary.LittleEndian, uint32(105))   // 802.11 linktype

	// Packet 1: 802.11 Beacon frame (FC = 0x80 -> subtype 8, type 0)
	payload1 := []byte{0x80, 0x00, 0x00, 0x00}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1700000000)) // sec
	_ = binary.Write(&buf, binary.LittleEndian, uint32(100))        // usec
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(payload1)))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(payload1)))
	buf.Write(payload1)

	// Packet 2: 802.11 Deauth frame (FC = 0xc0 -> subtype 12, type 0)
	payload2 := []byte{0xc0, 0x00, 0x00, 0x00}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1700000001))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(200))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(payload2)))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(payload2)))
	buf.Write(payload2)

	return buf.Bytes()
}

func TestStreamingPCAPReader(t *testing.T) {
	data := buildSyntheticPCAP()
	reader, err := NewStreamingPCAPReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("NewStreamingPCAPReader failed: %v", err)
	}

	if reader.Header.Network != 105 {
		t.Errorf("Expected network 105, got %d", reader.Header.Network)
	}

	// Read packet 1
	p1, _, err := reader.NextPacket()
	if err != nil {
		t.Fatalf("NextPacket 1 failed: %v", err)
	}
	if p1.Type != "Management" || p1.Subtype != "Beacon" {
		t.Errorf("Packet 1 expected Management Beacon, got %s %s", p1.Type, p1.Subtype)
	}

	// Read packet 2
	p2, _, err := reader.NextPacket()
	if err != nil {
		t.Fatalf("NextPacket 2 failed: %v", err)
	}
	if p2.Type != "Management" || p2.Subtype != "Deauthentication" {
		t.Errorf("Packet 2 expected Management Deauthentication, got %s %s", p2.Type, p2.Subtype)
	}

	// Read EOF
	_, _, err = reader.NextPacket()
	if err != io.EOF {
		t.Errorf("Expected io.EOF on end of stream, got %v", err)
	}
}

func TestDiscoverSDCaptures(t *testing.T) {
	tmpDir := t.TempDir()
	pcapFile := filepath.Join(tmpDir, "test_capture.pcap")
	_ = os.WriteFile(pcapFile, buildSyntheticPCAP(), 0644)

	caps, err := DiscoverSDCaptures(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverSDCaptures failed: %v", err)
	}
	if len(caps) != 1 {
		t.Fatalf("Expected 1 discovered capture, got %d", len(caps))
	}
	if caps[0].FileName != "test_capture.pcap" {
		t.Errorf("Expected test_capture.pcap, got %s", caps[0].FileName)
	}
}
