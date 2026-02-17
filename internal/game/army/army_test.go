package army

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// --- Warscroll Tests ---

func TestWarscroll_BaseSizeInches(t *testing.T) {
	ws := Warscroll{BaseSizeMM: 32}
	inches := ws.BaseSizeInches()
	if inches < 1.25 || inches > 1.27 {
		t.Errorf("expected ~1.26 inches for 32mm, got %.4f", inches)
	}

	ws2 := Warscroll{BaseSizeMM: 25}
	if ws2.BaseSizeInches() < 0.98 || ws2.BaseSizeInches() > 0.99 {
		t.Errorf("expected ~0.98 inches for 25mm, got %.4f", ws2.BaseSizeInches())
	}
}

func TestWarscroll_HasKeyword(t *testing.T) {
	ws := Warscroll{Keywords: []string{"Hero", "Infantry", "Wizard"}}
	if !ws.HasKeyword("Hero") {
		t.Error("expected HasKeyword('Hero') = true")
	}
	if ws.HasKeyword("Monster") {
		t.Error("expected HasKeyword('Monster') = false")
	}
}

func TestWarscroll_ToCoreKeywords(t *testing.T) {
	ws := Warscroll{Keywords: []string{"Hero", "Infantry", "Wizard", "Fly", "SomeFactionKeyword"}}
	coreKW := ws.ToCoreKeywords()
	// Hero, Infantry, Wizard, Fly = 4 core keywords; SomeFactionKeyword is ignored
	if len(coreKW) != 4 {
		t.Errorf("expected 4 core keywords, got %d", len(coreKW))
	}

	hasHero := false
	for _, kw := range coreKW {
		if kw == core.KeywordHero {
			hasHero = true
		}
	}
	if !hasHero {
		t.Error("expected core keywords to contain Hero")
	}
}

func TestWarscroll_ToCoreStats(t *testing.T) {
	ws := Warscroll{Stats: WarscrollStats{Move: 5, Save: 4, Control: 1, Health: 2}}
	stats := ws.ToCoreStats()
	if stats.Move != 5 || stats.Save != 4 || stats.Control != 1 || stats.Health != 2 {
		t.Errorf("stats mismatch: got %+v", stats)
	}
}

func TestWarscroll_ToCoreWeapons(t *testing.T) {
	ws := Warscroll{
		Weapons: []WarscrollWeapon{
			{Name: "Sword", Range: 0, Attacks: 2, ToHit: 3, ToWound: 4, Rend: 1, Damage: 1, Abilities: []string{"Anti-Infantry"}},
			{Name: "Bow", Range: 24, Attacks: 1, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1, Abilities: []string{"Crit(2 Hits)"}},
		},
	}
	weapons := ws.ToCoreWeapons()
	if len(weapons) != 2 {
		t.Fatalf("expected 2 weapons, got %d", len(weapons))
	}
	if weapons[0].Name != "Sword" {
		t.Errorf("expected 'Sword', got '%s'", weapons[0].Name)
	}
	if !weapons[0].HasAbility(core.AbilityAntiInfantry) {
		t.Error("expected Sword to have Anti-Infantry")
	}
	if weapons[0].HasAbility(core.AbilityCrit2Hits) {
		t.Error("expected Sword NOT to have Crit(2 Hits)")
	}
	if !weapons[1].HasAbility(core.AbilityCrit2Hits) {
		t.Error("expected Bow to have Crit(2 Hits)")
	}
	if !weapons[1].IsRanged() {
		t.Error("expected Bow to be ranged")
	}
}

func TestWarscroll_ToCoreSpells(t *testing.T) {
	ws := Warscroll{
		Spells: []WarscrollSpell{
			{Name: "Fireball", CastingValue: 7, Range: 18, Effect: "damage", EffectValue: 3, TargetAlly: false},
			{Name: "Heal", CastingValue: 5, Range: 12, Effect: "heal", EffectValue: 2, TargetAlly: true},
		},
	}
	spells := ws.ToCoreSpells()
	if len(spells) != 2 {
		t.Fatalf("expected 2 spells, got %d", len(spells))
	}
	if spells[0].Name != "Fireball" || spells[0].CastingValue != 7 {
		t.Errorf("spell 0 mismatch: %+v", spells[0])
	}
	if spells[0].Effect != core.SpellEffectDamage {
		t.Error("expected damage effect")
	}
	if spells[1].Effect != core.SpellEffectHeal {
		t.Error("expected heal effect")
	}
	if !spells[1].TargetFriendly {
		t.Error("expected heal to target friendly")
	}
}

