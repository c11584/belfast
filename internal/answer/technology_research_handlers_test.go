package answer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
)

func newTechnologyTestClient(t *testing.T) *connection.Client {
	t.Helper()
	orm.InitDatabase()
	commanderID := uint32(time.Now().UnixNano())
	if err := orm.CreateCommanderRoot(commanderID, commanderID, fmt.Sprintf("tech-%d", commanderID), 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := &orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: commander}
}

func decodeTechnologyResponse(t *testing.T, client *connection.Client, expectedID int, out proto.Message) {
	t.Helper()
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response packet")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	client.Buffer.Reset()
}

func seedTechnologyState(t *testing.T, commanderID uint32) orm.TechnologyRefreshPoolState {
	t.Helper()
	pools, err := orm.BuildTechnologyRefreshPools(0)
	if err != nil {
		t.Fatalf("build pools: %v", err)
	}
	if len(pools) == 0 || len(pools[0].Technologies) == 0 {
		t.Fatalf("expected seeded technology pools")
	}
	state := &orm.TechnologyResearchState{
		CommanderID:  commanderID,
		RefreshDay:   orm.CurrentTechnologyDay(time.Now().UTC()),
		RefreshPools: pools,
		Queue:        []orm.TechnologyQueueState{},
	}
	if err := orm.SaveTechnologyResearchState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	return pools[0]
}

func TestStartAndStopTechnologyResearch(t *testing.T) {
	client := newTechnologyTestClient(t)
	pool := seedTechnologyState(t, client.Commander.CommanderID)
	techID := pool.Technologies[0].TechID

	startReq := &protobuf.CS_63001{TechId: proto.Uint32(techID), RefreshId: proto.Uint32(pool.ID)}
	startBuf, _ := proto.Marshal(startReq)
	if _, _, err := StartTechnologyResearch(&startBuf, client); err != nil {
		t.Fatalf("start technology: %v", err)
	}
	startResp := &protobuf.SC_63002{}
	decodeTechnologyResponse(t, client, 63002, startResp)
	if startResp.GetResult() != 0 || startResp.GetTime() == 0 {
		t.Fatalf("unexpected start response: %+v", startResp)
	}

	stopReq := &protobuf.CS_63005{TechId: proto.Uint32(techID), RefreshId: proto.Uint32(pool.ID)}
	stopBuf, _ := proto.Marshal(stopReq)
	if _, _, err := StopTechnologyResearch(&stopBuf, client); err != nil {
		t.Fatalf("stop technology: %v", err)
	}
	stopResp := &protobuf.SC_63006{}
	decodeTechnologyResponse(t, client, 63006, stopResp)
	if stopResp.GetResult() != 0 {
		t.Fatalf("expected stop success, got %d", stopResp.GetResult())
	}
}

func TestRefreshTechnologyProjectsAndTendency(t *testing.T) {
	client := newTechnologyTestClient(t)
	pool := seedTechnologyState(t, client.Commander.CommanderID)

	refreshReq := &protobuf.CS_63007{Type: proto.Uint32(1)}
	refreshBuf, _ := proto.Marshal(refreshReq)
	if _, _, err := RefreshTechnologyProjects(&refreshBuf, client); err != nil {
		t.Fatalf("refresh technologies: %v", err)
	}
	refreshResp := &protobuf.SC_63008{}
	decodeTechnologyResponse(t, client, 63008, refreshResp)
	if refreshResp.GetResult() != 0 || len(refreshResp.GetRefreshList()) == 0 {
		t.Fatalf("unexpected refresh response: %+v", refreshResp)
	}

	tendencyReq := &protobuf.CS_63009{Id: proto.Uint32(pool.ID), Target: proto.Uint32(1)}
	tendencyBuf, _ := proto.Marshal(tendencyReq)
	if _, _, err := ChangeRefreshTechnologyTendency(&tendencyBuf, client); err != nil {
		t.Fatalf("change tendency: %v", err)
	}
	tendencyResp := &protobuf.SC_63010{}
	decodeTechnologyResponse(t, client, 63010, tendencyResp)
	if tendencyResp.GetResult() != 0 {
		t.Fatalf("expected tendency success, got %d", tendencyResp.GetResult())
	}
}

