package answer_test

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestFeastRandomShipsSuccessPersistsState(t *testing.T) {
	commanderID := uint32(9012)
	actID := uint32(60020)
	cleanupFeastData(t, commanderID, actID)
	seedFeastActivityTemplate(t, actID, time.Now().Add(24*time.Hour), "[3001,3002,3003]")
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupFeastData(t, commanderID, actID)

	payload := &protobuf.CS_26158{ActId: proto.Uint32(actID), ShipGroupId: []uint32{3002, 3001}}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := answer.FeastRandomShips(&buf, client); err != nil {
		t.Fatalf("FeastRandomShips failed: %v", err)
	}

	response := &protobuf.SC_26159{}
	decodeTestPacket(t, client, 26159, response)
	if response.GetRet() != 0 {
		t.Fatalf("expected success ret, got %d", response.GetRet())
	}
	if len(response.GetPartyRoles()) != 2 {
		t.Fatalf("expected 2 party roles, got %d", len(response.GetPartyRoles()))
	}
	if response.GetPartyRoles()[0].GetTid() != 3001 || response.GetPartyRoles()[1].GetTid() != 3002 {
		t.Fatalf("expected sorted party roles, got %+v", response.GetPartyRoles())
	}
	if response.GetRefreshTime() <= uint32(time.Now().Unix()) {
		t.Fatalf("expected refresh_time in the future")
	}

	state, err := orm.GetFeastState(commanderID, actID)
	if err != nil {
		t.Fatalf("failed to load feast state: %v", err)
	}
	if len(state.PartyRoles) != 2 || state.PartyRoles[0].Tid != 3001 || state.PartyRoles[1].Tid != 3002 {
		t.Fatalf("unexpected persisted party roles: %+v", state.PartyRoles)
	}
}

func TestFeastRandomShipsValidationFailureNoMutation(t *testing.T) {
	commanderID := uint32(9013)
	actID := uint32(60021)
	cleanupFeastData(t, commanderID, actID)
	seedFeastActivityTemplate(t, actID, time.Now().Add(24*time.Hour), "[4001]")
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupFeastData(t, commanderID, actID)

	if err := orm.SaveFeastState(&orm.FeastState{
		CommanderID: commanderID,
		ActID:       actID,
		RefreshTime: 123,
		PartyRoles:  []orm.FeastPartyRole{{Tid: 4001, Bubble: 1, SpeechBubble: 1}},
	}); err != nil {
		t.Fatalf("seed feast state: %v", err)
	}

	payload := &protobuf.CS_26158{ActId: proto.Uint32(actID), ShipGroupId: []uint32{9999}}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := answer.FeastRandomShips(&buf, client); err != nil {
		t.Fatalf("FeastRandomShips failed: %v", err)
	}

	response := &protobuf.SC_26159{}
	decodeTestPacket(t, client, 26159, response)
	if response.GetRet() == 0 {
		t.Fatalf("expected validation failure")
	}

	state, err := orm.GetFeastState(commanderID, actID)
	if err != nil {
		t.Fatalf("failed to load feast state: %v", err)
	}
	if state.RefreshTime != 123 || len(state.PartyRoles) != 1 || state.PartyRoles[0].Tid != 4001 {
		t.Fatalf("expected state unchanged after failure, got %+v", state)
	}
}
