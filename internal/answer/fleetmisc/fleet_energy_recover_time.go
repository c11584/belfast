package fleetmisc

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FleetEnergyRecoverTime(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_12031
	nowUnix := uint32(time.Now().Unix())
	nextTick, err := orm.ApplyCommanderMoraleRecovery(client.Commander.CommanderID, nowUnix)
	if err != nil {
		return 0, 12031, err
	}
	if nextTick == 0 {
		nextTick = nowUnix + 6*60
	}
	response.EnergyAutoIncreaseTime = proto.Uint32(nextTick)
	return client.SendMessage(12031, &response)
}