func TestCatchupQueueAndFinishQueue(t *testing.T) {
	client := newTechnologyTestClient(t)
	pool := seedTechnologyState(t, client.Commander.CommanderID)
	techID := pool.Technologies[0].TechID
	version, target := firstCatchupTarget(t)

	selectReq := &protobuf.CS_63011{Version: proto.Uint32(version), Target: proto.Uint32(target)}
	selectBuf, _ := proto.Marshal(selectReq)
	if _, _, err := SelectTechnologyCatchupTarget(&selectBuf, client); err != nil {
		t.Fatalf("select catchup target: %v", err)
	}
	selectResp := &protobuf.SC_63012{}
	decodeTechnologyResponse(t, client, 63012, selectResp)
	if selectResp.GetResult() != 0 {
		t.Fatalf("expected catchup select success, got %d", selectResp.GetResult())
	}

	state, err := orm.GetTechnologyResearchState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	state.RefreshPools[0].Technologies[0].FinishTime = uint32(time.Now().Add(time.Hour).Unix())
	if err := orm.SaveTechnologyResearchState(state); err != nil {
		t.Fatalf("save active state: %v", err)
	}

	queueReq := &protobuf.CS_63013{TechId: proto.Uint32(techID), RefreshId: proto.Uint32(pool.ID)}
	queueBuf, _ := proto.Marshal(queueReq)
	if _, _, err := JoinTechnologyQueue(&queueBuf, client); err != nil {
		t.Fatalf("join queue: %v", err)
	}
	queueResp := &protobuf.SC_63014{}
	decodeTechnologyResponse(t, client, 63014, queueResp)
	if queueResp.GetResult() != 0 {
		t.Fatalf("expected queue success, got %d", queueResp.GetResult())
	}

	state, err = orm.GetTechnologyResearchState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("reload state: %v", err)
	}
	if len(state.Queue) == 0 {
		t.Fatalf("expected queued entry")
	}
	state.Queue[0].FinishTime = uint32(time.Now().Add(-time.Minute).Unix())
	if err := orm.SaveTechnologyResearchState(state); err != nil {
		t.Fatalf("save queue complete: %v", err)
	}

	finishReq := &protobuf.CS_63015{Id: proto.Uint32(0)}
	finishBuf, _ := proto.Marshal(finishReq)
	if _, _, err := FinishQueueTechnology(&finishBuf, client); err != nil {
		t.Fatalf("finish queue: %v", err)
	}
	finishResp := &protobuf.SC_63016{}
	decodeTechnologyResponse(t, client, 63016, finishResp)
	if finishResp.GetResult() != 0 || len(finishResp.GetDrops()) == 0 {
		t.Fatalf("unexpected finish queue response: %+v", finishResp)
	}
}

