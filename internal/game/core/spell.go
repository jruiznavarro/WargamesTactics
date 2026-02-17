package core

// SpellEffect defines what a spell or prayer does when it succeeds.
type SpellEffect int

const (
	SpellEffectDamage SpellEffect = iota // Deal mortal wounds to enemy unit
	SpellEffectHeal                      // Heal wounds on friendly unit
	SpellEffectBuff                      // Apply buff to friendly unit (e.g. +1 save)
)

// Spell represents a Wizard's spell ability. AoS4 Rule 19.0:
// Roll 2D6 >= CastingValue to cast. Miscast on double 1s (D3 mortal, no more spells).
// All spells are warscroll/faction-specific - there are no universal spells in AoS4.
// Unbind: enemy wizard within 30" rolls 2D6, must exceed the casting roll.
type Spell struct {
	Name           string
	CastingValue   int         // 2D6 >= this to succeed
	Range          int         // Range in inches (target must be wholly within)
	Effect         SpellEffect // What the spell does
	EffectValue    int         // Effect magnitude (e.g. D3 for damage, +1 for buff)
	TargetFriendly bool        // true = targets friendly, false = targets enemy
	Unlimited      bool        // If true, multiple wizards can cast this same spell per turn
}

// Prayer represents a Priest's prayer ability. AoS4 Rule 19.2:
// Priests build ritual points over turns. Roll D6 each hero phase:
//   - Roll of 1: fail, remove D3 ritual points
//   - Roll of 2+: either bank points (= roll) or spend (add ritual points to roll)
//     If spending and total >= ChantingValue, the prayer is answered.
//
// All prayers are warscroll/faction-specific - there are no universal prayers in AoS4.
type Prayer struct {
	Name           string
	ChantingValue  int         // Ritual points + roll must >= this to answer
	Range          int         // Range in inches
	Effect         SpellEffect // What the prayer does
	EffectValue    int         // Effect magnitude
	TargetFriendly bool        // true = targets friendly, false = targets enemy
	Unlimited      bool        // If true, multiple priests can chant this same prayer per turn
}
