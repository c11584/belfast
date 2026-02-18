package minigameshop

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
)

const gameRoomShopCategory = "ShareCfg/gameroom_shop_template.json"

const miniGameShopTicketResourceID = uint32(12)

var (
	ErrInvalidPurchasePayload = errors.New("invalid minigame shop purchase payload")
	ErrInsufficientTickets    = errors.New("insufficient minigame tickets")
	ErrSoldOut                = errors.New("minigame shop good sold out")
	ErrUnsupportedReward      = errors.New("unsupported minigame reward type")
)

type shopEntry struct {
	ID                 uint32     `json:"id"`
	GoodsPurchaseLimit uint32     `json:"goods_purchase_limit"`
	Goods              []uint32   `json:"goods"`
	DropType           uint32     `json:"drop_type"`
	Price              uint32     `json:"price"`
	Num                uint32     `json:"num"`
	Time               [][][3]int `json:"time"`
	Order              uint32     `json:"order"`
}

type PurchaseSelection struct {
	ID  uint32
	Num uint32
}

type PurchaseDrop struct {
	Type   uint32
	ID     uint32
	Number uint32
}

type Config struct {
	Goods []shopEntry
}

type RefreshOptions struct {
	NextRefreshTime uint32
}

func LoadConfig(now time.Time) (*Config, error) {
	entries, err := orm.ListConfigEntries(gameRoomShopCategory)
	if err != nil {
		return nil, err
	}
	goods := make([]shopEntry, 0, len(entries))
	for _, entry := range entries {
		var configEntry shopEntry
		if err := json.Unmarshal(entry.Data, &configEntry); err != nil {
			return nil, err
		}
		if configEntry.ID == 0 {
			continue
		}
		if !isWithinTime(now, configEntry.Time) {
			continue
		}
		goods = append(goods, configEntry)
	}
	sort.Slice(goods, func(i, j int) bool {
		if goods[i].Order != goods[j].Order {
			return goods[i].Order < goods[j].Order
		}
		return goods[i].ID < goods[j].ID
	})
	return &Config{Goods: goods}, nil
}

func EnsureState(commanderID uint32, now time.Time, config *Config) (*orm.MiniGameShopState, []orm.MiniGameShopGood, error) {
	state, err := orm.GetMiniGameShopState(commanderID)
	if err != nil {
		if !db.IsNotFound(err) {
			return nil, nil, err
		}
		state = &orm.MiniGameShopState{
			CommanderID:     commanderID,
			NextRefreshTime: nextDailyReset(now),
		}
		if err := orm.CreateMiniGameShopState(*state); err != nil {
			return nil, nil, err
		}
		goods, err := RefreshGoods(commanderID, config, RefreshOptions{
			NextRefreshTime: state.NextRefreshTime,
		})
		if err != nil {
			return nil, nil, err
		}
		return state, goods, nil
	}
	goods, err := LoadGoods(commanderID)
	if err != nil {
		return nil, nil, err
	}
	return state, goods, nil
}

func RefreshIfNeeded(commanderID uint32, now time.Time, config *Config) (*orm.MiniGameShopState, []orm.MiniGameShopGood, error) {
	state, goods, err := EnsureState(commanderID, now, config)
	if err != nil {
		return nil, nil, err
	}
	if now.Unix() >= int64(state.NextRefreshTime) || len(goods) == 0 {
		goods, err = RefreshGoods(commanderID, config, RefreshOptions{
			NextRefreshTime: nextDailyReset(now),
		})
		if err != nil {
			return nil, nil, err
		}
		state, err = orm.GetMiniGameShopState(commanderID)
		if err != nil {
			return nil, nil, err
		}
	}
	return state, goods, nil
}

func RefreshGoods(commanderID uint32, config *Config, options RefreshOptions) ([]orm.MiniGameShopGood, error) {
	goods := buildGoods(commanderID, config)
	if err := orm.RefreshMiniGameShopGoods(commanderID, goods, options.NextRefreshTime); err != nil {
		return nil, err
	}
	return goods, nil
}

func ForceRefresh(commanderID uint32, now time.Time, config *Config) (*orm.MiniGameShopState, []orm.MiniGameShopGood, error) {
	if _, _, err := EnsureState(commanderID, now, config); err != nil {
		return nil, nil, err
	}
	goods, err := RefreshGoods(commanderID, config, RefreshOptions{NextRefreshTime: nextDailyReset(now)})
	if err != nil {
		return nil, nil, err
	}
	state, err := orm.GetMiniGameShopState(commanderID)
	if err != nil {
		return nil, nil, err
	}
	return state, goods, nil
}

