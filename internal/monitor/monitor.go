package monitor

import (
	"sync"
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
)

// AlertSeverity defines the risk level of an observation.
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// AnomalyAlert represents an observed defensive indicator.
type AnomalyAlert struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Severity  AlertSeverity `json:"severity"`
	Category  string        `json:"category"` // "deauth_flood", "evil_twin", "beacon_anomaly"
	Title     string        `json:"title"`
	Detail    string        `json:"detail"`
	SourceMAC string        `json:"source_mac,omitempty"`
	TargetMAC string        `json:"target_mac,omitempty"`
	Channel   int           `json:"channel,omitempty"`
}

// ChannelStats captures activity metrics for a specific Wi-Fi channel.
type ChannelStats struct {
	Channel   int `json:"channel"`
	APCount   int `json:"ap_count"`
	BeaconRate int `json:"beacon_rate"`
	DeauthRate int `json:"deauth_rate"`
}

// MonitorEngine analyzes live observations and raises defensive alerts without transmitting.
type MonitorEngine struct {
	mu           sync.RWMutex
	alerts       []AnomalyAlert
	channelStats map[int]*ChannelStats
	knownSSIDs   map[string]string // SSID -> First seen BSSID

	// Deauth rate tracking
	deauthHistory []time.Time
}

// NewMonitorEngine returns a new defensive MonitorEngine.
func NewMonitorEngine() *MonitorEngine {
	me := &MonitorEngine{
		alerts:        make([]AnomalyAlert, 0),
		channelStats:  make(map[int]*ChannelStats),
		knownSSIDs:    make(map[string]string),
		deauthHistory: make([]time.Time, 0),
	}
	for ch := 1; ch <= 14; ch++ {
		me.channelStats[ch] = &ChannelStats{Channel: ch}
	}
	return me
}

// ProcessAP analyzes an AP observation for potential Evil Twin anomalies.
func (m *MonitorEngine) ProcessAP(ap marauder.APScanEventData) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update channel stats
	stats, ok := m.channelStats[ap.Channel]
	if !ok {
		stats = &ChannelStats{Channel: ap.Channel}
		m.channelStats[ap.Channel] = stats
	}
	stats.APCount++

	if ap.SSID == "" {
		return
	}

	// Check for Evil Twin (same SSID broadcast from different BSSID)
	firstBSSID, exists := m.knownSSIDs[ap.SSID]
	if exists && firstBSSID != ap.BSSID {
		m.alerts = append(m.alerts, AnomalyAlert{
			ID:        time.Now().Format(time.RFC3339Nano),
			Timestamp: time.Now(),
			Severity:  SeverityWarning,
			Category:  "evil_twin",
			Title:     "Potential Evil Twin / Rogue AP Detected",
			Detail:    "SSID '" + ap.SSID + "' observed from new BSSID " + ap.BSSID + " (first seen at " + firstBSSID + ")",
			SourceMAC: ap.BSSID,
			Channel:   ap.Channel,
		})
	} else if !exists {
		m.knownSSIDs[ap.SSID] = ap.BSSID
	}
}

// ProcessDeauth tracks deauthentication frame rate and flags bursts.
func (m *MonitorEngine) ProcessDeauth(deauth marauder.SniffDeauthEventData) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.deauthHistory = append(m.deauthHistory, now)

	// Prune history older than 5 seconds
	cutoff := now.Add(-5 * time.Second)
	valid := make([]time.Time, 0, len(m.deauthHistory))
	for _, t := range m.deauthHistory {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	m.deauthHistory = valid

	// Alert if > 10 deauths seen in 5s window
	if len(m.deauthHistory) >= 10 {
		m.alerts = append(m.alerts, AnomalyAlert{
			ID:        now.Format(time.RFC3339Nano),
			Timestamp: now,
			Severity:  SeverityCritical,
			Category:  "deauth_flood",
			Title:     "Deauthentication Burst Anomaly Detected",
			Detail:    "Observed >10 deauthentication management frames in 5s window",
			SourceMAC: deauth.SourceMAC,
			TargetMAC: deauth.DestMAC,
		})
	}
}

// GetAlerts returns recorded anomaly alerts.
func (m *MonitorEngine) GetAlerts() []AnomalyAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]AnomalyAlert, len(m.alerts))
	copy(res, m.alerts)
	return res
}

// GetChannelStats returns channel distribution statistics.
func (m *MonitorEngine) GetChannelStats() []*ChannelStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*ChannelStats, 0, len(m.channelStats))
	for _, cs := range m.channelStats {
		res = append(res, cs)
	}
	return res
}
