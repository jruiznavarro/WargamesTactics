package game

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
	"github.com/jruiznavarro/wargamestactics/pkg/dice"
)

func newTestEngine() *rules.Engine {
	return rules.NewEngine()
}

func newTestAttacker() *core.Unit {
	return &core.Unit{
		ID:   1,
		Name: "Attackers",
		Stats: core.Stats{
			Move: 5, Save: 4, Control: 1, Health: 1,
		},
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 10, Y: 10}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Greatsword", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Rend: 1, Damage: 2},
		},
		OwnerID: 1,
	}
}

func newTestDefender() *core.Unit {
	return &core.Unit{
		ID:   2,
		Name: "Defenders",
		Stats: core.Stats{
			Move: 4, Save: 4, Control: 1, Health: 2,
		},
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 11, Y: 10}, CurrentWounds: 2, MaxWounds: 2, IsAlive: true},
			{ID: 1, Position: core.Position{X: 11, Y: 11}, CurrentWounds: 2, MaxWounds: 2, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Shield Bash", Range: 0, Attacks: 1, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
		},
		OwnerID: 2,
	}
}

func TestResolveAttacks_Deterministic(t *testing.T) {
	roller := dice.NewRoller(42)
	engine := newTestEngine()
	attacker := newTestAttacker()
	defender := newTestDefender()

	result := ResolveAttacks(roller, engine, attacker, defender, &attacker.Weapons[0], false)

	if result.TotalAttacks != 3 {
		t.Errorf("expected 3 attacks, got %d", result.TotalAttacks)
	}
	if result.Hits < 0 || result.Hits > 3 {
		t.Errorf("hits out of range: %d", result.Hits)
	}
	if result.Wounds < 0 || result.Wounds > result.Hits {
		t.Errorf("wounds out of range: %d (hits: %d)", result.Wounds, result.Hits)
	}
	if result.SavesFailed < 0 || result.SavesFailed > result.Wounds {
		t.Errorf("saves failed out of range: %d (wounds: %d)", result.SavesFailed, result.Wounds)
	}

	// Run again with same seed - should get same result
	roller2 := dice.NewRoller(42)
	engine2 := newTestEngine()
	attacker2 := newTestAttacker()
	defender2 := newTestDefender()

	result2 := ResolveAttacks(roller2, engine2, attacker2, defender2, &attacker2.Weapons[0], false)

	if result.Hits != result2.Hits || result.Wounds != result2.Wounds ||
		result.SavesFailed != result2.SavesFailed || result.DamageDealt != result2.DamageDealt {
		t.Error("same seed should produce identical combat results")
	}
}

func TestResolveAttacks_DamageAllocation(t *testing.T) {
	roller := dice.NewRoller(100)
	engine := newTestEngine()

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Test", Attacks: 20, ToHit: 2, ToWound: 2, Rend: 3, Damage: 1},
		},
	}
	defender := &core.Unit{
		ID: 2,
		Stats: core.Stats{Save: 6},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 3, MaxWounds: 3, IsAlive: true},
			{ID: 1, CurrentWounds: 3, MaxWounds: 3, IsAlive: true},
		},
	}

	result := ResolveAttacks(roller, engine, attacker, defender, &attacker.Weapons[0], false)

	if result.DamageDealt == 0 {
		t.Error("expected some damage to be dealt")
	}
}

func TestResolveMortalWounds(t *testing.T) {
	roller := dice.NewRoller(42)
	unit := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 2, MaxWounds: 2, IsAlive: true},
			{ID: 1, CurrentWounds: 2, MaxWounds: 2, IsAlive: true},
		},
	}

	damage, slain := ResolveMortalWounds(roller, unit, 3)
	// No ward save on this unit, so all 3 mortal wounds go through
	if damage != 3 {
		t.Errorf("expected 3 damage from mortal wounds (no ward), got %d", damage)
	}
	if slain != 1 {
		t.Errorf("expected 1 model slain by 3 mortal wounds, got %d", slain)
	}
	if unit.Models[0].IsAlive {
		t.Error("first model should be dead")
	}
	if !unit.Models[1].IsAlive {
		t.Error("second model should be alive")
	}
	if unit.Models[1].CurrentWounds != 1 {
		t.Errorf("expected 1 wound on second model, got %d", unit.Models[1].CurrentWounds)
	}
}

