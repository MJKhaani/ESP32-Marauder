package integration

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/app"
	"github.com/MJKhaani/ESP32-Marauder/internal/policy"
	"github.com/MJKhaani/ESP32-Marauder/internal/serialtransport"
)

func TestEndToEndPipeline(t *testing.T) {
	enum := serialtransport.NewMockPortEnumerator()
	mockPort := serialtransport.NewMockPort()

	infoData, _ := os.ReadFile("../../testdata/info_v120.txt")
	helpData, _ := os.ReadFile("../../testdata/help_v120.txt")
	scanData, _ := os.ReadFile("../../testdata/scanall_output.txt")

	mockPort.RegisterResponse("info", infoData)
	mockPort.RegisterResponse("help", helpData)
	mockPort.RegisterResponse("scanall", scanData)
	mockPort.RegisterResponse("stopscan", []byte("Scan stopped\r\n"))

	targetPath := "/dev/tty.usbserial-0001"
	enum.RegisterPort(targetPath, mockPort)

	application, err := app.NewApp(enum, "../../firmware/catalog/targets.json", ":memory:")
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	// 1. Scan ports & connect
	devices, err := application.ScanPorts()
	if err != nil || len(devices) == 0 {
		t.Fatalf("ScanPorts failed or found 0 devices")
	}

	d, err := application.ConnectDevice(targetPath)
	if err != nil {
		t.Fatalf("ConnectDevice failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if d.Identity.FirmwareVer != "v1.20.0" {
		t.Errorf("Expected firmware v1.20.0, got %s", d.Identity.FirmwareVer)
	}

	// 2. Start Survey
	job, err := application.StartSurvey(d.ID)
	if err != nil {
		t.Fatalf("StartSurvey failed: %v", err)
	}
	if job == nil || job.Type != "wifi_survey" {
		t.Fatalf("Invalid survey job returned")
	}

	time.Sleep(50 * time.Millisecond)

	// Verify APs parsed into store
	aps := application.GetSurveyAPs(false)
	if len(aps) < 3 {
		t.Errorf("Expected 3 APs recorded from scanall output, got %d", len(aps))
	}

	// 3. Stop survey job
	err = application.StopJob(job.ID)
	if err != nil {
		t.Fatalf("StopJob failed: %v", err)
	}

	// 4. Authorized Lab testing with policy gate
	req := policy.ActiveTestRequest{
		DeviceID:        d.ID,
		Command:         "attack -t deauth",
		TargetBSSIDs:    []string{"00:11:22:33:44:55"},
		DurationSeconds: 10,
	}

	// Unconsented -> reject
	_, err = application.StartAuthorizedLabTest(req)
	if err == nil {
		t.Errorf("Expected error for unconsented lab test, got nil")
	}

	// Sign consent and allowlist
	application.SetLabConsent("Lead Researcher", "Wireless Security Lab")
	application.SetLabAllowlist([]string{"00:11:22:33:44:55"})

	record, err := application.StartAuthorizedLabTest(req)
	if err != nil {
		t.Fatalf("StartAuthorizedLabTest failed: %v", err)
	}
	if !record.Approved {
		t.Errorf("Expected approved audit record")
	}

	// 5. Emergency Stop
	application.StopAll()

	// 6. Report Export
	htmlReport, err := application.ExportReportHTML("Final Verification Report")
	if err != nil {
		t.Fatalf("ExportReportHTML failed: %v", err)
	}
	if !strings.Contains(htmlReport, "00:11:22:xx:xx:xx") {
		t.Errorf("HTML report missing sanitized BSSID")
	}

	_ = application.DisconnectDevice(d.ID)
}