func TestStartTechnologyResearchRollsBackConsumeFailure(t *testing.T) {
	client := newTechnologyTestClient(t)
	const (
		techID      = uint32(991001)
		refreshID   = uint32(77)
		resourceID  = uint32(900001)
		resourceKey = int64(900001)
	)
	execAnswerTestSQLT(t, `
INSERT INTO resources (id, item_id, name)
VALUES ($1, 0, 'test-resource-tech')
ON CONFLICT (id) DO NOTHING
`, resourceKey)
	if err := client.Commander.AddResource(resourceID, 1); err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	payload := fmt.Sprintf(`{"id":991001,"type":77,"time":60,"condition":0,"consume":[[1,%d,1],[1,%d,1]],"drop_client":[[1,%d,1]]}`,
		resourceID,
		resourceID,
		resourceID,
	)
	seedConfigEntry(t, "ShareCfg/technology_data_template.json", fmt.Sprintf("%d", techID), payload)
	state := &orm.TechnologyResearchState{
		CommanderID: client.Commander.CommanderID,
		RefreshDay:  orm.CurrentTechnologyDay(time.Now().UTC()),
		RefreshPools: []orm.TechnologyRefreshPoolState{{
			ID: refreshID,
			Technologies: []orm.TechnologyProjectState{{
				TechID: techID,
			}},
		}},
		Queue: []orm.TechnologyQueueState{},
	}
	if err := orm.SaveTechnologyResearchState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	initialAmount := uint32(1)

	startReq := &protobuf.CS_63001{TechId: proto.Uint32(techID), RefreshId: proto.Uint32(refreshID)}
	startBuf, _ := proto.Marshal(startReq)
	if _, _, err := StartTechnologyResearch(&startBuf, client); err != nil {
		t.Fatalf("start technology: %v", err)
	}
	startResp := &protobuf.SC_63002{}
	decodeTechnologyResponse(t, client, 63002, startResp)
	if startResp.GetResult() == technologyOK {
		t.Fatalf("expected start to fail when consume cannot fully complete")
	}

	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	reloadedResource, ok := client.Commander.OwnedResourcesMap[resourceID]
	if !ok {
		t.Fatalf("expected reloaded resource")
	}
	if reloadedResource.Amount != initialAmount {
		t.Fatalf("expected resource rollback to %d, got %d", initialAmount, reloadedResource.Amount)
	}
}

func TestFinishTechnologyResearchAllowsConditionTemplates(t *testing.T) {
	client := newTechnologyTestClient(t)
	const (
		techID      = uint32(991002)
		refreshID   = uint32(78)
		resourceID  = uint32(900002)
		resourceKey = int64(900002)
	)
	execAnswerTestSQLT(t, `
INSERT INTO resources (id, item_id, name)
VALUES ($1, 0, 'test-resource-tech-finish')
ON CONFLICT (id) DO NOTHING
`, resourceKey)
	seedConfigEntry(t, "ShareCfg/technology_data_template.json", fmt.Sprintf("%d", techID), fmt.Sprintf(`{"id":991002,"type":78,"time":60,"condition":1,"consume":[],"drop_client":[[1,%d,1]]}`,
		resourceID,
	))
	state := &orm.TechnologyResearchState{
		CommanderID: client.Commander.CommanderID,
		RefreshDay:  orm.CurrentTechnologyDay(time.Now().UTC()),
		RefreshPools: []orm.TechnologyRefreshPoolState{{
			ID: refreshID,
			Technologies: []orm.TechnologyProjectState{{
				TechID:     techID,
				FinishTime: uint32(time.Now().Add(-time.Minute).Unix()),
			}},
		}},
		Queue: []orm.TechnologyQueueState{},
	}
	if err := orm.SaveTechnologyResearchState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	finishReq := &protobuf.CS_63003{TechId: proto.Uint32(techID), RefreshId: proto.Uint32(refreshID)}
	finishBuf, _ := proto.Marshal(finishReq)
	if _, _, err := FinishTechnologyResearch(&finishBuf, client); err != nil {
		t.Fatalf("finish technology: %v", err)
	}
	finishResp := &protobuf.SC_63004{}
	decodeTechnologyResponse(t, client, 63004, finishResp)
	if finishResp.GetResult() != technologyOK {
		t.Fatalf("expected finish success for condition template, got %d", finishResp.GetResult())
	}
	if len(finishResp.GetCommonList()) == 0 {
		t.Fatalf("expected finish rewards")
	}
}

func firstCatchupTarget(t *testing.T) (uint32, uint32) {
	t.Helper()
	type catchupTemplate struct {
		ID         uint32   `json:"id"`
		CharChoice []uint32 `json:"char_choice"`
	}
	entries, err := orm.ListConfigEntries("ShareCfg/technology_catchup_template.json")
	if err != nil || len(entries) == 0 {
		return 1, 29901
	}
	for _, entry := range entries {
		var parsed catchupTemplate
		if err := json.Unmarshal(entry.Data, &parsed); err != nil {
			continue
		}
		if parsed.ID != 0 && len(parsed.CharChoice) > 0 {
			return parsed.ID, parsed.CharChoice[0]
		}
	}
	return 1, 29901
}
