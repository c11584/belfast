package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	friendOperationSuccess uint32 = 0
	friendOperationFailure uint32 = 1
	friendOperationMaxed   uint32 = 6
	maxFriendCount         uint32 = 50
)

func buildFriendInfo(profile orm.CommanderSocialProfile, online bool) *protobuf.FRIEND_INFO {
	onlineValue := uint32(0)
	if online {
		onlineValue = 1
	}
	now := uint32(time.Now().UTC().Unix())
	return &protobuf.FRIEND_INFO{
		Id:            proto.Uint32(profile.CommanderID),
		Name:          proto.String(profile.Name),
		Lv:            proto.Uint32(profile.Level),
		Adv:           proto.String(""),
		Online:        proto.Uint32(onlineValue),
		PreOnlineTime: proto.Uint32(now),
	}
}

func buildPlayerInfo(profile orm.CommanderSocialProfile) *protobuf.PLAYER_INFO_P50 {
	return &protobuf.PLAYER_INFO_P50{
		Id:   proto.Uint32(profile.CommanderID),
		Name: proto.String(profile.Name),
		Lv:   proto.Uint32(profile.Level),
	}
}
