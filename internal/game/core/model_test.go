package core

import "testing"

func TestModel_TakeDamage(t *testing.T) {
	m := Model{ID: 1, CurrentWounds: 3, MaxWounds: 3, IsAlive: true}

	// Take 1 damage
	overflow := m.TakeDamage(1)
	if overflow != 0 {
		t.Errorf("expected 0 overflow, got %d", overflow)
	}
	if m.CurrentWounds != 2 {
		t.Errorf("expected 2 wounds, got %d", m.CurrentWounds)
	}
	if !m.IsAlive {
		t.Error("model should still be alive")
	}

	// Take 3 damage (1 overflow)
	overflow = m.TakeDamage(3)
	if overflow != 1 {
		t.Errorf("expected 1 overflow, got %d", overflow)
	}
	if m.CurrentWounds != 0 {
		t.Errorf("expected 0 wounds, got %d", m.CurrentWounds)
	}
	if m.IsAlive {
		t.Error("model should be dead")
	}

	// Damage to dead model returns full damage
	overflow = m.TakeDamage(5)
	if overflow != 5 {
		t.Errorf("expected 5 overflow from dead model, got %d", overflow)
	}
}

func TestModel_Heal(t *testing.T) {
	m := Model{ID: 1, CurrentWounds: 1, MaxWounds: 3, IsAlive: true}

	m.Heal(1)
	if m.CurrentWounds != 2 {
		t.Errorf("expected 2 wounds, got %d", m.CurrentWounds)
	}

	// Heal beyond max
	m.Heal(5)
	if m.CurrentWounds != 3 {
		t.Errorf("expected 3 wounds (max), got %d", m.CurrentWounds)
	}

	// Heal dead model does nothing
	dead := Model{ID: 2, CurrentWounds: 0, MaxWounds: 3, IsAlive: false}
	dead.Heal(3)
	if dead.CurrentWounds != 0 {
		t.Errorf("dead model should not heal, got %d wounds", dead.CurrentWounds)
	}
}
