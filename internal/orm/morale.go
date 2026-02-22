package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

const (
	moraleTickSeconds        = uint32(6 * 60)
	moraleBaseCap            = uint32(119)
	moraleMarriageCapBonus   = uint32(10)
	moraleDormCap            = uint32(150)
	moraleOnsenCap           = uint32(150)
	moraleBaseTickGain       = uint32(1)
	moraleOnsenExtraTickGain = uint32(1)

	shipStateDormRest     = uint32(2)
	shipStateDormTraining = uint32(5)
	shipStateOnsen        = uint32(6)
)

func ApplyCommanderMoraleRecovery(commanderID uint32, nowUnix uint32) (uint32, error) {
	if nowUnix == 0 {
		return 0, nil
	}

	activeEventCount, err := GetActiveEventCount(nil, commanderID)
	if err != nil {
		return 0, err
	}
	onsenEnabled := activeEventCount > 0

	rows, err := db.DefaultStore.Pool.Query(context.Background(), `
SELECT id, energy, state, state_info1, propose
FROM owned_ships
WHERE owner_id = $1
  AND deleted_at IS NULL
`, int64(commanderID))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type moraleShipRow struct {
		id       uint32
		energy   uint32
		state    uint32
		anchor   uint32
		proposed bool
	}

	ships := make([]moraleShipRow, 0)
	for rows.Next() {
		var row moraleShipRow
		if err := rows.Scan(&row.id, &row.energy, &row.state, &row.anchor, &row.proposed); err != nil {
			return 0, err
		}
		ships = append(ships, row)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	var nextTick uint32
	ctx := context.Background()
	for i := range ships {
		ship := ships[i]
		anchor := ship.anchor
		if anchor == 0 {
			anchor = nowUnix
		}

		updatedEnergy := ship.energy
		updatedAnchor := anchor
		ticks := uint32(0)
		if nowUnix > anchor {
			ticks = (nowUnix - anchor) / moraleTickSeconds
			if ticks > 0 {
				updatedAnchor = anchor + ticks*moraleTickSeconds
				gain, cap := moraleRecoveryProfile(ship.state, ship.proposed, onsenEnabled)
				if updatedEnergy < cap && gain > 0 {
					maxGain := cap - updatedEnergy
					totalGain := ticks * gain
					if totalGain > maxGain {
						totalGain = maxGain
					}
					updatedEnergy += totalGain
				}
			}
		}

		if updatedEnergy != ship.energy || updatedAnchor != ship.anchor {
			_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE owned_ships
SET energy = $3,
    state_info1 = $4
WHERE owner_id = $1
  AND id = $2
  AND deleted_at IS NULL
`, int64(commanderID), int64(ship.id), int64(updatedEnergy), int64(updatedAnchor))
			if err != nil {
				return 0, err
			}
		}

		_, cap := moraleRecoveryProfile(ship.state, ship.proposed, onsenEnabled)
		if updatedEnergy < cap {
			candidate := updatedAnchor + moraleTickSeconds
			if nextTick == 0 || candidate < nextTick {
				nextTick = candidate
			}
		}
	}

	return nextTick, nil
}

func moraleRecoveryProfile(state uint32, proposed bool, onsenEnabled bool) (uint32, uint32) {
	cap := moraleBaseCap
	gain := moraleBaseTickGain

	if isDormShipState(state) {
		cap = moraleDormCap
	}
	if proposed {
		cap += moraleMarriageCapBonus
		if cap > moraleDormCap {
			cap = moraleDormCap
		}
	}
	if onsenEnabled && state == shipStateOnsen {
		cap = moraleOnsenCap
		gain += moraleOnsenExtraTickGain
	}

	return gain, cap
}

func isDormShipState(state uint32) bool {
	return state == shipStateDormRest || state == shipStateDormTraining
}
