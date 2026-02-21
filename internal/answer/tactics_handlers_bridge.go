package answer

import (
	answertactics "github.com/ggmolly/belfast/internal/answer/tactics"
	"github.com/ggmolly/belfast/internal/connection"
)

const (
	lessonResultOK     = 0
	lessonResultFailed = 1

	skillCancelTypeAuto   = 0
	skillCancelTypeManual = 1

	lessonQuickFinishResultOK                = 0
	lessonQuickFinishResultInvalidRoom       = 1
	lessonQuickFinishResultSessionNotFound   = 2
	lessonQuickFinishResultAllowanceExceeded = 3
	lessonQuickFinishResultInvalidState      = 4

	itemConfigCategory    = "sharecfgdata/item_data_statistics.json"
	shipTemplateCategory  = "sharecfgdata/ship_data_template.json"
	skillTemplateCategory = "sharecfgdata/skill_data_template.json"
)

func StartLearnTactics(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answertactics.StartLearnTactics(buffer, client)
}

func CancelLearnTactics(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answertactics.CancelLearnTactics(buffer, client)
}

func QuickFinishLearnTactics(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answertactics.QuickFinishLearnTactics(buffer, client)
}
