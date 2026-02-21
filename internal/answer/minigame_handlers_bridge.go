package answer

import (
	"github.com/ggmolly/belfast/internal/answer/minigame"
	"github.com/ggmolly/belfast/internal/connection"
)

func MiniGameHubData(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.MiniGameHubData(buffer, client)
}

func GetMiniGameShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.GetMiniGameShop(buffer, client)
}
