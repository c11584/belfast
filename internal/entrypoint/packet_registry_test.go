package entrypoint

import (
	"testing"

	"github.com/ggmolly/belfast/internal/packets"
)

func TestRegisterPacketsIncludes14004(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[14004]; !ok {
		t.Fatalf("expected handler for CS_14004 to be registered")
	}
}

func TestRegisterPacketsIncludesLoveLetterGetAll(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[12406]; !ok {
		t.Fatalf("expected handler for CS_12406 to be registered")
	}

}

func TestRegisterPacketsIncludesIslandDelegationAwardClaim(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[21505]; !ok {
		t.Fatalf("expected handler for CS_21505 to be registered")
	}
}

func TestRegisterPacketsIncludesIslandAgoraThemeDelete(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[21319]; !ok {
		t.Fatalf("expected handler for CS_21319 to be registered")
	}
}

func TestRegisterPacketsIncludesIslandShipBreakout(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[21601]; !ok {
		t.Fatalf("expected handler for CS_21601 to be registered")
	}
}

func TestRegisterPacketsIncludesIslandFollowerOp(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[21630]; !ok {
		t.Fatalf("expected handler for CS_21630 to be registered")
	}
}

func TestRegisterPacketsIncludesIslandShipOrderFlow(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	for _, packetID := range []int{21408, 21416, 21429} {
		if _, ok := packets.PacketDecisionFn[packetID]; !ok {
			t.Fatalf("expected handler for CS_%d to be registered", packetID)
		}
	}
}
