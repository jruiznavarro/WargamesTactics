package rules

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/pkg/dice"
)

// Context carries all the information a rule needs to evaluate its condition
// and apply its effect. Fields are populated depending on the trigger point.
type Context struct {
	// Units involved
	Attacker *core.Unit // The acting unit (shooter, fighter, charger, mover)
	Defender *core.Unit // The target unit (if applicable)

	// Weapon being used (combat pipeline triggers)
	Weapon *core.Weapon

	// Current phase
	PhaseType string

	// Whether this is a shooting attack (for Cover/Obscuring differentiation)
	IsShooting bool

	// Position data (movement/charge triggers)
	Origin      core.Position // Starting position
	Destination core.Position // Intended destination
	Distance    float64       // Distance being moved/charged

	// Game state for faction rules
	BattleRound int            // Current battle round number
	AllUnits    []*core.Unit   // All alive units in the game (for proximity checks)
	PlayerID    int            // Player whose turn it is

	// Modifier accumulator -- rules write their modifiers here.
	Modifiers Modifiers

	// Re-roll flags (Rule 2.2, Errata Jan 2026): re-rolls happen before modifiers.
	// A die cannot be re-rolled more than once.
	RerollHit   dice.RerollType // Re-roll type for hit rolls
	RerollWound dice.RerollType // Re-roll type for wound rolls
	RerollSave  dice.RerollType // Re-roll type for save rolls

	// Ward override (for dynamic ward effects from faction rules)
	WardOverride int // If > 0, overrides the unit's base ward save (lower = better)

	// Control flags -- rules can set these to block an action.
	Blocked      bool   // If true, the action is prevented
	BlockMessage string // Reason for blocking
}

// Modifiers holds numeric modifiers that rules can accumulate.
type Modifiers struct {
	AttacksMod   int // Added to number of attacks (can be negative)
	HitMod       int // Added to hit roll result (positive = easier to hit)
	WoundMod     int // Added to wound roll result (positive = easier to wound)
	SaveMod      int // Added to save roll result (positive = easier to save, negative = harder)
	RendMod      int // Added to Rend characteristic (positive = more rend)
	DamageMod    int // Added to damage per unsaved wound
	MoveMod      int // Added to movement distance in inches
	ChargeMod    int // Added to charge roll
	PileInMod    int // Added to pile-in distance
	MortalWounds int // Mortal wounds to deal (bypasses saves)
}

// Merge combines two modifier sets by adding them together.
func (m *Modifiers) Merge(other Modifiers) {
	m.AttacksMod += other.AttacksMod
	m.HitMod += other.HitMod
	m.WoundMod += other.WoundMod
	m.SaveMod += other.SaveMod
	m.RendMod += other.RendMod
	m.DamageMod += other.DamageMod
	m.MoveMod += other.MoveMod
	m.ChargeMod += other.ChargeMod
	m.PileInMod += other.PileInMod
	m.MortalWounds += other.MortalWounds
}
