package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EducateGetTargetAward(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_27035
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 27036, err
	}

	response := &protobuf.SC_27036{Result: proto.Uint32(educateResultFailed), Drops: []*protobuf.CHILD_DROP{}}
	if client.Commander == nil || payload.GetType() != 0 {
		return client.SendMessage(27036, response)
	}

	targets, tasks, err := loadEducateTargetAndTaskConfigs()
	if err != nil {
		return 0, 27036, err
	}
	targetID := chooseEducateTargetID(targets)
	if targetID == 0 {
		return client.SendMessage(27036, response)
	}
	target := targets[targetID]
	claimedFlag := educateFlagID(educateFlagTargetAwardBase, targetID)
	claimed, err := hasEducateFlag(client.Commander.CommanderID, claimedFlag)
	if err != nil {
		return 0, 27036, err
	}
	if claimed {
		return client.SendMessage(27036, response)
	}

	progress := uint32(0)
	for _, taskID := range target.IDs {
		if task, ok := tasks[taskID]; ok {
			progress += task.TaskTargetProgress
		}
	}
	if progress < target.TargetProgress {
		return client.SendMessage(27036, response)
	}

	if err := setEducateFlag(client.Commander.CommanderID, claimedFlag); err != nil {
		return 0, 27036, err
	}
	if drop := toChildDrop(target.DropDisplay); drop != nil {
		response.Drops = append(response.Drops, drop)
	}
	response.Result = proto.Uint32(educateResultOK)
	return client.SendMessage(27036, response)
}
