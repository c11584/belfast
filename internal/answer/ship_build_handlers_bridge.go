package answer

import (
	answershipbuild "github.com/ggmolly/belfast/internal/answer/shipbuild"
	"github.com/ggmolly/belfast/internal/connection"
)

func ShipBuild(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answershipbuild.ShipBuild(buffer, client)
}

func BuildFinish(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answershipbuild.BuildFinish(buffer, client)
}

func BuildQuickFinish(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answershipbuild.BuildQuickFinish(buffer, client)
}

func OngoingBuilds(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answershipbuild.OngoingBuilds(buffer, client)
}
