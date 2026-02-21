package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupFleetTechHandlerTest(t *testing.T) {
	t.Helper()
	seedConfigEntry(t, "ShareCfg/fleet_tech_group.json", "1", `{"id":1,"techs":[1001,1002]}`)
	seedConfigEntry(t, "ShareCfg/fleet_tech_group.json", "2", `{"id":2,"techs":[2001]}`)
	seedConfigEntry(t, "ShareCfg/fleet_tech_template.json", "1001", `{"id":1001,"groupid":1,"cost":10,"time":0,"add":[[[1],2,5]],"level_award_display":[[1,1,10]]}`)
	seedConfigEntry(t, "ShareCfg/fleet_tech_template.json", "1002", `{"id":1002,"groupid":1,"cost":20,"time":0,"add":[[[1],2,4]],"level_award_display":[[1,1,20]]}`)
	seedConfigEntry(t, "ShareCfg/fleet_tech_template.json", "2001", `{"id":2001,"groupid":2,"cost":15,"time":0,"add":[[[2],3,7]],"level_award_display":[[1,1,30]]}`)
}

func TestStartCampTechAndFinishFlow(t *testing.T) {
	client := setupConfigTest(t)
	setupFleetTechHandlerTest(t)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if err := client.Commander.SetResource(1, 100); err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if !client.Commander.HasEnoughResource(1, 10) {
		t.Fatalf("expected seeded gold to be available")
	}

	start := &protobuf.CS_64001{TechGroupId: proto.Uint32(1), TechId: proto.Uint32(1001)}
	startBuf, _ := proto.Marshal(start)
	if _, _, err := StartCampTech(&startBuf, client); err != nil {
		t.Fatalf("start camp tech: %v", err)
	}

	var startResp protobuf.SC_64002
	decodePacketAt(t, client, 0, 64002, &startResp)
	if startResp.GetResult() != 0 {
		t.Fatalf("expected start success, got %d", startResp.GetResult())
	}

	state, err := orm.GetOrCreateCommanderFleetTechState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	group, ok := state.GetGroup(1)
	if !ok || group.StudyTechID != 1001 {
		t.Fatalf("expected group 1 study to be 1001")
	}

	client.Buffer.Reset()
	finish := &protobuf.CS_64003{TechGroupId: proto.Uint32(1)}
	finishBuf, _ := proto.Marshal(finish)
	if _, _, err := FinishCampTechnology(&finishBuf, client); err != nil {
		t.Fatalf("finish camp tech: %v", err)
	}

	var finishResp protobuf.SC_64004
	decodePacketAt(t, client, 0, 64004, &finishResp)
	if finishResp.GetResult() != 0 {
		t.Fatalf("expected finish success")
	}

	client.Buffer.Reset()
	empty := []byte{}
	if _, _, err := TechnologyNationProxy(&empty, client); err != nil {
		t.Fatalf("technology nation proxy: %v", err)
	}
	var proxyResp protobuf.SC_64000
	decodePacketAt(t, client, 0, 64000, &proxyResp)
	if len(proxyResp.GetTechList()) == 0 || proxyResp.GetTechList()[0].GetEffectTechId() != 1001 {
		t.Fatalf("expected bootstrap to include completed tech")
	}
}

func TestFinishCampTechnologyTooEarlyFailure(t *testing.T) {
	client := setupConfigTest(t)
	setupFleetTechHandlerTest(t)
	state, err := orm.GetOrCreateCommanderFleetTechState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	group := state.UpsertGroup(1)
	group.StudyTechID = 1001
	group.StudyFinishTime = 4102444800
	if err := orm.SaveCommanderFleetTechState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	finish := &protobuf.CS_64003{TechGroupId: proto.Uint32(1)}
	finishBuf, _ := proto.Marshal(finish)
	if _, _, err := FinishCampTechnology(&finishBuf, client); err != nil {
		t.Fatalf("finish camp tech: %v", err)
	}

	var finishResp protobuf.SC_64004
	decodePacketAt(t, client, 0, 64004, &finishResp)
	if finishResp.GetResult() == 0 {
		t.Fatalf("expected finish failure while timer active")
	}
}

