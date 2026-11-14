package jobs

import (
	"context"
	"sync"
	"time"
)

// JobState represents the lifecycle status of a background job.
type JobState string

const (
	StateQueued      JobState = "queued"
	StateRunning     JobState = "running"
	StateStopping    JobState = "stopping"
	StateSucceeded   JobState = "succeeded"
	StateFailed      JobState = "failed"
	StateInterrupted JobState = "interrupted"
)

// Job represents an execution unit running on a connected device.
type Job struct {
	ID        string                 `json:"id"`
	DeviceID  string                 `json:"device_id"`
	Type      string                 `json:"type"` // e.g. "survey", "sniff_beacon", "flash"
	State     JobState               `json:"state"`
	Progress  int                    `json:"progress"` // 0-100
	StartedAt time.Time              `json:"started_at"`
	EndedAt   *time.Time             `json:"ended_at,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// UpdateProgress safely updates the percentage progress of the job.
func (j *Job) UpdateProgress(p int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	j.Progress = p
}

// SetState updates the job's state.
func (j *Job) SetState(s JobState, errStr string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.State = s
	j.Error = errStr
	if s == StateSucceeded || s == StateFailed || s == StateInterrupted {
		now := time.Now()
		j.EndedAt = &now
	}
}

// Context returns the job's execution context.
func (j *Job) Context() context.Context {
	return j.ctx
}

// Cancel requests premature termination of the job.
func (j *Job) Cancel() {
	j.mu.Lock()
	if j.State == StateRunning {
		j.State = StateStopping
	}
	j.mu.Unlock()

	if j.cancel != nil {
		j.cancel()
	}
}
