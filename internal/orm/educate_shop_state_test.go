package orm

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

func TestEducateShopStateRoundTripAndList(t *testing.T) {
	os.Setenv("MODE", "test")
	InitDatabase()
	clearTable(t, &EducateShopState{})
	clearTable(t, &Commander{})
	if err := CreateCommanderRoot(9201, 9201, "educate shop orm", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}

	state := &EducateShopState{
		CommanderID: 9201,
		ShopID:      2,
		RefreshKey:  12,
		Goods: []EducateShopGoodsState{
			{ID: 11, Num: 1},
			{ID: 12, Num: 0},
		},
	}
	if err := UpsertEducateShopState(state); err != nil {
		t.Fatalf("upsert state: %v", err)
	}

	err := db.DefaultStore.WithPGXTx(context.Background(), func(tx pgx.Tx) error {
		loaded, err := GetEducateShopStateTx(context.Background(), tx, 9201, 2)
		if err != nil {
			return err
		}
		if loaded.RefreshKey != 12 || len(loaded.Goods) != 2 || loaded.Goods[1].Num != 0 {
			t.Fatalf("unexpected loaded state: %+v", loaded)
		}
		loaded.Goods[0].Num = 0
		return UpsertEducateShopStateTx(context.Background(), tx, loaded)
	})
	if err != nil {
		t.Fatalf("tx roundtrip: %v", err)
	}

	rows, err := ListEducateShopStates(9201)
	if err != nil {
		t.Fatalf("list states: %v", err)
	}
	if len(rows) != 1 || rows[0].Goods[0].Num != 0 {
		t.Fatalf("unexpected listed states: %+v", rows)
	}
}
