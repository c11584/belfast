package answer

import (
	answerpermanentactivity "github.com/ggmolly/belfast/internal/answer/permanentactivity"
	"github.com/ggmolly/belfast/internal/connection"
)

func PermanentActivites(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answerpermanentactivity.PermanentActivites(buffer, client)
}
