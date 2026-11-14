package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CaptureFileMetadata describes an imported or discovered PCAP capture file.
type CaptureFileMetadata struct {
	Path        string    `json:"path"`
	FileName    string    `json:"file_name"`
	SizeBytes   int64     `json:"size_bytes"`
	ModTime     time.Time `json:"mod_time"`
	PacketCount int       `json:"packet_count"`
	DurationSec float64   `json:"duration_sec"`
}

// DiscoverSDCaptures scans a mounted directory (e.g. SD card volume) for PCAP files.
func DiscoverSDCaptures(mountDir string) ([]CaptureFileMetadata, error) {
	cleanPath := filepath.Clean(mountDir)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("unable to access directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("specified path is not a directory")
	}

	results := make([]CaptureFileMetadata, 0)
	err = filepath.Walk(cleanPath, func(path string, f os.FileInfo, walkErr error) error {
		if walkErr != nil || f.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".pcap" || ext == ".cap" {
			results = append(results, CaptureFileMetadata{
				Path:      path,
				FileName:  filepath.Base(path),
				SizeBytes: f.Size(),
				ModTime:   f.ModTime(),
			})
		}
		return nil
	})

	return results, err
}
