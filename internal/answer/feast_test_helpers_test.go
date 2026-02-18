package answer_test

import (
	"fmt"
	"testing"
	"time"
)

func cleanupFeastData(t *testing.T, commanderID uint32, actID uint32) {
	t.Helper()
	execAnswerExternalTestSQLT(t, "DELETE FROM feast_states WHERE commander_id = $1 AND act_id = $2", int64(commanderID), int64(actID))
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries WHERE category = $1 AND key = $2", "ShareCfg/activity_template.json", fmt.Sprintf("%d", actID))
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
}

func seedFeastActivityTemplate(t *testing.T, actID uint32, end time.Time, configData string) {
	t.Helper()
	start := end.Add(-24 * time.Hour).UTC()
	end = end.UTC()
	if configData == "" {
		configData = "[]"
	}
	payload := fmt.Sprintf(`{"id":%d,"type":999,"config_data":%s,"time":["timer",[[%d,%d,%d],[%d,%d,%d]],[[%d,%d,%d],[%d,%d,%d]]]}`,
		actID,
		configData,
		start.Year(), int(start.Month()), start.Day(), start.Hour(), start.Minute(), start.Second(),
		end.Year(), int(end.Month()), end.Day(), end.Hour(), end.Minute(), end.Second(),
	)
	seedConfigEntry(t, "ShareCfg/activity_template.json", fmt.Sprintf("%d", actID), payload)
}
