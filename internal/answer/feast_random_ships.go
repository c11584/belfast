package answer

import (
	"errors"
	"sort"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FeastRandomShips(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_26158
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 26159, err
	}

	response := &protobuf.SC_26159{
		Ret:        proto.Uint32(feastFailureResult),
		PartyRoles: []*protobuf.P_PARTY_ROLE{},
	}
	actID := payload.GetActId()
	if actID == 0 {
		return client.SendMessage(26159, response)
	}

	active, template, err := isFeastActivityActive(actID, time.Now())
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return client.SendMessage(26159, response)
		}
		return 0, 26159, err
	}
	if !active {
		return client.SendMessage(26159, response)
	}

	partyRoles, ok := buildFeastPartyRoles(payload.GetShipGroupId(), flattenUintSetFromJSON(template.ConfigData))
	if !ok {
		return client.SendMessage(26159, response)
	}

	state, err := orm.GetOrCreateFeastState(client.Commander.CommanderID, actID)
	if err != nil {
		return 0, 26159, err
	}
	state.PartyRoles = partyRoles
	state.RefreshTime = uint32(time.Now().Add(time.Hour).Unix())
	if err := orm.SaveFeastState(state); err != nil {
		return 0, 26159, err
	}

	response.Ret = proto.Uint32(0)
	response.RefreshTime = proto.Uint32(state.RefreshTime)
	response.PartyRoles = feastPartyRolesToProto(state.PartyRoles)
	return client.SendMessage(26159, response)
}

func buildFeastPartyRoles(shipGroupIDs []uint32, allowed map[uint32]struct{}) ([]orm.FeastPartyRole, bool) {
	if len(shipGroupIDs) == 0 {
		return nil, false
	}
	seen := make(map[uint32]struct{}, len(shipGroupIDs))
	roles := make([]orm.FeastPartyRole, 0, len(shipGroupIDs))
	for _, shipGroupID := range shipGroupIDs {
		if shipGroupID == 0 {
			return nil, false
		}
		if _, exists := seen[shipGroupID]; exists {
			return nil, false
		}
		if len(allowed) > 0 {
			if _, ok := allowed[shipGroupID]; !ok {
				return nil, false
			}
		}
		seen[shipGroupID] = struct{}{}
		roles = append(roles, orm.FeastPartyRole{Tid: shipGroupID, Bubble: 0, SpeechBubble: 0})
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i].Tid < roles[j].Tid })
	return roles, true
}