func TestWarscroll_ToCorePrayers(t *testing.T) {
	ws := Warscroll{
		Prayers: []WarscrollPrayer{
			{Name: "Blessing", ChantingValue: 4, Range: 12, Effect: "buff", EffectValue: 1, TargetAlly: true},
		},
	}
	prayers := ws.ToCorePrayers()
	if len(prayers) != 1 {
		t.Fatalf("expected 1 prayer, got %d", len(prayers))
	}
	if prayers[0].Name != "Blessing" || prayers[0].ChantingValue != 4 {
		t.Errorf("prayer mismatch: %+v", prayers[0])
	}
	if prayers[0].Effect != core.SpellEffectBuff {
		t.Error("expected buff effect")
	}
}

// --- Faction Tests ---

func TestFaction_GetWarscroll(t *testing.T) {
	faction := &Faction{
		ID:   "test",
		Name: "Test Faction",
		Warscrolls: []Warscroll{
			{ID: "test_warrior", Name: "Warriors"},
			{ID: "test_hero", Name: "Hero"},
		},
	}

	ws := faction.GetWarscroll("test_warrior")
	if ws == nil || ws.Name != "Warriors" {
		t.Error("expected to find Warriors warscroll")
	}

	ws2 := faction.GetWarscroll("nonexistent")
	if ws2 != nil {
		t.Error("expected nil for nonexistent warscroll")
	}
}

func TestFaction_GetWarscrollByName(t *testing.T) {
	faction := &Faction{
		Warscrolls: []Warscroll{
			{ID: "a", Name: "Alpha"},
			{ID: "b", Name: "Beta"},
		},
	}
	ws := faction.GetWarscrollByName("Beta")
	if ws == nil || ws.ID != "b" {
		t.Error("expected to find Beta warscroll")
	}
}

func TestFaction_Heroes(t *testing.T) {
	faction := &Faction{
		Warscrolls: []Warscroll{
			{ID: "hero1", Keywords: []string{"Hero", "Infantry"}},
			{ID: "unit1", Keywords: []string{"Infantry"}},
			{ID: "hero2", Keywords: []string{"Hero", "Monster"}},
		},
	}
	heroes := faction.Heroes()
	if len(heroes) != 2 {
		t.Errorf("expected 2 heroes, got %d", len(heroes))
	}
	nonHeroes := faction.NonHeroes()
	if len(nonHeroes) != 1 {
		t.Errorf("expected 1 non-hero, got %d", len(nonHeroes))
	}
}

// --- Roster Validation Tests ---

func makeTestFaction() *Faction {
	return &Faction{
		ID:   "test",
		Name: "Test",
		Warscrolls: []Warscroll{
			{ID: "hero_a", Name: "Hero A", Points: 150, UnitSize: 1, Keywords: []string{"Hero", "Infantry"}},
			{ID: "hero_b", Name: "Hero B", Points: 200, UnitSize: 1, Keywords: []string{"Hero", "Monster"}, Unique: true},
			{ID: "infantry", Name: "Infantry", Points: 100, UnitSize: 10, MaxSize: 20, BaseSizeMM: 32, Keywords: []string{"Infantry"},
				Stats: WarscrollStats{Move: 5, Save: 4, Control: 1, Health: 1}},
			{ID: "cavalry", Name: "Cavalry", Points: 180, UnitSize: 5, MaxSize: 10, BaseSizeMM: 75, Keywords: []string{"Cavalry"},
				Stats: WarscrollStats{Move: 10, Save: 3, Control: 2, Health: 3}},
			{ID: "elite", Name: "Elite", Points: 300, UnitSize: 3, MaxSize: 0, Keywords: []string{"Infantry"}},
		},
	}
}

func TestRoster_ValidArmy(t *testing.T) {
	faction := makeTestFaction()
	roster := &ArmyRoster{
		FactionID:   "test",
		PointsLimit: 2000,
		Entries: []RosterEntry{
			{WarscrollID: "hero_a", IsGeneral: true},
			{WarscrollID: "infantry"},
			{WarscrollID: "cavalry"},
		},
	}
	errs := roster.Validate(faction)
	if len(errs) > 0 {
		t.Errorf("expected valid army, got errors: %v", errs)
	}
}

func TestRoster_OverPoints(t *testing.T) {
	faction := makeTestFaction()
	roster := &ArmyRoster{
		FactionID:   "test",
		PointsLimit: 200,
		Entries: []RosterEntry{
			{WarscrollID: "hero_a", IsGeneral: true},
			{WarscrollID: "infantry"},
		},
	}
	errs := roster.Validate(faction)
	hasPointsError := false
	for _, e := range errs {
		if e.Error() == "army costs 250 points, exceeds limit of 200" {
			hasPointsError = true
		}
	}
	if !hasPointsError {
		t.Error("expected points limit error")
	}
}

