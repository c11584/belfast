package orm

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ggmolly/belfast/internal/db"
)

type EducateShopGoodsState struct {
	ID  uint32 `json:"id"`
	Num uint32 `json:"num"`
}

type EducateShopState struct {
	CommanderID uint32
	ShopID      uint32
	RefreshKey  uint32
	Goods       []EducateShopGoodsState
}

func (EducateShopState) TableName() string {
	return "educate_shop_states"
}

func GetEducateShopStateTx(ctx context.Context, tx pgx.Tx, commanderID uint32, shopID uint32) (*EducateShopState, error) {
	state := &EducateShopState{}
	var goodsJSON []byte
	err := tx.QueryRow(ctx, `
SELECT commander_id, shop_id, refresh_key, goods
FROM educate_shop_states
WHERE commander_id = $1 AND shop_id = $2
FOR UPDATE
`, int64(commanderID), int64(shopID)).Scan(&state.CommanderID, &state.ShopID, &state.RefreshKey, &goodsJSON)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(goodsJSON, &state.Goods); err != nil {
		return nil, err
	}
	if state.Goods == nil {
		state.Goods = []EducateShopGoodsState{}
	}
	return state, nil
}

func UpsertEducateShopStateTx(ctx context.Context, tx pgx.Tx, state *EducateShopState) error {
	return upsertEducateShopState(ctx, tx, state)
}

func UpsertEducateShopState(state *EducateShopState) error {
	return upsertEducateShopState(context.Background(), db.DefaultStore.Pool, state)
}

func ListEducateShopStates(commanderID uint32) ([]EducateShopState, error) {
	rows, err := db.DefaultStore.Pool.Query(context.Background(), `
SELECT commander_id, shop_id, refresh_key, goods
FROM educate_shop_states
WHERE commander_id = $1
ORDER BY shop_id
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	states := []EducateShopState{}
	for rows.Next() {
		var state EducateShopState
		var goodsJSON []byte
		if err := rows.Scan(&state.CommanderID, &state.ShopID, &state.RefreshKey, &goodsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(goodsJSON, &state.Goods); err != nil {
			return nil, err
		}
		if state.Goods == nil {
			state.Goods = []EducateShopGoodsState{}
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return states, nil
}

type educateShopStateExecer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func upsertEducateShopState(ctx context.Context, execer educateShopStateExecer, state *EducateShopState) error {
	goods := state.Goods
	if goods == nil {
		goods = []EducateShopGoodsState{}
	}
	goodsJSON, err := json.Marshal(goods)
	if err != nil {
		return err
	}
	_, err = execer.Exec(ctx, `
INSERT INTO educate_shop_states (commander_id, shop_id, refresh_key, goods)
VALUES ($1, $2, $3, $4)
ON CONFLICT (commander_id, shop_id)
DO UPDATE SET
	refresh_key = EXCLUDED.refresh_key,
	goods = EXCLUDED.goods,
	updated_at = CURRENT_TIMESTAMP
`, int64(state.CommanderID), int64(state.ShopID), int64(state.RefreshKey), goodsJSON)
	return err
}
