package answer

import (
	answersecondarypwd "github.com/ggmolly/belfast/internal/answer/secondarypwd"
	"github.com/ggmolly/belfast/internal/connection"
)

const secondaryPasswordMaxFailures = 5

func ConfirmSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersecondarypwd.ConfirmSecondaryPasswordCommandResponse(buffer, client)
}

func FetchSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersecondarypwd.FetchSecondaryPasswordCommandResponse(buffer, client)
}

func SetSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersecondarypwd.SetSecondaryPasswordCommandResponse(buffer, client)
}

func SetSecondaryPasswordSettingsCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answersecondarypwd.SetSecondaryPasswordSettingsCommandResponse(buffer, client)
}
