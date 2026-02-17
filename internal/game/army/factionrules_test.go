package army

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
)

// --- Helper functions ---

func makeSeraphonSaurusUnit(id core.UnitID, ownerID int) *core.Unit {
	return &core.Unit{
		ID:             id,
		Name:           "Saurus Warriors",
		OwnerID:        ownerID,
		FactionKeyword: "seraphon",
		Tags:           []string{"Saurus"},
		Keywords:       []core.Keyword{core.KeywordInfantry},
		Stats:          core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		Weapons: []core.Weapon{
			{Name: "Celestite Weapon", Range: 0, Attacks: 2, ToHit: 4, ToWound: 3, Rend: 1, Damage: 1},
		},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 1, MaxWounds: 1, Position: core.Position{X: 10, Y: 10}},
			{ID: 1, IsAlive: true, CurrentWounds: 1, MaxWounds: 1, Position: core.Position{X: 10, Y: 10}},
			{ID: 2, IsAlive: true, CurrentWounds: 1, MaxWounds: 1, Position: core.Position{X: 10, Y: 10}},
		},
	}
}

func makeSeraphonHero(id core.UnitID, ownerID int) *core.Unit {
	return &core.Unit{
		ID:             id,
		Name:           "Saurus Oldblood",
		OwnerID:        ownerID,
		FactionKeyword: "seraphon",
		Tags:           []string{"Saurus"},
		Keywords:       []core.Keyword{core.KeywordHero, core.KeywordCavalry},
		Stats:          core.Stats{Move: 10, Save: 4, Control: 5, Health: 14},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 14, MaxWounds: 14, Position: core.Position{X: 12, Y: 10}},
		},
	}
}

func makeSeraphonWizard(id core.UnitID, ownerID int) *core.Unit {
	return &core.Unit{
		ID:             id,
		Name:           "Slann Starmaster",
		OwnerID:        ownerID,
		FactionKeyword: "seraphon",
		Tags:           []string{"Slann"},
		Keywords:       []core.Keyword{core.KeywordHero, core.KeywordWizard},
		Stats:          core.Stats{Move: 5, Save: 4, Control: 2, Health: 12},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 12, MaxWounds: 12, Position: core.Position{X: 15, Y: 10}},
		},
	}
}

func makeTzeentchDaemonUnit(id core.UnitID, ownerID int) *core.Unit {
	return &core.Unit{
		ID:             id,
		Name:           "Pink Horrors",
		OwnerID:        ownerID,
		FactionKeyword: "tzeentch",
		Tags:           []string{"Daemon", "Horror"},
		Keywords:       []core.Keyword{core.KeywordInfantry},
		Stats:          core.Stats{Move: 5, Save: 5, Control: 1, Health: 1},
		Weapons: []core.Weapon{
			{Name: "Magical Flames", Range: 12, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 1, Damage: 1},
		},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 1, MaxWounds: 1, Position: core.Position{X: 30, Y: 10}},
		},
	}
}

func makeTzeentchHero(id core.UnitID, ownerID int) *core.Unit {
	return &core.Unit{
		ID:             id,
		Name:           "Lord of Change",
		OwnerID:        ownerID,
		FactionKeyword: "tzeentch",
		Tags:           []string{"Daemon"},
		Keywords:       []core.Keyword{core.KeywordHero, core.KeywordMonster, core.KeywordWizard},
		Stats:          core.Stats{Move: 12, Save: 4, Control: 5, Health: 14},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 14, MaxWounds: 14, Position: core.Position{X: 32, Y: 10}},
		},
	}
}

func makeFlamerUnit(id core.UnitID, ownerID int) *core.Unit {
	return &core.Unit{
		ID:             id,
		Name:           "Flamers of Tzeentch",
		OwnerID:        ownerID,
		FactionKeyword: "tzeentch",
		Tags:           []string{"Daemon", "Flamer"},
		Keywords:       []core.Keyword{core.KeywordInfantry},
		Stats:          core.Stats{Move: 9, Save: 5, Control: 1, Health: 3},
		Weapons: []core.Weapon{
			{Name: "Warpflame", Range: 12, Attacks: 3, ToHit: 3, ToWound: 4, Rend: 1, Damage: 1},
		},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 3, MaxWounds: 3, Position: core.Position{X: 34, Y: 10}},
		},
	}
}

