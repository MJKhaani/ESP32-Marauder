package policy

import (
	"errors"
	"testing"
)

func TestPolicyGateEnforcement(t *testing.T) {
	gate := NewPolicyGate()

	req := ActiveTestRequest{
		DeviceID:        "dev-01",
		Command:         "attack -t deauth",
		TargetBSSIDs:    []string{"00:11:22:33:44:55"},
		DurationSeconds: 30,
	}

	// 1. Without consent -> ErrConsentRequired
	_, err := gate.ValidateRequest(req)
	if !errors.Is(err, ErrConsentRequired) {
		t.Errorf("Expected ErrConsentRequired, got %v", err)
	}

	// 2. Add consent, but target not in allowlist
	gate.SetConsent("Researcher Bob", "Lab Room 402")
	_, err = gate.ValidateRequest(req)
	if err == nil {
		t.Errorf("Expected error for non-allowlisted target, got nil")
	}

	// 3. Set allowlist with target
	gate.SetAllowlist([]string{"00:11:22:33:44:55"})
	record, err := gate.ValidateRequest(req)
	if err != nil {
		t.Fatalf("ValidateRequest failed: %v", err)
	}
	if record.Operator != "Researcher Bob" {
		t.Errorf("Expected operator Bob, got %s", record.Operator)
	}

	// 4. Wildcard rejection
	wildcardReq := req
	wildcardReq.TargetBSSIDs = []string{"*"}
	_, err = gate.ValidateRequest(wildcardReq)
	if !errors.Is(err, ErrWildcardProhibited) {
		t.Errorf("Expected ErrWildcardProhibited, got %v", err)
	}

	// 5. Exceeded duration rejection (> 120s)
	longReq := req
	longReq.DurationSeconds = 300
	_, err = gate.ValidateRequest(longReq)
	if !errors.Is(err, ErrDurationExceeded) {
		t.Errorf("Expected ErrDurationExceeded, got %v", err)
	}
}
