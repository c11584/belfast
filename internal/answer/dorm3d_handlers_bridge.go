package answer

import (
	"github.com/ggmolly/belfast/internal/answer/dorm3d"
	"github.com/ggmolly/belfast/internal/connection"
)

func Dorm3dApartmentData(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dApartmentData(buffer, client)
}

func Dorm3dCollectionItem(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dCollectionItem(buffer, client)
}

func Dorm3dChangeSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dChangeSkin(buffer, client)
}

func Dorm3dTalk(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dTalk(buffer, client)
}

func Dorm3dSetCall(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dSetCall(buffer, client)
}

func Dorm3dSetSkinHiddenParts(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dSetSkinHiddenParts(buffer, client)
}

func Dorm3dChatSetBackground(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dChatSetBackground(buffer, client)
}

func Dorm3dChatSetCare(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dChatSetCare(buffer, client)
}

func Dorm3dInstagramSetTopic(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dInstagramSetTopic(buffer, client)
}

func Dorm3dRecordVisit(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dRecordVisit(buffer, client)
}

func Dorm3dChatTriggerEvent(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dChatTriggerEvent(buffer, client)
}

func Dorm3dTriggerFavor(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dTriggerFavor(buffer, client)
}

func Dorm3dApartmentLevelUp(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dApartmentLevelUp(buffer, client)
}

func HandleDorm3dGiveGift(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.HandleDorm3dGiveGift(buffer, client)
}

func HandleDorm3dInstagramAction(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.HandleDorm3dInstagramAction(buffer, client)
}

func Dorm3dInstagramDiscuss(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dInstagramDiscuss(buffer, client)
}

func SelectDorm3dEnter(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.SelectDorm3dEnter(buffer, client)
}

func Dorm3dRoomUnlock(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dRoomUnlock(buffer, client)
}

func Dorm3dReplaceFurniture(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dReplaceFurniture(buffer, client)
}

func Dorm3dRoomInviteUnlock(buffer *[]byte, client *connection.Client) (int, int, error) {
	return dorm3d.Dorm3dRoomInviteUnlock(buffer, client)
}
