package orm

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestDeleteIslandAgoraThemeScopeAndSlotReuse(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &IslandAgoraTheme{})
	clearTable(t, &Commander{})

	if err := CreateCommanderRoot(9401, 9401, "Island Theme Commander A", 0, 0); err != nil {
		t.Fatalf("create commander A: %v", err)
	}
	if err := CreateCommanderRoot(9402, 9402, "Island Theme Commander B", 0, 0); err != nil {
		t.Fatalf("create commander B: %v", err)
	}

	ctx := context.Background()
	err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		if err := UpsertIslandAgoraThemeTx(ctx, tx, 9401, 1, "A-1", []byte{1}); err != nil {
			return err
		}
		if err := UpsertIslandAgoraThemeTx(ctx, tx, 9401, 2, "A-2", []byte{2}); err != nil {
			return err
		}
		if err := UpsertIslandAgoraThemeTx(ctx, tx, 9402, 1, "B-1", []byte{3}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed themes: %v", err)
	}

	err = WithPGXTx(ctx, func(tx pgx.Tx) error {
		return DeleteIslandAgoraThemeTx(ctx, tx, 9401, 1)
	})
	if err != nil {
		t.Fatalf("delete theme: %v", err)
	}

	themesA, err := ListIslandAgoraThemes(9401)
	if err != nil {
		t.Fatalf("list commander A themes: %v", err)
	}
	if len(themesA) != 1 || themesA[0].ThemeSlotID != 2 {
		t.Fatalf("expected commander A to keep only slot 2, got %+v", themesA)
	}

	themesB, err := ListIslandAgoraThemes(9402)
	if err != nil {
		t.Fatalf("list commander B themes: %v", err)
	}
	if len(themesB) != 1 || themesB[0].ThemeSlotID != 1 {
		t.Fatalf("expected commander B slot 1 untouched, got %+v", themesB)
	}

	err = WithPGXTx(ctx, func(tx pgx.Tx) error {
		return UpsertIslandAgoraThemeTx(ctx, tx, 9401, 1, "A-1-reused", []byte{9})
	})
	if err != nil {
		t.Fatalf("reuse slot after delete: %v", err)
	}

	themesA, err = ListIslandAgoraThemes(9401)
	if err != nil {
		t.Fatalf("list commander A themes after reuse: %v", err)
	}
	if len(themesA) != 2 {
		t.Fatalf("expected two themes after reuse, got %d", len(themesA))
	}
	if themesA[0].ThemeSlotID != 1 || themesA[0].Name != "A-1-reused" {
		t.Fatalf("expected reused slot 1 to be present first, got %+v", themesA[0])
	}
}