func makeSkinkUnit(id core.UnitID, ownerID int) *core.Unit {
	return &core.Unit{
		ID:             id,
		Name:           "Skinks",
		OwnerID:        ownerID,
		FactionKeyword: "seraphon",
		Tags:           []string{"Skink"},
		Keywords:       []core.Keyword{core.KeywordInfantry},
		Stats:          core.Stats{Move: 8, Save: 6, Control: 1, Health: 1},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 1, MaxWounds: 1, Position: core.Position{X: 10, Y: 10}},
		},
	}
}

func makeEnemyUnit(id core.UnitID, ownerID int) *core.Unit {
	return &core.Unit{
		ID:      id,
		Name:    "Enemy Warriors",
		OwnerID: ownerID,
		Stats:   core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		Weapons: []core.Weapon{
			{Name: "Sword", Range: 0, Attacks: 2, ToHit: 3, ToWound: 4, Rend: 0, Damage: 1},
		},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 2, MaxWounds: 2, Position: core.Position{X: 50, Y: 10}},
		},
	}
}

// --- Seraphon Battle Traits ---

func TestScalySkin_WardOverride(t *testing.T) {
	engine := rules.NewEngine()
	registerSeraphonRules(engine, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)
	enemy := makeEnemyUnit(10, 2)

	ctx := &rules.Context{
		Defender: saurus,
		Attacker: enemy,
	}
	engine.Evaluate(rules.BeforeWardSave, ctx)

	if ctx.WardOverride != 6 {
		t.Errorf("Expected Scaly Skin to set WardOverride=6, got %d", ctx.WardOverride)
	}
}

func TestScalySkin_NoEffectOnNonSaurus(t *testing.T) {
	engine := rules.NewEngine()
	registerSeraphonRules(engine, 1)

	skink := makeSkinkUnit(2, 1)
	enemy := makeEnemyUnit(10, 2)

	ctx := &rules.Context{
		Defender: skink,
		Attacker: enemy,
	}
	engine.Evaluate(rules.BeforeWardSave, ctx)

	if ctx.WardOverride != 0 {
		t.Errorf("Expected no WardOverride for Skink, got %d", ctx.WardOverride)
	}
}

func TestScalySkin_NoEffectOnEnemySaurus(t *testing.T) {
	engine := rules.NewEngine()
	registerSeraphonRules(engine, 1)

	// Saurus belonging to player 2 should not get ward from player 1's rules
	enemySaurus := makeSeraphonSaurusUnit(3, 2)
	attacker := makeEnemyUnit(10, 1)

	ctx := &rules.Context{
		Defender: enemySaurus,
		Attacker: attacker,
	}
	engine.Evaluate(rules.BeforeWardSave, ctx)

	if ctx.WardOverride != 0 {
		t.Errorf("Expected no WardOverride for enemy Saurus, got %d", ctx.WardOverride)
	}
}

func TestPredatoryFighters_ChargeBonus(t *testing.T) {
	engine := rules.NewEngine()
	registerSeraphonRules(engine, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)
	saurus.HasCharged = true
	enemy := makeEnemyUnit(10, 2)
	weapon := &saurus.Weapons[0]

	ctx := &rules.Context{
		Attacker:   saurus,
		Defender:   enemy,
		Weapon:     weapon,
		IsShooting: false,
	}
	engine.Evaluate(rules.BeforeAttackCount, ctx)

	// +1 attack per alive model = +3 for 3 models
	if ctx.Modifiers.AttacksMod != 3 {
		t.Errorf("Expected Predatory Fighters to add +3 attacks (3 alive models), got %d", ctx.Modifiers.AttacksMod)
	}
}

