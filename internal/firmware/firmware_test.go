package firmware

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogAndValidator(t *testing.T) {
	cat, err := LoadCatalog("../../firmware/catalog/targets.json")
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	target, found := cat.FindTargetByModel("marauder_v6")
	if !found {
		t.Fatalf("Target marauder_v6 not found in catalog")
	}
	if target.ChipFamily != "ESP32" {
		t.Errorf("Expected chip ESP32, got %s", target.ChipFamily)
	}

	// Create dummy binary
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "firmware.bin")
	_ = os.WriteFile(binPath, []byte("ESP32_MARAUDER_TEST_IMAGE_PAYLOAD"), 0644)

	validator := NewValidator()
	computedHash, err := validator.ComputeSHA256(binPath)
	if err != nil {
		t.Fatalf("ComputeSHA256 failed: %v", err)
	}

	// 1. Valid preflight
	preflight, err := validator.ValidatePreflight(*target, "ESP32", 4, binPath, computedHash)
	if err != nil {
		t.Fatalf("ValidatePreflight failed: %v", err)
	}
	if !preflight.IsValid {
		t.Errorf("Expected preflight to be valid")
	}

	// 2. Chip mismatch -> error
	_, err = validator.ValidatePreflight(*target, "ESP32-S3", 16, binPath, computedHash)
	if !errors.Is(err, ErrMismatchChipFamily) {
		t.Errorf("Expected ErrMismatchChipFamily, got %v", err)
	}

	// 3. ESP32-C5 Layout Guardrail
	c5Target, found := cat.FindTargetByModel("marauder_v8")
	if !found {
		t.Fatalf("marauder_v8 target not found")
	}
	// Try flashing 8MB target on 16MB board
	_, err = validator.ValidatePreflight(*c5Target, "ESP32-C5", 16, binPath, computedHash)
	if !errors.Is(err, ErrMismatchC5Layout) {
		t.Errorf("Expected ErrMismatchC5Layout, got %v", err)
	}

	// 4. Test Flasher Args
	flasher := NewFlasherAdapter()
	args := flasher.BuildEsptoolArgs(preflight, "/dev/tty.usbserial", 115200)
	if len(args) == 0 || args[1] != "esp32" {
		t.Errorf("Unexpected esptool args: %v", args)
	}

	// 5. Test DryRunFlash
	stepCount := 0
	err = flasher.DryRunFlash(context.Background(), preflight, func(step string, pct int) {
		stepCount++
	})
	if err != nil {
		t.Fatalf("DryRunFlash failed: %v", err)
	}
	if stepCount != 6 {
		t.Errorf("Expected 6 steps, got %d", stepCount)
	}
}