func Purchase(commander *orm.Commander, goodsID uint32, selected []PurchaseSelection, now time.Time, config *Config) ([]PurchaseDrop, error) {
	if commander == nil || goodsID == 0 || config == nil {
		return nil, ErrInvalidPurchasePayload
	}
	if _, _, err := RefreshIfNeeded(commander.CommanderID, now, config); err != nil {
		return nil, err
	}
	entry, ok := findGood(config, goodsID)
	if !ok {
		return nil, ErrInvalidPurchasePayload
	}
	rewards, totalUnits, err := resolvePurchaseRewards(entry, selected)
	if err != nil {
		return nil, err
	}
	totalCost := entry.Price * totalUnits

	ctx := context.Background()
	err = orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		var stock uint32
		err := tx.QueryRow(ctx, `
SELECT count
FROM mini_game_shop_goods
WHERE commander_id = $1
  AND goods_id = $2
FOR UPDATE
`, int64(commander.CommanderID), int64(goodsID)).Scan(&stock)
		err = db.MapNotFound(err)
		if err != nil {
			if db.IsNotFound(err) {
				return ErrInvalidPurchasePayload
			}
			return err
		}
		if stock < totalUnits {
			return ErrSoldOut
		}

		if !commander.HasEnoughResource(miniGameShopTicketResourceID, totalCost) {
			return ErrInsufficientTickets
		}
		if err := commander.ConsumeResourceTx(ctx, tx, miniGameShopTicketResourceID, totalCost); err != nil {
			return ErrInsufficientTickets
		}

		res, err := tx.Exec(ctx, `
UPDATE mini_game_shop_goods
SET count = count - $3
WHERE commander_id = $1
  AND goods_id = $2
  AND count >= $3
`, int64(commander.CommanderID), int64(goodsID), int64(totalUnits))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return ErrSoldOut
		}

		for _, reward := range rewards {
			switch reward.Type {
			case consts.DROP_TYPE_RESOURCE:
				if err := commander.AddResourceTx(ctx, tx, reward.ID, reward.Number); err != nil {
					return err
				}
			case consts.DROP_TYPE_ITEM:
				if err := commander.AddItemTx(ctx, tx, reward.ID, reward.Number); err != nil {
					return err
				}
			case consts.DROP_TYPE_SHIP:
				for i := uint32(0); i < reward.Number; i++ {
					if _, err := commander.AddShipTx(ctx, tx, reward.ID); err != nil {
						return err
					}
				}
			case consts.DROP_TYPE_SKIN:
				for i := uint32(0); i < reward.Number; i++ {
					if err := commander.GiveSkinTx(ctx, tx, reward.ID); err != nil {
						return err
					}
				}
			default:
				return ErrUnsupportedReward
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rewards, nil
}

func LoadGoods(commanderID uint32) ([]orm.MiniGameShopGood, error) {
	return orm.LoadMiniGameShopGoods(commanderID)
}

func buildGoods(commanderID uint32, config *Config) []orm.MiniGameShopGood {
	if config == nil {
		return nil
	}
	goods := make([]orm.MiniGameShopGood, 0, len(config.Goods))
	for _, entry := range config.Goods {
		count := entry.GoodsPurchaseLimit
		if count == 0 {
			count = 1
		}
		goods = append(goods, orm.MiniGameShopGood{
			CommanderID: commanderID,
			GoodsID:     entry.ID,
			Count:       count,
		})
	}
	return goods
}

func findGood(config *Config, goodsID uint32) (shopEntry, bool) {
	if config == nil {
		return shopEntry{}, false
	}
	for _, good := range config.Goods {
		if good.ID == goodsID {
			return good, true
		}
	}
	return shopEntry{}, false
}

func resolvePurchaseRewards(entry shopEntry, selected []PurchaseSelection) ([]PurchaseDrop, uint32, error) {
	if len(selected) == 0 {
		return nil, 0, ErrInvalidPurchasePayload
	}
	allowed := make(map[uint32]struct{}, len(entry.Goods))
	for _, id := range entry.Goods {
		if id == 0 {
			continue
		}
		allowed[id] = struct{}{}
	}

	rewardUnits := make(map[uint32]uint32, len(selected))
	totalUnits := uint32(0)
	for _, pick := range selected {
		if pick.ID == 0 || pick.Num == 0 {
			return nil, 0, ErrInvalidPurchasePayload
		}
		if len(allowed) > 0 {
			if _, ok := allowed[pick.ID]; !ok {
				return nil, 0, ErrInvalidPurchasePayload
			}
		}
		totalUnits += pick.Num
		rewardUnits[pick.ID] += pick.Num
	}
	if totalUnits == 0 {
		return nil, 0, ErrInvalidPurchasePayload
	}

	rewardMultiplier := entry.Num
	if rewardMultiplier == 0 {
		rewardMultiplier = 1
	}
	rewardIDs := make([]uint32, 0, len(rewardUnits))
	for id := range rewardUnits {
		rewardIDs = append(rewardIDs, id)
	}
	sort.Slice(rewardIDs, func(i, j int) bool { return rewardIDs[i] < rewardIDs[j] })
	rewards := make([]PurchaseDrop, 0, len(rewardIDs))
	for _, id := range rewardIDs {
		rewards = append(rewards, PurchaseDrop{
			Type:   entry.DropType,
			ID:     id,
			Number: rewardUnits[id] * rewardMultiplier,
		})
	}
	return rewards, totalUnits, nil
}

func isWithinTime(now time.Time, ranges [][][3]int) bool {
	if len(ranges) == 0 {
		return true
	}
	current := now.UTC()
	for _, window := range ranges {
		if len(window) != 2 {
			continue
		}
		start := timeFromConfig(current.Location(), window[0])
		end := timeFromConfig(current.Location(), window[1])
		if !start.IsZero() && !end.IsZero() {
			if !current.Before(start) && !current.After(end) {
				return true
			}
		}
	}
	return false
}

func timeFromConfig(loc *time.Location, parts [3]int) time.Time {
	if parts[0] == 0 && parts[1] == 0 && parts[2] == 0 {
		return time.Time{}
	}
	return time.Date(parts[0], time.Month(parts[1]), parts[2], 0, 0, 0, 0, loc)
}

func nextDailyReset(now time.Time) uint32 {
	utc := now.UTC()
	next := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return uint32(next.Unix())
}
