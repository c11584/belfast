package medalshop

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/shopreset"
)

const (
	monthShopConfigCategory = "ShareCfg/month_shop_template.json"
	shopTemplateCategory    = "ShareCfg/shop_template.json"
)

type monthShopTemplate struct {
	ID                  uint32   `json:"id"`
	HonorMedalShopGoods []uint32 `json:"honormedal_shop_goods"`
}

type shopTemplateEntry struct {
	ID                 uint32 `json:"id"`
	GoodsPurchaseLimit uint32 `json:"goods_purchase_limit"`
}

type Config struct {
	GoodsIDs      []uint32
	PurchaseLimit map[uint32]uint32
}

type RefreshOptions struct {
	NextRefreshTime uint32
}

func LoadConfig() (*Config, error) {
	return LoadConfigAt(time.Now())
}

func LoadConfigAt(now time.Time) (*Config, error) {
	monthEntries, err := orm.ListConfigEntries(monthShopConfigCategory)
	if err != nil {
		return nil, err
	}
	template, err := selectMonthTemplate(monthEntries, now)
	if err != nil {
		return nil, err
	}
	purchaseLimit := map[uint32]uint32{}
	shopEntries, err := orm.ListConfigEntries(shopTemplateCategory)
	if err != nil {
		return nil, err
	}
	for _, entry := range shopEntries {
		var shopItem shopTemplateEntry
		if err := json.Unmarshal(entry.Data, &shopItem); err != nil {
			return nil, err
		}
		if shopItem.ID == 0 {
			continue
		}
		purchaseLimit[shopItem.ID] = shopItem.GoodsPurchaseLimit
	}
	return &Config{
		GoodsIDs:      template,
		PurchaseLimit: purchaseLimit,
	}, nil
}

func EnsureState(commanderID uint32, now time.Time, config *Config) (*orm.MedalShopState, []orm.MedalShopGood, error) {
	state, err := orm.GetMedalShopState(commanderID)
	if err != nil {
		if !db.IsNotFound(err) {
			return nil, nil, err
		}
		state = &orm.MedalShopState{
			CommanderID:     commanderID,
			NextRefreshTime: nextMonthlyReset(now),
		}
		if err := orm.CreateMedalShopState(*state); err != nil {
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

func RefreshIfNeeded(commanderID uint32, now time.Time, config *Config) (*orm.MedalShopState, []orm.MedalShopGood, error) {
	state, goods, err := EnsureState(commanderID, now, config)
	if err != nil {
		return nil, nil, err
	}
	if now.Unix() >= int64(state.NextRefreshTime) || len(goods) == 0 {
		goods, err = RefreshGoods(commanderID, config, RefreshOptions{
			NextRefreshTime: nextMonthlyReset(now),
		})
		if err != nil {
			return nil, nil, err
		}
		state, err = orm.GetMedalShopState(commanderID)
		if err != nil {
			return nil, nil, err
		}
	}
	return state, goods, nil
}

func RefreshGoods(commanderID uint32, config *Config, options RefreshOptions) ([]orm.MedalShopGood, error) {
	goods := buildGoods(commanderID, config)
	if err := orm.RefreshMedalShopGoods(commanderID, goods, options.NextRefreshTime); err != nil {
		return nil, err
	}
	return goods, nil
}

func LoadGoods(commanderID uint32) ([]orm.MedalShopGood, error) {
	return orm.LoadMedalShopGoods(commanderID)
}

func buildGoods(commanderID uint32, config *Config) []orm.MedalShopGood {
	if config == nil || len(config.GoodsIDs) == 0 {
		return nil
	}
	goods := make([]orm.MedalShopGood, 0, len(config.GoodsIDs))
	for i, id := range config.GoodsIDs {
		count := config.PurchaseLimit[id]
		if count == 0 {
			count = 1
		}
		goods = append(goods, orm.MedalShopGood{
			CommanderID: commanderID,
			Index:       uint32(i + 1),
			GoodsID:     id,
			Count:       count,
		})
	}
	return goods
}

func nextMonthlyReset(now time.Time) uint32 {
	window, err := shopreset.MonthlyWindow(now)
	if err == nil {
		return uint32(window.End.Unix())
	}
	utc := now.UTC()
	next := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	return uint32(next.Unix())
}

func NextMonthlyReset(now time.Time) uint32 {
	return nextMonthlyReset(now)
}

func NextDailyReset(now time.Time) uint32 {
	return nextMonthlyReset(now)
}

func selectMonthTemplate(entries []orm.ConfigEntry, now time.Time) ([]uint32, error) {
	templates := make([]monthShopTemplate, 0, len(entries))
	for _, entry := range entries {
		var single monthShopTemplate
		if err := json.Unmarshal(entry.Data, &single); err == nil && len(single.HonorMedalShopGoods) > 0 {
			templates = append(templates, single)
			continue
		}
		var list []monthShopTemplate
		if err := json.Unmarshal(entry.Data, &list); err != nil {
			return nil, err
		}
		templates = append(templates, list...)
	}
	if len(templates) == 0 {
		return nil, nil
	}

	window, err := shopreset.MonthlyWindow(now)
	month := uint32(now.UTC().Month())
	if err == nil {
		month = window.Key % 100
	}
	for _, template := range templates {
		if template.ID == month {
			return template.HonorMedalShopGoods, nil
		}
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].ID < templates[j].ID
	})
	index := int((month - 1) % uint32(len(templates)))
	return templates[index].HonorMedalShopGoods, nil
}
