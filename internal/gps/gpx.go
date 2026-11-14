package gps

import (
	"encoding/xml"
	"io"
	"time"
)

// GPXRoot models the top-level GPX XML structure.
type GPXRoot struct {
	XMLName xml.Name  `xml:"gpx"`
	Version string    `xml:"version,attr"`
	Creator string    `xml:"creator,attr"`
	Trk     GPXTrack  `xml:"trk"`
}

// GPXTrack models a track container.
type GPXTrack struct {
	Name   string      `xml:"name"`
	TrkSeg GPXTrackSeg `xml:"trkseg"`
}

// GPXTrackSeg contains coordinate points.
type GPXTrackSeg struct {
	Points []GPXPoint `xml:"trkpt"`
}

// GPXPoint models a single trackpoint.
type GPXPoint struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Ele  float64 `xml:"ele,omitempty"`
	Time string  `xml:"time,omitempty"`
}

// ExportGPX writes a collection of GPS coordinates into standard GPX format.
func ExportGPX(w io.Writer, trackName string, coords []GPSCoordinate) error {
	pts := make([]GPXPoint, len(coords))
	for i, c := range coords {
		pts[i] = GPXPoint{
			Lat:  c.Latitude,
			Lon:  c.Longitude,
			Ele:  c.Altitude,
			Time: c.Timestamp.Format(time.RFC3339),
		}
	}

	root := GPXRoot{
		Version: "1.1",
		Creator: "Marauder Studio",
		Trk: GPXTrack{
			Name: trackName,
			TrkSeg: GPXTrackSeg{
				Points: pts,
			},
		},
	}

	io.WriteString(w, xml.Header)
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(root)
}

// SanitizeCoord truncates coordinate precision for privacy preservation.
func SanitizeCoord(lat, lon float64) (float64, float64) {
	// Truncate to ~1km precision (2 decimal places) for anonymized exports
	sanLat := float64(int(lat*100)) / 100.0
	sanLon := float64(int(lon*100)) / 100.0
	return sanLat, sanLon
}
