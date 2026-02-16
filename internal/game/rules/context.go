package rules

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// Context carries all the information a rule needs to evaluate its condition
// and apply its effect. Fields are populated depending on the trigger point.
// Not all fields are set for every trigger -- only the ones relevant to the
// current game event.
type Context struct {
	// Units involved
	Attacker *core.Unit // The acting unit (shooter, fighter, charger, mover)
	Defender *core.Unit // The target unit (if applicable)

	// Weapon being used (combat pipeline triggers)
	Weapon *core.Weapon

	// Current phase
	PhaseType string

	// Position data (movement/charge triggers)
	Origin      core.Position // Starting position
	Destination core.Position // Intended destination
	Distance    float64       // Distance being moved/charged

	// Modifier accumulator -- rules write their modifiers here.
	// The game engine reads these after all rules have been evaluated.
	Modifiers Modifiers

	// Control flags -- rules can set these to block an action.
	Blocked      bool   // If true, the action is prevented
	BlockMessage string // Reason for blocking
}

// Modifiers holds numeric modifiers that rules can accumulate.
// The game engine applies these when resolving actions.
type Modifiers struct {
	AttacksMod  int // Added to number of attacks (can be negative)
	HitMod      int // Added to hit roll result (positive = easier to hit)
	WoundMod    int // Added to wound roll result (positive = easier to wound)
	SaveMod     int // Added to save roll result (positive = easier to save, negative = harder)
	DamageMod   int // Added to damage per unsaved wound
	MoveMod     int // Added to movement distance in inches
	ChargeMod   int // Added to charge roll
	PileInMod   int // Added to pile-in distance (in tenths of inches for precision)
	MortalWounds int // Mortal wounds to deal (bypasses saves)
}

// Merge combines two modifier sets by adding them together.
func (m *Modifiers) Merge(other Modifiers) {
	m.AttacksMod += other.AttacksMod
	m.HitMod += other.HitMod
	m.WoundMod += other.WoundMod
	m.SaveMod += other.SaveMod
	m.DamageMod += other.DamageMod
	m.MoveMod += other.MoveMod
	m.ChargeMod += other.ChargeMod
	m.PileInMod += other.PileInMod
	m.MortalWounds += other.MortalWounds
}