func TestRoster_NoGeneral(t *testing.T) {
	faction := makeTestFaction()
	roster := &ArmyRoster{
		FactionID:   "test",
		PointsLimit: 2000,
		Entries: []RosterEntry{
			{WarscrollID: "hero_a"}, // Not general
			{WarscrollID: "infantry"},
		},
	}
	errs := roster.Validate(faction)
	hasGeneralError := false
	for _, e := range errs {
		if e.Error() == "army must designate a general" {
			hasGeneralError = true
		}
	}
	if !hasGeneralError {
		t.Error("expected general requirement error")
	}
}

func TestRoster_DuplicateUnique(t *testing.T) {
	faction := makeTestFaction()
	roster := &ArmyRoster{
		FactionID:   "test",
		PointsLimit: 2000,
		Entries: []RosterEntry{
			{WarscrollID: "hero_b", IsGeneral: true},
			{WarscrollID: "hero_b"}, // Duplicate unique
		},
	}
	errs := roster.Validate(faction)
	hasDuplicateError := false
	for _, e := range errs {
		t.Logf("error: %s", e.Error())
		if e != nil && len(e.Error()) > 10 {
			hasDuplicateError = true
		}
	}
	if !hasDuplicateError {
		t.Error("expected duplicate unique error")
	}
}

func TestRoster_ReinforcedDoublesCost(t *testing.T) {
	faction := makeTestFaction()
	roster := &ArmyRoster{
		FactionID:   "test",
		PointsLimit: 2000,
		Entries: []RosterEntry{
			{WarscrollID: "hero_a", IsGeneral: true},
			{WarscrollID: "infantry", Reinforced: true},
		},
	}
	total := roster.TotalPoints(faction)
	// hero_a=150 + infantry reinforced=200 = 350
	if total != 350 {
		t.Errorf("expected 350 total points, got %d", total)
	}
}

func TestRoster_CannotReinforceSingleModel(t *testing.T) {
	faction := makeTestFaction()
	roster := &ArmyRoster{
		FactionID:   "test",
		PointsLimit: 2000,
		Entries: []RosterEntry{
			{WarscrollID: "hero_a", IsGeneral: true},
			{WarscrollID: "elite", Reinforced: true}, // MaxSize=0, cannot reinforce
		},
	}
	errs := roster.Validate(faction)
	hasReinforceError := false
	for _, e := range errs {
		if e.Error() == "entry 1: 'Elite' cannot be reinforced" {
			hasReinforceError = true
		}
	}
	if !hasReinforceError {
		t.Error("expected reinforcement error for Elite")
	}
}

func TestRoster_TooManyHeroes(t *testing.T) {
	faction := makeTestFaction()
	entries := make([]RosterEntry, 0)
	for i := 0; i < 7; i++ {
		e := RosterEntry{WarscrollID: "hero_a"}
		if i == 0 {
			e.IsGeneral = true
		}
		entries = append(entries, e)
	}
	roster := &ArmyRoster{
		FactionID:   "test",
		PointsLimit: 5000,
		Entries:     entries,
	}
	errs := roster.Validate(faction)
	hasHeroError := false
	for _, e := range errs {
		if e.Error() == "too many heroes: 7 (max 6)" {
			hasHeroError = true
		}
	}
	if !hasHeroError {
		t.Error("expected too many heroes error")
	}
}

func TestRoster_BuildUnits(t *testing.T) {
	faction := makeTestFaction()
	roster := &ArmyRoster{
		FactionID: "test",
		Entries: []RosterEntry{
			{WarscrollID: "hero_a", IsGeneral: true},
			{WarscrollID: "infantry"},
			{WarscrollID: "infantry", Reinforced: true},
		},
	}
	positions := []core.Position{
		{X: 10, Y: 10},
		{X: 20, Y: 10},
		{X: 30, Y: 10},
	}
	specs := roster.BuildUnits(faction, 1, positions)
	if len(specs) != 3 {
		t.Fatalf("expected 3 unit specs, got %d", len(specs))
	}
	if specs[0].Warscroll.Name != "Hero A" || !specs[0].IsGeneral {
		t.Error("expected Hero A as general")
	}
	if specs[1].NumModels != 10 {
		t.Errorf("expected 10 models for infantry, got %d", specs[1].NumModels)
	}
	if specs[2].NumModels != 20 {
		t.Errorf("expected 20 models for reinforced infantry, got %d", specs[2].NumModels)
	}
}

