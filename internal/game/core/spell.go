package core

// SpellEffect defines what a spell or prayer does when it succeeds.
type SpellEffect int

const (
	SpellEffectDamage SpellEffect = iota // Deal mortal wounds to enemy unit
	SpellEffectHeal                      // Heal wounds on friendly unit
	SpellEffectBuff                      // Apply +1 save to friendly unit (Mystic Shield)
)

// Spell represents a Wizard's spell. AoS4 Rule 19.0:
// Roll 2D6 >= CastingValue to cast. Natural doubles = empowered (cannot be unbound).
type Spell struct {
	Name           string
	CastingValue   int         // 2D6 >= this to succeed
	Range          int         // Range in inches
	Effect         SpellEffect // What the spell does
	EffectValue    int         // Amount (D3 for damage/heal, +1 for buff)
	TargetFriendly bool        // true = targets friendly, false = targets enemy
}

// Prayer represents a Priest's prayer. AoS4 Rule 19.2:
// Roll D6 >= ChantingValue to chant successfully.
type Prayer struct {
	Name           string
	ChantingValue  int         // D6 >= this to succeed
	Range          int         // Range in inches
	Effect         SpellEffect // What the prayer does
	EffectValue    int         // Amount (D3 for damage/heal)
	TargetFriendly bool        // true = targets friendly, false = targets enemy
}

// DefaultWizardSpells returns the two universal spells every Wizard knows.
// AoS4 Rule 19.0: Arcane Bolt and Mystic Shield.
func DefaultWizardSpells() []Spell {
	return []Spell{
		{
			Name:           "Arcane Bolt",
			CastingValue:   5,
			Range:          18,
			Effect:         SpellEffectDamage,
			EffectValue:    0, // D3 mortal wounds (rolled at resolution)
			TargetFriendly: false,
		},
		{
			Name:           "Mystic Shield",
			CastingValue:   5,
			Range:          12,
			Effect:         SpellEffectBuff,
			EffectValue:    1, // +1 to save rolls
			TargetFriendly: true,
		},
	}
}

// DefaultPriestPrayers returns the two universal prayers every Priest knows.
// AoS4 Rule 19.2: Heal and Smite.
func DefaultPriestPrayers() []Prayer {
	return []Prayer{
		{
			Name:           "Heal",
			ChantingValue:  4,
			Range:          12,
			Effect:         SpellEffectHeal,
			EffectValue:    0, // D3 healing (rolled at resolution)
			TargetFriendly: true,
		},
		{
			Name:           "Smite",
			ChantingValue:  4,
			Range:          12,
			Effect:         SpellEffectDamage,
			EffectValue:    0, // D3 mortal wounds (rolled at resolution)
			TargetFriendly: false,
		},
	}
}
