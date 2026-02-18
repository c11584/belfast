package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FeastGetData(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_26156
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 26157, err
	}

	response := &protobuf.SC_26157{
		Ret:          proto.Uint32(feastFailureResult),
		PartyRoles:   []*protobuf.P_PARTY_ROLE{},
		SpecialRoles: []*protobuf.P_SPECIAL_ROLE{},
	}
	if payload.GetActId() == 0 {
		return client.SendMessage(26157, response)
	}

	active, _, err := isFeastActivityActive(payload.GetActId(), time.Now())
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return client.SendMessage(26157, response)
		}
		return 0, 26157, err
	}
	if !active {
		return client.SendMessage(26157, response)
	}

	state, err := orm.GetOrCreateFeastState(client.Commander.CommanderID, payload.GetActId())
	if err != nil {
		return 0, 26157, err
	}

	response.Ret = proto.Uint32(0)
	response.PartyRoles = feastPartyRolesToProto(state.PartyRoles)
	response.SpecialRoles = feastSpecialRolesToProto(state.SpecialRoles)
	if state.RefreshTime > 0 {
		response.RefreshTime = proto.Uint32(state.RefreshTime)
	}
	return client.SendMessage(26157, response)
}
