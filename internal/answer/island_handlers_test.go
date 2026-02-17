package answer

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestIslandGetDataSelfIncludesPrivateFollowShips(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.IslandSnapshot{})
	clearTable(t, &orm.IslandTechnologyState{})
	clearTable(t, &orm.IslandCommanderDressState{})
	clearTable(t, &orm.IslandShopState{})

	if err := orm.UpsertIslandSnapshot(&orm.IslandSnapshot{CommanderID: client.Commander.CommanderID, Name: "Home", Level: 3, StorageLevel: 2, FollowShips: []uint32{1001, 1002}}); err != nil {
		t.Fatalf("seed island snapshot: %v", err)
	}

	payload := &protobuf.CS_21200{IslandId: proto.Uint32(client.Commander.CommanderID)}
	buffer, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := IslandGetData(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_21201
	decodePacketAt(t, client, 0, 21201, &response)
	if response.GetIsland().GetPrivateData() == nil {
		t.Fatalf("expected private data")
	}
	if len(response.GetIsland().GetPrivateData().GetFollowShips()) != 2 {
		t.Fatalf("expected persisted follow ships, got %+v", response.GetIsland().GetPrivateData().GetFollowShips())
	}
	if response.GetPlayerPosition().GetMapId() == 0 {
		t.Fatalf("expected non-zero fallback map id")
	}
}

func TestIslandSignInGiftClaimSignInAndClaim(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.IslandSignInState{})

	seedConfigEntry(t, islandSetCategory, "daily_gift_drop_num", `{"key_value_int":6}`)
	seedConfigEntry(t, islandSetCategory, "daily_gift_get_max", `{"key_value_int":3}`)
	seedConfigEntry(t, islandSetCategory, "daily_gift", `{"key_value_int":20001}`)

	signInPayload := &protobuf.CS_21310{IslandId: proto.Uint32(0), Pos: proto.Uint32(0)}
	buffer, _ := proto.Marshal(signInPayload)
	client.Buffer.Reset()
	if _, _, err := IslandSignInGiftClaim(&buffer, client); err != nil {
		t.Fatalf("sign-in failed: %v", err)
	}
	var signInResponse protobuf.SC_21311
	decodePacketAt(t, client, 0, 21311, &signInResponse)
	if signInResponse.GetResult() != 0 {
		t.Fatalf("expected sign-in success")
	}

	claimPayload := &protobuf.CS_21310{IslandId: proto.Uint32(client.Commander.CommanderID), Pos: proto.Uint32(1)}
	buffer, _ = proto.Marshal(claimPayload)
	client.Buffer.Reset()
	if _, _, err := IslandSignInGiftClaim(&buffer, client); err != nil {
		t.Fatalf("claim failed: %v", err)
	}
	var claimResponse protobuf.SC_21311
	decodePacketAt(t, client, 0, 21311, &claimResponse)
	if claimResponse.GetResult() != 0 || len(claimResponse.GetDropList()) != 1 {
		t.Fatalf("expected claim success with one drop, got result=%d drops=%d", claimResponse.GetResult(), len(claimResponse.GetDropList()))
	}
}

func TestIslandUnlockAndFinishTech(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.IslandSnapshot{})
	clearTable(t, &orm.IslandTechnologyState{})

	seedConfigEntry(t, islandTechCategory, "1", `{"id":1,"formula_id":101,"island_level":1,"sys_unlock":[],"tech_repeat":2}`)
	seedConfigEntry(t, islandFormulaCategory, "101", `{"id":101,"unlock_type":7,"drop_list":[[2,20001,2]]}`)
	if err := orm.UpsertIslandSnapshot(&orm.IslandSnapshot{CommanderID: client.Commander.CommanderID, Level: 2, StorageLevel: 1}); err != nil {
		t.Fatalf("seed island snapshot: %v", err)
	}

	unlockPayload := &protobuf.CS_21520{TechId: proto.Uint32(1)}
	buffer, _ := proto.Marshal(unlockPayload)
	client.Buffer.Reset()
	if _, _, err := IslandUnlockTech(&buffer, client); err != nil {
		t.Fatalf("unlock handler failed: %v", err)
	}
	var unlockResponse protobuf.SC_21521
	decodePacketAt(t, client, 0, 21521, &unlockResponse)
	if unlockResponse.GetResult() != 0 {
		t.Fatalf("expected unlock success")
	}

	finishPayload := &protobuf.CS_21522{TechId: proto.Uint32(1)}
	buffer, _ = proto.Marshal(finishPayload)
	client.Buffer.Reset()
	if _, _, err := IslandFinishTechImmediate(&buffer, client); err != nil {
		t.Fatalf("finish handler failed: %v", err)
	}
	var finishResponse protobuf.SC_21523
	decodePacketAt(t, client, 0, 21523, &finishResponse)
	if finishResponse.GetResult() != 0 {
		t.Fatalf("expected immediate finish success")
	}
}

