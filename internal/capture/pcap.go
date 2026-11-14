package capture

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	PCAPMagicSameEndian = 0xa1b2c3d4
	PCAPMagicSwapped    = 0xd4c3b2a1

	LinkTypeEthernet = 1
	LinkType80211    = 105
	LinkTypeRadiotap = 127
)

var (
	ErrInvalidPCAPHeader = errors.New("invalid or corrupt PCAP header")
	ErrPacketTooLarge    = errors.New("packet length exceeds max snaplen")
)

// GlobalHeader represents the standard 24-byte PCAP file header.
type GlobalHeader struct {
	MagicNumber  uint32
	VersionMajor uint16
	VersionMinor uint16
	ThisZone     int32
	SigFigs      uint32
	SnapLen      uint32
	Network      uint32
}

// PacketHeader represents the 16-byte record header before each frame.
type PacketHeader struct {
	TimestampSec  uint32
	TimestampUsec uint32
	IncludedLen   uint32
	OriginalLen   uint32
}

// PacketSummary provides a high-level representation of a parsed frame.
type PacketSummary struct {
	Index       int       `json:"index"`
	Timestamp   time.Time `json:"timestamp"`
	Length      uint32    `json:"length"`
	Type        string    `json:"type"`        // "Management", "Control", "Data", "Unknown"
	Subtype     string    `json:"subtype"`     // "Beacon", "Deauth", "ProbeReq", "EAPOL", etc.
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	BSSID       string    `json:"bssid"`
	Info        string    `json:"info"`
}

// StreamingPCAPReader reads PCAP files record by record in bounded memory.
type StreamingPCAPReader struct {
	r          io.Reader
	byteOrder  binary.ByteOrder
	Header     GlobalHeader
	packetIdx  int
}

// NewStreamingPCAPReader initializes and validates the PCAP global header.
func NewStreamingPCAPReader(r io.Reader) (*StreamingPCAPReader, error) {
	var magicBuf [4]byte
	if _, err := io.ReadFull(r, magicBuf[:]); err != nil {
		return nil, fmt.Errorf("failed reading magic number: %w", err)
	}

	magic := binary.BigEndian.Uint32(magicBuf[:])
	var bo binary.ByteOrder

	if magic == PCAPMagicSameEndian {
		bo = binary.BigEndian
	} else if binary.LittleEndian.Uint32(magicBuf[:]) == PCAPMagicSameEndian {
		bo = binary.LittleEndian
	} else {
		return nil, ErrInvalidPCAPHeader
	}

	header := GlobalHeader{
		MagicNumber: PCAPMagicSameEndian,
	}

	if err := binary.Read(r, bo, &header.VersionMajor); err != nil {
		return nil, err
	}
	if err := binary.Read(r, bo, &header.VersionMinor); err != nil {
		return nil, err
	}
	if err := binary.Read(r, bo, &header.ThisZone); err != nil {
		return nil, err
	}
	if err := binary.Read(r, bo, &header.SigFigs); err != nil {
		return nil, err
	}
	if err := binary.Read(r, bo, &header.SnapLen); err != nil {
		return nil, err
	}
	if err := binary.Read(r, bo, &header.Network); err != nil {
		return nil, err
	}

	if header.SnapLen == 0 || header.SnapLen > 262144 {
		header.SnapLen = 65535
	}

	return &StreamingPCAPReader{
		r:         r,
		byteOrder: bo,
		Header:    header,
		packetIdx: 0,
	}, nil
}

// NextPacket reads the next packet record without buffering past SnapLen.
func (p *StreamingPCAPReader) NextPacket() (*PacketSummary, []byte, error) {
	var ph PacketHeader
	if err := binary.Read(p.r, p.byteOrder, &ph.TimestampSec); err != nil {
		return nil, nil, err
	}
	if err := binary.Read(p.r, p.byteOrder, &ph.TimestampUsec); err != nil {
		return nil, nil, err
	}
	if err := binary.Read(p.r, p.byteOrder, &ph.IncludedLen); err != nil {
		return nil, nil, err
	}
	if err := binary.Read(p.r, p.byteOrder, &ph.OriginalLen); err != nil {
		return nil, nil, err
	}

	if ph.IncludedLen > p.Header.SnapLen {
		return nil, nil, ErrPacketTooLarge
	}

	buf := make([]byte, ph.IncludedLen)
	if _, err := io.ReadFull(p.r, buf); err != nil {
		return nil, nil, err
	}

	p.packetIdx++
	t := time.Unix(int64(ph.TimestampSec), int64(ph.TimestampUsec)*1000)

	summary := &PacketSummary{
		Index:     p.packetIdx,
		Timestamp: t,
		Length:    ph.OriginalLen,
		Type:      "Data",
		Subtype:   "Frame",
	}

	// Basic 802.11 frame decoding if linktype is 802.11 or radiotap
	payload := buf
	if p.Header.Network == LinkTypeRadiotap && len(buf) >= 4 {
		// Radiotap header length is at offset 2 (uint16)
		rtLen := p.byteOrder.Uint16(buf[2:4])
		if int(rtLen) < len(buf) {
			payload = buf[rtLen:]
		}
	}

	if len(payload) >= 2 {
		fc := payload[0]
		frameType := (fc >> 2) & 0x03
		frameSubtype := (fc >> 4) & 0x0f

		switch frameType {
		case 0:
			summary.Type = "Management"
			switch frameSubtype {
			case 8:
				summary.Subtype = "Beacon"
			case 4:
				summary.Subtype = "Probe Request"
			case 5:
				summary.Subtype = "Probe Response"
			case 12:
				summary.Subtype = "Deauthentication"
			case 10:
				summary.Subtype = "Disassociation"
			default:
				summary.Subtype = fmt.Sprintf("Mgmt (%d)", frameSubtype)
			}
		case 1:
			summary.Type = "Control"
			summary.Subtype = "Control"
		case 2:
			summary.Type = "Data"
			summary.Subtype = "Data"
		}
	}

	return summary, buf, nil
}

// ReadFileSummary pages through a PCAP file to gather aggregate statistics.
func ReadFileSummary(filePath string, maxPackets int) ([]PacketSummary, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader, err := NewStreamingPCAPReader(f)
	if err != nil {
		return nil, err
	}

	summaries := make([]PacketSummary, 0)
	for i := 0; maxPackets <= 0 || i < maxPackets; i++ {
		sum, _, err := reader.NextPacket()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return nil, err
		}
		summaries = append(summaries, *sum)
	}

	return summaries, nil
}
