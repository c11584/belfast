package dorm3d

import (
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func parseJSONUint(value any) (uint32, bool) {
	if number, ok := value.(float64); ok {
		return uint32(number), true
	}
	if number, ok := value.(int); ok {
		return uint32(number), true
	}
	if number, ok := value.(uint32); ok {
		return number, true
	}
	return 0, false
}

func newDropInfo(dropType uint32, dropID uint32, count uint32) *protobuf.DROPINFO {
	return &protobuf.DROPINFO{
		Type:   proto.Uint32(dropType),
		Id:     proto.Uint32(dropID),
		Number: proto.Uint32(count),
	}
}
