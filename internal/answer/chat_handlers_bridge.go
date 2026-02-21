package answer

import (
	answerchat "github.com/ggmolly/belfast/internal/answer/chat"
	"github.com/ggmolly/belfast/internal/connection"
)

func ChatRoomChange(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerchat.ChatRoomChange(buffer, client)
}

func ReceiveChatMessage(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerchat.ReceiveChatMessage(buffer, client)
}
