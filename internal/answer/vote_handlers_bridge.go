package answer

import (
	answervote "github.com/ggmolly/belfast/internal/answer/vote"
	"github.com/ggmolly/belfast/internal/connection"
)

func FetchVoteInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answervote.FetchVoteInfo(buffer, client)
}

func FetchVoteTicketInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answervote.FetchVoteTicketInfo(buffer, client)
}
