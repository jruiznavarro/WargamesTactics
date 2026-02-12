package core

// Model represents a single miniature within a unit.
type Model struct {
	ID            int
	Position      Position
	BaseSize      float64 // Base diameter in inches (for collision)
	CurrentWounds int
	MaxWounds     int
	IsAlive       bool
}

// TakeDamage applies damage to the model, killing it if wounds reach 0.
// Returns the amount of overflow damage (damage beyond what killed the model).
func (m *Model) TakeDamage(damage int) int {
	if !m.IsAlive {
		return damage
	}

	m.CurrentWounds -= damage
	if m.CurrentWounds <= 0 {
		overflow := -m.CurrentWounds
		m.CurrentWounds = 0
		m.IsAlive = false
		return overflow
	}
	return 0
}

// Heal restores wounds to the model up to its maximum.
func (m *Model) Heal(amount int) {
	if !m.IsAlive {
		return
	}
	m.CurrentWounds += amount
	if m.CurrentWounds > m.MaxWounds {
		m.CurrentWounds = m.MaxWounds
	}
}
