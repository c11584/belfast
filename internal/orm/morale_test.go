package orm

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestApplyCommanderMoraleRecoverySixMinuteTick(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Commander{})

	seedMoraleCommanderAndShip(t, 7001, 8001, 118, 0, 1, false)

	nextTick, err := ApplyCommanderMoraleRecovery(7001, 7201)
	if err != nil {
		t.Fatalf("apply morale recovery: %v", err)
	}
	if nextTick != 0 {
		t.Fatalf("expected no next tick at cap, got %d", nextTick)
	}

	energy, anchor := loadMoraleShipState(t, 7001, 8001)
	if energy != 119 {
		t.Fatalf("expected recovered energy 119, got %d", energy)
	}
	if anchor != 7201 {
		t.Fatalf("expected updated anchor 7201, got %d", anchor)
	}
}

func TestApplyCommanderMoraleRecoveryMarriageCapBonus(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Commander{})

	seedMoraleCommanderAndShip(t, 7002, 8002, 120, 0, 1, true)

	nextTick, err := ApplyCommanderMoraleRecovery(7002, 3601)
	if err != nil {
		t.Fatalf("apply morale recovery: %v", err)
	}
	if nextTick != 0 {
		t.Fatalf("expected no next tick at marriage cap, got %d", nextTick)
	}

	energy, _ := loadMoraleShipState(t, 7002, 8002)
	if energy != 129 {
		t.Fatalf("expected marriage cap 129, got %d", energy)
	}
}

func TestApplyCommanderMoraleRecoveryOnsenEventBonus(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Commander{})

	seedMoraleCommanderAndShip(t, 7003, 8003, 100, shipStateOnsen, 1, false)
	if err := SaveEventCollection(nil, &EventCollection{CommanderID: 7003, CollectionID: 1, StartTime: 1, FinishTime: 2, ShipIDs: Int64List{}}); err != nil {
		t.Fatalf("seed active event: %v", err)
	}

	nextTick, err := ApplyCommanderMoraleRecovery(7003, moraleTickSeconds*10+1)
	if err != nil {
		t.Fatalf("apply morale recovery: %v", err)
	}
	if nextTick != 3961 {
		t.Fatalf("expected next onsen tick at 3961, got %d", nextTick)
	}

	energy, anchor := loadMoraleShipState(t, 7003, 8003)
	if energy != 120 {
		t.Fatalf("expected onsen bonus recovery to 120, got %d", energy)
	}
	if anchor != moraleTickSeconds*10+1 {
		t.Fatalf("expected updated anchor %d, got %d", moraleTickSeconds*10+1, anchor)
	}
}

func TestApplyCommanderMoraleRecoveryPreservesAboveCap(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Commander{})

	seedMoraleCommanderAndShip(t, 7004, 8004, 150, 0, 1200, false)

	nextTick, err := ApplyCommanderMoraleRecovery(7004, 3600)
	if err != nil {
		t.Fatalf("apply morale recovery: %v", err)
	}
	if nextTick != 0 {
		t.Fatalf("expected no next tick for above-cap ship, got %d", nextTick)
	}

	energy, anchor := loadMoraleShipState(t, 7004, 8004)
	if energy != 150 {
		t.Fatalf("expected above-cap energy to stay 150, got %d", energy)
	}
	if anchor != 3360 {
		t.Fatalf("expected anchor to advance to 3360, got %d", anchor)
	}
}

func TestApplyCommanderMoraleRecoveryAnchorBoundaryAndNextTick(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &EventCollection{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Commander{})

	seedMoraleCommanderAndShip(t, 7005, 8005, 100, 0, 0, false)

	now := moraleTickSeconds - 1
	nextTick, err := ApplyCommanderMoraleRecovery(7005, now)
	if err != nil {
		t.Fatalf("apply morale recovery: %v", err)
	}
	if nextTick != now+moraleTickSeconds {
		t.Fatalf("expected next tick %d, got %d", now+moraleTickSeconds, nextTick)
	}

	energy, anchor := loadMoraleShipState(t, 7005, 8005)
	if energy != 100 {
		t.Fatalf("expected energy to stay at 100 before first full tick, got %d", energy)
	}
	if anchor != now {
		t.Fatalf("expected zero anchor to initialize to now (%d), got %d", now, anchor)
	}
}

func seedMoraleCommanderAndShip(t *testing.T, commanderID uint32, shipID uint32, energy uint32, state uint32, anchor uint32, proposed bool) {
	t.Helper()
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO commanders (commander_id, account_id, name)
VALUES ($1, $2, $3)
`, int64(commanderID), int64(commanderID), "Morale Commander"); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO owned_ships (id, owner_id, ship_id, energy, state, state_info1, propose)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, int64(shipID), int64(commanderID), int64(1), int64(energy), int64(state), int64(anchor), proposed); err != nil {
		t.Fatalf("seed ship: %v", err)
	}
}

func loadMoraleShipState(t *testing.T, commanderID uint32, shipID uint32) (uint32, uint32) {
	t.Helper()
	var energy uint32
	var anchor uint32
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `
SELECT energy, state_info1
FROM owned_ships
WHERE owner_id = $1 AND id = $2 AND deleted_at IS NULL
`, int64(commanderID), int64(shipID)).Scan(&energy, &anchor); err != nil {
		t.Fatalf("load ship state: %v", err)
	}
	return energy, anchor
}