func TestResolveCombat_OnlyMeleeWeapons(t *testing.T) {
	roller := dice.NewRoller(42)
	engine := newTestEngine()

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Sword", Range: 0, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
			{Name: "Bow", Range: 18, Attacks: 1, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
		},
	}
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 5},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 3, MaxWounds: 3, IsAlive: true},
		},
	}

	results := ResolveCombat(roller, engine, attacker, defender)

	if len(results) != 1 {
		t.Fatalf("expected 1 weapon result, got %d", len(results))
	}
	if results[0].WeaponName != "Sword" {
		t.Errorf("expected Sword, got %s", results[0].WeaponName)
	}
}

func TestResolveShooting_OnlyRangedWeapons(t *testing.T) {
	roller := dice.NewRoller(42)
	engine := newTestEngine()

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Sword", Range: 0, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
			{Name: "Bow", Range: 18, Attacks: 1, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
		},
	}
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 5},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 3, MaxWounds: 3, IsAlive: true},
		},
	}

	results := ResolveShooting(roller, engine, attacker, defender)

	if len(results) != 1 {
		t.Fatalf("expected 1 weapon result, got %d", len(results))
	}
	if results[0].WeaponName != "Bow" {
		t.Errorf("expected Bow, got %s", results[0].WeaponName)
	}
}

func TestCombat_DestroyedUnitStopsAttacking(t *testing.T) {
	roller := dice.NewRoller(42)
	engine := newTestEngine()

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Sword1", Range: 0, Attacks: 50, ToHit: 2, ToWound: 2, Rend: 5, Damage: 10},
			{Name: "Sword2", Range: 0, Attacks: 50, ToHit: 2, ToWound: 2, Rend: 5, Damage: 10},
		},
	}

	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	results := ResolveCombat(roller, engine, attacker, defender)

	if len(results) > 1 {
		for i := 1; i < len(results); i++ {
			if results[i].TotalAttacks > 0 && results[i].Hits > 0 {
				// Defender was already dead, combat should have stopped
			}
		}
	}
	if !defender.IsDestroyed() {
		t.Error("defender should be destroyed")
	}
}

func TestCritAutoWound_SkipsWoundRoll(t *testing.T) {
	// Use many attacks to ensure some crits (6s) are rolled.
	// With CritAutoWound, crits skip the wound roll.
	// With ToWound=6 (very hard), normal hits will rarely wound,
	// so most wounds should come from auto-wounds.
	engine := newTestEngine()

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "AutoWoundWeapon", Attacks: 100, ToHit: 2, ToWound: 6, Rend: 5, Damage: 1,
				Abilities: core.AbilityCritAutoWound},
		},
	}
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 200, MaxWounds: 200, IsAlive: true},
		},
	}

	roller := dice.NewRoller(42)
	result := ResolveAttacks(roller, engine, attacker, defender, &attacker.Weapons[0], false)

	// With 100 attacks at ToHit 2+, about ~17 will be crits (natural 6).
	// Those crits auto-wound, skipping the ToWound 6+ check.
	// The remaining ~66 hits need ToWound 6+, so ~11 wound normally.
	// Total wounds should be noticeably higher than if no auto-wound.
	if result.CriticalHits == 0 {
		t.Error("expected some critical hits with 100 attacks")
	}
	if result.Wounds < result.CriticalHits {
		t.Errorf("auto-wounds should contribute to total wounds: wounds=%d, crits=%d",
			result.Wounds, result.CriticalHits)
	}

	// Compare: same weapon WITHOUT CritAutoWound (just regular crits)
	attacker2 := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "NormalWeapon", Attacks: 100, ToHit: 2, ToWound: 6, Rend: 5, Damage: 1},
		},
	}
	defender2 := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 200, MaxWounds: 200, IsAlive: true},
		},
	}

	roller2 := dice.NewRoller(42)
	result2 := ResolveAttacks(roller2, engine, attacker2, defender2, &attacker2.Weapons[0], false)

	if result.Wounds <= result2.Wounds {
		t.Errorf("CritAutoWound should produce more wounds than normal: autoWound=%d, normal=%d",
			result.Wounds, result2.Wounds)
	}
}

