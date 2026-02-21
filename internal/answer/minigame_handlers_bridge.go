package answer

import (
	"github.com/ggmolly/belfast/internal/answer/minigame"
	"github.com/ggmolly/belfast/internal/connection"
)

const (
	miniGameOpResultSuccess = uint32(0)
	miniGameOpResultFailure = uint32(1)

	miniGameCmdComplete    = uint32(1)
	miniGameCmdUltimate    = uint32(2)
	miniGameCmdSpecialGame = uint32(3)
	miniGameCmdHighScore   = uint32(4)
	miniGameCmdPlay        = uint32(5)
	miniGameCmdSuccessData = uint32(101)
)

func MiniGameHubData(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.MiniGameHubData(buffer, client)
}

func GetMiniGameShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.GetMiniGameShop(buffer, client)
}

func MiniGameShopBuy(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.MiniGameShopBuy(buffer, client)
}

func MiniGameShopRefresh(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.MiniGameShopRefresh(buffer, client)
}

func MiniGameOperation(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.MiniGameOperation(buffer, client)
}

func MiniGameOperationBatch(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.MiniGameOperationBatch(buffer, client)
}

func MiniGameFriendRank(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.MiniGameFriendRank(buffer, client)
}

func MiniGameTimeSubmit(buffer *[]byte, client *connection.Client) (int, int, error) {
	return minigame.MiniGameTimeSubmit(buffer, client)
}
