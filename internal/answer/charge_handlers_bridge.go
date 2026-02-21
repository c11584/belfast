package answer

import (
	answercharge "github.com/ggmolly/belfast/internal/answer/charge"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
)

type ChargeSuccessEvent = answercharge.ChargeSuccessEvent

func HandleChargeStart(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercharge.HandleChargeStart(buffer, client)
}

func RefundHandleChargeStart(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercharge.RefundHandleChargeStart(buffer, client)
}

func HandleChargeConfirm(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercharge.HandleChargeConfirm(buffer, client)
}

func HandleChargeFailure(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercharge.HandleChargeFailure(buffer, client)
}

func ApplyChargeSuccessEvent(commander *orm.Commander, client *connection.Client, event ChargeSuccessEvent) error {
	return answercharge.ApplyChargeSuccessEvent(commander, client, event)
}

func GetChargeList(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercharge.GetChargeList(buffer, client)
}

func GetRefundInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answercharge.GetRefundInfo(buffer, client)
}
