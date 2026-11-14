package firmware

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// BoardTarget defines the hardware specifications and partition geometry for a Marauder board.
type BoardTarget struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ChipFamily       string `json:"chip_family"`
	PCBRevision      string `json:"pcb_revision"`
	FlashSizeMB      int    `json:"flash_size_mb"`
	PartitionScheme  string `json:"partition_scheme"`
	AppOffset        string `json:"app_offset"`
	BootloaderOffset string `json:"bootloader_offset"`
	PartitionOffset  string `json:"partition_offset"`
	HasSD            bool   `json:"has_sd"`
	HasGPS           bool   `json:"has_gps"`
	HasBLE           bool   `json:"has_ble"`
	HasDisplay       bool   `json:"has_display"`
	DisplayType      string `json:"display_type"`
	BinarySuffix     string `json:"binary_suffix"`
}

// Catalog manages the supported hardware targets.
type Catalog struct {
	Targets map[string]BoardTarget
}

// LoadCatalog reads targets from a JSON file.
func LoadCatalog(filePath string) (*Catalog, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog file: %w", err)
	}

	var targets []BoardTarget
	if err := json.Unmarshal(data, &targets); err != nil {
		return nil, fmt.Errorf("failed to parse catalog JSON: %w", err)
	}

	cat := &Catalog{Targets: make(map[string]BoardTarget)}
	for _, t := range targets {
		cat.Targets[t.ID] = t
	}
	return cat, nil
}

// FindTargetByModel matches a detected hardware string or binary suffix to a known BoardTarget.
func (c *Catalog) FindTargetByModel(model string) (*BoardTarget, bool) {
	norm := strings.ToLower(strings.TrimSpace(model))
	for _, t := range c.Targets {
		if strings.EqualFold(t.ID, norm) || strings.EqualFold(t.Name, norm) ||
			strings.Contains(strings.ToLower(t.Name), norm) || strings.Contains(strings.ToLower(t.BinarySuffix), norm) {
			return &t, true
		}
	}
	return nil, false
}
