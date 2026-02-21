package entrypoint

import (
	"testing"

	"github.com/ggmolly/belfast/internal/packets"
)

func TestRegisterPacketsIncludesGuildShopPurchaseHandler(t *testing.T) {
	previous := packets.PacketDecisionFn
	t.Cleanup(func() {
		packets.PacketDecisionFn = previous
	})

	packets.PacketDecisionFn = map[int][]packets.PacketHandler{}
	registerPackets()

	handlers, ok := packets.PacketDecisionFn[60035]
	if !ok {
		t.Fatalf("expected packet 60035 to be registered")
	}
	if len(handlers) == 0 {
		t.Fatalf("expected packet 60035 to have handlers")
	}
}
