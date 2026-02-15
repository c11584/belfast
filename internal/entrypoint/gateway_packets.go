package entrypoint

import (
	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/packets"
)

func registerGatewayPackets() {
	packets.RegisterPacketHandler(10800, []packets.PacketHandler{answer.HandleGatewayUpdateCheck})
	packets.RegisterPacketHandler(10700, []packets.PacketHandler{answer.GatewayPackInfo})
	packets.RegisterPacketHandler(8239, []packets.PacketHandler{answer.WriteServerListHTTPResponse})
	packets.RegisterPacketHandler(10018, []packets.PacketHandler{answer.HandleServerStateCheck})
	packets.RegisterPacketHandler(10001, []packets.PacketHandler{answer.RegisterAccount})
	packets.RegisterPacketHandler(10020, []packets.PacketHandler{answer.HandleGatewayAuthConfirm})
}