func TestIslandUnlockTechUsesDefaultLevelWhenSnapshotMissing(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.IslandSnapshot{})
	clearTable(t, &orm.IslandTechnologyState{})

	seedConfigEntry(t, islandTechCategory, "1", `{"id":1,"formula_id":101,"island_level":1,"sys_unlock":[],"tech_repeat":1}`)
	seedConfigEntry(t, islandTechCategory, "2", `{"id":2,"formula_id":102,"island_level":2,"sys_unlock":[],"tech_repeat":1}`)
	seedConfigEntry(t, islandFormulaCategory, "101", `{"id":101,"drop_list":[[2,20001,1]]}`)
	seedConfigEntry(t, islandFormulaCategory, "102", `{"id":102,"drop_list":[[2,20001,1]]}`)

	successPayload := &protobuf.CS_21520{TechId: proto.Uint32(1)}
	buffer, _ := proto.Marshal(successPayload)
	client.Buffer.Reset()
	if _, _, err := IslandUnlockTech(&buffer, client); err != nil {
		t.Fatalf("unlock level-1 tech failed: %v", err)
	}
	var successResponse protobuf.SC_21521
	decodePacketAt(t, client, 0, 21521, &successResponse)
	if successResponse.GetResult() != 0 {
		t.Fatalf("expected level-1 unlock success without snapshot")
	}

	failPayload := &protobuf.CS_21520{TechId: proto.Uint32(2)}
	buffer, _ = proto.Marshal(failPayload)
	client.Buffer.Reset()
	if _, _, err := IslandUnlockTech(&buffer, client); err != nil {
		t.Fatalf("unlock level-2 tech failed: %v", err)
	}
	var failResponse protobuf.SC_21521
	decodePacketAt(t, client, 0, 21521, &failResponse)
	if failResponse.GetResult() == 0 {
		t.Fatalf("expected level-2 unlock to fail without snapshot")
	}
}

func TestIslandFinishTechImmediateAppliesAllDropTypes(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.IslandSnapshot{})
	clearTable(t, &orm.IslandTechnologyState{})

	seedConfigEntry(t, islandTechCategory, "5", `{"id":5,"formula_id":105,"island_level":1,"sys_unlock":[],"tech_repeat":1}`)
	seedConfigEntry(t, islandFormulaCategory, "105", `{"id":105,"drop_list":[[1,1,33],[2,20001,2]]}`)
	if err := orm.UpsertIslandSnapshot(&orm.IslandSnapshot{CommanderID: client.Commander.CommanderID, Level: 1, StorageLevel: 1}); err != nil {
		t.Fatalf("seed island snapshot: %v", err)
	}

	unlockPayload := &protobuf.CS_21520{TechId: proto.Uint32(5)}
	buffer, _ := proto.Marshal(unlockPayload)
	client.Buffer.Reset()
	if _, _, err := IslandUnlockTech(&buffer, client); err != nil {
		t.Fatalf("unlock tech failed: %v", err)
	}

	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(20001)

	finishPayload := &protobuf.CS_21522{TechId: proto.Uint32(5)}
	buffer, _ = proto.Marshal(finishPayload)
	client.Buffer.Reset()
	if _, _, err := IslandFinishTechImmediate(&buffer, client); err != nil {
		t.Fatalf("finish tech failed: %v", err)
	}

	var response protobuf.SC_21523
	decodePacketAt(t, client, 0, 21523, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected finish success")
	}
	if len(response.GetDropList()) != 2 {
		t.Fatalf("expected two configured drops, got %d", len(response.GetDropList()))
	}
	if client.Commander.GetResourceCount(1) != goldBefore+33 {
		t.Fatalf("expected gold increase by 33, got %d", client.Commander.GetResourceCount(1)-goldBefore)
	}
	if client.Commander.GetItemCount(20001) != itemBefore+2 {
		t.Fatalf("expected item increase by 2, got %d", client.Commander.GetItemCount(20001)-itemBefore)
	}
}

