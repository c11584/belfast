package scheduler

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/region"
)

func TestNewResetClockRejectsUnsupportedRegion(t *testing.T) {
	if _, err := NewResetClock("US"); err == nil {
		t.Fatalf("expected unsupported region error")
	}
}

func TestResetClockBoundariesCN(t *testing.T) {
	clock, err := NewResetClock("CN")
	if err != nil {
		t.Fatalf("create clock: %v", err)
	}

	now := time.Date(2026, 2, 22, 15, 30, 0, 0, time.UTC)

	if got := clock.CurrentDailyReset(now); !got.Equal(time.Date(2026, 2, 21, 16, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected daily reset: %v", got)
	}
	if got := clock.CurrentWeeklyReset(now); !got.Equal(time.Date(2026, 2, 15, 16, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected weekly reset: %v", got)
	}
	if got := clock.CurrentMonthlyReset(now); !got.Equal(time.Date(2026, 1, 31, 16, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected monthly reset: %v", got)
	}
	if got := clock.CurrentMonthKey(now); got != 202602 {
		t.Fatalf("unexpected month key: %d", got)
	}
}

func TestResetClockBoundariesEN(t *testing.T) {
	clock, err := NewResetClock("EN")
	if err != nil {
		t.Fatalf("create clock: %v", err)
	}

	now := time.Date(2026, 2, 22, 7, 30, 0, 0, time.UTC)

	if got := clock.CurrentDailyReset(now); !got.Equal(time.Date(2026, 2, 22, 7, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected daily reset: %v", got)
	}
	if got := clock.CurrentWeeklyReset(now); !got.Equal(time.Date(2026, 2, 16, 7, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected weekly reset: %v", got)
	}
	if got := clock.CurrentMonthlyReset(now); !got.Equal(time.Date(2026, 2, 1, 7, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected monthly reset: %v", got)
	}
	if got := clock.CurrentMonthKey(now); got != 202602 {
		t.Fatalf("unexpected month key: %d", got)
	}
}

func TestWeeklyResetMatchesMondayAnchor(t *testing.T) {
	regions := []string{"CN", "EN"}
	for _, regionID := range regions {
		clock, err := NewResetClock(regionID)
		if err != nil {
			t.Fatalf("create clock for %s: %v", regionID, err)
		}

		anchor := time.Unix(int64(consts.Monday_0OclockTimestamps[regionID]), 0).UTC()
		now := anchor.Add(2 * time.Hour)
		if got := clock.CurrentWeeklyReset(now); !got.Equal(anchor) {
			t.Fatalf("expected weekly reset to match anchor for %s: got %v want %v", regionID, got, anchor)
		}
	}
}

func TestNewCurrentRegionResetClock(t *testing.T) {
	region.ResetCurrentForTest()
	t.Setenv("AL_REGION", "JP")

	clock, err := NewCurrentRegionResetClock()
	if err != nil {
		t.Fatalf("create current-region clock: %v", err)
	}

	now := time.Date(2026, 2, 22, 0, 30, 0, 0, time.UTC)
	if got := clock.CurrentDailyReset(now); !got.Equal(time.Date(2026, 2, 21, 15, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected JP daily reset: %v", got)
	}
}
