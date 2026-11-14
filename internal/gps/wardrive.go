package gps

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"
)

// WardriveEntry represents an Access Point logged during a wardrive session.
type WardriveEntry struct {
	BSSID     string    `json:"bssid"`
	SSID      string    `json:"ssid"`
	AuthMode  string    `json:"auth_mode"`
	FirstSeen time.Time `json:"first_seen"`
	Channel   int       `json:"channel"`
	RSSI      int       `json:"rssi"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  float64   `json:"altitude"`
	Accuracy  float64   `json:"accuracy"`
	Type      string    `json:"type"` // "WIFI", "BLE"
}

// ParseWardriveCSV parses a WiGLE/Marauder CSV file.
func ParseWardriveCSV(r io.Reader) ([]WardriveEntry, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 // Allow variable fields

	entries := make([]WardriveEntry, 0)
	isHeader := true

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if len(record) == 0 {
			continue
		}

		// Skip WiGLE pre-header comments or metadata
		if strings.HasPrefix(record[0], "WigleWifi") || strings.HasPrefix(record[0], "#") {
			continue
		}

		if isHeader {
			isHeader = false
			continue
		}

		if len(record) < 11 {
			continue
		}

		// WiGLE format: MAC,SSID,AuthMode,FirstSeen,Channel,RSSI,CurrentLatitude,CurrentLongitude,AltitudeMeters,AccuracyMeters,Type
		bssid := strings.ToLower(strings.TrimSpace(record[0]))
		ssid := record[1]
		auth := record[2]
		ch, _ := strconv.Atoi(record[4])
		rssi, _ := strconv.Atoi(record[5])
		lat, _ := strconv.ParseFloat(record[6], 64)
		lon, _ := strconv.ParseFloat(record[7], 64)
		alt, _ := strconv.ParseFloat(record[8], 64)
		acc, _ := strconv.ParseFloat(record[9], 64)
		entryType := "WIFI"
		if len(record) > 10 {
			entryType = record[10]
		}

		entries = append(entries, WardriveEntry{
			BSSID:     bssid,
			SSID:      ssid,
			AuthMode:  auth,
			Channel:   ch,
			RSSI:      rssi,
			Latitude:  lat,
			Longitude: lon,
			Altitude:  alt,
			Accuracy:  acc,
			Type:      entryType,
			FirstSeen: time.Now(),
		})
	}

	return entries, nil
}
