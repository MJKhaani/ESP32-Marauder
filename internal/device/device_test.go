package device

import (
	"os"
	"testing"
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/serialtransport"
)

func TestDeviceManagerConnectAndInterrogate(t *testing.T) {
	enum := serialtransport.NewMockPortEnumerator()
	mockPort := serialtransport.NewMockPort()

	// Register info and help fixtures
	infoData, _ := os.ReadFile("../../testdata/info_v120.txt")
	helpData, _ := os.ReadFile("../../testdata/help_v120.txt")

	mockPort.RegisterResponse("info", infoData)
	mockPort.RegisterResponse("help", helpData)

	targetPath := "/dev/tty.usbserial-0001"
	enum.RegisterPort(targetPath, mockPort)

	mgr := NewManager(enum)

	// 1. Scan ports
	devs, err := mgr.ScanPorts()
	if err != nil {
		t.Fatalf("ScanPorts failed: %v", err)
	}
	if len(devs) < 2 {
		t.Fatalf("Expected at least 2 discovered devices, got %d", len(devs))
	}

	// 2. Connect
	d, err := mgr.Connect(targetPath)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	if d.State != StateConnected {
		t.Errorf("Expected device state connected, got %s", d.State)
	}

	// Wait briefly for interrogation to finish
	time.Sleep(50 * time.Millisecond)

	if d.Identity.FirmwareVer != "v1.20.0" {
		t.Errorf("Expected firmware v1.20.0, got %s", d.Identity.FirmwareVer)
	}
	if d.Identity.BoardModel != "Marauder v6" {
		t.Errorf("Expected board 'Marauder v6', got %s", d.Identity.BoardModel)
	}
	if d.Capabilities == nil {
		t.Fatalf("Expected non-nil capabilities matrix")
	}
	if !d.Capabilities.IsAvailable("wifi_survey") {
		t.Errorf("Expected wifi_survey to be available")
	}

	// 3. Disconnect
	err = mgr.Disconnect(targetPath)
	if err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}
	if d.State != StateDisconnected {
		t.Errorf("Expected state disconnected, got %s", d.State)
	}
}
