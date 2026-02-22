package scheduler

import (
	"context"
	"testing"
	"time"
)

type fakeTicker struct {
	ch      chan time.Time
	stopped bool
}

func newFakeTicker() *fakeTicker {
	return &fakeTicker{ch: make(chan time.Time)}
}

func (t *fakeTicker) Chan() <-chan time.Time {
	return t.ch
}

func (t *fakeTicker) Stop() {
	t.stopped = true
}

func TestNewRunnerRejectsNonPositiveInterval(t *testing.T) {
	if _, err := NewRunner(0); err == nil {
		t.Fatalf("expected error for non-positive interval")
	}
}

func TestRunnerRunsAllJobsOnTick(t *testing.T) {
	jobCalls := 0
	var calledAt time.Time

	runner, err := NewRunner(time.Second,
		func(_ context.Context, at time.Time) {
			jobCalls++
			calledAt = at
		},
		func(_ context.Context, _ time.Time) {
			jobCalls++
		},
	)
	if err != nil {
		t.Fatalf("create runner: %v", err)
	}

	fake := newFakeTicker()
	tickTime := time.Date(2026, 2, 22, 12, 0, 0, 0, time.UTC)
	runner.newTicker = func(time.Duration) ticker { return fake }
	runner.now = func() time.Time { return tickTime }

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runner.Run(ctx)
		close(done)
	}()

	fake.ch <- tickTime

	deadline := time.After(2 * time.Second)
	for jobCalls < 2 {
		select {
		case <-deadline:
			t.Fatalf("expected 2 job calls, got %d", jobCalls)
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	if !calledAt.Equal(tickTime) {
		t.Fatalf("expected calledAt %v, got %v", tickTime, calledAt)
	}

	cancel()
	<-done
	if !fake.stopped {
		t.Fatalf("expected ticker to be stopped")
	}
}
