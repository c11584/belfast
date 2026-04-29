package debug

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
)

// InsertPacket 将收到的包记录到数据库用于调试
// 跳过 8239（心跳包）和 10999（断开连接包），避免产生大量无用记录
func InsertPacket(packetId int, payload *[]uint8) {
	if packetId == 8239 || packetId == 10999 {
		return
	}
	err := orm.InsertDebugPacket(len(*payload), packetId, *payload)
	if err != nil {
		logger.LogEvent("Debug", "InsertPacket", fmt.Sprintf("Failed to insert packet %d", packetId), logger.LOG_LEVEL_ERROR)
		logger.LogEvent("Debug", "InsertPacket", err.Error(), logger.LOG_LEVEL_ERROR)
	}
}
