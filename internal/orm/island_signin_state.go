package orm

import (
	"context"
	"encoding/json"

	"github.com/ggmolly/belfast/internal/db"
)

type IslandSignInState struct {
	CommanderID        uint32
	DayStartUnix       uint32
	SignedIn           bool
	ExternalClaimCount uint32
	ClaimedSlots       []string
}

func (IslandSignInState) TableName() string {
	return "island_signin_states"
}

func GetIslandSignInState(commanderID uint32) (*IslandSignInState, error) {
	var (
		commanderIDRaw        int64
		dayStartUnixRaw       int64
		externalClaimCountRaw int64
		claimedSlotsJSON      []byte
		state                 IslandSignInState
	)
	err := db.DefaultStore.Pool.QueryRow(context.Background(), `
SELECT commander_id, day_start_unix, signed_in, external_claim_count, claimed_slots
FROM island_signin_states
WHERE commander_id = $1
`, int64(commanderID)).Scan(
		&commanderIDRaw,
		&dayStartUnixRaw,
		&state.SignedIn,
		&externalClaimCountRaw,
		&claimedSlotsJSON,
	)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(claimedSlotsJSON, &state.ClaimedSlots); err != nil {
		return nil, err
	}
	if state.ClaimedSlots == nil {
		state.ClaimedSlots = []string{}
	}
	state.CommanderID = uint32(commanderIDRaw)
	state.DayStartUnix = uint32(dayStartUnixRaw)
	state.ExternalClaimCount = uint32(externalClaimCountRaw)
	return &state, nil
}

func UpsertIslandSignInState(state *IslandSignInState) error {
	claimedSlots := state.ClaimedSlots
	if claimedSlots == nil {
		claimedSlots = []string{}
	}
	claimedSlotsJSON, err := json.Marshal(claimedSlots)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO island_signin_states (commander_id, day_start_unix, signed_in, external_claim_count, claimed_slots)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (commander_id)
DO UPDATE SET
	day_start_unix = EXCLUDED.day_start_unix,
	signed_in = EXCLUDED.signed_in,
	external_claim_count = EXCLUDED.external_claim_count,
	claimed_slots = EXCLUDED.claimed_slots,
	updated_at = CURRENT_TIMESTAMP
`, int64(state.CommanderID), int64(state.DayStartUnix), state.SignedIn, int64(state.ExternalClaimCount), claimedSlotsJSON)
	return err
}
