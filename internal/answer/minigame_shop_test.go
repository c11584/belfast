package answer_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const miniGameShopCategory = "ShareCfg/gameroom_shop_template.json"

type miniGameShopEntry struct {
	ID                 uint32     `json:"id"`
	GoodsPurchaseLimit uint32     `json:"goods_purchase_limit"`
	Goods              []uint32   `json:"goods"`
	DropType           uint32     `json:"drop_type"`
	Price              uint32     `json:"price"`
	Num                uint32     `json:"num"`
	Order              uint32     `json:"order"`
	Time               [][][3]int `json:"time"`
}

func seedMiniGameShopConfig(t *testing.T, entries []miniGameShopEntry) {
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries WHERE category = $1", miniGameShopCategory)
	for _, entry := range entries {
		payload, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("failed to marshal minigame shop entry: %v", err)
		}
		execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", miniGameShopCategory, fmt.Sprintf("%d", entry.ID), string(payload))
	}
}

func setupMiniGameCommander(t *testing.T, commanderID uint32) *orm.Commander {
	name := fmt.Sprintf("MiniGame Commander %d", commanderID)
	if err := orm.CreateCommanderRoot(commanderID, commanderID, name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func cleanupMiniGameShopData(t *testing.T, commanderID uint32) {
	execAnswerExternalTestSQLT(t, "DELETE FROM mini_game_shop_goods WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM mini_game_shop_states WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
}

func TestGetMiniGameShopFiltersByTime(t *testing.T) {
	commanderID := uint32(9001)
	cleanupMiniGameShopData(t, commanderID)
	now := time.Now().UTC()
	seedMiniGameShopConfig(t, []miniGameShopEntry{
		{
			ID:                 1,
			GoodsPurchaseLimit: 2,
			Order:              1,
			Time:               [][][3]int{{{now.Year() - 1, int(now.Month()), now.Day()}, {now.Year() + 1, int(now.Month()), now.Day()}}},
		},
		{
			ID:                 2,
			GoodsPurchaseLimit: 3,
			Order:              2,
			Time:               [][][3]int{{{now.Year() - 2, int(now.Month()), now.Day()}, {now.Year() - 1, int(now.Month()), now.Day()}}},
		},
	})
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupMiniGameShopData(t, commanderID)

	payload := &protobuf.CS_26150{Type: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetMiniGameShop(&buf, client); err != nil {
		t.Fatalf("GetMiniGameShop failed: %v", err)
	}
	response := &protobuf.SC_26151{}
	decodeTestPacket(t, client, 26151, response)
	if len(response.GetGoods()) != 1 {
		t.Fatalf("expected 1 good, got %d", len(response.GetGoods()))
	}
	if response.GetGoods()[0].GetCount() != 2 {
		t.Fatalf("expected count 2, got %d", response.GetGoods()[0].GetCount())
	}
}

func TestGetMiniGameShopResetsOnExpiry(t *testing.T) {
	commanderID := uint32(9002)
	cleanupMiniGameShopData(t, commanderID)
	now := time.Now().UTC()
	seedMiniGameShopConfig(t, []miniGameShopEntry{
		{
			ID:                 3,
			GoodsPurchaseLimit: 1,
			Order:              1,
			Time:               [][][3]int{{{now.Year() - 1, int(now.Month()), now.Day()}, {now.Year() + 1, int(now.Month()), now.Day()}}},
		},
	})
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupMiniGameShopData(t, commanderID)

	execAnswerExternalTestSQLT(t, "INSERT INTO mini_game_shop_states (commander_id, next_refresh_time) VALUES ($1, $2)", int64(commanderID), int64(time.Now().Add(-time.Hour).Unix()))
	payload := &protobuf.CS_26150{Type: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetMiniGameShop(&buf, client); err != nil {
		t.Fatalf("GetMiniGameShop failed: %v", err)
	}
	response := &protobuf.SC_26151{}
	decodeTestPacket(t, client, 26151, response)
	if response.GetNextFlashTime() <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next_flash_time refreshed")
	}
}

func TestMiniGameShopBuySuccess(t *testing.T) {
	commanderID := uint32(9003)
	cleanupMiniGameShopData(t, commanderID)
	now := time.Now().UTC()
	seedMiniGameShopConfig(t, []miniGameShopEntry{{
		ID:                 10,
		GoodsPurchaseLimit: 3,
		Goods:              []uint32{1},
		DropType:           1,
		Price:              5,
		Num:                2,
		Order:              1,
		Time:               [][][3]int{{{now.Year() - 1, int(now.Month()), now.Day()}, {now.Year() + 1, int(now.Month()), now.Day()}}},
	}})
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupMiniGameShopData(t, commanderID)
	if err := client.Commander.SetResource(1, 10); err != nil {
		t.Fatalf("failed to seed gold: %v", err)
	}
	if err := client.Commander.SetResource(12, 30); err != nil {
		t.Fatalf("failed to seed minigame ticket: %v", err)
	}

	payload := &protobuf.CS_26152{
		Goodsid: proto.Uint32(10),
		Selected: []*protobuf.SELECT_INFO{{
			Id:  proto.Uint32(1),
			Num: proto.Uint32(2),
		}},
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.MiniGameShopBuy(&buf, client); err != nil {
		t.Fatalf("MiniGameShopBuy failed: %v", err)
	}

	response := &protobuf.SC_26153{}
	decodeTestPacket(t, client, 26153, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}
	if len(response.GetDropList()) != 1 {
		t.Fatalf("expected 1 drop entry, got %d", len(response.GetDropList()))
	}
	if response.GetDropList()[0].GetId() != 1 || response.GetDropList()[0].GetNumber() != 4 {
		t.Fatalf("unexpected drop payload: %+v", response.GetDropList()[0])
	}

	remainingStock := queryAnswerExternalTestInt64(t, "SELECT count FROM mini_game_shop_goods WHERE commander_id = $1 AND goods_id = $2", int64(commanderID), int64(10))
	if remainingStock != 1 {
		t.Fatalf("expected remaining stock 1, got %d", remainingStock)
	}
	ticketAmount := queryAnswerExternalTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(commanderID), int64(12))
	if ticketAmount != 20 {
		t.Fatalf("expected 20 tickets after purchase, got %d", ticketAmount)
	}
}

func TestMiniGameShopBuyFailureDoesNotMutate(t *testing.T) {
	commanderID := uint32(9004)
	cleanupMiniGameShopData(t, commanderID)
	now := time.Now().UTC()
	seedMiniGameShopConfig(t, []miniGameShopEntry{{
		ID:                 11,
		GoodsPurchaseLimit: 2,
		Goods:              []uint32{1},
		DropType:           1,
		Price:              7,
		Num:                1,
		Order:              1,
		Time:               [][][3]int{{{now.Year() - 1, int(now.Month()), now.Day()}, {now.Year() + 1, int(now.Month()), now.Day()}}},
	}})
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupMiniGameShopData(t, commanderID)
	if err := client.Commander.SetResource(12, 3); err != nil {
		t.Fatalf("failed to seed minigame ticket: %v", err)
	}

	payload := &protobuf.CS_26152{
		Goodsid: proto.Uint32(11),
		Selected: []*protobuf.SELECT_INFO{{
			Id:  proto.Uint32(1),
			Num: proto.Uint32(1),
		}},
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.MiniGameShopBuy(&buf, client); err != nil {
		t.Fatalf("MiniGameShopBuy failed: %v", err)
	}

	response := &protobuf.SC_26153{}
	decodeTestPacket(t, client, 26153, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}

	stock := queryAnswerExternalTestInt64(t, "SELECT count FROM mini_game_shop_goods WHERE commander_id = $1 AND goods_id = $2", int64(commanderID), int64(11))
	if stock != 2 {
		t.Fatalf("expected stock unchanged, got %d", stock)
	}
	tickets := queryAnswerExternalTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(commanderID), int64(12))
	if tickets != 3 {
		t.Fatalf("expected tickets unchanged, got %d", tickets)
	}
}