func TestPredatoryFighters_NoEffectWithoutCharge(t *testing.T) {
	engine := rules.NewEngine()
	registerSeraphonRules(engine, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)
	saurus.HasCharged = false
	enemy := makeEnemyUnit(10, 2)
	weapon := &saurus.Weapons[0]

	ctx := &rules.Context{
		Attacker:   saurus,
		Defender:   enemy,
		Weapon:     weapon,
		IsShooting: false,
	}
	engine.Evaluate(rules.BeforeAttackCount, ctx)

	if ctx.Modifiers.AttacksMod != 0 {
		t.Errorf("Expected no attack bonus without charge, got %d", ctx.Modifiers.AttacksMod)
	}
}

func TestPredatoryFighters_NoEffectOnShooting(t *testing.T) {
	engine := rules.NewEngine()
	registerSeraphonRules(engine, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)
	saurus.HasCharged = true
	enemy := makeEnemyUnit(10, 2)
	weapon := &core.Weapon{Name: "Javelin", Range: 12, Attacks: 1, ToHit: 4, ToWound: 5}

	ctx := &rules.Context{
		Attacker:   saurus,
		Defender:   enemy,
		Weapon:     weapon,
		IsShooting: true,
	}
	engine.Evaluate(rules.BeforeAttackCount, ctx)

	if ctx.Modifiers.AttacksMod != 0 {
		t.Errorf("Expected no attack bonus for shooting, got %d", ctx.Modifiers.AttacksMod)
	}
}

func TestColdBlooded_IgnoreNegativeMods(t *testing.T) {
	engine := rules.NewEngine()
	registerSeraphonRules(engine, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)
	hero := makeSeraphonHero(2, 1) // Within 12" of saurus
	enemy := makeEnemyUnit(10, 2)

	ctx := &rules.Context{
		Attacker: saurus,
		Defender: enemy,
		AllUnits: []*core.Unit{saurus, hero, enemy},
	}
	ctx.Modifiers.HitMod = -1 // Simulate negative modifier
	engine.Evaluate(rules.BeforeHitRoll, ctx)

	if ctx.Modifiers.HitMod != 0 {
		t.Errorf("Expected Cold-blooded to remove negative hit mod, got %d", ctx.Modifiers.HitMod)
	}
}

func TestColdBlooded_NoEffectWithoutHero(t *testing.T) {
	engine := rules.NewEngine()
	registerSeraphonRules(engine, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)
	// Put saurus far from any hero
	saurus.Models[0].Position = core.Position{X: 100, Y: 100}
	saurus.Models[1].Position = core.Position{X: 100, Y: 100}
	saurus.Models[2].Position = core.Position{X: 100, Y: 100}
	enemy := makeEnemyUnit(10, 2)

	ctx := &rules.Context{
		Attacker: saurus,
		Defender: enemy,
		AllUnits: []*core.Unit{saurus, enemy}, // No hero nearby
	}
	ctx.Modifiers.HitMod = -1
	engine.Evaluate(rules.BeforeHitRoll, ctx)

	// Should stay at -1 since no hero nearby
	if ctx.Modifiers.HitMod != -1 {
		t.Errorf("Expected hit mod to remain -1 without hero nearby, got %d", ctx.Modifiers.HitMod)
	}
}

// --- Tzeentch Battle Traits ---

func TestLocusOfChange_WoundPenalty(t *testing.T) {
	engine := rules.NewEngine()
	registerTzeentchRules(engine, 2)

	daemon := makeTzeentchDaemonUnit(5, 2)
	hero := makeTzeentchHero(6, 2)
	attacker := makeEnemyUnit(10, 1)

	ctx := &rules.Context{
		Attacker: attacker,
		Defender: daemon,
		AllUnits: []*core.Unit{daemon, hero, attacker},
	}
	engine.Evaluate(rules.BeforeWoundRoll, ctx)

	if ctx.Modifiers.WoundMod != -1 {
		t.Errorf("Expected Locus of Change to apply -1 wound mod, got %d", ctx.Modifiers.WoundMod)
	}
}

