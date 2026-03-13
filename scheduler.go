package cronkit

import (
	"context"
	"sync"
	"time"
)

// Job represents a scheduled job.
type Job struct {
	Name     string
	Schedule *Schedule
	Handler  func(ctx context.Context)
	running  bool
}

// Scheduler manages cron jobs with overlap prevention.
type Scheduler struct {
	mu       sync.Mutex
	jobs     []*Job
	stop     chan struct{}
	stopOnce sync.Once
}

// NewScheduler creates a new Scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		stop: make(chan struct{}),
	}
}

// Add registers a new job with the given cron expression.
func (s *Scheduler) Add(name string, expr string, handler func(ctx context.Context)) error {
	sched, err := Parse(expr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs = append(s.jobs, &Job{
		Name:     name,
		Schedule: sched,
		Handler:  handler,
	})
	return nil
}

// Start begins the scheduler. It checks for due jobs every minute.
// It blocks until the context is cancelled or Stop is called.
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// Check immediately on start
	s.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stop:
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// Stop signals the scheduler to stop. Safe to call multiple times.
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() { close(s.stop) })
}

// NextRun returns the next scheduled run time for the named job.
func (s *Scheduler) NextRun(name string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.jobs {
		if job.Name == name {
			return job.Schedule.Next(time.Now()), true
		}
	}
	return time.Time{}, false
}

// Jobs returns a copy of the job names.
func (s *Scheduler) Jobs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, len(s.jobs))
	for i, j := range s.jobs {
		names[i] = j.Name
	}
	return names
}

func (s *Scheduler) tick(ctx context.Context) {
	now := time.Now().Truncate(time.Minute)
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.jobs {
		if job.running {
			continue // overlap prevention
		}
		if job.Schedule.matches(now) {
			job.running = true
			go func(j *Job) {
				defer func() {
					recover()
					s.mu.Lock()
					j.running = false
					s.mu.Unlock()
				}()
				j.Handler(ctx)
			}(job)
		}
	}
}
