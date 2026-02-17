package orm

import (
	"context"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/jackc/pgx/v5"
)

func TestLoadIslandTaskProgressCreatesDefault(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &IslandTaskProgress{})
	clearTable(t, &Commander{})

	if err := CreateCommanderRoot(9301, 9301, "Island Default", 0, 0); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	now := time.Date(2026, time.February, 14, 15, 0, 0, 0, time.UTC)
	state, err := LoadIslandTaskProgress(9301, now)
	if err != nil {
		t.Fatalf("load island state: %v", err)
	}

	if state.WeekStartUnix != CurrentWeeklyResetUnix(now) {
		t.Fatalf("expected week start %d, got %d", CurrentWeeklyResetUnix(now), state.WeekStartUnix)
	}
	if state.LastRefreshDayUnix != CurrentDayResetUnix(now) {
		t.Fatalf("expected day start %d, got %d", CurrentDayResetUnix(now), state.LastRefreshDayUnix)
	}
	if state.WeekDailyTaskNum != 0 || len(state.ActiveTasks) != 0 || len(state.FinishedTaskIDs) != 0 || len(state.RandomTaskWindows) != 0 {
		t.Fatalf("expected empty default state, got %+v", state)
	}
}

func TestLoadIslandTaskProgressResetsDailyCounter(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &IslandTaskProgress{})
	clearTable(t, &Commander{})

	if err := CreateCommanderRoot(9302, 9302, "Island Day", 0, 0); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	firstDay := time.Date(2026, time.February, 14, 8, 0, 0, 0, time.UTC)
	err := db.DefaultStore.WithPGXTx(context.Background(), func(tx pgx.Tx) error {
		state, err := LoadIslandTaskProgressForUpdateTx(context.Background(), tx, 9302, firstDay)
		if err != nil {
			return err
		}
		state.WeekDailyTaskNum = 7
		state.ActiveTasks = []IslandTaskEntry{{TaskID: 40101001, Timestamp: uint32(firstDay.Unix())}}
		state.FinishedTaskIDs = []uint32{40102001}
		state.RandomTaskWindows = []IslandTaskEntry{{TaskID: 40103001, Timestamp: uint32(firstDay.Unix()) + 50}}
		state.FutureTaskWindows = state.RandomTaskWindows
		return SaveIslandTaskProgressTx(context.Background(), tx, state)
	})
	if err != nil {
		t.Fatalf("seed island state: %v", err)
	}

	nextDay := firstDay.Add(24 * time.Hour)
	state, err := LoadIslandTaskProgress(9302, nextDay)
	if err != nil {
		t.Fatalf("load next day island state: %v", err)
	}
	if state.LastRefreshDayUnix != CurrentDayResetUnix(nextDay) {
		t.Fatalf("expected day reset %d, got %d", CurrentDayResetUnix(nextDay), state.LastRefreshDayUnix)
	}
	if state.WeekDailyTaskNum != 0 {
		t.Fatalf("expected daily counter reset, got %d", state.WeekDailyTaskNum)
	}
	if len(state.ActiveTasks) != 1 || state.ActiveTasks[0].TaskID != 40101001 {
		t.Fatalf("expected active tasks preserved, got %+v", state.ActiveTasks)
	}
}

func TestWithIslandTaskProgressTxPersistsOrderedState(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &IslandTaskProgress{})
	clearTable(t, &Commander{})

	if err := CreateCommanderRoot(9303, 9303, "Island Save", 0, 0); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	now := time.Date(2026, time.February, 15, 9, 0, 0, 0, time.UTC)
	err := WithIslandTaskProgressTx(9303, now, func(state *IslandTaskProgress) error {
		state.WeekDailyTaskNum = 2
		state.TraceTaskID = 40104001
		state.TraceDailyTaskID = 40105001
		state.ActiveTasks = []IslandTaskEntry{{TaskID: 40105001, Timestamp: 20}, {TaskID: 40104001, Timestamp: 10}}
		state.FinishedTaskIDs = []uint32{40107001, 40106001}
		state.RandomTaskWindows = []IslandTaskEntry{{TaskID: 40109001, Timestamp: 200}, {TaskID: 40108001, Timestamp: 100}}
		state.FutureTaskWindows = append([]IslandTaskEntry{}, state.RandomTaskWindows...)
		return nil
	})
	if err != nil {
		t.Fatalf("update island state: %v", err)
	}

	state, err := LoadIslandTaskProgress(9303, now)
	if err != nil {
		t.Fatalf("load island state: %v", err)
	}
	if state.WeekDailyTaskNum != 2 || state.TraceTaskID != 40104001 || state.TraceDailyTaskID != 40105001 {
		t.Fatalf("unexpected scalar state: %+v", state)
	}
	if len(state.ActiveTasks) != 2 || state.ActiveTasks[0].TaskID != 40104001 || state.ActiveTasks[1].TaskID != 40105001 {
		t.Fatalf("expected sorted active tasks, got %+v", state.ActiveTasks)
	}
	if len(state.FinishedTaskIDs) != 2 || state.FinishedTaskIDs[0] != 40106001 || state.FinishedTaskIDs[1] != 40107001 {
		t.Fatalf("expected sorted finished ids, got %+v", state.FinishedTaskIDs)
	}
	if len(state.RandomTaskWindows) != 2 || state.RandomTaskWindows[0].TaskID != 40108001 || state.RandomTaskWindows[1].TaskID != 40109001 {
		t.Fatalf("expected sorted random windows, got %+v", state.RandomTaskWindows)
	}
}