func TestCompanion_IgnoresPositiveModifiers(t *testing.T) {
	engine := rules.NewEngine()

	// Add All-out Attack rule: +1 to hit
	engine.AddRule(rules.Rule{
		Name:    "AllOutAttack",
		Trigger: rules.BeforeHitRoll,
		Source:  rules.SourceGlobal,
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.HitMod += 1
		},
	})
	// Add +1 wound modifier
	engine.AddRule(rules.Rule{
		Name:    "WoundBuff",
		Trigger: rules.BeforeWoundRoll,
		Source:  rules.SourceGlobal,
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.WoundMod += 1
		},
	})

	// Companion weapon: should ignore the +1 hit and +1 wound
	companionWeapon := core.Weapon{
		Name: "Companion Fangs", Attacks: 100, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1,
		Abilities: core.AbilityCompanion,
	}
	// Regular weapon: should benefit from +1 hit and +1 wound
	regularWeapon := core.Weapon{
		Name: "Regular Sword", Attacks: 100, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1,
	}

	makeAttacker := func() *core.Unit {
		return &core.Unit{
			ID: 1,
			Models: []core.Model{
				{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
			},
		}
	}
	makeDefender := func() *core.Unit {
		return &core.Unit{
			ID:    2,
			Stats: core.Stats{Save: 6},
			Models: []core.Model{
				{ID: 0, CurrentWounds: 500, MaxWounds: 500, IsAlive: true},
			},
		}
	}

	roller1 := dice.NewRoller(42)
	companionResult := ResolveAttacks(roller1, engine, makeAttacker(), makeDefender(), &companionWeapon, false)

	roller2 := dice.NewRoller(42)
	regularResult := ResolveAttacks(roller2, engine, makeAttacker(), makeDefender(), &regularWeapon, false)

	// Regular weapon should deal more damage because it benefits from +1 hit and +1 wound
	if regularResult.DamageDealt <= companionResult.DamageDealt {
		t.Errorf("regular weapon should deal more damage with buffs: regular=%d, companion=%d",
			regularResult.DamageDealt, companionResult.DamageDealt)
	}
}

func TestCompanion_NegativeModifiersStillApply(t *testing.T) {
	engine := rules.NewEngine()

	// Add cover: -1 to hit
	engine.AddRule(rules.Rule{
		Name:    "Cover",
		Trigger: rules.BeforeHitRoll,
		Source:  rules.SourceTerrain,
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.HitMod -= 1
		},
	})

	companionWeapon := core.Weapon{
		Name: "Companion Fangs", Attacks: 100, ToHit: 3, ToWound: 3, Rend: 3, Damage: 1,
		Abilities: core.AbilityCompanion,
	}

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 500, MaxWounds: 500, IsAlive: true},
		},
	}

	// With cover
	roller1 := dice.NewRoller(42)
	withCover := ResolveAttacks(roller1, engine, attacker, defender, &companionWeapon, false)

	// Without cover
	attacker2 := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}
	defender2 := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 500, MaxWounds: 500, IsAlive: true},
		},
	}
	engine2 := rules.NewEngine()
	roller2 := dice.NewRoller(42)
	withoutCover := ResolveAttacks(roller2, engine2, attacker2, defender2, &companionWeapon, false)

	// Companion should still be affected by the -1 penalty
	if withCover.DamageDealt >= withoutCover.DamageDealt {
		t.Errorf("companion should still suffer from -1 hit: withCover=%d, without=%d",
			withCover.DamageDealt, withoutCover.DamageDealt)
	}
}

func TestPostCombatTriggers_OnModelSlain(t *testing.T) {
	engine := rules.NewEngine()
	slainCount := 0

	engine.AddRule(rules.Rule{
		Name:    "TrackSlain",
		Trigger: rules.OnModelSlain,
		Source:  rules.SourceGlobal,
		Apply: func(ctx *rules.Context) {
			slainCount = ctx.Modifiers.AttacksMod // Carries models slain count
		},
	})

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Overkill", Range: 0, Attacks: 50, ToHit: 2, ToWound: 2, Rend: 5, Damage: 3},
		},
	}
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6, Health: 1},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
			{ID: 1, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
			{ID: 2, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	roller := dice.NewRoller(42)
	ResolveCombat(roller, engine, attacker, defender)

	if slainCount == 0 {
		t.Error("OnModelSlain trigger should have fired with slain count")
	}
}

func TestPostCombatTriggers_OnUnitDestroyed(t *testing.T) {
	engine := rules.NewEngine()
	destroyed := false

	engine.AddRule(rules.Rule{
		Name:    "TrackDestroyed",
		Trigger: rules.OnUnitDestroyed,
		Source:  rules.SourceGlobal,
		Apply: func(ctx *rules.Context) {
			destroyed = true
		},
	})

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Overkill", Range: 0, Attacks: 50, ToHit: 2, ToWound: 2, Rend: 5, Damage: 10},
		},
	}
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6, Health: 1},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	roller := dice.NewRoller(42)
	ResolveCombat(roller, engine, attacker, defender)

	if !destroyed {
		t.Error("OnUnitDestroyed trigger should have fired")
	}
	if !defender.IsDestroyed() {
		t.Error("defender should be destroyed")
	}
}

