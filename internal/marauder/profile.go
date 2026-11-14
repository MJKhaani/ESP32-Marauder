package marauder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FirmwareProfile models the identified Marauder firmware capabilities and runtime attributes.
type FirmwareProfile struct {
	RawVersion        string          `json:"raw_version"`
	Major             int             `json:"major"`
	Minor             int             `json:"minor"`
	Patch             int             `json:"patch"`
	HardwareName      string          `json:"hardware_name"`
	HasSD             bool            `json:"has_sd"`
	HasGPS            bool            `json:"has_gps"`
	HasBattery        bool            `json:"has_battery"`
	BatteryStatus     string          `json:"battery_status"`
	SupportedCommands map[string]bool `json:"supported_commands"`
}

var versionRegex = regexp.MustCompile(`v?(\d+)\.(\d+)(?:\.(\d+))?`)

// ParseVersion extracts semver components from a firmware version string.
func ParseVersion(v string) (major, minor, patch int, err error) {
	matches := versionRegex.FindStringSubmatch(v)
	if len(matches) < 3 {
		return 0, 0, 0, fmt.Errorf("invalid version string: %s", v)
	}

	major, _ = strconv.Atoi(matches[1])
	minor, _ = strconv.Atoi(matches[2])
	if len(matches) > 3 && matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}
	return major, minor, patch, nil
}

// IsModern returns true if the firmware is v1.20.0 or higher.
func (p *FirmwareProfile) IsModern() bool {
	if p.Major > 1 {
		return true
	}
	if p.Major == 1 && p.Minor >= 20 {
		return true
	}
	return false
}

// HasCommand checks if a specific CLI command is supported.
func (p *FirmwareProfile) HasCommand(cmd string) bool {
	if p.SupportedCommands == nil {
		return false
	}
	return p.SupportedCommands[strings.ToLower(cmd)]
}

// PreferredScanCommand returns "scanall" for modern firmware or "scanap" for older builds.
func (p *FirmwareProfile) PreferredScanCommand() string {
	if p.HasCommand("scanall") || p.IsModern() {
		return "scanall"
	}
	return "scanap"
}
