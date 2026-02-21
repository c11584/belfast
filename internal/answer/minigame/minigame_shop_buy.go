package minigame

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/minigameshop"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const miniGameShopPurchaseFailureResult = uint32(1)

func MiniGameShopBuy(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_26152
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 26153, err
	}

	response := &protobuf.SC_26153{
		Result:   proto.Uint32(miniGameShopPurchaseFailureResult),
		DropList: []*protobuf.DROPINFO{},
	}
	if payload.GetGoodsid() == 0 {
		return client.SendMessage(26153, response)
	}

	config, err := minigameshop.LoadConfig(time.Now())
	if err != nil {
		return 0, 26153, err
	}
	selected := buildMiniGameShopSelections(payload.GetSelected())
	drops, err := minigameshop.Purchase(client.Commander, payload.GetGoodsid(), selected, time.Now(), config)
	if err != nil {
		if errors.Is(err, minigameshop.ErrInvalidPurchasePayload) || errors.Is(err, minigameshop.ErrInsufficientTickets) || errors.Is(err, minigameshop.ErrSoldOut) || errors.Is(err, minigameshop.ErrUnsupportedReward) {
			return client.SendMessage(26153, response)
		}
		return 0, 26153, err
	}

	response.Result = proto.Uint32(0)
	response.DropList = buildMiniGameShopDrops(drops)
	return client.SendMessage(26153, response)
}

func buildMiniGameShopSelections(selected []*protobuf.SELECT_INFO) []minigameshop.PurchaseSelection {
	out := make([]minigameshop.PurchaseSelection, 0, len(selected))
	for _, pick := range selected {
		out = append(out, minigameshop.PurchaseSelection{ID: pick.GetId(), Num: pick.GetNum()})
	}
	return out
}

func buildMiniGameShopDrops(drops []minigameshop.PurchaseDrop) []*protobuf.DROPINFO {
	out := make([]*protobuf.DROPINFO, 0, len(drops))
	for _, drop := range drops {
		out = append(out, &protobuf.DROPINFO{
			Type:   proto.Uint32(drop.Type),
			Id:     proto.Uint32(drop.ID),
			Number: proto.Uint32(drop.Number),
		})
	}
	return out
}
