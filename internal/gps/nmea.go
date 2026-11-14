package gps

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FixStatus describes the GPS satellite lock state.
type FixStatus string

const (
	FixNone FixStatus = "none"
	Fix2D   FixStatus = "2D"
	Fix3D   FixStatus = "3D"
	FixDGPS FixStatus = "DGPS"
)

// GPSCoordinate represents a geographic fix.
type GPSCoordinate struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  float64   `json:"altitude"`
	SpeedKnots float64  `json:"speed_knots"`
	Fix       FixStatus `json:"fix"`
	Satellites int      `json:"satellites"`
	Timestamp time.Time `json:"timestamp"`
}

// NMEAParser decodes standard $GPRMC and $GPGGA sentences.
type NMEAParser struct{}

// NewNMEAParser creates a new NMEAParser.
func NewNMEAParser() *NMEAParser {
	return &NMEAParser{}
}

// ParseSentence decodes an NMEA string into a GPSCoordinate.
func (p *NMEAParser) ParseSentence(line string) (*GPSCoordinate, error) {
	clean := strings.TrimSpace(line)
	if !strings.HasPrefix(clean, "$") {
		return nil, errors.New("not an NMEA sentence")
	}

	parts := strings.Split(clean, ",")
	if len(parts) < 7 {
		return nil, errors.New("sentence too short")
	}

	sentenceType := parts[0]
	coord := &GPSCoordinate{Fix: FixNone, Timestamp: time.Now()}

	switch sentenceType {
	case "$GPRMC", "$GNRMC":
		// $GPRMC,hhmmss.ss,A,llll.ll,a,yyyyy.yy,a,x.x,x.x,ddmmyy,,,a*hh
		if len(parts) < 10 {
			return nil, errors.New("invalid RMC length")
		}
		status := parts[2]
		if status != "A" {
			return coord, nil // Void / no fix
		}
		lat, err := parseNMEACoord(parts[3], parts[4])
		if err != nil {
			return nil, err
		}
		lon, err := parseNMEACoord(parts[5], parts[6])
		if err != nil {
			return nil, err
		}
		spd, _ := strconv.ParseFloat(parts[7], 64)

		coord.Latitude = lat
		coord.Longitude = lon
		coord.SpeedKnots = spd
		coord.Fix = Fix3D
		return coord, nil

	case "$GPGGA", "$GNGGA":
		// $GPGGA,hhmmss.ss,llll.ll,a,yyyyy.yy,a,x,xx,x.x,x.x,M,x.x,M,x.x,xxxx*hh
		if len(parts) < 10 {
			return nil, errors.New("invalid GGA length")
		}
		qual := parts[6]
		if qual == "0" {
			return coord, nil
		}
		lat, err := parseNMEACoord(parts[2], parts[3])
		if err != nil {
			return nil, err
		}
		lon, err := parseNMEACoord(parts[4], parts[5])
		if err != nil {
			return nil, err
		}
		sats, _ := strconv.Atoi(parts[7])
		alt, _ := strconv.ParseFloat(parts[9], 64)

		coord.Latitude = lat
		coord.Longitude = lon
		coord.Altitude = alt
		coord.Satellites = sats
		coord.Fix = Fix3D
		return coord, nil
	}

	return nil, fmt.Errorf("unsupported NMEA sentence: %s", sentenceType)
}

func parseNMEACoord(raw, hemisphere string) (float64, error) {
	if raw == "" || hemisphere == "" {
		return 0, errors.New("empty coordinate fields")
	}
	dotIdx := strings.Index(raw, ".")
	if dotIdx < 2 {
		return 0, errors.New("malformed coordinate format")
	}

	degStr := raw[:dotIdx-2]
	minStr := raw[dotIdx-2:]

	deg, err := strconv.ParseFloat(degStr, 64)
	if err != nil {
		return 0, err
	}
	min, err := strconv.ParseFloat(minStr, 64)
	if err != nil {
		return 0, err
	}

	val := deg + (min / 60.0)
	if hemisphere == "S" || hemisphere == "W" {
		val = -val
	}
	return val, nil
}
