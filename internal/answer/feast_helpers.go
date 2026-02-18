package answer

import (
	"encoding/json"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const feastFailureResult = uint32(1)

func isFeastActivityActive(actID uint32, now time.Time) (bool, activityTemplate, error) {
	template, err := loadActivityTemplate(actID)
	if err != nil {
		return false, activityTemplate{}, err
	}
	stopTime := activityStopTime(template.Time)
	if stopTime != 0 && uint32(now.Unix()) > stopTime {
		return false, activityTemplate{}, nil
	}
	return true, template, nil
}

func flattenUintSetFromJSON(raw json.RawMessage) map[uint32]struct{} {
	if len(raw) == 0 {
		return map[uint32]struct{}{}
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return map[uint32]struct{}{}
	}
	out := make(map[uint32]struct{})
	flattenUintSet(out, value)
	return out
}

func flattenUintSet(out map[uint32]struct{}, value any) {
	if out == nil {
		return
	}
	if number, ok := parseJSONUint(value); ok {
		if number > 0 {
			out[number] = struct{}{}
		}
		return
	}
	list, ok := value.([]any)
	if !ok {
		return
	}
	for _, entry := range list {
		flattenUintSet(out, entry)
	}
}

func feastPartyRolesToProto(roles []orm.FeastPartyRole) []*protobuf.P_PARTY_ROLE {
	out := make([]*protobuf.P_PARTY_ROLE, 0, len(roles))
	for _, role := range roles {
		out = append(out, &protobuf.P_PARTY_ROLE{
			Tid:          proto.Uint32(role.Tid),
			Bubble:       proto.Uint32(role.Bubble),
			SpeechBubble: proto.Uint32(role.SpeechBubble),
		})
	}
	return out
}

func feastSpecialRolesToProto(roles []orm.FeastSpecialRole) []*protobuf.P_SPECIAL_ROLE {
	out := make([]*protobuf.P_SPECIAL_ROLE, 0, len(roles))
	for _, role := range roles {
		out = append(out, &protobuf.P_SPECIAL_ROLE{
			Tid:   proto.Uint32(role.Tid),
			State: proto.Uint32(role.State),
			Gift:  proto.Uint32(role.Gift),
		})
	}
	return out
}
