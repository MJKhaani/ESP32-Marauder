package survey

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
)

// AccessPoint models an observed 802.11 Access Point.
type AccessPoint struct {
	BSSID       string    `json:"bssid"`
	SSID        string    `json:"ssid"`
	Channel     int       `json:"channel"`
	Band        string    `json:"band"` // "2.4 GHz" or "5 GHz"
	RSSI        int       `json:"rssi"`
	Encryption  string    `json:"encryption"`
	Vendor      string    `json:"vendor,omitempty"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	FrameCount  int       `json:"frame_count"`
	IsAllowlisted bool    `json:"is_allowlisted"`
}

// Station models an observed 802.11 client device.
type Station struct {
	MAC             string    `json:"mac"`
	RSSI            int       `json:"rssi"`
	AssociatedBSSID string    `json:"associated_bssid"`
	Vendor          string    `json:"vendor,omitempty"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
}

// SurveyStore manages in-memory survey findings with thread-safety and bounded storage.
type SurveyStore struct {
	mu        sync.RWMutex
	aps       map[string]*AccessPoint
	stations  map[string]*Station
	allowlist map[string]bool // normalized lowercase BSSIDs / MACs
}

// NewSurveyStore creates a new SurveyStore.
func NewSurveyStore() *SurveyStore {
	return &SurveyStore{
		aps:       make(map[string]*AccessPoint),
		stations:  make(map[string]*Station),
		allowlist: make(map[string]bool),
	}
}

// SetAllowlist updates the active lab target allowlist.
func (s *SurveyStore) SetAllowlist(targets []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.allowlist = make(map[string]bool)
	for _, t := range targets {
		norm := strings.ToLower(strings.TrimSpace(t))
		if norm != "" {
			s.allowlist[norm] = true
		}
	}

	// Update existing AP allowlist flags
	for _, ap := range s.aps {
		ap.IsAllowlisted = s.allowlist[ap.BSSID]
	}
}

// RecordAP registers or updates an Access Point observation.
func (s *SurveyStore) RecordAP(apData marauder.APScanEventData) *AccessPoint {
	s.mu.Lock()
	defer s.mu.Unlock()

	bssid := strings.ToLower(apData.BSSID)
	now := time.Now()

	band := "2.4 GHz"
	if apData.Channel > 14 {
		band = "5 GHz"
	}

	ap, exists := s.aps[bssid]
	if !exists {
		ap = &AccessPoint{
			BSSID:         bssid,
			SSID:          apData.SSID,
			Channel:       apData.Channel,
			Band:          band,
			RSSI:          apData.RSSI,
			Encryption:    apData.Encryption,
			Vendor:        ResolveOUI(bssid),
			FirstSeen:     now,
			LastSeen:      now,
			FrameCount:    1,
			IsAllowlisted: s.allowlist[bssid],
		}
		s.aps[bssid] = ap
	} else {
		ap.LastSeen = now
		ap.FrameCount++
		ap.RSSI = apData.RSSI
		if apData.SSID != "" && ap.SSID == "" {
			ap.SSID = apData.SSID
		}
		if apData.Encryption != "" && ap.Encryption == "" {
			ap.Encryption = apData.Encryption
		}
	}
	return ap
}

// RecordStation registers or updates a client station observation.
func (s *SurveyStore) RecordStation(staData marauder.StationScanEventData) *Station {
	s.mu.Lock()
	defer s.mu.Unlock()

	mac := strings.ToLower(staData.MAC)
	assoc := strings.ToLower(staData.AssociatedBSSID)
	now := time.Now()

	sta, exists := s.stations[mac]
	if !exists {
		sta = &Station{
			MAC:             mac,
			RSSI:            staData.RSSI,
			AssociatedBSSID: assoc,
			Vendor:          ResolveOUI(mac),
			FirstSeen:       now,
			LastSeen:        now,
		}
		s.stations[mac] = sta
	} else {
		sta.LastSeen = now
		sta.RSSI = staData.RSSI
		if assoc != "" && assoc != "unknown" {
			sta.AssociatedBSSID = assoc
		}
	}
	return sta
}

// GetAPs returns all recorded Access Points, sorted by signal strength.
func (s *SurveyStore) GetAPs(onlyAllowlisted bool) []*AccessPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*AccessPoint, 0, len(s.aps))
	for _, ap := range s.aps {
		if onlyAllowlisted && !ap.IsAllowlisted {
			continue
		}
		res = append(res, ap)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].RSSI > res[j].RSSI
	})
	return res
}

// GetStations returns all recorded stations.
func (s *SurveyStore) GetStations() []*Station {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*Station, 0, len(s.stations))
	for _, sta := range s.stations {
		res = append(res, sta)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].RSSI > res[j].RSSI
	})
	return res
}

// Clear resets the survey findings.
func (s *SurveyStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aps = make(map[string]*AccessPoint)
	s.stations = make(map[string]*Station)
}

// ExportJSON writes the survey results in JSON format.
func (s *SurveyStore) ExportJSON(w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := struct {
		Timestamp time.Time      `json:"timestamp"`
		APs       []*AccessPoint `json:"access_points"`
		Stations  []*Station     `json:"stations"`
	}{
		Timestamp: time.Now(),
		APs:       s.GetAPs(false),
		Stations:  s.GetStations(),
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// ExportCSV writes the AP survey results in CSV format.
func (s *SurveyStore) ExportCSV(w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"BSSID", "SSID", "Channel", "Band", "RSSI", "Encryption", "Vendor", "FirstSeen", "LastSeen", "Frames"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, ap := range s.GetAPs(false) {
		row := []string{
			ap.BSSID,
			ap.SSID,
			fmt.Sprintf("%d", ap.Channel),
			ap.Band,
			fmt.Sprintf("%d", ap.RSSI),
			ap.Encryption,
			ap.Vendor,
			ap.FirstSeen.Format(time.RFC3339),
			ap.LastSeen.Format(time.RFC3339),
			fmt.Sprintf("%d", ap.FrameCount),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// ResolveOUI extracts high-level vendor attribution from the first 3 octets of a MAC address.
func ResolveOUI(mac string) string {
	clean := strings.ReplaceAll(strings.ToLower(mac), ":", "")
	if len(clean) < 6 {
		return "Unknown"
	}
	prefix := clean[:6]
	switch prefix {
	case "001122":
		return "Cisco Systems"
	case "001a11", "0024d7", "0026c6":
		return "Intel Corporate"
	case "001788", "001e52":
		return "Philips / Signify"
	case "00259c", "002608":
		return "Apple, Inc."
	case "303a64", "240ac4", "246f28", "84f3eb":
		return "Espressif Inc."
	case "b827eb", "dca632", "e45f01":
		return "Raspberry Pi"
	default:
		return "Private / Other"
	}
}
