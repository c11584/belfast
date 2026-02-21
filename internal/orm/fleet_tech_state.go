package orm

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

const fleetTechStateCategory = "runtime/fleet_tech_state"

type FleetTechGroupState struct {
	GroupID         uint32 `json:"group_id"`
	EffectTechID    uint32 `json:"effect_tech_id"`
	StudyTechID     uint32 `json:"study_tech_id"`
	StudyFinishTime uint32 `json:"study_finish_time"`
	RewardedTechID  uint32 `json:"rewarded_tech_id"`
}

type FleetTechAttrOverride struct {
	ShipType uint32 `json:"ship_type"`
	AttrType uint32 `json:"attr_type"`
	SetValue uint32 `json:"set_value"`
}

type CommanderFleetTechState struct {
	CommanderID   uint32                  `json:"commander_id"`
	Groups        []FleetTechGroupState   `json:"groups"`
	AttrOverrides []FleetTechAttrOverride `json:"attr_overrides"`
}

func GetCommanderFleetTechState(commanderID uint32) (*CommanderFleetTechState, error) {
	entry, err := GetConfigEntry(fleetTechStateCategory, strconv.FormatUint(uint64(commanderID), 10))
	if err != nil {
		return nil, err
	}
	state := &CommanderFleetTechState{}
	if len(entry.Data) != 0 {
		if err := json.Unmarshal(entry.Data, state); err != nil {
			return nil, err
		}
	}
	state.ensureDefaults(commanderID)
	return state, nil
}

func GetOrCreateCommanderFleetTechState(commanderID uint32) (*CommanderFleetTechState, error) {
	state, err := GetCommanderFleetTechState(commanderID)
	if err == nil {
		return state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}
	state = &CommanderFleetTechState{CommanderID: commanderID, Groups: []FleetTechGroupState{}, AttrOverrides: []FleetTechAttrOverride{}}
	if err := SaveCommanderFleetTechState(state); err != nil {
		return nil, err
	}
	return state, nil
}

func SaveCommanderFleetTechState(state *CommanderFleetTechState) error {
	if state == nil {
		return nil
	}
	state.ensureDefaults(state.CommanderID)
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return UpsertConfigEntry(fleetTechStateCategory, strconv.FormatUint(uint64(state.CommanderID), 10), payload)
}

func GetOrCreateCommanderFleetTechStateTx(ctx context.Context, tx pgx.Tx, commanderID uint32) (*CommanderFleetTechState, error) {
	state, err := GetCommanderFleetTechStateTx(ctx, tx, commanderID)
	if err == nil {
		return state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}
	state = &CommanderFleetTechState{CommanderID: commanderID, Groups: []FleetTechGroupState{}, AttrOverrides: []FleetTechAttrOverride{}}
	if err := SaveCommanderFleetTechStateTx(ctx, tx, state); err != nil {
		return nil, err
	}
	return state, nil
}

func GetCommanderFleetTechStateTx(ctx context.Context, tx pgx.Tx, commanderID uint32) (*CommanderFleetTechState, error) {
	category := fleetTechStateCategory
	key := strconv.FormatUint(uint64(commanderID), 10)
	var payload []byte
	err := tx.QueryRow(ctx, `
SELECT data
FROM config_entries
WHERE category = $1 AND key = $2
FOR UPDATE
`, category, key).Scan(&payload)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	state := &CommanderFleetTechState{}
	if len(payload) != 0 {
		if err := json.Unmarshal(payload, state); err != nil {
			return nil, err
		}
	}
	state.ensureDefaults(commanderID)
	return state, nil
}

func SaveCommanderFleetTechStateTx(ctx context.Context, tx pgx.Tx, state *CommanderFleetTechState) error {
	if state == nil {
		return nil
	}
	state.ensureDefaults(state.CommanderID)
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO config_entries (category, key, data)
VALUES ($1, $2, $3)
ON CONFLICT (category, key)
DO UPDATE SET data = EXCLUDED.data
`, fleetTechStateCategory, strconv.FormatUint(uint64(state.CommanderID), 10), payload)
	return err
}

func (state *CommanderFleetTechState) ensureDefaults(commanderID uint32) {
	if state.CommanderID == 0 {
		state.CommanderID = commanderID
	}
	if state.Groups == nil {
		state.Groups = []FleetTechGroupState{}
	}
	if state.AttrOverrides == nil {
		state.AttrOverrides = []FleetTechAttrOverride{}
	}
	sort.Slice(state.Groups, func(i int, j int) bool {
		return state.Groups[i].GroupID < state.Groups[j].GroupID
	})
	sort.Slice(state.AttrOverrides, func(i int, j int) bool {
		if state.AttrOverrides[i].ShipType != state.AttrOverrides[j].ShipType {
			return state.AttrOverrides[i].ShipType < state.AttrOverrides[j].ShipType
		}
		return state.AttrOverrides[i].AttrType < state.AttrOverrides[j].AttrType
	})
}

func (state *CommanderFleetTechState) UpsertGroup(groupID uint32) *FleetTechGroupState {
	for i := range state.Groups {
		if state.Groups[i].GroupID == groupID {
			return &state.Groups[i]
		}
	}
	state.Groups = append(state.Groups, FleetTechGroupState{GroupID: groupID})
	state.ensureDefaults(state.CommanderID)
	for i := range state.Groups {
		if state.Groups[i].GroupID == groupID {
			return &state.Groups[i]
		}
	}
	return nil
}

func (state *CommanderFleetTechState) GetGroup(groupID uint32) (*FleetTechGroupState, bool) {
	for i := range state.Groups {
		if state.Groups[i].GroupID == groupID {
			return &state.Groups[i], true
		}
	}
	return nil, false
}

func (state *CommanderFleetTechState) SetAttrOverrides(overrides []FleetTechAttrOverride) {
	if overrides == nil {
		overrides = []FleetTechAttrOverride{}
	}
	state.AttrOverrides = overrides
	state.ensureDefaults(state.CommanderID)
}
