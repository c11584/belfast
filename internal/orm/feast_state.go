package orm

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

type FeastPartyRole struct {
	Tid          uint32 `json:"tid"`
	Bubble       uint32 `json:"bubble"`
	SpeechBubble uint32 `json:"speech_bubble"`
}

type FeastSpecialRole struct {
	Tid   uint32 `json:"tid"`
	State uint32 `json:"state"`
	Gift  uint32 `json:"gift"`
}

type FeastState struct {
	CommanderID  uint32
	ActID        uint32
	RefreshTime  uint32
	PartyRoles   []FeastPartyRole
	SpecialRoles []FeastSpecialRole
}

func GetFeastState(commanderID uint32, actID uint32) (*FeastState, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, act_id, refresh_time, party_roles, special_roles
FROM feast_states
WHERE commander_id = $1
  AND act_id = $2
`, int64(commanderID), int64(actID))

	state := &FeastState{}
	var partyRaw []byte
	var specialRaw []byte
	err := row.Scan(&state.CommanderID, &state.ActID, &state.RefreshTime, &partyRaw, &specialRaw)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(partyRaw, &state.PartyRoles); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(specialRaw, &state.SpecialRoles); err != nil {
		return nil, err
	}
	if state.PartyRoles == nil {
		state.PartyRoles = []FeastPartyRole{}
	}
	if state.SpecialRoles == nil {
		state.SpecialRoles = []FeastSpecialRole{}
	}
	return state, nil
}

func GetOrCreateFeastState(commanderID uint32, actID uint32) (*FeastState, error) {
	state, err := GetFeastState(commanderID, actID)
	if err == nil {
		return state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}
	state = &FeastState{
		CommanderID:  commanderID,
		ActID:        actID,
		RefreshTime:  0,
		PartyRoles:   []FeastPartyRole{},
		SpecialRoles: []FeastSpecialRole{},
	}
	if err := SaveFeastState(state); err != nil {
		return nil, err
	}
	return GetFeastState(commanderID, actID)
}

func SaveFeastState(state *FeastState) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		return SaveFeastStateTx(ctx, tx, state)
	})
}

func SaveFeastStateTx(ctx context.Context, tx pgx.Tx, state *FeastState) error {
	if state == nil {
		return errors.New("feast state is nil")
	}
	if state.PartyRoles == nil {
		state.PartyRoles = []FeastPartyRole{}
	}
	if state.SpecialRoles == nil {
		state.SpecialRoles = []FeastSpecialRole{}
	}
	partyRaw, err := json.Marshal(state.PartyRoles)
	if err != nil {
		return err
	}
	specialRaw, err := json.Marshal(state.SpecialRoles)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO feast_states (commander_id, act_id, refresh_time, party_roles, special_roles)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (commander_id, act_id)
DO UPDATE SET
  refresh_time = EXCLUDED.refresh_time,
  party_roles = EXCLUDED.party_roles,
  special_roles = EXCLUDED.special_roles,
  updated_at = NOW()
`, int64(state.CommanderID), int64(state.ActID), int64(state.RefreshTime), partyRaw, specialRaw)
	return err
}