func TestLocusOfChange_NoEffectOnNonDaemon(t *testing.T) {
	engine := rules.NewEngine()
	registerTzeentchRules(engine, 2)

	// Arcanite unit (non-Daemon) should not get Locus of Change
	arcanite := &core.Unit{
		ID: 5, Name: "Kairic Acolytes", OwnerID: 2,
		FactionKeyword: "tzeentch", Tags: []string{"Arcanite"},
		Keywords: []core.Keyword{core.KeywordInfantry},
		Models:   []core.Model{{ID: 0, IsAlive: true, CurrentWounds: 1, MaxWounds: 1, Position: core.Position{X: 30, Y: 10}}},
	}
	hero := makeTzeentchHero(6, 2)
	attacker := makeEnemyUnit(10, 1)

	ctx := &rules.Context{
		Attacker: attacker,
		Defender: arcanite,
		AllUnits: []*core.Unit{arcanite, hero, attacker},
	}
	engine.Evaluate(rules.BeforeWoundRoll, ctx)

	if ctx.Modifiers.WoundMod != 0 {
		t.Errorf("Expected no wound penalty for non-Daemon, got %d", ctx.Modifiers.WoundMod)
	}
}

func TestLocusOfChange_NoEffectWithoutHero(t *testing.T) {
	engine := rules.NewEngine()
	registerTzeentchRules(engine, 2)

	daemon := makeTzeentchDaemonUnit(5, 2)
	daemon.Models[0].Position = core.Position{X: 100, Y: 100} // Far from hero
	attacker := makeEnemyUnit(10, 1)

	ctx := &rules.Context{
		Attacker: attacker,
		Defender: daemon,
		AllUnits: []*core.Unit{daemon, attacker}, // No hero
	}
	engine.Evaluate(rules.BeforeWoundRoll, ctx)

	if ctx.Modifiers.WoundMod != 0 {
		t.Errorf("Expected no wound penalty without hero, got %d", ctx.Modifiers.WoundMod)
	}
}

// --- Seraphon Formations ---

func TestSunclawTempleHost_ChargeRendBonus(t *testing.T) {
	engine := rules.NewEngine()
	formation := &BattleFormation{Name: "Sunclaw Temple-host"}
	registerSeraphonFormation(engine, formation, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)
	saurus.HasCharged = true
	enemy := makeEnemyUnit(10, 2)

	ctx := &rules.Context{
		Attacker:   saurus,
		Defender:   enemy,
		IsShooting: false,
	}
	engine.Evaluate(rules.BeforeSaveRoll, ctx)

	if ctx.Modifiers.RendMod != 1 {
		t.Errorf("Expected Sunclaw Temple-host to add +1 Rend, got %d", ctx.Modifiers.RendMod)
	}
}

func TestStarborneHost_WardNearWizard(t *testing.T) {
	engine := rules.NewEngine()
	formation := &BattleFormation{Name: "Starborne Host"}
	registerSeraphonFormation(engine, formation, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)
	wizard := makeSeraphonWizard(2, 1)
	enemy := makeEnemyUnit(10, 2)

	ctx := &rules.Context{
		Defender: saurus,
		Attacker: enemy,
		AllUnits: []*core.Unit{saurus, wizard, enemy},
	}
	engine.Evaluate(rules.BeforeWardSave, ctx)

	if ctx.WardOverride != 6 {
		t.Errorf("Expected Starborne Host Ward 6+ near wizard, got %d", ctx.WardOverride)
	}
}

func TestShadowstrikeStarhost_ChargeBonus(t *testing.T) {
	engine := rules.NewEngine()
	formation := &BattleFormation{Name: "Shadowstrike Starhost"}
	registerSeraphonFormation(engine, formation, 1)

	skink := makeSkinkUnit(1, 1)

	ctx := &rules.Context{
		Attacker: skink,
	}
	engine.Evaluate(rules.BeforeCharge, ctx)

	if ctx.Modifiers.ChargeMod != 1 {
		t.Errorf("Expected Shadowstrike +1 charge for Skinks, got %d", ctx.Modifiers.ChargeMod)
	}
}

func TestShadowstrikeStarhost_NoEffectOnSaurus(t *testing.T) {
	engine := rules.NewEngine()
	formation := &BattleFormation{Name: "Shadowstrike Starhost"}
	registerSeraphonFormation(engine, formation, 1)

	saurus := makeSeraphonSaurusUnit(1, 1)

	ctx := &rules.Context{
		Attacker: saurus,
	}
	engine.Evaluate(rules.BeforeCharge, ctx)

	if ctx.Modifiers.ChargeMod != 0 {
		t.Errorf("Expected no charge bonus for Saurus, got %d", ctx.Modifiers.ChargeMod)
	}
}

