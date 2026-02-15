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

func TestRegisterPacketsIncludesActivityTaskCluster(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[20205]; !ok {
		t.Fatalf("expected handler for CS_20205 to be registered")
	}
	if _, ok := packets.PacketDecisionFn[20207]; !ok {
		t.Fatalf("expected handler for CS_20207 to be registered")
	}
	if _, ok := packets.PacketDecisionFn[20209]; !ok {
		t.Fatalf("expected handler for CS_20209 to be registered")
	}
}
