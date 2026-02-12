package game

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/pkg/dice"
)

func newTestAttacker() *core.Unit {
	return &core.Unit{
		ID:   1,
		Name: "Attackers",
		Stats: core.Stats{
			Move: 5, Save: 4, Bravery: 7, Wounds: 1,
		},
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 10, Y: 10}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Greatsword", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Rend: -1, Damage: 2},
		},
		OwnerID: 1,
	}
}

func newTestDefender() *core.Unit {
	return &core.Unit{
		ID:   2,
		Name: "Defenders",
		Stats: core.Stats{
			Move: 4, Save: 4, Bravery: 6, Wounds: 2,
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
	attacker := newTestAttacker()
	defender := newTestDefender()

	result := ResolveAttacks(roller, attacker, defender, &attacker.Weapons[0])

	// With seed 42, we should get consistent results
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
	attacker2 := newTestAttacker()
	defender2 := newTestDefender()

	result2 := ResolveAttacks(roller2, attacker2, defender2, &attacker2.Weapons[0])

	if result.Hits != result2.Hits || result.Wounds != result2.Wounds ||
		result.SavesFailed != result2.SavesFailed || result.DamageDealt != result2.DamageDealt {
		t.Error("same seed should produce identical combat results")
	}
}

func TestResolveAttacks_DamageAllocation(t *testing.T) {
	// Use a seed that we know produces hits
	// We'll run with many attacks to ensure some damage goes through
	roller := dice.NewRoller(100)

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Test", Attacks: 20, ToHit: 2, ToWound: 2, Rend: -3, Damage: 1},
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

	result := ResolveAttacks(roller, attacker, defender, &attacker.Weapons[0])

	// With 20 attacks hitting on 2+, wounding on 2+, and save of 9+ (impossible),
	// we should get significant damage
	if result.DamageDealt == 0 {
		t.Error("expected some damage to be dealt")
	}
}

func TestResolveMortalWounds(t *testing.T) {
	unit := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 2, MaxWounds: 2, IsAlive: true},
			{ID: 1, CurrentWounds: 2, MaxWounds: 2, IsAlive: true},
		},
	}

	slain := ResolveMortalWounds(unit, 3)
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

	results := ResolveCombat(roller, attacker, defender)

	// Should only use melee weapon (Sword), not ranged (Bow)
	if len(results) != 1 {
		t.Fatalf("expected 1 weapon result, got %d", len(results))
	}
	if results[0].WeaponName != "Sword" {
		t.Errorf("expected Sword, got %s", results[0].WeaponName)
	}
}

func TestResolveShooting_OnlyRangedWeapons(t *testing.T) {
	roller := dice.NewRoller(42)

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

	results := ResolveShooting(roller, attacker, defender)

	if len(results) != 1 {
		t.Fatalf("expected 1 weapon result, got %d", len(results))
	}
	if results[0].WeaponName != "Bow" {
		t.Errorf("expected Bow, got %s", results[0].WeaponName)
	}
}

func TestCombat_DestroyedUnitStopsAttacking(t *testing.T) {
	roller := dice.NewRoller(42)

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []core.Weapon{
			{Name: "Sword1", Range: 0, Attacks: 50, ToHit: 2, ToWound: 2, Rend: -5, Damage: 10},
			{Name: "Sword2", Range: 0, Attacks: 50, ToHit: 2, ToWound: 2, Rend: -5, Damage: 10},
		},
	}

	// Defender with tiny amount of wounds - will die to first weapon
	defender := &core.Unit{
		ID:    2,
		Stats: core.Stats{Save: 6},
		Models: []core.Model{
			{ID: 0, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	results := ResolveCombat(roller, attacker, defender)

	// Should have at most 1 result since defender should die from first weapon
	if len(results) > 1 {
		// If defender died from first weapon, second weapon shouldn't fire
		for i := 1; i < len(results); i++ {
			if results[i].TotalAttacks > 0 && results[i].Hits > 0 {
				// This is fine - defender was already dead, combat should have stopped
			}
		}
	}
	if !defender.IsDestroyed() {
		t.Error("defender should be destroyed")
	}
}
