package answer

import (
	answerplayerops "github.com/ggmolly/belfast/internal/answer/playerops"
	"github.com/ggmolly/belfast/internal/connection"
)

func AttireApply(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.AttireApply(buffer, client)
}

func ChangeManifesto(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.ChangeManifesto(buffer, client)
}

func ChangeSelectedSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.ChangeSelectedSkin(buffer, client)
}

func ChangeShipLockState(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.ChangeShipLockState(buffer, client)
}

func LastLogin(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.LastLogin(buffer, client)
}

func LastOnlineInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.LastOnlineInfo(buffer, client)
}

func SendHeartbeat(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.SendHeartbeat(buffer, client)
}

func ResourcesInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.ResourcesInfo(buffer, client)
}

func PlayerExist(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerplayerops.PlayerExist(buffer, client)
}
