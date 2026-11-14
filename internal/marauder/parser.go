package marauder

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// EventType identifies the category of parsed serial event.
type EventType string

const (
	EventRawLine EventType = "raw_line"
	EventInfo    EventType = "info"
	EventHelp    EventType = "help"
	EventAP      EventType = "ap"
	EventSTA     EventType = "sta"
	EventBeacon  EventType = "beacon"
	EventDeauth  EventType = "deauth"
	EventPMKID   EventType = "pmkid"
	EventStatus  EventType = "status"
)

// ParsedEvent wraps any parsed message from Marauder CLI.
type ParsedEvent struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	RawLine   string      `json:"raw_line"`
}

// InfoEventData holds parsed hardware/firmware details from the `info` command.
type InfoEventData struct {
	Version       string `json:"version"`
	Major         int    `json:"major"`
	Minor         int    `json:"minor"`
	Patch         int    `json:"patch"`
	Hardware      string `json:"hardware"`
	HasSD         bool   `json:"has_sd"`
	HasGPS        bool   `json:"has_gps"`
	BatteryStatus string `json:"battery_status"`
}

// HelpEventData holds discovered command verbs from the `help` menu.
type HelpEventData struct {
	Commands []string `json:"commands"`
}

// APScanEventData represents an Access Point observation.
type APScanEventData struct {
	Index      int    `json:"index"`
	SSID       string `json:"ssid"`
	BSSID      string `json:"bssid"`
	RSSI       int    `json:"rssi"`
	Channel    int    `json:"channel"`
	Encryption string `json:"encryption"`
}

// StationScanEventData represents a Client Station observation.
type StationScanEventData struct {
	Index          int    `json:"index"`
	MAC            string `json:"mac"`
	RSSI           int    `json:"rssi"`
	AssociatedBSSID string `json:"associated_bssid"`
}

// SniffDeauthEventData represents an observed deauthentication frame.
type SniffDeauthEventData struct {
	SourceMAC string `json:"source_mac"`
	DestMAC   string `json:"dest_mac"`
	BSSID     string `json:"bssid"`
	Reason    string `json:"reason"`
}

// SniffPMKIDEventData represents a captured PMKID key exchange.
type SniffPMKIDEventData struct {
	BSSID     string `json:"bssid"`
	ClientMAC string `json:"client_mac"`
	PMKID     string `json:"pmkid"`
}

