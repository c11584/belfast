package shopreset

import (
	"time"

	"github.com/ggmolly/belfast/internal/region"
)

type Window struct {
	Start time.Time
	End   time.Time
	Key   uint32
}

func DailyWindow(now time.Time) (Window, error) {
	location, err := currentRegionLocation()
	if err != nil {
		return Window{}, err
	}
	local := now.In(location)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location).UTC()
	end := start.Add(24 * time.Hour)
	return Window{Start: start, End: end, Key: uint32(start.Unix())}, nil
}

func MonthlyWindow(now time.Time) (Window, error) {
	location, err := currentRegionLocation()
	if err != nil {
		return Window{}, err
	}
	local := now.In(location)
	start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, location).UTC()
	end := start.In(location).AddDate(0, 1, 0).UTC()
	key := uint32(local.Year()*100 + int(local.Month()))
	return Window{Start: start, End: end, Key: key}, nil
}

func DeterministicSeed(commanderID uint32, parts ...uint32) uint64 {
	seed := uint64(1469598103934665603)
	mix := func(v uint64) {
		seed ^= v
		seed *= 1099511628211
	}
	mix(uint64(commanderID))
	for _, part := range parts {
		mix(uint64(part))
	}
	return seed
}

func currentRegionLocation() (*time.Location, error) {
	regionID := region.Current()
	offset, ok := map[string]int{
		"CN": 8 * 60 * 60,
		"EN": -7 * 60 * 60,
		"JP": 9 * 60 * 60,
		"KR": 9 * 60 * 60,
		"TW": 8 * 60 * 60,
	}[regionID]
	if !ok {
		return time.UTC, nil
	}
	return time.FixedZone(regionID, offset), nil
}