// --- Tzeentch Formations ---

func TestWyrdflameHost_RendBonus(t *testing.T) {
	engine := rules.NewEngine()
	formation := &BattleFormation{Name: "Wyrdflame Host"}
	registerTzeentchFormation(engine, formation, 2)

	flamer := makeFlamerUnit(5, 2)
	enemy := makeEnemyUnit(10, 1)

	ctx := &rules.Context{
		Attacker:   flamer,
		Defender:   enemy,
		IsShooting: true,
	}
	engine.Evaluate(rules.BeforeSaveRoll, ctx)

	if ctx.Modifiers.RendMod != 1 {
		t.Errorf("Expected Wyrdflame +1 Rend for Flamers, got %d", ctx.Modifiers.RendMod)
	}
}

func TestWyrdflameHost_NoEffectOnMelee(t *testing.T) {
	engine := rules.NewEngine()
	formation := &BattleFormation{Name: "Wyrdflame Host"}
	registerTzeentchFormation(engine, formation, 2)

	flamer := makeFlamerUnit(5, 2)
	enemy := makeEnemyUnit(10, 1)

	ctx := &rules.Context{
		Attacker:   flamer,
		Defender:   enemy,
		IsShooting: false,
	}
	engine.Evaluate(rules.BeforeSaveRoll, ctx)

	if ctx.Modifiers.RendMod != 0 {
		t.Errorf("Expected no Rend bonus on melee, got %d", ctx.Modifiers.RendMod)
	}
}

// --- Destiny Dice ---

func TestDestinyDicePool_Basic(t *testing.T) {
	pool := NewDestinyDicePool(1, []int{1, 2, 3, 4, 5, 6, 3, 4, 5})

	if pool.Count() != 9 {
		t.Errorf("Expected 9 dice, got %d", pool.Count())
	}

	if !pool.HasValue(6) {
		t.Error("Expected pool to have value 6")
	}
	if pool.HasValue(7) {
		t.Error("Expected pool to not have value 7")
	}
}

func TestDestinyDicePool_UseValue(t *testing.T) {
	pool := NewDestinyDicePool(1, []int{1, 3, 5, 6, 2})

	ok := pool.UseValue(5)
	if !ok {
		t.Error("Expected UseValue(5) to succeed")
	}
	if pool.Count() != 4 {
		t.Errorf("Expected 4 dice after using one, got %d", pool.Count())
	}
	if pool.HasValue(5) {
		// Could still have 5 if there were multiple, but we only had one
		t.Error("Expected no more 5s in pool")
	}

	ok = pool.UseValue(9)
	if ok {
		t.Error("Expected UseValue(9) to fail")
	}
}

func TestDestinyDicePool_UseBest(t *testing.T) {
	pool := NewDestinyDicePool(1, []int{2, 4, 6, 1, 3})

	best := pool.UseBest()
	if best != 6 {
		t.Errorf("Expected UseBest to return 6, got %d", best)
	}
	if pool.Count() != 4 {
		t.Errorf("Expected 4 dice after UseBest, got %d", pool.Count())
	}
}

func TestDestinyDicePool_UseWorst(t *testing.T) {
	pool := NewDestinyDicePool(1, []int{2, 4, 6, 1, 3})

	worst := pool.UseWorst()
	if worst != 1 {
		t.Errorf("Expected UseWorst to return 1, got %d", worst)
	}
}

func TestDestinyDicePool_AddDie(t *testing.T) {
	pool := NewDestinyDicePool(1, []int{3})
	pool.AddDie(5)

	if pool.Count() != 2 {
		t.Errorf("Expected 2 dice after adding, got %d", pool.Count())
	}
	if !pool.HasValue(5) {
		t.Error("Expected pool to have added value 5")
	}
}

func TestDestinyDicePool_EmptyPool(t *testing.T) {
	pool := NewDestinyDicePool(1, []int{})

	if pool.UseBest() != 0 {
		t.Error("Expected UseBest on empty pool to return 0")
	}
	if pool.UseWorst() != 0 {
		t.Error("Expected UseWorst on empty pool to return 0")
	}
}