func TestClaimFleetTechCampAwardSingleAndOneStep(t *testing.T) {
	client := setupConfigTest(t)
	setupFleetTechHandlerTest(t)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	state, err := orm.GetOrCreateCommanderFleetTechState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	group1 := state.UpsertGroup(1)
	group1.EffectTechID = 1002
	group1.RewardedTechID = 0
	group2 := state.UpsertGroup(2)
	group2.EffectTechID = 2001
	group2.RewardedTechID = 0
	if err := orm.SaveCommanderFleetTechState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	claim := &protobuf.CS_64005{GroupId: proto.Uint32(1), TechId: proto.Uint32(1001)}
	claimBuf, _ := proto.Marshal(claim)
	if _, _, err := ClaimFleetTechCampAward(&claimBuf, client); err != nil {
		t.Fatalf("claim camp tech award: %v", err)
	}
	var claimResp protobuf.SC_64006
	decodePacketAt(t, client, 0, 64006, &claimResp)
	if claimResp.GetResult() != 0 || len(claimResp.GetRewards()) == 0 {
		t.Fatalf("expected single claim success with rewards")
	}

	client.Buffer.Reset()
	oneStep := &protobuf.CS_64007{Type: proto.Uint32(1)}
	oneStepBuf, _ := proto.Marshal(oneStep)
	if _, _, err := ClaimTechnologyCampAwardsOneStep(&oneStepBuf, client); err != nil {
		t.Fatalf("claim one-step awards: %v", err)
	}
	var oneStepResp protobuf.SC_64008
	decodePacketAt(t, client, 0, 64008, &oneStepResp)
	if oneStepResp.GetResult() != 0 {
		t.Fatalf("expected one-step success")
	}
	if len(oneStepResp.GetRewards()) != 1 || oneStepResp.GetRewards()[0].GetNumber() != 50 {
		t.Fatalf("expected merged reward amount 50, got %+v", oneStepResp.GetRewards())
	}

	reloaded, err := orm.GetOrCreateCommanderFleetTechState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("reload state: %v", err)
	}
	g1, _ := reloaded.GetGroup(1)
	g2, _ := reloaded.GetGroup(2)
	if g1.RewardedTechID != g1.EffectTechID || g2.RewardedTechID != g2.EffectTechID {
		t.Fatalf("expected one-step to advance rewarded tech pointers")
	}
}

func TestSetFleetTechAttrAdditionPersistsOverrides(t *testing.T) {
	client := setupConfigTest(t)
	setupFleetTechHandlerTest(t)

	state, err := orm.GetOrCreateCommanderFleetTechState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	group := state.UpsertGroup(1)
	group.EffectTechID = 1002
	if err := orm.SaveCommanderFleetTechState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	setReq := &protobuf.CS_64009{TechsetList: []*protobuf.TECHSET{
		{ShipType: proto.Uint32(1), AttrType: proto.Uint32(2), SetValue: proto.Uint32(8)},
		{ShipType: proto.Uint32(1), AttrType: proto.Uint32(2), SetValue: proto.Uint32(7)},
	}}
	setBuf, _ := proto.Marshal(setReq)
	if _, _, err := SetFleetTechAttrAddition(&setBuf, client); err != nil {
		t.Fatalf("set fleet tech attr addition: %v", err)
	}
	var setResp protobuf.SC_64010
	decodePacketAt(t, client, 0, 64010, &setResp)
	if setResp.GetResult() != 0 {
		t.Fatalf("expected techset save success")
	}

	client.Buffer.Reset()
	empty := []byte{}
	if _, _, err := TechnologyNationProxy(&empty, client); err != nil {
		t.Fatalf("technology nation proxy: %v", err)
	}
	var proxyResp protobuf.SC_64000
	decodePacketAt(t, client, 0, 64000, &proxyResp)
	if len(proxyResp.GetTechsetList()) != 1 || proxyResp.GetTechsetList()[0].GetSetValue() != 7 {
		t.Fatalf("expected persisted override in bootstrap")
	}

	client.Buffer.Reset()
	invalidReq := &protobuf.CS_64009{TechsetList: []*protobuf.TECHSET{{ShipType: proto.Uint32(1), AttrType: proto.Uint32(2), SetValue: proto.Uint32(99)}}}
	invalidBuf, _ := proto.Marshal(invalidReq)
	if _, _, err := SetFleetTechAttrAddition(&invalidBuf, client); err != nil {
		t.Fatalf("set fleet tech attr addition: %v", err)
	}
	var invalidResp protobuf.SC_64010
	decodePacketAt(t, client, 0, 64010, &invalidResp)
	if invalidResp.GetResult() == 0 {
		t.Fatalf("expected invalid override failure")
	}
}
