package answer

import (
	"github.com/ggmolly/belfast/internal/answer/instagramchat"
	"github.com/ggmolly/belfast/internal/connection"
)

func InstagramChatActivateTopic(buffer *[]byte, client *connection.Client) (int, int, error) {
	return instagramchat.InstagramChatActivateTopic(buffer, client)
}

func InstagramChatReply(buffer *[]byte, client *connection.Client) (int, int, error) {
	return instagramchat.InstagramChatReply(buffer, client)
}

func InstagramChatSetCare(buffer *[]byte, client *connection.Client) (int, int, error) {
	return instagramchat.InstagramChatSetCare(buffer, client)
}

func InstagramChatSetSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	return instagramchat.InstagramChatSetSkin(buffer, client)
}

func InstagramChatSetTopic(buffer *[]byte, client *connection.Client) (int, int, error) {
	return instagramchat.InstagramChatSetTopic(buffer, client)
}
