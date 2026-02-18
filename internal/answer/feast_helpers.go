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
	startTime, stopTime, ok := parseActivityTimerWindow(template.Time)
	if !ok {
		return false, activityTemplate{}, nil
	}
	nowUnix := uint32(now.Unix())
	if nowUnix < startTime || nowUnix > stopTime {
		return false, activityTemplate{}, nil
	}
	return true, template, nil
}

func parseActivityTimerWindow(raw json.RawMessage) (uint32, uint32, bool) {
	var value []any
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0, 0, false
	}
	if len(value) < 3 {
		return 0, 0, false
	}
	typeTag, ok := value[0].(string)
	if !ok || typeTag != "timer" {
		return 0, 0, false
	}

	start, ok := parseActivityTimerPoint(value[1])
	if !ok {
		return 0, 0, false
	}
	stop, ok := parseActivityTimerPoint(value[2])
	if !ok {
		return 0, 0, false
	}
	if stop < start {
		return 0, 0, false
	}
	return start, stop, true
}

func parseActivityTimerPoint(raw any) (uint32, bool) {
	point, ok := raw.([]any)
	if !ok || len(point) != 2 {
		return 0, false
	}
	date, ok := point[0].([]any)
	if !ok || len(date) != 3 {
		return 0, false
	}
	clock, ok := point[1].([]any)
	if !ok || len(clock) != 3 {
		return 0, false
	}

	year, ok := parseJSONInt(date[0])
	if !ok {
		return 0, false
	}
	month, ok := parseJSONInt(date[1])
	if !ok {
		return 0, false
	}
	day, ok := parseJSONInt(date[2])
	if !ok {
		return 0, false
	}
	hour, ok := parseJSONInt(clock[0])
	if !ok {
		return 0, false
	}
	minute, ok := parseJSONInt(clock[1])
	if !ok {
		return 0, false
	}
	second, ok := parseJSONInt(clock[2])
	if !ok {
		return 0, false
	}
	return uint32(time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC).Unix()), true
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
