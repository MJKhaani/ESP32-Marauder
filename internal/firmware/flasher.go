package firmware

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FlasherAdapter manages firmware burning through isolated subprocess calls or dry-run simulation.
type FlasherAdapter struct{}

// NewFlasherAdapter creates a new FlasherAdapter.
func NewFlasherAdapter() *FlasherAdapter {
	return &FlasherAdapter{}
}

// BuildEsptoolArgs produces a safe, structured argument slice for esptool.py.
// NEVER string-interpolated into a shell!
func (f *FlasherAdapter) BuildEsptoolArgs(preflight *FlashPreflight, port string, baud int) []string {
	if baud <= 0 {
		baud = 921600
	}

	chipFlag := strings.ToLower(preflight.Target.ChipFamily)
	chipFlag = strings.ReplaceAll(chipFlag, "-", "") // "esp32s3", "esp32c5", "esp32"

	args := []string{
		"--chip", chipFlag,
		"--port", port,
		"--baud", fmt.Sprintf("%d", baud),
		"--before", "default_reset",
		"--after", "hard_reset",
		"write_flash",
		"-z",
		"--flash_mode", "dio",
		"--flash_freq", "80m",
		"--flash_size", fmt.Sprintf("%dMB", preflight.Target.FlashSizeMB),
		preflight.Target.AppOffset, preflight.FirmwareBinary,
	}

	return args
}

// DryRunFlash simulates the step-by-step flashing workflow for validation without touching hardware.
func (f *FlasherAdapter) DryRunFlash(ctx context.Context, preflight *FlashPreflight, progressCb func(step string, pct int)) error {
	steps := []struct {
		name string
		pct  int
	}{
		{"Preflight verification & checksum match", 10},
		{"Entering ROM bootloader & syncing baud", 25},
		{"Verifying chip revision and partition table", 45},
		{"Writing firmware binary payload to flash offset " + preflight.Target.AppOffset, 75},
		{"Verifying flash readback hash", 90},
		{"Issuing hard reset and awaiting runtime CLI", 100},
	}

	for _, s := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if progressCb != nil {
				progressCb(s.name, s.pct)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}

	return nil
}