// --- Enhancement System ---

func TestApplyEnhancement_Ward(t *testing.T) {
	u := makeSeraphonHero(1, 1)
	enh := &Enhancement{
		Name:   "Amulet of Destiny",
		Effect: "ward",
		Value:  5,
	}
	ApplyEnhancement(u, enh)
	if u.WardSave != 5 {
		t.Errorf("Expected ward 5 after enhancement, got %d", u.WardSave)
	}
}

func TestApplyEnhancement_WardDoesNotOverwriteBetter(t *testing.T) {
	u := makeSeraphonHero(1, 1)
	u.WardSave = 4 // Already has ward 4+
	enh := &Enhancement{
		Name:   "Amulet of Destiny",
		Effect: "ward",
		Value:  5,
	}
	ApplyEnhancement(u, enh)
	if u.WardSave != 4 {
		t.Errorf("Expected ward to remain 4, got %d", u.WardSave)
	}
}

func TestApplyEnhancement_ExtraRend(t *testing.T) {
	u := makeSeraphonSaurusUnit(1, 1)
	originalRend := u.Weapons[0].Rend
	enh := &Enhancement{
		Name:   "Serpent God Dagger",
		Effect: "extraRend",
		Value:  1,
	}
	ApplyEnhancement(u, enh)
	if u.Weapons[0].Rend != originalRend+1 {
		t.Errorf("Expected Rend %d, got %d", originalRend+1, u.Weapons[0].Rend)
	}
}

func TestApplyEnhancement_ExtraDamage(t *testing.T) {
	u := makeSeraphonSaurusUnit(1, 1)
	originalDmg := u.Weapons[0].Damage
	enh := &Enhancement{
		Name:   "Wicked Shard",
		Effect: "extraDamage",
		Value:  1,
	}
	ApplyEnhancement(u, enh)
	if u.Weapons[0].Damage != originalDmg+1 {
		t.Errorf("Expected Damage %d, got %d", originalDmg+1, u.Weapons[0].Damage)
	}
}

func TestApplyEnhancement_ExtraCast(t *testing.T) {
	u := makeSeraphonWizard(1, 1)
	u.PowerLevel = 2
	enh := &Enhancement{
		Name:   "Arch-sorcerer",
		Effect: "extraCast",
		Value:  1,
	}
	ApplyEnhancement(u, enh)
	if u.PowerLevel != 3 {
		t.Errorf("Expected PowerLevel 3, got %d", u.PowerLevel)
	}
}

// --- Scourge of Ghyran ---

func TestDefaultScourgeOfGhyran(t *testing.T) {
	sog := DefaultScourgeOfGhyran()

	if len(sog.HeroicTraits) != 3 {
		t.Errorf("Expected 3 GHB heroic traits, got %d", len(sog.HeroicTraits))
	}
	if len(sog.Artefacts) != 3 {
		t.Errorf("Expected 3 GHB artefacts, got %d", len(sog.Artefacts))
	}
	if len(sog.SpellLore) != 2 {
		t.Errorf("Expected 2 GHB spells, got %d", len(sog.SpellLore))
	}

	// Verify specific traits exist
	found := false
	for _, ht := range sog.HeroicTraits {
		if ht.Name == "Battle-hardened" && ht.Effect == "ward" && ht.Value == 6 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Battle-hardened heroic trait in Scourge of Ghyran")
	}
}

// --- Unit HasTag ---

func TestUnit_HasTag(t *testing.T) {
	u := makeSeraphonSaurusUnit(1, 1)
	if !u.HasTag("Saurus") {
		t.Error("Expected Saurus unit to have Saurus tag")
	}
	if u.HasTag("Skink") {
		t.Error("Expected Saurus unit to NOT have Skink tag")
	}
}

// --- JSON Loading with New Fields ---