func TestUnitSpec_ToUnitParams(t *testing.T) {
	faction := makeTestFaction()
	ws := faction.GetWarscroll("infantry")
	spec := &UnitSpec{
		Warscroll: ws,
		NumModels: 10,
		Position:  core.Position{X: 20, Y: 10},
		OwnerID:   1,
	}
	name, ownerID, stats, weapons, numModels, pos, baseSize := spec.ToUnitParams()
	if name != "Infantry" {
		t.Errorf("expected 'Infantry', got '%s'", name)
	}
	if ownerID != 1 {
		t.Errorf("expected ownerID 1, got %d", ownerID)
	}
	if stats.Move != 5 || stats.Save != 4 {
		t.Errorf("stats mismatch: %+v", stats)
	}
	if len(weapons) != 0 {
		t.Errorf("expected 0 weapons (infantry has none), got %d", len(weapons))
	}
	if numModels != 10 {
		t.Errorf("expected 10 models, got %d", numModels)
	}
	if pos.X != 20 || pos.Y != 10 {
		t.Errorf("position mismatch: %+v", pos)
	}
	if baseSize == 0 {
		t.Error("expected non-zero base size")
	}
}

// --- JSON Loading Tests ---

func TestLoadFaction_Seraphon(t *testing.T) {
	path := filepath.Join("..", "..", "..", "data", "factions", "seraphon.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("seraphon.json not found, skipping")
	}

	registry := NewRegistry()
	faction, err := registry.LoadFaction(path)
	if err != nil {
		t.Fatalf("failed to load seraphon: %v", err)
	}

	if faction.ID != "seraphon" {
		t.Errorf("expected id 'seraphon', got '%s'", faction.ID)
	}
	if faction.Name != "Seraphon" {
		t.Errorf("expected name 'Seraphon', got '%s'", faction.Name)
	}
	if len(faction.Warscrolls) < 15 {
		t.Errorf("expected at least 15 warscrolls, got %d", len(faction.Warscrolls))
	}

	// Verify a known unit
	warriors := faction.GetWarscroll("seraphon_saurus_warriors")
	if warriors == nil {
		t.Fatal("expected to find Saurus Warriors")
	}
	if warriors.Points != 120 {
		t.Errorf("expected 120 points, got %d", warriors.Points)
	}
	if warriors.UnitSize != 10 {
		t.Errorf("expected unit size 10, got %d", warriors.UnitSize)
	}
	if warriors.MaxSize != 20 {
		t.Errorf("expected max size 20, got %d", warriors.MaxSize)
	}
	if warriors.Stats.Move != 5 || warriors.Stats.Save != 4 {
		t.Errorf("stats mismatch: %+v", warriors.Stats)
	}

	// Verify unique hero
	kroak := faction.GetWarscroll("seraphon_lord_kroak")
	if kroak == nil {
		t.Fatal("expected to find Lord Kroak")
	}
	if !kroak.Unique {
		t.Error("expected Lord Kroak to be unique")
	}
	if kroak.PowerLevel != 3 {
		t.Errorf("expected power level 3, got %d", kroak.PowerLevel)
	}
	if kroak.WardSave != 4 {
		t.Errorf("expected ward save 4, got %d", kroak.WardSave)
	}
	if !kroak.HasKeyword("Wizard") {
		t.Error("expected Lord Kroak to have Wizard keyword")
	}
}

func TestLoadFaction_Tzeentch(t *testing.T) {
	path := filepath.Join("..", "..", "..", "data", "factions", "tzeentch.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("tzeentch.json not found, skipping")
	}

	registry := NewRegistry()
	faction, err := registry.LoadFaction(path)
	if err != nil {
		t.Fatalf("failed to load tzeentch: %v", err)
	}

	if faction.ID != "tzeentch" {
		t.Errorf("expected id 'tzeentch', got '%s'", faction.ID)
	}
	if len(faction.Warscrolls) < 15 {
		t.Errorf("expected at least 15 warscrolls, got %d", len(faction.Warscrolls))
	}

	// Verify Pink Horrors
	pinks := faction.GetWarscroll("tzeentch_pink_horrors")
	if pinks == nil {
		t.Fatal("expected to find Pink Horrors")
	}
	if pinks.Points != 130 {
		t.Errorf("expected 130 points, got %d", pinks.Points)
	}
	if !pinks.HasKeyword("Daemon") {
		t.Error("expected Pink Horrors to have Daemon keyword")
	}

	// Verify Kairos
	kairos := faction.GetWarscroll("tzeentch_kairos_fateweaver")
	if kairos == nil {
		t.Fatal("expected to find Kairos Fateweaver")
	}
	if !kairos.Unique {
		t.Error("expected Kairos to be unique")
	}
	if kairos.PowerLevel != 3 {
		t.Errorf("expected power level 3, got %d", kairos.PowerLevel)
	}
}

