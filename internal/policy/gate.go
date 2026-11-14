package policy

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrConsentRequired   = errors.New("user has not acknowledged explicit lab ownership and authorization")
	ErrAllowlistEmpty     = errors.New("operation requires an explicit allowlist of authorized lab BSSIDs")
	ErrWildcardProhibited = errors.New("wildcard, broadcast, or unscoped target identifiers are strictly prohibited")
	ErrDurationExceeded   = errors.New("requested duration exceeds maximum authorized test window (max 120s)")
)

// LabConsent records the user's manual authorization confirmation.
type LabConsent struct {
	OperatorName string    `json:"operator_name"`
	LabName      string    `json:"lab_name"`
	SignedAt     time.Time `json:"signed_at"`
	IsActive     bool      `json:"is_active"`
}

// ActiveTestRequest represents an authorized test invocation.
type ActiveTestRequest struct {
	DeviceID        string        `json:"device_id"`
	Command         string        `json:"command"`
	TargetBSSIDs    []string      `json:"target_bssids"`
	DurationSeconds int           `json:"duration_seconds"`
	Justification   string        `json:"justification"`
}

// PolicyGate enforces safety rules for potentially disruptive RF operations.
type PolicyGate struct {
	mu           sync.RWMutex
	consent      *LabConsent
	allowlist    map[string]bool
	auditTrail   []AuditRecord
	maxDurationS int
}

// NewPolicyGate creates a new PolicyGate.
func NewPolicyGate() *PolicyGate {
	return &PolicyGate{
		allowlist:    make(map[string]bool),
		auditTrail:   make([]AuditRecord, 0),
		maxDurationS: 120, // 2 minutes maximum bounded duration
	}
}

// SetConsent registers explicit user authorization.
func (p *PolicyGate) SetConsent(operatorName, labName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consent = &LabConsent{
		OperatorName: operatorName,
		LabName:      labName,
		SignedAt:     time.Now(),
		IsActive:     true,
	}
}

// RevokeConsent invalidates the current lab authorization.
func (p *PolicyGate) RevokeConsent() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.consent != nil {
		p.consent.IsActive = false
	}
}

// SetAllowlist registers approved target BSSIDs.
func (p *PolicyGate) SetAllowlist(bssids []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.allowlist = make(map[string]bool)
	for _, b := range bssids {
		norm := strings.ToLower(strings.TrimSpace(b))
		if norm != "" && norm != "*" && norm != "ff:ff:ff:ff:ff:ff" {
			p.allowlist[norm] = true
		}
	}
}

// ValidateRequest checks if an active test request satisfies safety policy.
func (p *PolicyGate) ValidateRequest(req ActiveTestRequest) (*AuditRecord, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 1. Consent check
	if p.consent == nil || !p.consent.IsActive {
		return nil, ErrConsentRequired
	}

	// 2. Allowlist targets check
	if len(req.TargetBSSIDs) == 0 {
		return nil, ErrAllowlistEmpty
	}

	for _, target := range req.TargetBSSIDs {
		norm := strings.ToLower(strings.TrimSpace(target))
		if norm == "" || norm == "*" || norm == "ff:ff:ff:ff:ff:ff" {
			return nil, ErrWildcardProhibited
		}
		if !p.allowlist[norm] {
			return nil, fmt.Errorf("target BSSID %s is not in the verified lab allowlist", norm)
		}
	}

	// 3. Duration check
	if req.DurationSeconds <= 0 || req.DurationSeconds > p.maxDurationS {
		return nil, fmt.Errorf("%w: requested %ds", ErrDurationExceeded, req.DurationSeconds)
	}

	// 4. Create and append audit record
	record := AuditRecord{
		ID:              uuid.New().String(),
		Timestamp:       time.Now(),
		Operator:        p.consent.OperatorName,
		Lab:             p.consent.LabName,
		DeviceID:        req.DeviceID,
		Command:         req.Command,
		Targets:         req.TargetBSSIDs,
		DurationSeconds: req.DurationSeconds,
		Approved:        true,
	}
	p.auditTrail = append(p.auditTrail, record)

	return &record, nil
}

// GetAuditTrail returns all recorded audit records.
func (p *PolicyGate) GetAuditTrail() []AuditRecord {
	p.mu.RLock()
	defer p.mu.RUnlock()
	res := make([]AuditRecord, len(p.auditTrail))
	copy(res, p.auditTrail)
	return res
}