func TestIslandShopRefreshAndDressRead(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.IslandShopState{})
	clearTable(t, &orm.IslandCommanderDressState{})

	seedConfigEntry(t, islandShopTemplateCategory, "10", `{"id":10,"goods_id":[101,102]}`)
	seedConfigEntry(t, islandShopNormalCategory, "10", `{"id":10,"refresh_set":3,"refresh_player":[2,20001,1],"refresh_free":1,"refresh_time":120,"exist_time":3600}`)

	refreshPayload := &protobuf.CS_21020{ShopId: proto.Uint32(10)}
	buffer, _ := proto.Marshal(refreshPayload)
	client.Buffer.Reset()
	if _, _, err := IslandShopPlayerRefresh(&buffer, client); err != nil {
		t.Fatalf("shop refresh failed: %v", err)
	}
	var refreshResponse protobuf.SC_21021
	decodePacketAt(t, client, 0, 21021, &refreshResponse)
	if refreshResponse.GetResult() != 0 || refreshResponse.GetShopInfo() == nil {
		t.Fatalf("expected shop refresh success")
	}

	dressPayload := &protobuf.CS_21621{DressId: []uint32{5001, 5001, 5002}}
	buffer, _ = proto.Marshal(dressPayload)
	client.Buffer.Reset()
	if _, _, err := IslandSetCommanderDressRead(&buffer, client); err != nil {
		t.Fatalf("dress read failed: %v", err)
	}
	var dressResponse protobuf.SC_21622
	decodePacketAt(t, client, 0, 21622, &dressResponse)
	if dressResponse.GetResult() != 0 {
		t.Fatalf("expected dress read success")
	}

	states, err := orm.ListIslandCommanderDressStates(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("list dress states: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("expected deduplicated dress rows, got %d", len(states))
	}
}

func TestIslandGoFishingSuccess(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntry(t, islandFishPointCategory, "300", `{"id":300}`)
	seedConfigEntry(t, islandFishCategory, "9001", `{"id":9001,"min_weight":10,"max_weight":10,"gold_state":1}`)

	payload := &protobuf.CS_21060{IslandId: proto.Uint32(client.Commander.CommanderID), PointId: proto.Uint32(300)}
	buffer, _ := proto.Marshal(payload)
	client.Buffer.Reset()
	if _, _, err := IslandGoFishing(&buffer, client); err != nil {
		t.Fatalf("go fishing failed: %v", err)
	}
	var response protobuf.SC_21061
	decodePacketAt(t, client, 0, 21061, &response)
	if response.GetResult() != 0 || response.GetFishId() == 0 || response.GetWeight() == 0 {
		t.Fatalf("expected fishing success payload, got %+v", response)
	}
}

func TestIslandFishingIntnConcurrent(t *testing.T) {
	islandFishingRNGMu.Lock()
	previousRNG := islandFishingRNG
	islandFishingRNG = rand.New(rand.NewSource(1))
	islandFishingRNGMu.Unlock()
	t.Cleanup(func() {
		islandFishingRNGMu.Lock()
		islandFishingRNG = previousRNG
		islandFishingRNGMu.Unlock()
	})

	var wg sync.WaitGroup
	failures := make(chan struct{}, 1)

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				value := islandFishingIntn(7)
				if value >= 0 && value < 7 {
					continue
				}
				select {
				case failures <- struct{}{}:
				default:
				}
				return
			}
		}()
	}

	wg.Wait()
	if len(failures) > 0 {
		t.Fatalf("expected islandFishingIntn to stay within bounds under concurrency")
	}
}