func TestLoadAllFactions(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "data", "factions")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("data/factions not found, skipping")
	}

	registry := NewRegistry()
	err := registry.LoadAllFactions(dir)
	if err != nil {
		t.Fatalf("failed to load factions: %v", err)
	}

	factions := registry.AllFactions()
	if len(factions) < 2 {
		t.Errorf("expected at least 2 factions, got %d", len(factions))
	}

	sera := registry.GetFaction("seraphon")
	if sera == nil {
		t.Error("expected to find seraphon faction")
	}
	tzee := registry.GetFaction("tzeentch")
	if tzee == nil {
		t.Error("expected to find tzeentch faction")
	}
}

func TestParseFactionJSON(t *testing.T) {
	data := []byte(`{
		"id": "mini",
		"name": "Mini Faction",
		"grandAlliance": "Order",
		"warscrolls": [
			{
				"id": "mini_unit",
				"name": "Mini Unit",
				"points": 100,
				"unitSize": 5,
				"maxSize": 10,
				"baseSizeMM": 32,
				"keywords": ["Infantry"],
				"stats": {"move": 5, "save": 4, "control": 1, "health": 1},
				"weapons": [{"name": "Sword", "range": 0, "attacks": 2, "hit": 3, "wound": 4, "rend": 1, "damage": 1}]
			}
		]
	}`)

	faction, err := ParseFactionJSON(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if faction.ID != "mini" {
		t.Errorf("expected id 'mini', got '%s'", faction.ID)
	}
	if len(faction.Warscrolls) != 1 {
		t.Fatalf("expected 1 warscroll, got %d", len(faction.Warscrolls))
	}
	if faction.Warscrolls[0].Faction != "mini" {
		t.Errorf("expected faction 'mini' set on warscroll, got '%s'", faction.Warscrolls[0].Faction)
	}
}

// --- Weapon Ability Parsing Tests ---

func TestParseWeaponAbilities(t *testing.T) {
	abilities := parseWeaponAbilities([]string{"Anti-Infantry", "Charge", "Crit(Mortal)"})
	if abilities&core.AbilityAntiInfantry == 0 {
		t.Error("expected Anti-Infantry")
	}
	if abilities&core.AbilityCharge == 0 {
		t.Error("expected Charge")
	}
	if abilities&core.AbilityCritMortal == 0 {
		t.Error("expected Crit(Mortal)")
	}
	if abilities&core.AbilityAntiCavalry != 0 {
		t.Error("expected no Anti-Cavalry")
	}
}

func TestParseWeaponAbilities_Empty(t *testing.T) {
	abilities := parseWeaponAbilities(nil)
	if abilities != core.AbilityNone {
		t.Errorf("expected AbilityNone, got %d", abilities)
	}
}

// --- JSON Roundtrip ---

func TestWarscroll_JSONRoundtrip(t *testing.T) {
	original := Warscroll{
		ID:       "test_unit",
		Name:     "Test Unit",
		Faction:  "test",
		Points:   150,
		UnitSize: 5,
		MaxSize:  10,
		Keywords: []string{"Infantry"},
		Stats:    WarscrollStats{Move: 5, Save: 4, Control: 1, Health: 2},
		Weapons: []WarscrollWeapon{
			{Name: "Sword", Range: 0, Attacks: 2, ToHit: 3, ToWound: 4, Rend: 1, Damage: 1, Abilities: []string{"Anti-Infantry"}},
		},
		WardSave:   6,
		PowerLevel: 0,
		Abilities: []WarscrollAbility{
			{Name: "Test Ability", Description: "Does something", Phase: "passive", Effect: "ward", Value: 6},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Warscroll
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ID != original.ID || decoded.Name != original.Name {
		t.Errorf("roundtrip failed: ID/Name mismatch")
	}
	if decoded.Points != original.Points || decoded.UnitSize != original.UnitSize {
		t.Errorf("roundtrip failed: Points/UnitSize mismatch")
	}
	if len(decoded.Weapons) != 1 || decoded.Weapons[0].Name != "Sword" {
		t.Errorf("roundtrip failed: Weapons mismatch")
	}
	if len(decoded.Abilities) != 1 || decoded.Abilities[0].Effect != "ward" {
		t.Errorf("roundtrip failed: Abilities mismatch")
	}
}
