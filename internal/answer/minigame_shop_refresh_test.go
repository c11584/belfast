package answer_test

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestMiniGameShopRefreshSuccess(t *testing.T) {
	commanderID := uint32(9005)
	cleanupMiniGameShopData(t, commanderID)
	now := time.Now().UTC()
	seedMiniGameShopConfig(t, []miniGameShopEntry{{
		ID:                 12,
		GoodsPurchaseLimit: 2,
		Goods:              []uint32{1},
		DropType:           1,
		Price:              2,
		Num:                1,
		Order:              1,
		Time:               [][][3]int{{{now.Year() - 1, int(now.Month()), now.Day()}, {now.Year() + 1, int(now.Month()), now.Day()}}},
	}})
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupMiniGameShopData(t, commanderID)

	payload := &protobuf.CS_26154{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.MiniGameShopRefresh(&buf, client); err != nil {
		t.Fatalf("MiniGameShopRefresh failed: %v", err)
	}

	response := &protobuf.SC_26155{}
	decodeTestPacket(t, client, 26155, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}
	if len(response.GetNextFlashTime()) == 0 || response.GetNextFlashTime()[0] <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next flash time in future")
	}
}

func TestMiniGameShopRefreshRejectsUnsupportedType(t *testing.T) {
	commanderID := uint32(9006)
	cleanupMiniGameShopData(t, commanderID)
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupMiniGameShopData(t, commanderID)

	payload := &protobuf.CS_26154{Type: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.MiniGameShopRefresh(&buf, client); err != nil {
		t.Fatalf("MiniGameShopRefresh failed: %v", err)
	}

	response := &protobuf.SC_26155{}
	decodeTestPacket(t, client, 26155, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result for unsupported type")
	}
	if len(response.GetNextFlashTime()) != 0 {
		t.Fatalf("expected empty next flash time on failure")
	}
}
