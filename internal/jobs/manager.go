package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDeviceBusy        = errors.New("device is currently executing another job")
	ErrJobNotFound       = errors.New("job not found")
	ErrJobNotCancellable = errors.New("job cannot be cancelled in its current state")
)

// JobEventListener receives notifications when job state or progress updates.
type JobEventListener func(job *Job)

// JobManager coordinates job lifecycles and enforces per-device radio locks.
type JobManager struct {
	mu          sync.RWMutex
	jobs        map[string]*Job
	activeOnDev map[string]string // deviceID -> jobID

	listenersMu sync.RWMutex
	listeners   []JobEventListener
}

// NewJobManager creates a new JobManager.
func NewJobManager() *JobManager {
	return &JobManager{
		jobs:        make(map[string]*Job),
		activeOnDev: make(map[string]string),
		listeners:   make([]JobEventListener, 0),
	}
}

// AddListener registers an event callback for job updates.
func (m *JobManager) AddListener(l JobEventListener) {
	m.listenersMu.Lock()
	defer m.listenersMu.Unlock()
	m.listeners = append(m.listeners, l)
}

func (m *JobManager) notify(j *Job) {
	m.listenersMu.RLock()
	listeners := append([]JobEventListener(nil), m.listeners...)
	m.listenersMu.RUnlock()

	for _, l := range listeners {
		l(j)
	}
}

// CreateJob creates and queues a new job, enforcing single active job per device.
func (m *JobManager) CreateJob(parentCtx context.Context, deviceID, jobType string, timeout time.Duration) (*Job, error) {
	m.mu.Lock()

	if activeID, busy := m.activeOnDev[deviceID]; busy {
		if activeJob, ok := m.jobs[activeID]; ok && (activeJob.State == StateRunning || activeJob.State == StateStopping) {
			m.mu.Unlock()
			return nil, fmt.Errorf("%w: active job %s (%s)", ErrDeviceBusy, activeJob.ID, activeJob.Type)
		}
	}

	if parentCtx == nil {
		parentCtx = context.Background()
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(parentCtx, timeout)
	} else {
		ctx, cancel = context.WithCancel(parentCtx)
	}

	id := uuid.New().String()
	job := &Job{
		ID:        id,
		DeviceID:  deviceID,
		Type:      jobType,
		State:     StateRunning,
		Progress:  0,
		StartedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
		ctx:       ctx,
		cancel:    cancel,
	}

	m.jobs[id] = job
	m.activeOnDev[deviceID] = id
	m.mu.Unlock()

	m.notify(job)
	return job, nil
}

// FinishJob completes a job and releases the device lock.
func (m *JobManager) FinishJob(jobID string, state JobState, errStr string) error {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return ErrJobNotFound
	}

	job.SetState(state, errStr)
	delete(m.activeOnDev, job.DeviceID)
	m.mu.Unlock()

	m.notify(job)
	return nil
}

// StopJob cancels an active job.
func (m *JobManager) StopJob(jobID string) error {
	m.mu.RLock()
	job, ok := m.jobs[jobID]
	m.mu.RUnlock()

	if !ok {
		return ErrJobNotFound
	}

	job.Cancel()
	m.notify(job)
	return nil
}

// StopAllJobs terminates all currently active jobs across all devices.
func (m *JobManager) StopAllJobs() {
	m.mu.RLock()
	activeList := make([]*Job, 0)
	for _, job := range m.jobs {
		if job.State == StateRunning {
			activeList = append(activeList, job)
		}
	}
	m.mu.RUnlock()

	for _, job := range activeList {
		job.Cancel()
		m.notify(job)
	}
}

// GetJob returns a copy or pointer to the job.
func (m *JobManager) GetJob(jobID string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[jobID]
	return j, ok
}

// ListJobs returns all recorded jobs.
func (m *JobManager) ListJobs() []*Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		list = append(list, j)
	}
	return list
}
