package scheduler

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/region"
)

var regionOffsets = map[string]int{
	"CN": 8 * 60 * 60,
	"EN": -7 * 60 * 60,
	"JP": 9 * 60 * 60,
	"KR": 9 * 60 * 60,
	"TW": 8 * 60 * 60,
}

type ResetClock struct {
	regionID string
	location *time.Location
}

func NewResetClock(regionID string) (*ResetClock, error) {
	offset, ok := regionOffsets[regionID]
	if !ok {
		return nil, fmt.Errorf("unsupported region %q", regionID)
	}

	clock := &ResetClock{
		regionID: regionID,
		location: time.FixedZone(regionID, offset),
	}
	return clock, nil
}

func NewCurrentRegionResetClock() (*ResetClock, error) {
	return NewResetClock(region.Current())
}

func (c *ResetClock) CurrentDailyReset(now time.Time) time.Time {
	local := now.In(c.location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, c.location).UTC()
}

func (c *ResetClock) NextDailyReset(now time.Time) time.Time {
	return c.CurrentDailyReset(now).Add(24 * time.Hour)
}

func (c *ResetClock) CurrentWeeklyReset(now time.Time) time.Time {
	local := now.In(c.location)
	startOfDay := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, c.location)
	offset := (int(startOfDay.Weekday()) + 6) % 7
	return startOfDay.AddDate(0, 0, -offset).UTC()
}

func (c *ResetClock) NextWeeklyReset(now time.Time) time.Time {
	return c.CurrentWeeklyReset(now).AddDate(0, 0, 7)
}

func (c *ResetClock) CurrentMonthlyReset(now time.Time) time.Time {
	local := now.In(c.location)
	return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, c.location).UTC()
}

func (c *ResetClock) NextMonthlyReset(now time.Time) time.Time {
	current := c.CurrentMonthlyReset(now).In(c.location)
	return current.AddDate(0, 1, 0).UTC()
}

func (c *ResetClock) CurrentMonthKey(now time.Time) uint32 {
	local := now.In(c.location)
	return uint32(local.Year()*100 + int(local.Month()))
}
