package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestCommanderFriendBatchGetProfilesAndPresence(t *testing.T) {
	client := setupHandlerCommander(t)
	onlineProfile := seedSocialCommander(t, 911001, "Batch Online")
	offlineProfile := seedSocialCommander(t, 911002, "Batch Offline")

	onlineCommander := orm.Commander{CommanderID: onlineProfile.CommanderID}
	if err := onlineCommander.Load(); err != nil {
		t.Fatalf("load online commander: %v", err)
	}
	client.Server.AddClient(&connection.Client{Commander: &onlineCommander, Hash: onlineProfile.CommanderID})

	request := &protobuf.CS_50018{UserIdList: []uint32{onlineProfile.CommanderID, 999999, offlineProfile.CommanderID}}
	buffer, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	if _, _, err := CommanderFriendBatchGet(&buffer, client); err != nil {
		t.Fatalf("CommanderFriendBatchGet failed: %v", err)
	}

	response := &protobuf.SC_50019{}
	decodePacketAt(t, client, 0, 50019, response)
	if len(response.GetUserList()) != 2 {
		t.Fatalf("expected 2 user profiles, got %d", len(response.GetUserList()))
	}

	first := response.GetUserList()[0]
	if first.GetId() != onlineProfile.CommanderID {
		t.Fatalf("expected first id %d, got %d", onlineProfile.CommanderID, first.GetId())
	}
	if first.GetName() != onlineProfile.Name || first.GetLv() != onlineProfile.Level || first.GetAdv() != onlineProfile.Manifesto {
		t.Fatalf("unexpected first profile payload: %+v", first)
	}
	if first.GetOnline() != 1 || first.GetPreOnlineTime() != 0 {
		t.Fatalf("expected online profile state, got online=%d pre_online_time=%d", first.GetOnline(), first.GetPreOnlineTime())
	}
	if first.GetDisplay() == nil {
		t.Fatalf("expected display for online profile")
	}
	if first.GetDisplay().GetIcon() != onlineProfile.DisplayIconID ||
		first.GetDisplay().GetSkin() != onlineProfile.DisplaySkinID ||
		first.GetDisplay().GetIconFrame() != onlineProfile.SelectedIconFrameID ||
		first.GetDisplay().GetChatFrame() != onlineProfile.SelectedChatFrameID ||
		first.GetDisplay().GetIconTheme() != onlineProfile.DisplayIconThemeID {
		t.Fatalf("unexpected online display payload: %+v", first.GetDisplay())
	}

	second := response.GetUserList()[1]
	if second.GetId() != offlineProfile.CommanderID {
		t.Fatalf("expected second id %d, got %d", offlineProfile.CommanderID, second.GetId())
	}
	if second.GetOnline() != 0 {
		t.Fatalf("expected offline profile online=0, got %d", second.GetOnline())
	}
	if second.GetPreOnlineTime() == 0 || second.GetPreOnlineTime() != offlineProfile.LastLoginUnix {
		t.Fatalf("expected offline pre_online_time=%d, got %d", offlineProfile.LastLoginUnix, second.GetPreOnlineTime())
	}
}

func TestCommanderFriendBatchGetEmptyInput(t *testing.T) {
	client := setupHandlerCommander(t)

	request := &protobuf.CS_50018{UserIdList: nil}
	buffer, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	if _, _, err := CommanderFriendBatchGet(&buffer, client); err != nil {
		t.Fatalf("CommanderFriendBatchGet failed: %v", err)
	}

	response := &protobuf.SC_50019{}
	decodePacketAt(t, client, 0, 50019, response)
	if len(response.GetUserList()) != 0 {
		t.Fatalf("expected empty user list, got %d", len(response.GetUserList()))
	}
}

func TestCommanderFriendBatchGetDecodeFailure(t *testing.T) {
	client := setupHandlerCommander(t)
	invalid := []byte{0xff, 0x00}
	if _, outID, err := CommanderFriendBatchGet(&invalid, client); err == nil || outID != 50019 {
		t.Fatalf("expected decode error with outgoing id 50019, got err=%v out=%d", err, outID)
	}
}
