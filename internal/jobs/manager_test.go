package jobs

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestJobManagerLifecycle(t *testing.T) {
	mgr := NewJobManager()

	job, err := mgr.CreateJob(context.Background(), "dev-01", "survey", 2*time.Second)
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}
	if job.State != StateRunning {
		t.Errorf("Expected job state running, got %s", job.State)
	}

	// Try creating second job on same device -> must fail
	_, err = mgr.CreateJob(context.Background(), "dev-01", "sniff", 0)
	if !errors.Is(err, ErrDeviceBusy) {
		t.Errorf("Expected ErrDeviceBusy, got %v", err)
	}

	// Stop job
	err = mgr.StopJob(job.ID)
	if err != nil {
		t.Fatalf("StopJob failed: %v", err)
	}

	// Verify context cancelled
	select {
	case <-job.Context().Done():
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Expected job context to be cancelled")
	}

	// Finish job
	_ = mgr.FinishJob(job.ID, StateSucceeded, "")

	// Now a second job on dev-01 should succeed
	job2, err := mgr.CreateJob(context.Background(), "dev-01", "sniff", 0)
	if err != nil {
		t.Fatalf("CreateJob after finish failed: %v", err)
	}
	if job2 == nil {
		t.Fatalf("Expected valid job2")
	}
}

func TestJobManagerStopAll(t *testing.T) {
	mgr := NewJobManager()

	j1, _ := mgr.CreateJob(context.Background(), "dev-01", "survey", 0)
	j2, _ := mgr.CreateJob(context.Background(), "dev-02", "sniff", 0)

	mgr.StopAllJobs()

	if j1.State != StateStopping {
		t.Errorf("Expected j1 stopping, got %s", j1.State)
	}
	if j2.State != StateStopping {
		t.Errorf("Expected j2 stopping, got %s", j2.State)
	}
}