var (
	ansiRegex    = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	apLineRegex  = regexp.MustCompile(`^AP:\s*(\d+)\s*\|\s*(.*?)\s*\|\s*([0-9a-fA-F:]{17})\s*\|\s*(-?\d+)\s*\|\s*(\d+)\s*\|\s*(.*)$`)
	staLineRegex = regexp.MustCompile(`^STA:\s*(\d+)\s*\|\s*([0-9a-fA-F:]{17})\s*\|\s*(-?\d+)\s*\|\s*(.*)$`)
	helpCmdRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9_\-]+)\s+-\s+(.*)$`)
	deauthRegex  = regexp.MustCompile(`(?i)deauth.*?([0-9a-fA-F:]{17}).*?([0-9a-fA-F:]{17})`)
	pmkidRegex   = regexp.MustCompile(`(?i)pmkid.*?([0-9a-fA-F:]{17}).*?([0-9a-fA-F:]{17})`)
)

// StripANSI removes terminal control sequences from input text.
func StripANSI(line string) string {
	return ansiRegex.ReplaceAllString(line, "")
}

// WireParser maintains parsing state across serial stream lines.
type WireParser struct {
	inInfoBlock bool
	currentInfo InfoEventData
	inHelpBlock bool
	helpCmds    []string
}

// NewWireParser creates a fresh WireParser.
func NewWireParser() *WireParser {
	return &WireParser{
		helpCmds: make([]string, 0),
	}
}

// ParseLine processes a single line from the serial connection and returns a parsed event.
func (p *WireParser) ParseLine(raw string) *ParsedEvent {
	clean := strings.TrimSpace(StripANSI(raw))
	if clean == "" {
		return nil
	}

	now := time.Now()

	// Check for start/end of info block
	if strings.Contains(clean, "ESP32 Marauder") {
		p.inInfoBlock = true
		p.currentInfo = InfoEventData{}

		// Check for inline version like "ESP32 Marauder v0.13.4"
		if vIdx := strings.Index(clean, "v"); vIdx != -1 {
			vStr := strings.Fields(clean[vIdx:])[0]
			p.currentInfo.Version = vStr
			p.currentInfo.Major, p.currentInfo.Minor, p.currentInfo.Patch, _ = ParseVersion(vStr)
		}
		return &ParsedEvent{
			Type:      EventRawLine,
			Timestamp: now,
			RawLine:   clean,
		}
	}

	if p.inInfoBlock {
		if strings.HasPrefix(clean, "Version:") {
			vStr := strings.TrimSpace(strings.TrimPrefix(clean, "Version:"))
			p.currentInfo.Version = vStr
			p.currentInfo.Major, p.currentInfo.Minor, p.currentInfo.Patch, _ = ParseVersion(vStr)
		} else if strings.HasPrefix(clean, "Hardware:") {
			p.currentInfo.Hardware = strings.TrimSpace(strings.TrimPrefix(clean, "Hardware:"))
		} else if strings.HasPrefix(clean, "SD Card:") || strings.HasPrefix(clean, "SD:") {
			p.currentInfo.HasSD = strings.Contains(strings.ToLower(clean), "connected")
		} else if strings.HasPrefix(clean, "GPS:") {
			p.currentInfo.HasGPS = strings.Contains(strings.ToLower(clean), "connected")
		} else if strings.HasPrefix(clean, "Battery:") {
			p.currentInfo.BatteryStatus = strings.TrimSpace(strings.TrimPrefix(clean, "Battery:"))
		} else if strings.HasPrefix(clean, "====") || strings.HasPrefix(clean, "----") {
			// End of info block
			p.inInfoBlock = false
			if p.currentInfo.Version != "" || p.currentInfo.Hardware != "" {
				return &ParsedEvent{
					Type:      EventInfo,
					Timestamp: now,
					Data:      p.currentInfo,
					RawLine:   clean,
				}
			}
		}
	}

	// Check for help menu
	if strings.EqualFold(clean, "Commands:") {
		p.inHelpBlock = true
		p.helpCmds = make([]string, 0)
		return &ParsedEvent{
			Type:      EventRawLine,
			Timestamp: now,
			RawLine:   clean,
		}
	}

	if p.inHelpBlock {
		matches := helpCmdRegex.FindStringSubmatch(clean)
		if len(matches) >= 3 {
			p.helpCmds = append(p.helpCmds, strings.ToLower(matches[1]))
			return &ParsedEvent{
				Type:      EventRawLine,
				Timestamp: now,
				RawLine:   clean,
			}
		} else if len(p.helpCmds) > 0 {
			// End of help block
			cmds := p.helpCmds
			p.inHelpBlock = false
			p.helpCmds = make([]string, 0)
			return &ParsedEvent{
				Type:      EventHelp,
				Timestamp: now,
				Data:      HelpEventData{Commands: cmds},
				RawLine:   clean,
			}
		}
	}

	// Check for AP scan line: AP: 0 | Lab_Secure_5G | 00:11:22:33:44:55 | -45 | 36 | WPA2
	if matches := apLineRegex.FindStringSubmatch(clean); len(matches) == 7 {
		idx, _ := strconv.Atoi(matches[1])
		rssi, _ := strconv.Atoi(matches[4])
		ch, _ := strconv.Atoi(matches[5])
		return &ParsedEvent{
			Type:      EventAP,
			Timestamp: now,
			Data: APScanEventData{
				Index:      idx,
				SSID:       matches[2],
				BSSID:      strings.ToLower(matches[3]),
				RSSI:       rssi,
				Channel:    ch,
				Encryption: matches[6],
			},
			RawLine: clean,
		}
	}

	// Check for STA scan line: STA: 0 | 12:34:56:78:9a:bc | -55 | 00:11:22:33:44:55
	if matches := staLineRegex.FindStringSubmatch(clean); len(matches) == 5 {
		idx, _ := strconv.Atoi(matches[1])
		rssi, _ := strconv.Atoi(matches[3])
		return &ParsedEvent{
			Type:      EventSTA,
			Timestamp: now,
			Data: StationScanEventData{
				Index:           idx,
				MAC:             strings.ToLower(matches[2]),
				RSSI:            rssi,
				AssociatedBSSID: strings.ToLower(matches[4]),
			},
			RawLine: clean,
		}
	}

	// Check for Deauth events
	if matches := deauthRegex.FindStringSubmatch(clean); len(matches) >= 3 {
		return &ParsedEvent{
			Type:      EventDeauth,
			Timestamp: now,
			Data: SniffDeauthEventData{
				SourceMAC: strings.ToLower(matches[1]),
				DestMAC:   strings.ToLower(matches[2]),
				BSSID:     strings.ToLower(matches[1]),
				Reason:    clean,
			},
			RawLine: clean,
		}
	}

	// Check for PMKID events
	if matches := pmkidRegex.FindStringSubmatch(clean); len(matches) >= 3 {
		return &ParsedEvent{
			Type:      EventPMKID,
			Timestamp: now,
			Data: SniffPMKIDEventData{
				BSSID:     strings.ToLower(matches[1]),
				ClientMAC: strings.ToLower(matches[2]),
				PMKID:     clean,
			},
			RawLine: clean,
		}
	}

	// Default fallback to raw line
	return &ParsedEvent{
		Type:      EventRawLine,
		Timestamp: now,
		RawLine:   clean,
	}
}
