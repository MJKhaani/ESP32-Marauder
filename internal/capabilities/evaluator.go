package capabilities

import (
	"fmt"
	"strings"

	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
)

// Evaluator evaluates hardware evidence and firmware discovery into a CapabilityMatrix.
type Evaluator struct{}

// NewEvaluator creates a new Evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate constructs the complete capability matrix for a connected device.
func (e *Evaluator) Evaluate(deviceID string, profile *marauder.FirmwareProfile) *CapabilityMatrix {
	matrix := NewCapabilityMatrix(deviceID, profile.HardwareName, profile.RawVersion)

	// 1. Wi-Fi AP & Station Survey
	surveyCap := FeatureCapability{
		ID:                 "wifi_survey",
		Label:              "Wi-Fi AP & Station Survey",
		Category:           "Survey",
		RiskLevel:          "passive",
		HardwareSupported:  true,
		FirmwareReported:   true,
		CLIExposed:         profile.HasCommand("scanall") || profile.HasCommand("scanap"),
		DesktopImplemented: true,
	}
	if surveyCap.CLIExposed {
		surveyCap.Status = StatusAvailable
		surveyCap.Evidence = []string{fmt.Sprintf("CLI command '%s' discovered in help menu", profile.PreferredScanCommand())}
	} else {
		surveyCap.Status = StatusFirmwareTooOld
		surveyCap.Reason = "Neither 'scanall' nor 'scanap' exposed by connected firmware."
		surveyCap.Evidence = []string{"Missing scan commands in CLI help"}
	}
	matrix.Add(surveyCap)

	// 2. Channel Monitoring & Anomaly Detection
	channelCap := FeatureCapability{
		ID:                 "channel_monitor",
		Label:              "Defensive Channel & Beacon Monitor",
		Category:           "Defensive",
		RiskLevel:          "passive",
		HardwareSupported:  true,
		FirmwareReported:   true,
		CLIExposed:         profile.HasCommand("sniffbeacon") || profile.HasCommand("channel"),
		DesktopImplemented: true,
	}
	if channelCap.CLIExposed {
		channelCap.Status = StatusAvailable
		channelCap.Evidence = []string{"Beacon sniffer and channel query commands supported"}
	} else {
		channelCap.Status = StatusFirmwareTooOld
		channelCap.Reason = "Channel or beacon sniffing verbs not found in running firmware"
	}
	matrix.Add(channelCap)

	// 3. PMKID Passive Sniffing
	pmkidCap := FeatureCapability{
		ID:                 "sniff_pmkid",
		Label:              "Passive PMKID Key Exchange Sniffer",
		Category:           "Capture",
		RiskLevel:          "observational",
		HardwareSupported:  true,
		FirmwareReported:   true,
		CLIExposed:         profile.HasCommand("sniffpmkid"),
		DesktopImplemented: true,
	}
	if pmkidCap.CLIExposed {
		pmkidCap.Status = StatusAvailable
		pmkidCap.Evidence = []string{"'sniffpmkid' exposed in help"}
	} else {
		pmkidCap.Status = StatusNotExposedCLI
		pmkidCap.Reason = "Firmware build lacks sniffpmkid CLI command"
	}
	matrix.Add(pmkidCap)

	// 4. Raw PCAP Capture to SD Card
	rawCap := FeatureCapability{
		ID:                 "sniff_raw_sd",
		Label:              "Raw 802.11 Capture to MicroSD",
		Category:           "Capture",
		RiskLevel:          "observational",
		HardwareSupported:  profile.HasSD,
		FirmwareReported:   profile.HasCommand("sniffraw"),
		CLIExposed:         profile.HasCommand("sniffraw"),
		DesktopImplemented: true,
	}
	if !profile.HasSD {
		rawCap.Status = StatusRequiresHardware
		rawCap.Reason = "MicroSD card not detected or not attached to device"
		rawCap.Evidence = []string{"'SD Card: Not connected' in info output"}
	} else if !rawCap.CLIExposed {
		rawCap.Status = StatusNotExposedCLI
		rawCap.Reason = "'sniffraw' not exposed in CLI"
	} else {
		rawCap.Status = StatusAvailable
		rawCap.Evidence = []string{"SD card connected and 'sniffraw' command available"}
	}
	matrix.Add(rawCap)

	// 5. Bluetooth / BLE Observation
	bleCap := FeatureCapability{
		ID:                 "ble_observation",
		Label:              "Bluetooth Low Energy Device Observation",
		Category:           "Bluetooth",
		RiskLevel:          "passive",
		HardwareSupported:  !strings.Contains(strings.ToLower(profile.HardwareName), "s2"), // ESP32-S2 lacks BLE
		FirmwareReported:   profile.HasCommand("sniffbt"),
		CLIExposed:         profile.HasCommand("sniffbt"),
		DesktopImplemented: true,
	}
	if !bleCap.HardwareSupported {
		bleCap.Status = StatusUnsupported
		bleCap.Reason = "Connected chip/board architecture does not possess a Bluetooth radio (e.g. ESP32-S2)"
		bleCap.Evidence = []string{fmt.Sprintf("Hardware identified as %s", profile.HardwareName)}
	} else if bleCap.CLIExposed {
		bleCap.Status = StatusAvailable
		bleCap.Evidence = []string{"Bluetooth hardware verified and 'sniffbt' command verified"}
	} else {
		bleCap.Status = StatusNotExposedCLI
		bleCap.Reason = "'sniffbt' not exposed in CLI"
	}
	matrix.Add(bleCap)

	// 6. GPS Telemetry & Wardrive Field Tracking
	gpsCap := FeatureCapability{
		ID:                 "gps_telemetry",
		Label:              "GPS Telemetry & Field Wardrive Tracking",
		Category:           "Field Artifacts",
		RiskLevel:          "passive",
		HardwareSupported:  profile.HasGPS,
		FirmwareReported:   profile.HasGPS,
		CLIExposed:         true,
		DesktopImplemented: true,
	}
	if profile.HasGPS {
		gpsCap.Status = StatusAvailable
		gpsCap.Evidence = []string{"GPS receiver connected and providing valid NMEA sentences"}
	} else {
		gpsCap.Status = StatusRequiresHardware
		gpsCap.Reason = "External or onboard GPS module not detected"
		gpsCap.Evidence = []string{"'GPS: Not connected' in info output"}
	}
	matrix.Add(gpsCap)

	// 7. Authorized Active Lab Tests (High Risk, isolated)
	labCap := FeatureCapability{
		ID:                 "authorized_lab",
		Label:              "Authorized Active RF Lab Evaluation",
		Category:           "Active Lab",
		RiskLevel:          "high_risk",
		HardwareSupported:  true,
		FirmwareReported:   profile.HasCommand("attack"),
		CLIExposed:         profile.HasCommand("attack"),
		DesktopImplemented: true,
	}
	if labCap.CLIExposed {
		labCap.Status = StatusAvailable
		labCap.Evidence = []string{"'attack' command verified in firmware CLI"}
		labCap.Reason = "Requires explicit lab ownership authorization, allowlisted target BSSID, and duration limits before activation"
	} else {
		labCap.Status = StatusNotExposedCLI
		labCap.Reason = "Active attack testing verbs disabled or absent in this firmware build"
	}
	matrix.Add(labCap)

	// 8. Firmware Updates
	flashCap := FeatureCapability{
		ID:                 "firmware_flashing",
		Label:              "Hardware-Matched Firmware Management",
		Category:           "Firmware",
		RiskLevel:          "high_risk",
		HardwareSupported:  true,
		FirmwareReported:   true,
		CLIExposed:         true,
		DesktopImplemented: true,
		Status:             StatusAvailable,
		Evidence:           []string{"Hardware identified; bootloader/ROM download mode supported"},
	}
	matrix.Add(flashCap)

	return matrix
}
