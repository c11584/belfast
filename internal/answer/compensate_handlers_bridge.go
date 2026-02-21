package answer

import (
	answercompensate "github.com/ggmolly/belfast/internal/answer/compensate"
	"github.com/ggmolly/belfast/internal/connection"
)

func CompensateNotification(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercompensate.CompensateNotification(buffer, client)
}

func GetCompensateList(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercompensate.GetCompensateList(buffer, client)
}

func GetCompensateReward(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercompensate.GetCompensateReward(buffer, client)
}