func TestFactionJSON_LoadFormations(t *testing.T) {
	jsonData := `{
		"id": "test",
		"name": "Test Faction",
		"grandAlliance": "Order",
		"battleTraits": [],
		"spellLore": [],
		"prayerLore": [],
		"formations": [
			{
				"name": "Test Formation",
				"description": "A test formation.",
				"effects": [
					{"description": "+1 Rend", "targetTag": "Saurus", "effect": "rendBonus", "value": 1, "condition": "always"}
				]
			}
		],
		"heroicTraits": [
			{"name": "Test Trait", "description": "Test.", "type": "heroicTrait", "effect": "ward", "value": 6}
		],
		"artefacts": [
			{"name": "Test Artefact", "description": "Test.", "type": "artefact", "effect": "extraDamage", "value": 1}
		],
		"warscrolls": [
			{
				"id": "test_unit",
				"name": "Test Unit",
				"faction": "test",
				"points": 100,
				"unitSize": 5,
				"maxSize": 10,
				"baseSizeMM": 32,
				"keywords": ["Infantry"],
				"tags": ["Saurus", "Elite"],
				"unique": false,
				"stats": {"move": 5, "save": 4, "control": 1, "health": 1},
				"weapons": [{"name": "Blade", "range": 0, "attacks": 2, "hit": 4, "wound": 3, "rend": 1, "damage": 1, "abilities": []}],
				"wardSave": 0,
				"powerLevel": 0,
				"spells": [],
				"prayers": [],
				"abilities": []
			}
		]
	}`

	faction, err := ParseFactionJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to parse faction JSON: %v", err)
	}

	if len(faction.Formations) != 1 {
		t.Errorf("Expected 1 formation, got %d", len(faction.Formations))
	}
	if faction.Formations[0].Name != "Test Formation" {
		t.Errorf("Expected formation name 'Test Formation', got '%s'", faction.Formations[0].Name)
	}
	if len(faction.Formations[0].Effects) != 1 {
		t.Errorf("Expected 1 formation effect, got %d", len(faction.Formations[0].Effects))
	}

	if len(faction.HeroicTraits) != 1 {
		t.Errorf("Expected 1 heroic trait, got %d", len(faction.HeroicTraits))
	}
	if len(faction.Artefacts) != 1 {
		t.Errorf("Expected 1 artefact, got %d", len(faction.Artefacts))
	}

	ws := faction.GetWarscroll("test_unit")
	if ws == nil {
		t.Fatal("Expected to find test_unit warscroll")
	}
	if len(ws.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(ws.Tags))
	}
	if ws.Tags[0] != "Saurus" || ws.Tags[1] != "Elite" {
		t.Errorf("Expected tags [Saurus, Elite], got %v", ws.Tags)
	}
}

// --- Warscroll Ability Rules ---

func TestRegisterMinusOneToBeHit(t *testing.T) {
	engine := rules.NewEngine()
	unit := makeSeraphonHero(1, 1)
	ws := &Warscroll{
		Abilities: []WarscrollAbility{
			{Name: "Terror", Effect: "minusOneToBeHit", Value: 1},
		},
	}
	RegisterWarscrollAbilityRules(engine, unit, ws)

	enemy := makeEnemyUnit(10, 2)
	ctx := &rules.Context{
		Attacker: enemy,
		Defender: unit,
	}
	engine.Evaluate(rules.BeforeHitRoll, ctx)

	if ctx.Modifiers.HitMod != -1 {
		t.Errorf("Expected -1 hit mod from minusOneToBeHit, got %d", ctx.Modifiers.HitMod)
	}
}

func TestRegisterMortalOnCharge(t *testing.T) {
	engine := rules.NewEngine()
	unit := makeSeraphonSaurusUnit(1, 1)
	unit.HasCharged = true
	ws := &Warscroll{
		Abilities: []WarscrollAbility{
			{Name: "Stampede", Effect: "mortalOnCharge", Value: 3},
		},
	}
	RegisterWarscrollAbilityRules(engine, unit, ws)

	enemy := makeEnemyUnit(10, 2)
	ctx := &rules.Context{
		Attacker:   unit,
		Defender:   enemy,
		IsShooting: false,
	}
	engine.Evaluate(rules.BeforeAttackCount, ctx)

	if ctx.Modifiers.MortalWounds != 3 {
		t.Errorf("Expected 3 mortal wounds on charge, got %d", ctx.Modifiers.MortalWounds)
	}
}

