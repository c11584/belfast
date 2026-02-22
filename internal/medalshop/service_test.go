package medalshop

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
)

func TestBuildGoods(t *testing.T) {
	config := &Config{GoodsIDs: []uint32{1, 2}, PurchaseLimit: map[uint32]uint32{2: 5}}
	goods := buildGoods(10, config)
	if len(goods) != 2 {
		t.Fatalf("expected 2 goods")
	}
	if goods[0].Count != 1 {
		t.Fatalf("expected default count 1")
	}
	if goods[1].Count != 5 {
		t.Fatalf("expected configured count 5")
	}
}

func TestNextDailyReset(t *testing.T) {
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	reset := nextMonthlyReset(now)
	if reset <= uint32(now.Unix()) {
		t.Fatalf("expected reset in the future")
	}
	resetAt := time.Unix(int64(reset), 0).UTC()
	if resetAt.Day() != 1 {
		t.Fatalf("expected reset to land on first day of month, got %v", resetAt)
	}
}

func TestNextDailyResetAtExactMonthBoundary(t *testing.T) {
	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	reset := nextMonthlyReset(now)
	resetAt := time.Unix(int64(reset), 0).UTC()
	if !resetAt.After(now) {
		t.Fatalf("expected reset after boundary instant, got %v", resetAt)
	}
	if resetAt.Day() != 1 {
		t.Fatalf("expected monthly boundary to land on day 1, got %v", resetAt)
	}
}

func TestSelectMonthTemplateAcceptsEmptySingleTemplate(t *testing.T) {
	payload, err := json.Marshal(monthShopTemplate{ID: 5, HonorMedalShopGoods: []uint32{}})
	if err != nil {
		t.Fatalf("marshal template: %v", err)
	}

	goods, err := selectMonthTemplate([]orm.ConfigEntry{{Data: payload}}, time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("select month template: %v", err)
	}
	if goods == nil {
		t.Fatalf("expected empty list, got nil")
	}
	if len(goods) != 0 {
		t.Fatalf("expected empty goods list, got %d", len(goods))
	}
}
