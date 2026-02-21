package answer

import (
	answermiscops "github.com/ggmolly/belfast/internal/answer/miscops"
	"github.com/ggmolly/belfast/internal/connection"
)

func CheaterMark(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answermiscops.CheaterMark(buffer, client)
}

func ClickMingShi(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answermiscops.ClickMingShi(buffer, client)
}

func PlayerBuffs(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answermiscops.PlayerBuffs(buffer, client)
}

func OwnedItems(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answermiscops.OwnedItems(buffer, client)
}

func HandleConsoleCommand(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answermiscops.HandleConsoleCommand(buffer, client)
}

func GiveResources(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answermiscops.GiveResources(buffer, client)
}

func GiveItem(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answermiscops.GiveItem(buffer, client)
}

func SellItem(buffer *[]byte, client *connection.Client) (int, int, error) {
	return answermiscops.SellItem(buffer, client)
}
