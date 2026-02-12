package core

import "testing"

func newTestUnit() *Unit {
	return &Unit{
		ID:   1,
		Name: "Test Warriors",
		Stats: Stats{
			Move:    5,
			Save:    4,
			Bravery: 7,
			Wounds:  1,
		},
		Models: []Model{
			{ID: 0, Position: Position{X: 10, Y: 10}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
			{ID: 1, Position: Position{X: 10, Y: 11}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
			{ID: 2, Position: Position{X: 11, Y: 10}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
		Weapons: []Weapon{
			{Name: "Sword", Range: 0, Attacks: 2, ToHit: 3, ToWound: 3, Rend: -1, Damage: 1},
			{Name: "Bow", Range: 18, Attacks: 1, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
		},
		OwnerID: 1,
	}
}

func TestUnit_Position(t *testing.T) {
	u := newTestUnit()
	pos := u.Position()
	if pos.X != 10 || pos.Y != 10 {
		t.Errorf("expected (10, 10), got (%f, %f)", pos.X, pos.Y)
	}
}

func TestUnit_AliveModels(t *testing.T) {
	u := newTestUnit()
	if u.AliveModels() != 3 {
		t.Errorf("expected 3 alive, got %d", u.AliveModels())
	}

	u.Models[1].IsAlive = false
	if u.AliveModels() != 2 {
		t.Errorf("expected 2 alive, got %d", u.AliveModels())
	}
}

func TestUnit_IsDestroyed(t *testing.T) {
	u := newTestUnit()
	if u.IsDestroyed() {
		t.Error("unit should not be destroyed")
	}

	for i := range u.Models {
		u.Models[i].IsAlive = false
	}
	if !u.IsDestroyed() {
		t.Error("unit should be destroyed")
	}
}

func TestUnit_MeleeAndRangedWeapons(t *testing.T) {
	u := newTestUnit()

	melee := u.MeleeWeapons()
	if len(melee) != 1 || melee[0] != 0 {
		t.Errorf("expected melee weapon at index 0, got %v", melee)
	}

	ranged := u.RangedWeapons()
	if len(ranged) != 1 || ranged[0] != 1 {
		t.Errorf("expected ranged weapon at index 1, got %v", ranged)
	}
}

func TestUnit_TotalAttacks(t *testing.T) {
	u := newTestUnit()
	// Sword: 2 attacks * 3 alive models = 6
	if u.TotalAttacks(0) != 6 {
		t.Errorf("expected 6 total attacks, got %d", u.TotalAttacks(0))
	}
	// Invalid index
	if u.TotalAttacks(99) != 0 {
		t.Errorf("expected 0 for invalid weapon index, got %d", u.TotalAttacks(99))
	}
}

func TestUnit_AllocateDamage(t *testing.T) {
	u := newTestUnit() // 3 models with 1 wound each

	u.AllocateDamage(2)
	if u.AliveModels() != 1 {
		t.Errorf("expected 1 alive after 2 damage, got %d", u.AliveModels())
	}

	u.AllocateDamage(1)
	if !u.IsDestroyed() {
		t.Error("unit should be destroyed after 3 total damage")
	}
}

func TestUnit_AllocateDamage_MultiWound(t *testing.T) {
	u := &Unit{
		ID: 2,
		Models: []Model{
			{ID: 0, CurrentWounds: 3, MaxWounds: 3, IsAlive: true},
			{ID: 1, CurrentWounds: 3, MaxWounds: 3, IsAlive: true},
		},
	}

	// 4 damage: kills first model (3 wounds), 1 spills to second
	u.AllocateDamage(4)
	if u.Models[0].IsAlive {
		t.Error("first model should be dead")
	}
	if !u.Models[1].IsAlive {
		t.Error("second model should be alive")
	}
	if u.Models[1].CurrentWounds != 2 {
		t.Errorf("expected 2 wounds on second model, got %d", u.Models[1].CurrentWounds)
	}
}

func TestUnit_ResetPhaseFlags(t *testing.T) {
	u := newTestUnit()
	u.HasMoved = true
	u.HasShot = true
	u.HasFought = true
	u.HasCharged = true

	u.ResetPhaseFlags()

	if u.HasMoved || u.HasShot || u.HasFought || u.HasCharged {
		t.Error("all phase flags should be reset")
	}
}
