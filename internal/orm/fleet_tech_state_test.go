package orm

import "testing"

func TestCommanderFleetTechStateRoundTrip(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ConfigEntry{})
	clearTable(t, &Commander{})

	if err := CreateCommanderRoot(9001, 9001, "fleet-tech", 0, 0); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	state := &CommanderFleetTechState{
		CommanderID: 9001,
		Groups: []FleetTechGroupState{
			{GroupID: 2, EffectTechID: 2002, StudyTechID: 0, StudyFinishTime: 0, RewardedTechID: 2001},
			{GroupID: 1, EffectTechID: 1001, StudyTechID: 1002, StudyFinishTime: 12345, RewardedTechID: 1000},
		},
		AttrOverrides: []FleetTechAttrOverride{
			{ShipType: 7, AttrType: 3, SetValue: 12},
			{ShipType: 1, AttrType: 9, SetValue: 2},
		},
	}
	if err := SaveCommanderFleetTechState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	reloaded, err := GetCommanderFleetTechState(9001)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if len(reloaded.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(reloaded.Groups))
	}
	if reloaded.Groups[0].GroupID != 1 || reloaded.Groups[1].GroupID != 2 {
		t.Fatalf("expected sorted groups, got %+v", reloaded.Groups)
	}
	if len(reloaded.AttrOverrides) != 2 {
		t.Fatalf("expected 2 overrides, got %d", len(reloaded.AttrOverrides))
	}
	if reloaded.AttrOverrides[0].ShipType != 1 || reloaded.AttrOverrides[1].ShipType != 7 {
		t.Fatalf("expected sorted overrides, got %+v", reloaded.AttrOverrides)
	}
}
