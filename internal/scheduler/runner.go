package scheduler

import (
	"context"
	"fmt"
	"time"
)

type TickJob func(context.Context, time.Time)

type ticker interface {
	Chan() <-chan time.Time
	Stop()
}

type wallTicker struct {
	ticker *time.Ticker
}

func newWallTicker(interval time.Duration) ticker {
	return &wallTicker{ticker: time.NewTicker(interval)}
}

func (t *wallTicker) Chan() <-chan time.Time {
	return t.ticker.C
}

func (t *wallTicker) Stop() {
	t.ticker.Stop()
}

type Runner struct {
	interval  time.Duration
	jobs      []TickJob
	now       func() time.Time
	newTicker func(time.Duration) ticker
}

func NewRunner(interval time.Duration, jobs ...TickJob) (*Runner, error) {
	if interval <= 0 {
		return nil, fmt.Errorf("interval must be positive")
	}

	runner := &Runner{
		interval:  interval,
		jobs:      jobs,
		now:       time.Now,
		newTicker: newWallTicker,
	}
	return runner, nil
}

func (r *Runner) Run(ctx context.Context) {
	ticker := r.newTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.Chan():
			now := r.now().UTC()
			for _, job := range r.jobs {
				job(ctx, now)
			}
		}
	}
}
