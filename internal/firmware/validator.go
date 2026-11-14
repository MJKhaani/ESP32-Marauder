package firmware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrMismatchChipFamily      = errors.New("detected chip family contradicts target binary layout")
	ErrMismatchFlashSize       = errors.New("flash size of connected hardware cannot accommodate partition scheme")
	ErrMismatchC5Layout        = errors.New("ESP32-C5 8MB layout cannot be flashed to 16MB board or vice versa")
	ErrChecksumVerificationFail = errors.New("firmware binary SHA-256 digest does not match expected checksum")
)

// FlashPreflight contains parameters validated before flashing.
type FlashPreflight struct {
	Target           BoardTarget `json:"target"`
	PortPath         string      `json:"port_path"`
	FirmwareBinary   string      `json:"firmware_binary"`
	ExpectedSHA256   string      `json:"expected_sha256,omitempty"`
	ComputedSHA256   string      `json:"computed_sha256"`
	EraseBeforeFlash bool        `json:"erase_before_flash"`
	IsValid          bool        `json:"is_valid"`
}

// Validator validates firmware images against board targets.
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ComputeSHA256 calculates the hex SHA-256 digest of a local file.
func (v *Validator) ComputeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ValidatePreflight ensures all safety criteria are satisfied prior to flash execution.
func (v *Validator) ValidatePreflight(target BoardTarget, detectedChip string, detectedFlashMB int, binaryPath, expectedHash string) (*FlashPreflight, error) {
	// 1. Verify file exists
	info, err := os.Stat(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("firmware file inaccessible: %w", err)
	}
	if info.Size() == 0 {
		return nil, errors.New("firmware file is empty")
	}

	// 2. Chip family check
	if detectedChip != "" && !strings.EqualFold(detectedChip, target.ChipFamily) {
		return nil, fmt.Errorf("%w: detected %s, target requires %s", ErrMismatchChipFamily, detectedChip, target.ChipFamily)
	}

	// 3. ESP32-C5 specific partition guardrail:
	// Marauder v8 (8MB default_8MB) vs T-Dongle C5 (16MB default_16MB)
	if strings.EqualFold(target.ChipFamily, "ESP32-C5") {
		if target.FlashSizeMB == 8 && detectedFlashMB == 16 {
			return nil, fmt.Errorf("%w: attempting to flash 8MB v8 image to 16MB T-Dongle", ErrMismatchC5Layout)
		}
		if target.FlashSizeMB == 16 && detectedFlashMB == 8 {
			return nil, fmt.Errorf("%w: attempting to flash 16MB T-Dongle image to 8MB v8 board", ErrMismatchC5Layout)
		}
	}

	// 4. Flash size bounds check
	if detectedFlashMB > 0 && target.FlashSizeMB > detectedFlashMB {
		return nil, fmt.Errorf("%w: board has %dMB, target requires %dMB", ErrMismatchFlashSize, detectedFlashMB, target.FlashSizeMB)
	}

	// 5. SHA-256 verification
	computedHash, err := v.ComputeSHA256(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed computing hash: %w", err)
	}

	if expectedHash != "" && !strings.EqualFold(computedHash, expectedHash) {
		return nil, fmt.Errorf("%w: got %s, expected %s", ErrChecksumVerificationFail, computedHash, expectedHash)
	}

	return &FlashPreflight{
		Target:         target,
		FirmwareBinary: binaryPath,
		ExpectedSHA256: expectedHash,
		ComputedSHA256: computedHash,
		IsValid:        true,
	}, nil
}
