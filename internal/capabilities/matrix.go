package capabilities

// FeatureStatus defines the exact readiness state of a feature on a specific connected device.
type FeatureStatus string

const (
	StatusAvailable          FeatureStatus = "Available"
	StatusRequiresHardware   FeatureStatus = "Requires hardware"
	StatusFirmwareTooOld     FeatureStatus = "Firmware too old"
	StatusNotExposedCLI      FeatureStatus = "Not exposed through CLI"
	StatusUnsupported        FeatureStatus = "Unsupported"
	StatusUnknown            FeatureStatus = "Unknown / needs verification"
	StatusBusy               FeatureStatus = "Temporarily busy"
)

// FeatureCapability models an individual capability's status, evidence, and diagnostics.
type FeatureCapability struct {
	ID                 string        `json:"id"`
	Label              string        `json:"label"`
	Category           string        `json:"category"`
	RiskLevel          string        `json:"risk_level"` // "passive", "observational", "high_risk"
	Status             FeatureStatus `json:"status"`
	HardwareSupported  bool          `json:"hardware_supported"`
	FirmwareReported   bool          `json:"firmware_reported"`
	CLIExposed         bool          `json:"cli_exposed"`
	DesktopImplemented bool          `json:"desktop_implemented"`
	Reason             string        `json:"reason"`
	Evidence           []string      `json:"evidence"`
}

// CapabilityMatrix stores all evaluated capabilities for a device.
type CapabilityMatrix struct {
	DeviceID     string                        `json:"device_id"`
	HardwareName string                        `json:"hardware_name"`
	FirmwareVer  string                        `json:"firmware_version"`
	Capabilities map[string]FeatureCapability  `json:"capabilities"`
}

// NewCapabilityMatrix initializes an empty CapabilityMatrix.
func NewCapabilityMatrix(deviceID, hardwareName, firmwareVer string) *CapabilityMatrix {
	return &CapabilityMatrix{
		DeviceID:     deviceID,
		HardwareName: hardwareName,
		FirmwareVer:  firmwareVer,
		Capabilities: make(map[string]FeatureCapability),
	}
}

// Add sets or updates a capability in the matrix.
func (m *CapabilityMatrix) Add(cap FeatureCapability) {
	m.Capabilities[cap.ID] = cap
}

// Get retrieves a capability by ID.
func (m *CapabilityMatrix) Get(id string) (FeatureCapability, bool) {
	c, ok := m.Capabilities[id]
	return c, ok
}

// IsAvailable returns true if the capability status is StatusAvailable.
func (m *CapabilityMatrix) IsAvailable(id string) bool {
	c, ok := m.Capabilities[id]
	return ok && c.Status == StatusAvailable
}
