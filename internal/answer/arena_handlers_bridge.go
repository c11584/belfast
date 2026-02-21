package answer

import (
	answerarena "github.com/ggmolly/belfast/internal/answer/arena"
	"github.com/ggmolly/belfast/internal/connection"
)

func GetArenaShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerarena.GetArenaShop(buffer, client)
}

func RefreshArenaShop(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerarena.RefreshArenaShop(buffer, client)
}