// --- Helper Functions ---

func TestIsNearFriendlyHero(t *testing.T) {
	unit := makeSeraphonSaurusUnit(1, 1)
	hero := makeSeraphonHero(2, 1)
	allUnits := []*core.Unit{unit, hero}

	// They're at (10,10) and (12,10) = 2" apart, well within 12"
	if !isNearFriendlyHero(unit, allUnits, "seraphon", 12.0) {
		t.Error("Expected unit to be near friendly hero")
	}

	// Move hero far away
	hero.Models[0].Position = core.Position{X: 100, Y: 100}
	if isNearFriendlyHero(unit, allUnits, "seraphon", 12.0) {
		t.Error("Expected unit to NOT be near friendly hero after moving far")
	}
}

func TestIsNearFriendlyWizard(t *testing.T) {
	unit := makeSeraphonSaurusUnit(1, 1)
	wizard := makeSeraphonWizard(3, 1)
	allUnits := []*core.Unit{unit, wizard}

	// (10,10) and (15,10) = 5" apart, within 12"
	if !isNearFriendlyWizard(unit, allUnits, 1, 12.0) {
		t.Error("Expected unit to be near friendly wizard")
	}
}

// --- Roster Enhancement Fields ---

func TestRosterEnhancementFields(t *testing.T) {
	roster := &ArmyRoster{
		FactionID:      "seraphon",
		FormationIndex: 1,
		HeroicTraitIdx: 0,
		ArtefactIdx:    2,
		ArtefactUnitID: "seraphon_saurus_oldblood_on_carnosaur",
	}

	if roster.FormationIndex != 1 {
		t.Errorf("Expected FormationIndex 1, got %d", roster.FormationIndex)
	}
	if roster.HeroicTraitIdx != 0 {
		t.Errorf("Expected HeroicTraitIdx 0, got %d", roster.HeroicTraitIdx)
	}
	if roster.ArtefactIdx != 2 {
		t.Errorf("Expected ArtefactIdx 2, got %d", roster.ArtefactIdx)
	}
}

// --- Integration: Full Seraphon Rules Registration ---

func TestFullSeraphonRulesRegistration(t *testing.T) {
	engine := rules.NewEngine()
	faction := &Faction{
		ID:   "seraphon",
		Name: "Seraphon",
		Formations: []BattleFormation{
			{Name: "Sunclaw Temple-host"},
		},
	}

	RegisterFactionRules(engine, faction, 1)
	RegisterFormationRules(engine, faction, 0, 1)

	// Should have rules registered for various triggers
	if !engine.HasRulesFor(rules.BeforeWardSave) {
		t.Error("Expected BeforeWardSave rules (Scaly Skin)")
	}
	if !engine.HasRulesFor(rules.BeforeAttackCount) {
		t.Error("Expected BeforeAttackCount rules (Predatory Fighters)")
	}
	if !engine.HasRulesFor(rules.BeforeHitRoll) {
		t.Error("Expected BeforeHitRoll rules (Cold-blooded)")
	}
	if !engine.HasRulesFor(rules.BeforeSaveRoll) {
		t.Error("Expected BeforeSaveRoll rules (Sunclaw Temple-host)")
	}
	if engine.RuleCount() < 4 {
		t.Errorf("Expected at least 4 rules, got %d", engine.RuleCount())
	}
}

func TestFullTzeentchRulesRegistration(t *testing.T) {
	engine := rules.NewEngine()
	faction := &Faction{
		ID:   "tzeentch",
		Name: "Disciples of Tzeentch",
		Formations: []BattleFormation{
			{Name: "Wyrdflame Host"},
		},
	}

	RegisterFactionRules(engine, faction, 2)
	RegisterFormationRules(engine, faction, 0, 2)

	if !engine.HasRulesFor(rules.BeforeWoundRoll) {
		t.Error("Expected BeforeWoundRoll rules (Locus of Change)")
	}
	if !engine.HasRulesFor(rules.BeforeSaveRoll) {
		t.Error("Expected BeforeSaveRoll rules (Wyrdflame Host)")
	}
}
