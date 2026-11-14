package app

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/MJKhaani/ESP32-Marauder/internal/ble"
	"github.com/MJKhaani/ESP32-Marauder/internal/device"
	"github.com/MJKhaani/ESP32-Marauder/internal/diagnostics"
	"github.com/MJKhaani/ESP32-Marauder/internal/firmware"
	"github.com/MJKhaani/ESP32-Marauder/internal/jobs"
	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
	"github.com/MJKhaani/ESP32-Marauder/internal/monitor"
	"github.com/MJKhaani/ESP32-Marauder/internal/policy"
	"github.com/MJKhaani/ESP32-Marauder/internal/reports"
	"github.com/MJKhaani/ESP32-Marauder/internal/serialtransport"
	"github.com/MJKhaani/ESP32-Marauder/internal/store"
	"github.com/MJKhaani/ESP32-Marauder/internal/survey"
)

// App is the central desktop application bridge.
type App struct {
	ctx           context.Context
	mu            sync.RWMutex
	DevManager    *device.Manager
	JobMgr        *jobs.JobManager
	SurveyStore   *survey.SurveyStore
	MonitorEngine *monitor.MonitorEngine
	BLEStore      *ble.BLEStore
	PolicyGate    *policy.PolicyGate
	FirmwareCat   *firmware.Catalog
	FirmwareVal   *firmware.Validator
	Flasher       *firmware.FlasherAdapter
	Store         *store.Store
	RecentErrors  []string
}

// NewApp creates a fully wired App instance.
func NewApp(enumerator serialtransport.PortEnumerator, catalogPath, dbPath string) (*App, error) {
	if enumerator == nil {
		enumerator = serialtransport.NewSystemPortEnumerator()
	}

	cat, err := firmware.LoadCatalog(catalogPath)
	if err != nil {
		// Fallback empty catalog if not found immediately
		cat = &firmware.Catalog{Targets: make(map[string]firmware.BoardTarget)}
	}

	st, err := store.NewStore(dbPath)
	if err != nil {
		// In-memory fallback
		st, _ = store.NewStore(":memory:")
	}

	app := &App{
		DevManager:    device.NewManager(enumerator),
		JobMgr:        jobs.NewJobManager(),
		SurveyStore:   survey.NewSurveyStore(),
		MonitorEngine: monitor.NewMonitorEngine(),
		BLEStore:      ble.NewBLEStore(),
		PolicyGate:    policy.NewPolicyGate(),
		FirmwareCat:   cat,
		FirmwareVal:   firmware.NewValidator(),
		Flasher:       firmware.NewFlasherAdapter(),
		Store:         st,
		RecentErrors:  make([]string, 0),
	}

	app.initEventBridges()
	return app, nil
}

// Startup is called by Wails when the application starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) emit(event string, data interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data)
	}
}

func (a *App) initEventBridges() {
	// Bridge device notifications to frontend events
	a.DevManager.AddListener(func(ev string, d *device.Device, data interface{}) {
		a.emit("device:"+ev, d)
		if ev == "connected" && a.Store != nil {
			_ = a.Store.SaveDevice(d.ID, "", d.Identity.BoardModel, d.Identity.ChipFamily, d.Identity.FirmwareVer)
		}
	})

	// Bridge job updates
	a.JobMgr.AddListener(func(j *jobs.Job) {
		a.emit("job:updated", j)
	})

	// Bridge incoming serial session events
	a.DevManager.AddSessionListener(func(ev *marauder.ParsedEvent) {
		a.handleWireEvent("", ev)
	})
}

// ScanPorts discovers attached serial devices.
func (a *App) ScanPorts() ([]*device.Device, error) {
	return a.DevManager.ScanPorts()
}

// ConnectDevice connects to a serial device.
func (a *App) ConnectDevice(path string) (*device.Device, error) {
	d, err := a.DevManager.Connect(path)
	if err != nil {
		a.recordError(err.Error())
		return nil, err
	}
	return d, nil
}

// DisconnectDevice disconnects from an active device.
func (a *App) DisconnectDevice(deviceID string) error {
	return a.DevManager.Disconnect(deviceID)
}

// ListDevices returns all known devices.
func (a *App) ListDevices() []*device.Device {
	return a.DevManager.ListDevices()
}

func (a *App) handleWireEvent(deviceID string, ev *marauder.ParsedEvent) {
	a.emit("serial:event", ev)

	switch ev.Type {
	case marauder.EventAP:
		if apData, ok := ev.Data.(marauder.APScanEventData); ok {
			ap := a.SurveyStore.RecordAP(apData)
			a.MonitorEngine.ProcessAP(apData)
			a.emit("survey:ap", ap)
		}
	case marauder.EventSTA:
		if staData, ok := ev.Data.(marauder.StationScanEventData); ok {
			sta := a.SurveyStore.RecordStation(staData)
			a.emit("survey:sta", sta)
		}
	case marauder.EventDeauth:
		if deauthData, ok := ev.Data.(marauder.SniffDeauthEventData); ok {
			a.MonitorEngine.ProcessDeauth(deauthData)
			a.emit("monitor:deauth", deauthData)
		}
	}
}

// StartSurvey starts a Wi-Fi survey job.
func (a *App) StartSurvey(deviceID string) (*jobs.Job, error) {
	session, ok := a.DevManager.GetSession(deviceID)
	if !ok {
		return nil, fmt.Errorf("device %s is not connected", deviceID)
	}

	// Check capability
	if session.Device.Capabilities != nil && !session.Device.Capabilities.IsAvailable("wifi_survey") {
		return nil, fmt.Errorf("wifi_survey is not currently available on this device")
	}

	job, err := a.JobMgr.CreateJob(context.Background(), deviceID, "wifi_survey", 0)
	if err != nil {
		return nil, err
	}

	scanCmd := session.Commands.ScanAll()
	if session.Device.Profile != nil && !session.Device.Profile.IsModern() {
		scanCmd = session.Commands.ScanAP()
	}

	if err := session.WriteCommand(scanCmd); err != nil {
		_ = a.JobMgr.FinishJob(job.ID, jobs.StateFailed, err.Error())
		return nil, err
	}

	return job, nil
}

