package cronkit

import (
	"context"
	"testing"
	"time"
)

func TestSchedulerAddAndJobs(t *testing.T) {
	s := NewScheduler()
	err := s.Add("job1", "* * * * *", func(ctx context.Context) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = s.Add("job2", "0 * * * *", func(ctx context.Context) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jobs := s.Jobs()
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0] != "job1" || jobs[1] != "job2" {
		t.Errorf("unexpected job names: %v", jobs)
	}
}

func TestSchedulerAddInvalidExpression(t *testing.T) {
	s := NewScheduler()
	err := s.Add("bad", "invalid", func(ctx context.Context) {})
	if err == nil {
		t.Fatal("expected error for invalid expression")
	}
}

func TestSchedulerNextRun(t *testing.T) {
	s := NewScheduler()
	s.Add("job", "0 * * * *", func(ctx context.Context) {})

	next, ok := s.NextRun("job")
	if !ok {
		t.Fatal("expected NextRun to find job")
	}
	if next.IsZero() {
		t.Fatal("expected non-zero next run time")
	}
	if next.Minute() != 0 {
		t.Errorf("expected next run at minute 0, got %d", next.Minute())
	}
}

func TestSchedulerNextRunNotFound(t *testing.T) {
	s := NewScheduler()
	_, ok := s.NextRun("nonexistent")
	if ok {
		t.Fatal("expected NextRun to return false for missing job")
	}
}

func TestSchedulerStopTwice(t *testing.T) {
	s := NewScheduler()

	// Should not panic
	s.Stop()
	s.Stop()
}

func TestSchedulerStartStop(t *testing.T) {
	s := NewScheduler()
	s.Add("job", "* * * * *", func(ctx context.Context) {})

	done := make(chan struct{})
	go func() {
		ctx := context.Background()
		s.Start(ctx)
		close(done)
	}()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop in time")
	}
}

func TestSchedulerHandlerPanicRecovery(t *testing.T) {
	s := NewScheduler()

	callCount := 0
	s.mu.Lock()
	s.jobs = append(s.jobs, &Job{
		Name: "panicker",
		Schedule: &Schedule{
			Minutes:     makeRange(0, 59, 1),
			Hours:       makeRange(0, 23, 1),
			DaysOfMonth: makeRange(1, 31, 1),
			Months:      makeRange(1, 12, 1),
			DaysOfWeek:  makeRange(0, 6, 1),
			domWild:     true,
			dowWild:     true,
		},
		Handler: func(ctx context.Context) {
			callCount++
			if callCount == 1 {
				panic("test panic")
			}
		},
	})
	s.mu.Unlock()

	// First tick — handler panics, but running flag should be reset
	ctx := context.Background()
	s.tick(ctx)
	time.Sleep(50 * time.Millisecond)

	// Verify the job is not stuck — running should be false
	s.mu.Lock()
	running := s.jobs[0].running
	s.mu.Unlock()
	if running {
		t.Fatal("expected job to not be stuck after panic")
	}

	// Second tick — handler should run again
	s.tick(ctx)
	time.Sleep(50 * time.Millisecond)

	if callCount < 2 {
		t.Fatalf("expected handler to be called again after panic, callCount=%d", callCount)
	}
}

func TestSchedulerContextCancel(t *testing.T) {
	s := NewScheduler()
	s.Add("job", "* * * * *", func(ctx context.Context) {})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop on context cancel")
	}
}