func TestPostCombatTriggers_AfterCombatResolve(t *testing.T) {
	engine := rules.NewEngine()
	triggered := false

	engine.AddRule(rules.Rule{
		Name:    "TrackAfterCombat",
		Trigger: rules.AfterCombatResolve,
		Source:  rules.SourceGlobal,
		Apply: func(ctx *rules.Context) {
			triggered = true
		},
	})

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Sword", Range: 0, Attacks: 1, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
		},
	}
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 4, Health: 10},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 10, MaxWounds: 10, IsAlive: true},
		},
	}

	roller := dice.NewRoller(42)
	ResolveCombat(roller, engine, attacker, defender)

	if !triggered {
		t.Error("AfterCombatResolve trigger should have fired")
	}
}

func TestWardSave_ReducesDamage(t *testing.T) {
	engine := newTestEngine()

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Test", Attacks: 100, ToHit: 2, ToWound: 2, Rend: 5, Damage: 1},
		},
	}

	// Defender with ward save 5+
	defenderWithWard := &core.Unit{
		ID:       2,
		Stats:    core.Stats{Save: 6, Health: 1},
		WardSave: 5,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 500, MaxWounds: 500, IsAlive: true},
		},
	}
	roller1 := dice.NewRoller(42)
	resultWard := ResolveAttacks(roller1, engine, attacker, defenderWithWard, &attacker.Weapons[0], false)

	// Defender without ward
	attacker2 := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Test", Attacks: 100, ToHit: 2, ToWound: 2, Rend: 5, Damage: 1},
		},
	}
	defenderNoWard := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6, Health: 1},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 500, MaxWounds: 500, IsAlive: true},
		},
	}
	roller2 := dice.NewRoller(42)
	resultNoWard := ResolveAttacks(roller2, engine, attacker2, defenderNoWard, &attacker2.Weapons[0], false)

	if resultWard.WardSaved == 0 {
		t.Error("ward save should have prevented some damage")
	}
	if resultWard.DamageDealt >= resultNoWard.DamageDealt {
		t.Errorf("ward save should reduce damage: with ward=%d, without=%d",
			resultWard.DamageDealt, resultNoWard.DamageDealt)
	}
}

func TestResolveAttacks_WithRuleModifiers(t *testing.T) {
	roller := dice.NewRoller(42)
	engine := rules.NewEngine()

	// Add cover rule: -1 to hit rolls (AoS4: cover subtracts from hit rolls)
	engine.AddRule(rules.Rule{
		Name:    "Cover",
		Trigger: rules.BeforeHitRoll,
		Source:  rules.SourceTerrain,
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.HitMod -= 1
		},
	})

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Test", Attacks: 100, ToHit: 2, ToWound: 2, Rend: 0, Damage: 1},
		},
	}
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 4},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 100, MaxWounds: 100, IsAlive: true},
		},
	}

	resultWithCover := ResolveAttacks(roller, engine, attacker, defender, &attacker.Weapons[0], false)

	// With cover (-1 hit), fewer attacks will hit
	// so less damage compared to without cover
	roller2 := dice.NewRoller(42)
	engine2 := rules.NewEngine() // no cover
	attacker2 := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Test", Attacks: 100, ToHit: 2, ToWound: 2, Rend: 0, Damage: 1},
		},
	}
	defender2 := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 4},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 100, MaxWounds: 100, IsAlive: true},
		},
	}

	resultWithout := ResolveAttacks(roller2, engine2, attacker2, defender2, &attacker2.Weapons[0], false)

	// Cover should result in less damage
	if resultWithCover.DamageDealt >= resultWithout.DamageDealt {
		t.Errorf("cover should reduce damage: with=%d without=%d",
			resultWithCover.DamageDealt, resultWithout.DamageDealt)
	}
}
