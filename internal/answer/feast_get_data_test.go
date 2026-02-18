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

func TestFeastGetDataSuccess(t *testing.T) {
	commanderID := uint32(9010)
	actID := uint32(60010)
	cleanupFeastData(t, commanderID, actID)
	seedFeastActivityTemplate(t, actID, time.Now().Add(24*time.Hour), "[101,102]")
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupFeastData(t, commanderID, actID)

	err := orm.SaveFeastState(&orm.FeastState{
		CommanderID: commanderID,
		ActID:       actID,
		RefreshTime: uint32(time.Now().Add(time.Hour).Unix()),
		PartyRoles: []orm.FeastPartyRole{{
			Tid:          101,
			Bubble:       3,
			SpeechBubble: 4,
		}},
		SpecialRoles: []orm.FeastSpecialRole{{
			Tid:   102,
			State: 1,
			Gift:  2,
		}},
	})
	if err != nil {
		t.Fatalf("failed to seed feast state: %v", err)
	}

	payload := &protobuf.CS_26156{ActId: proto.Uint32(actID)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := answer.FeastGetData(&buf, client); err != nil {
		t.Fatalf("FeastGetData failed: %v", err)
	}

	response := &protobuf.SC_26157{}
	decodeTestPacket(t, client, 26157, response)
	if response.GetRet() != 0 {
		t.Fatalf("expected success ret, got %d", response.GetRet())
	}
	if len(response.GetPartyRoles()) != 1 || response.GetPartyRoles()[0].GetTid() != 101 {
		t.Fatalf("unexpected party roles: %+v", response.GetPartyRoles())
	}
	if len(response.GetSpecialRoles()) != 1 || response.GetSpecialRoles()[0].GetTid() != 102 {
		t.Fatalf("unexpected special roles: %+v", response.GetSpecialRoles())
	}
	if response.RefreshTime == nil || response.GetRefreshTime() == 0 {
		t.Fatalf("expected refresh_time to be present")
	}
}

func TestFeastGetDataFailsForInvalidOrInactiveActivity(t *testing.T) {
	commanderID := uint32(9011)
	invalidActID := uint32(60011)
	inactiveActID := uint32(60012)
	cleanupFeastData(t, commanderID, invalidActID)
	cleanupFeastData(t, commanderID, inactiveActID)
	seedFeastActivityTemplate(t, inactiveActID, time.Now().Add(-time.Hour), "[]")
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupFeastData(t, commanderID, invalidActID)
	defer cleanupFeastData(t, commanderID, inactiveActID)

	for _, actID := range []uint32{invalidActID, inactiveActID} {
		payload := &protobuf.CS_26156{ActId: proto.Uint32(actID)}
		buf, err := proto.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		if _, _, err := answer.FeastGetData(&buf, client); err != nil {
			t.Fatalf("FeastGetData failed: %v", err)
		}
		response := &protobuf.SC_26157{}
		decodeTestPacket(t, client, 26157, response)
		if response.GetRet() == 0 {
			t.Fatalf("expected failure ret for act %d", actID)
		}
	}
}