// StopJob stops an active job.
func (a *App) StopJob(jobID string) error {
	job, ok := a.JobMgr.GetJob(jobID)
	if !ok {
		return jobs.ErrJobNotFound
	}

	// Send stopscan if it's a survey or sniff job
	if session, hasSession := a.DevManager.GetSession(job.DeviceID); hasSession {
		_ = session.WriteCommand(session.Commands.StopScan())
	}

	_ = a.JobMgr.StopJob(jobID)
	return a.JobMgr.FinishJob(jobID, jobs.StateSucceeded, "Stopped by user")
}

// StopAll halts all jobs across all devices immediately.
func (a *App) StopAll() {
	for _, d := range a.DevManager.ListDevices() {
		if session, ok := a.DevManager.GetSession(d.ID); ok {
			_ = session.WriteCommand(session.Commands.StopScan())
		}
	}
	a.JobMgr.StopAllJobs()
	a.emit("app:emergency_stop", "All operations halted")
}

// GetSurveyAPs returns recorded APs.
func (a *App) GetSurveyAPs(onlyAllowlisted bool) []*survey.AccessPoint {
	return a.SurveyStore.GetAPs(onlyAllowlisted)
}

// GetSurveyStations returns recorded Stations.
func (a *App) GetSurveyStations() []*survey.Station {
	return a.SurveyStore.GetStations()
}

// ClearSurvey clears in-memory survey data.
func (a *App) ClearSurvey() {
	a.SurveyStore.Clear()
}

// SetLabConsent sets authorization credentials.
func (a *App) SetLabConsent(operator, lab string) {
	a.PolicyGate.SetConsent(operator, lab)
}

// SetLabAllowlist sets the active target allowlist.
func (a *App) SetLabAllowlist(bssids []string) {
	a.PolicyGate.SetAllowlist(bssids)
	a.SurveyStore.SetAllowlist(bssids)
}

// StartAuthorizedLabTest initiates an authorized test under policy supervision.
func (a *App) StartAuthorizedLabTest(req policy.ActiveTestRequest) (*policy.AuditRecord, error) {
	record, err := a.PolicyGate.ValidateRequest(req)
	if err != nil {
		a.recordError(err.Error())
		return nil, err
	}

	session, ok := a.DevManager.GetSession(req.DeviceID)
	if !ok {
		return nil, fmt.Errorf("device %s not connected", req.DeviceID)
	}

	dur := time.Duration(req.DurationSeconds) * time.Second
	job, err := a.JobMgr.CreateJob(context.Background(), req.DeviceID, "authorized_lab", dur)
	if err != nil {
		return nil, err
	}

	cmdPayload := req.Command + "\r\n"
	if err := session.WriteCommand(cmdPayload); err != nil {
		_ = a.JobMgr.FinishJob(job.ID, jobs.StateFailed, err.Error())
		return nil, err
	}

	// Auto-stop routine after timeout
	go func() {
		<-job.Context().Done()
		_ = session.WriteCommand(session.Commands.StopScan())
		_ = a.JobMgr.FinishJob(job.ID, jobs.StateSucceeded, "Session duration expired")
	}()

	return record, nil
}

// GetFirmwareCatalog returns all board targets.
func (a *App) GetFirmwareCatalog() []firmware.BoardTarget {
	targets := make([]firmware.BoardTarget, 0, len(a.FirmwareCat.Targets))
	for _, t := range a.FirmwareCat.Targets {
		targets = append(targets, t)
	}
	return targets
}

// DryRunFlash simulates flashing for safety verification.
func (a *App) DryRunFlash(targetID, binaryPath string) error {
	target, ok := a.FirmwareCat.Targets[targetID]
	if !ok {
		return fmt.Errorf("unknown target ID: %s", targetID)
	}

	preflight, err := a.FirmwareVal.ValidatePreflight(target, target.ChipFamily, target.FlashSizeMB, binaryPath, "")
	if err != nil {
		return err
	}

	return a.Flasher.DryRunFlash(context.Background(), preflight, func(step string, pct int) {
		a.emit("flash:progress", map[string]interface{}{
			"step":    step,
			"percent": pct,
		})
	})
}

// ExportReportHTML generates a sanitized HTML report.
func (a *App) ExportReportHTML(title string) (string, error) {
	aps := a.SurveyStore.GetAPs(false)
	stas := a.SurveyStore.GetStations()
	alerts := a.MonitorEngine.GetAlerts()
	audit := a.PolicyGate.GetAuditTrail()

	report := reports.GenerateSanitizedReport(title, "Connected ESP32 Marauder", "v1.20.0", aps, stas, alerts, audit)
	var buf bytes.Buffer
	if err := report.ExportHTML(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetSupportBundle returns diagnostics.
func (a *App) GetSupportBundle() *diagnostics.SupportBundle {
	return diagnostics.GenerateSupportBundle("1.0.0", len(a.DevManager.ListDevices()), a.RecentErrors)
}

func (a *App) recordError(msg string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.RecentErrors = append(a.RecentErrors, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	if len(a.RecentErrors) > 50 {
		a.RecentErrors = a.RecentErrors[1:]
	}
}

// Ensure dummy import for filepath
var _ = filepath.Clean
