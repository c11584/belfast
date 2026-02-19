package answer

import (
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupEducateHandlerTest(t *testing.T, commanderID uint32) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderCommonFlag{})
	clearTable(t, &orm.EducateShopState{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.Commander{})
	if err := orm.CreateCommanderRoot(commanderID, commanderID, "Educate Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestEducateGetEventsSortedAndConsumedFiltered(t *testing.T) {
	client := setupEducateHandlerTest(t, 9101)
	seedConfigEntry(t, "ShareCfg/child_event_special.json", "rows", `[
		{"id":200,"show":1,"type":1},
		{"id":100,"show":1,"type":1},
		{"id":300,"show":0,"type":1}
	]`)
	if err := orm.SetCommanderCommonFlag(client.Commander.CommanderID, educateFlagID(educateFlagHomeEventBase, 200)); err != nil {
		t.Fatalf("seed consumed flag: %v", err)
	}

	buf, _ := proto.Marshal(&protobuf.CS_27014{Type: proto.Uint32(0)})
	if _, _, err := EducateGetEvents(&buf, client); err != nil {
		t.Fatalf("EducateGetEvents: %v", err)
	}
	var resp protobuf.SC_27015
	decodePacketAt(t, client, 0, 27015, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}
	if len(resp.GetEvents()) != 1 || resp.GetEvents()[0] != 100 {
		t.Fatalf("unexpected events: %v", resp.GetEvents())
	}
}

func TestEducateTriggerEventValidationAndSuccess(t *testing.T) {
	client := setupEducateHandlerTest(t, 9102)
	seedConfigEntry(t, "ShareCfg/child_event.json", "rows", `[{"id":110}]`)

	badBuf, _ := proto.Marshal(&protobuf.CS_27016{Eventid: proto.Uint32(999)})
	if _, _, err := EducateTriggerEvent(&badBuf, client); err != nil {
		t.Fatalf("EducateTriggerEvent bad: %v", err)
	}
	var badResp protobuf.SC_27017
	decodePacketAt(t, client, 0, 27017, &badResp)
	if badResp.GetResult() == 0 {
		t.Fatalf("expected failure for unknown event")
	}

	client.Buffer.Reset()
	okBuf, _ := proto.Marshal(&protobuf.CS_27016{Eventid: proto.Uint32(110)})
	if _, _, err := EducateTriggerEvent(&okBuf, client); err != nil {
		t.Fatalf("EducateTriggerEvent ok: %v", err)
	}
	var okResp protobuf.SC_27017
	decodePacketAt(t, client, 0, 27017, &okResp)
	if okResp.GetResult() != 0 {
		t.Fatalf("expected success, got %d", okResp.GetResult())
	}
	has, err := hasEducateFlag(client.Commander.CommanderID, educateFlagID(educateFlagHomeEventBase, 110))
	if err != nil {
		t.Fatalf("check flag: %v", err)
	}
	if !has {
		t.Fatalf("expected consumed flag after success")
	}
}

func TestEducateTriggerSpecEventSuccessDuplicateAndInvalid(t *testing.T) {
	client := setupEducateHandlerTest(t, 9103)
	seedConfigEntry(t, "ShareCfg/child_event_special.json", "rows", `[{"id":321,"show":1,"type":3,"drop_display":[3,201,2]}]`)

	buf, _ := proto.Marshal(&protobuf.CS_27027{SpecEventsId: proto.Uint32(321)})
	if _, _, err := EducateTriggerSpecEvent(&buf, client); err != nil {
		t.Fatalf("EducateTriggerSpecEvent: %v", err)
	}
	var resp protobuf.SC_27028
	decodePacketAt(t, client, 0, 27028, &resp)
	if resp.GetResult() != 0 || len(resp.GetDrops()) != 1 || resp.GetDrops()[0].GetId() != 201 {
		t.Fatalf("unexpected success response: %+v", resp)
	}

	client.Buffer.Reset()
	if _, _, err := EducateTriggerSpecEvent(&buf, client); err != nil {
		t.Fatalf("EducateTriggerSpecEvent duplicate: %v", err)
	}
	decodePacketAt(t, client, 0, 27028, &resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected duplicate failure")
	}

	client.Buffer.Reset()
	badBuf, _ := proto.Marshal(&protobuf.CS_27027{SpecEventsId: proto.Uint32(999)})
	if _, _, err := EducateTriggerSpecEvent(&badBuf, client); err != nil {
		t.Fatalf("EducateTriggerSpecEvent invalid: %v", err)
	}
	decodePacketAt(t, client, 0, 27028, &resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected invalid-id failure")
	}
}

func TestEducateShopRequestAndPurchaseFlow(t *testing.T) {
	client := setupEducateHandlerTest(t, 9104)
	seedConfigEntry(t, "ShareCfg/child_shop.json", "rows", `[{"id":2,"goods_num":2,"goods_pool":[[11,1,500,[]],[12,1,500,[]]],"goods_refresh_time":-1}]`)
	seedConfigEntry(t, "ShareCfg/child_shop_template.json", "rows", `[
		{"id":11,"item_id":500,"resource":1,"resource_num":10,"buy_num":1},
		{"id":12,"item_id":501,"resource":1,"resource_num":20,"buy_num":1}
	]`)
	if err := client.Commander.SetResource(1, 50); err != nil {
		t.Fatalf("seed resource: %v", err)
	}

	getBuf, _ := proto.Marshal(&protobuf.CS_27043{ShopId: proto.Uint32(2)})
	if _, _, err := EducateRequestShopData(&getBuf, client); err != nil {
		t.Fatalf("EducateRequestShopData: %v", err)
	}
	var getResp protobuf.SC_27044
	decodePacketAt(t, client, 0, 27044, &getResp)
	if getResp.GetResult() != 0 || len(getResp.GetShopData().GetGoods()) != 2 {
		t.Fatalf("unexpected shop data response: %+v", getResp)
	}

	client.Buffer.Reset()
	buyBuf, _ := proto.Marshal(&protobuf.CS_27033{ShopId: proto.Uint32(2), Goods: []*protobuf.CHILD_SHOP_GOODS{{Id: proto.Uint32(11), Num: proto.Uint32(1)}}})
	if _, _, err := EducateShopping(&buyBuf, client); err != nil {
		t.Fatalf("EducateShopping: %v", err)
	}
	var buyResp protobuf.SC_27034
	decodePacketAt(t, client, 0, 27034, &buyResp)
	if buyResp.GetResult() != 0 || len(buyResp.GetDrops()) != 1 || buyResp.GetDrops()[0].GetId() != 500 {
		t.Fatalf("unexpected purchase response: %+v", buyResp)
	}
	if got := client.Commander.GetResourceCount(1); got != 40 {
		t.Fatalf("expected resource 40, got %d", got)
	}

	client.Buffer.Reset()
	if _, _, err := EducateShopping(&buyBuf, client); err != nil {
		t.Fatalf("EducateShopping second: %v", err)
	}
	decodePacketAt(t, client, 0, 27034, &buyResp)
	if buyResp.GetResult() == 0 {
		t.Fatalf("expected sold-out failure")
	}

	client.Buffer.Reset()
	if _, _, err := EducateRequestShopData(&getBuf, client); err != nil {
		t.Fatalf("EducateRequestShopData replay: %v", err)
	}
	decodePacketAt(t, client, 0, 27044, &getResp)
	if getResp.GetShopData().GetGoods()[0].GetNum() != 0 {
		t.Fatalf("expected updated remaining count in shop data")
	}
}

func TestEducateShoppingFailurePathsAndDecodeError(t *testing.T) {
	client := setupEducateHandlerTest(t, 9105)
	seedConfigEntry(t, "ShareCfg/child_shop.json", "rows", `[{"id":3,"goods_num":1,"goods_pool":[[21,1,500,[]]],"goods_refresh_time":-1}]`)
	seedConfigEntry(t, "ShareCfg/child_shop_template.json", "rows", `[{"id":21,"item_id":600,"resource":1,"resource_num":30,"buy_num":1}]`)
	if err := client.Commander.SetResource(1, 5); err != nil {
		t.Fatalf("seed resource: %v", err)
	}

	buyBuf, _ := proto.Marshal(&protobuf.CS_27033{ShopId: proto.Uint32(3), Goods: []*protobuf.CHILD_SHOP_GOODS{{Id: proto.Uint32(21), Num: proto.Uint32(1)}}})
	if _, _, err := EducateShopping(&buyBuf, client); err != nil {
		t.Fatalf("EducateShopping insufficient: %v", err)
	}
	var buyResp protobuf.SC_27034
	decodePacketAt(t, client, 0, 27034, &buyResp)
	if buyResp.GetResult() == 0 {
		t.Fatalf("expected insufficient resource failure")
	}

	bad := []byte{0x01, 0x02}
	if _, packetID, err := EducateRequestShopData(&bad, client); err == nil || packetID != 27044 {
		t.Fatalf("expected decode error with packet 27044")
	}
}

func TestEducateTargetAwardSuccessDuplicateAndFailure(t *testing.T) {
	client := setupEducateHandlerTest(t, 9106)
	seedConfigEntry(t, "ShareCfg/child_target_set.json", "rows", `[{"id":1,"stage":1,"ids":[1001,1002],"target_progress":2,"drop_display":[3,201,1]}]`)
	seedConfigEntry(t, "ShareCfg/child_task.json", "rows", `[{"id":1001,"task_target_progress":1},{"id":1002,"task_target_progress":1}]`)

	buf, _ := proto.Marshal(&protobuf.CS_27035{Type: proto.Uint32(0)})
	if _, _, err := EducateGetTargetAward(&buf, client); err != nil {
		t.Fatalf("EducateGetTargetAward: %v", err)
	}
	var resp protobuf.SC_27036
	decodePacketAt(t, client, 0, 27036, &resp)
	if resp.GetResult() != 0 || len(resp.GetDrops()) != 1 {
		t.Fatalf("unexpected success response: %+v", resp)
	}

	client.Buffer.Reset()
	if _, _, err := EducateGetTargetAward(&buf, client); err != nil {
		t.Fatalf("EducateGetTargetAward duplicate: %v", err)
	}
	decodePacketAt(t, client, 0, 27036, &resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected duplicate claim failure")
	}

	client.Buffer.Reset()
	badTypeBuf, _ := proto.Marshal(&protobuf.CS_27035{Type: proto.Uint32(1)})
	if _, _, err := EducateGetTargetAward(&badTypeBuf, client); err != nil {
		t.Fatalf("EducateGetTargetAward bad type: %v", err)
	}
	decodePacketAt(t, client, 0, 27036, &resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected unsupported type failure")
	}

	client2 := setupEducateHandlerTest(t, 9107)
	seedConfigEntry(t, "ShareCfg/child_target_set.json", "rows", `[{"id":1,"stage":1,"ids":[1001,1002],"target_progress":3,"drop_display":[3,201,1]}]`)
	seedConfigEntry(t, "ShareCfg/child_task.json", "rows", `[{"id":1001,"task_target_progress":1},{"id":1002,"task_target_progress":1}]`)
	if _, _, err := EducateGetTargetAward(&buf, client2); err != nil {
		t.Fatalf("EducateGetTargetAward ineligible: %v", err)
	}
	decodePacketAt(t, client2, 0, 27036, &resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected ineligible failure")
	}
}
