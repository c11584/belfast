package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedIslandShipOrderConfig(t *testing.T) {
	t.Helper()
	seedConfigEntry(t, islandOrderSetCategory, "island_shiporder_refresh_cd", `{"key":"island_shiporder_refresh_cd","key_value_int":3600}`)
	seedConfigEntry(t, islandOrderListCategory, "all", `[{"id":301,"type":3},{"id":302,"type":3}]`)
	seedConfigEntry(t, islandOrderTemplateCategory, "all", `[{"id":100001,"type":3,"request":[[4001,4]],"award":[5001,2]},{"id":100002,"type":3,"request":[[4002,3]],"award":[5002,1]}]`)
}

func TestHandleIslandShipOrderRefreshCooldownAndPayload(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.IslandShipOrderState{})
	clearTable(t, &orm.IslandShipOrderSlot{})
	seedIslandShipOrderConfig(t)

	request := protobuf.CS_21429{SlotId: proto.Uint32(0)}
	buffer, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := HandleIslandShipOrderRefresh(&buffer, client); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	var response protobuf.SC_21430
	decodeResponse(t, client, &response)
	if response.GetResult() != islandOrderResultSuccess {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}
	if response.GetNextTime() <= uint32(time.Now().UTC().Unix()) {
		t.Fatalf("expected next_time in the future, got %d", response.GetNextTime())
	}
	if len(response.GetAppointList()) != 2 {
		t.Fatalf("expected 2 appoint entries, got %d", len(response.GetAppointList()))
	}

	if _, _, err := HandleIslandShipOrderRefresh(&buffer, client); err != nil {
		t.Fatalf("second refresh failed: %v", err)
	}
	decodeResponse(t, client, &response)
	if response.GetResult() != islandOrderResultInvalidState {
		t.Fatalf("expected cooldown failure result, got %d", response.GetResult())
	}
}

func TestIslandShipOrderOperateAndSubmitBaseline(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.IslandShipOrderState{})
	clearTable(t, &orm.IslandShipOrderSlot{})
	seedIslandShipOrderConfig(t)

	refresh := protobuf.CS_21429{SlotId: proto.Uint32(0)}
	refreshBuffer, _ := proto.Marshal(&refresh)
	if _, _, err := HandleIslandShipOrderRefresh(&refreshBuffer, client); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	op := protobuf.CS_21408{Type: proto.Uint32(1), ShipSlotId: proto.Uint32(301)}
	opBuffer, _ := proto.Marshal(&op)
	if _, _, err := IslandShipOrderOperate(&opBuffer, client); err != nil {
		t.Fatalf("operate failed: %v", err)
	}
	var opResp protobuf.SC_21409
	decodeResponse(t, client, &opResp)
	if opResp.GetResult() != islandOrderResultSuccess {
		t.Fatalf("expected operate success, got %d", opResp.GetResult())
	}
	if opResp.GetSlot().GetId() != 301 {
		t.Fatalf("expected slot 301, got %d", opResp.GetSlot().GetId())
	}

	submit := protobuf.CS_21416{ShipSlotId: proto.Uint32(301), ItemId: []uint32{4001}}
	submitBuffer, _ := proto.Marshal(&submit)
	if _, _, err := IslandShipOrderSubmit(&submitBuffer, client); err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	var submitResp protobuf.SC_21417
	decodeResponse(t, client, &submitResp)
	if submitResp.GetResult() != islandOrderResultSuccess {
		t.Fatalf("expected submit success, got %d", submitResp.GetResult())
	}
	if submitResp.GetGetTime() == 0 {
		t.Fatalf("expected get_time set")
	}
}
