package answer

import (
	answersimpleops "github.com/ggmolly/belfast/internal/answer/simpleops"
	"github.com/ggmolly/belfast/internal/connection"
)

func ApartmentTrackEvent(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.ApartmentTrackEvent(buffer, client)
}

func CancelCommonFlagCommand(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.CancelCommonFlagCommand(buffer, client)
}

func ChangeLivingAreaCover(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.ChangeLivingAreaCover(buffer, client)
}

func HandleLegacyItemOperation(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.HandleLegacyItemOperation(buffer, client)
}

func MainSceneTracking(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.MainSceneTracking(buffer, client)
}

func NewTracking(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.NewTracking(buffer, client)
}

func SetGuildDuty(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.SetGuildDuty(buffer, client)
}

func TrackCommand(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.TrackCommand(buffer, client)
}

func UpdateCommonFlagCommand(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.UpdateCommonFlagCommand(buffer, client)
}

func UpdateGuideIndex(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.UpdateGuideIndex(buffer, client)
}

func UpdateSecretaries(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.UpdateSecretaries(buffer, client)
}

func UpdateStory(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.UpdateStory(buffer, client)
}

func UpdateStoryList(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.UpdateStoryList(buffer, client)
}

func UrExchangeTracking(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersimpleops.UrExchangeTracking(buffer, client)
}
