package orm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ggmolly/belfast/internal/db"
)

type IslandTechnologyState struct {
	CommanderID     uint32
	UnlockedTechIDs []uint32
	AbilityIDs      []uint32
	FinishCounts    map[uint32]uint32
}

func (IslandTechnologyState) TableName() string {
	return "island_technology_states"
}

func GetIslandTechnologyState(commanderID uint32) (*IslandTechnologyState, error) {
	var (
		commanderIDRaw int64
		unlockedJSON   []byte
		abilityJSON    []byte
		finishJSON     []byte
		state          IslandTechnologyState
	)
	err := db.DefaultStore.Pool.QueryRow(context.Background(), `
SELECT commander_id, unlocked_tech_ids, ability_ids, finish_counts
FROM island_technology_states
WHERE commander_id = $1
`, int64(commanderID)).Scan(&commanderIDRaw, &unlockedJSON, &abilityJSON, &finishJSON)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(unlockedJSON, &state.UnlockedTechIDs); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(abilityJSON, &state.AbilityIDs); err != nil {
		return nil, err
	}
	finishCounts := map[string]uint32{}
	if err := json.Unmarshal(finishJSON, &finishCounts); err != nil {
		return nil, err
	}
	state.FinishCounts = make(map[uint32]uint32, len(finishCounts))
	for key, value := range finishCounts {
		var parsed uint32
		if _, err := fmt.Sscanf(key, "%d", &parsed); err == nil {
			state.FinishCounts[parsed] = value
		}
	}
	if state.UnlockedTechIDs == nil {
		state.UnlockedTechIDs = []uint32{}
	}
	if state.AbilityIDs == nil {
		state.AbilityIDs = []uint32{}
	}
	if state.FinishCounts == nil {
		state.FinishCounts = map[uint32]uint32{}
	}
	state.CommanderID = uint32(commanderIDRaw)
	return &state, nil
}

func NewIslandTechnologyState(commanderID uint32) *IslandTechnologyState {
	return &IslandTechnologyState{
		CommanderID:     commanderID,
		UnlockedTechIDs: []uint32{},
		AbilityIDs:      []uint32{},
		FinishCounts:    map[uint32]uint32{},
	}
}

func UpsertIslandTechnologyState(state *IslandTechnologyState) error {
	if state.UnlockedTechIDs == nil {
		state.UnlockedTechIDs = []uint32{}
	}
	if state.AbilityIDs == nil {
		state.AbilityIDs = []uint32{}
	}
	if state.FinishCounts == nil {
		state.FinishCounts = map[uint32]uint32{}
	}
	unlockedJSON, err := json.Marshal(state.UnlockedTechIDs)
	if err != nil {
		return err
	}
	abilityJSON, err := json.Marshal(state.AbilityIDs)
	if err != nil {
		return err
	}
	finishCounts := make(map[string]uint32, len(state.FinishCounts))
	for key, value := range state.FinishCounts {
		finishCounts[fmt.Sprintf("%d", key)] = value
	}
	finishJSON, err := json.Marshal(finishCounts)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO island_technology_states (commander_id, unlocked_tech_ids, ability_ids, finish_counts)
VALUES ($1, $2, $3, $4)
ON CONFLICT (commander_id)
DO UPDATE SET
	unlocked_tech_ids = EXCLUDED.unlocked_tech_ids,
	ability_ids = EXCLUDED.ability_ids,
	finish_counts = EXCLUDED.finish_counts,
	updated_at = CURRENT_TIMESTAMP
`, int64(state.CommanderID), unlockedJSON, abilityJSON, finishJSON)
	return err
}
