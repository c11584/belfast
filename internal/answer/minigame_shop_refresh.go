package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/minigameshop"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const miniGameShopRefreshFailureResult = uint32(1)

func MiniGameShopRefresh(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_26154
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 26155, err
	}

	response := &protobuf.SC_26155{
		Result:        proto.Uint32(miniGameShopRefreshFailureResult),
		NextFlashTime: []uint32{},
	}
	if payload.GetType() != 0 {
		return client.SendMessage(26155, response)
	}

	config, err := minigameshop.LoadConfig(time.Now())
	if err != nil {
		return 0, 26155, err
	}
	state, _, err := minigameshop.ForceRefresh(client.Commander.CommanderID, time.Now(), config)
	if err != nil {
		return 0, 26155, err
	}

	response.Result = proto.Uint32(0)
	response.NextFlashTime = []uint32{state.NextRefreshTime}
	return client.SendMessage(26155, response)
}
