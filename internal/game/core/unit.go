package core

// UnitID is a unique identifier for a unit.
type UnitID int

// StrikeOrder determines when a unit fights in the combat phase.
type StrikeOrder int

const (
	StrikeNormal StrikeOrder = 0  // Default: fights in the normal sub-phase
	StrikeFirst  StrikeOrder = 1  // Fights before normal units
	StrikeLast   StrikeOrder = -1 // Fights after normal units
)

// Keyword represents a unit keyword for ability targeting.
type Keyword string

const (
	KeywordInfantry  Keyword = "Infantry"
	KeywordCavalry   Keyword = "Cavalry"
	KeywordHero      Keyword = "Hero"
	KeywordMonster   Keyword = "Monster"
	KeywordWarMachine Keyword = "War Machine"
	KeywordWizard    Keyword = "Wizard"
	KeywordPriest    Keyword = "Priest"
	KeywordFly       Keyword = "Fly"
)

// Unit represents a group of models fighting together.
type Unit struct {
	ID      UnitID
	Name    string
	Stats   Stats
	Models  []Model
	Weapons []Weapon
	OwnerID int // Player ID of the owner

	Keywords    []Keyword   // Unit keywords (Infantry, Hero, Fly, etc.)
	WardSave    int         // Ward save value (0 = none, 6 = 6+, 5 = 5+)
	StrikeOrder StrikeOrder // Determines combat activation priority

	// Magic (AoS4 Rule 19.0 / 19.2)
	Spells       []Spell  // Known spells (warscroll/faction specific)
	Prayers      []Prayer // Known prayers (warscroll/faction specific)
	PowerLevel   int      // Wizard(X) or Priest(X) - abilities per phase (default 1)
	RitualPoints int      // Priests accumulate ritual points across turns

	HasMoved     bool
	HasRun       bool
	HasRetreated bool
	HasShot      bool
	HasFought    bool
	HasCharged   bool
	HasPiledIn   bool
	CastCount    int  // Spell/banish abilities used this phase
	ChantCount   int  // Prayer abilities used this phase
	UnbindCount  int  // Unbind attempts used this phase
	HasMiscast   bool // True if miscast this phase (no more spells)
}

// Position returns the position of the unit leader (first alive model).
func (u *Unit) Position() Position {
	for i := range u.Models {
		if u.Models[i].IsAlive {
			return u.Models[i].Position
		}
	}
	return Position{}
}

// AliveModels returns the number of models still alive.
func (u *Unit) AliveModels() int {
	count := 0
	for i := range u.Models {
		if u.Models[i].IsAlive {
			count++
		}
	}
	return count
}

// IsDestroyed returns true if no models in the unit are alive.
func (u *Unit) IsDestroyed() bool {
	return u.AliveModels() == 0
}

// TotalCurrentWounds returns the sum of current wounds across all alive models.
func (u *Unit) TotalCurrentWounds() int {
	total := 0
	for i := range u.Models {
		if u.Models[i].IsAlive {
			total += u.Models[i].CurrentWounds
		}
	}
	return total
}

// TotalAttacks returns the total number of attacks for a given weapon across alive models.
func (u *Unit) TotalAttacks(weaponIndex int) int {
	if weaponIndex < 0 || weaponIndex >= len(u.Weapons) {
		return 0
	}
	return u.AliveModels() * u.Weapons[weaponIndex].Attacks
}

// MeleeWeapons returns indices of melee weapons.
func (u *Unit) MeleeWeapons() []int {
	var indices []int
	for i, w := range u.Weapons {
		if w.IsMelee() {
			indices = append(indices, i)
		}
	}
	return indices
}

// RangedWeapons returns indices of ranged weapons.
func (u *Unit) RangedWeapons() []int {
	var indices []int
	for i, w := range u.Weapons {
		if w.IsRanged() {
			indices = append(indices, i)
		}
	}
	return indices
}

// HasKeyword returns true if the unit has the given keyword.
func (u *Unit) HasKeyword(k Keyword) bool {
	for _, kw := range u.Keywords {
		if kw == k {
			return true
		}
	}
	return false
}

// ResetPhaseFlags resets all per-turn action flags.
// Note: RitualPoints persist across turns and are NOT reset here.
func (u *Unit) ResetPhaseFlags() {
	u.HasMoved = false
	u.HasRun = false
	u.HasRetreated = false
	u.HasShot = false
	u.HasFought = false
	u.HasCharged = false
	u.HasPiledIn = false
	u.CastCount = 0
	u.ChantCount = 0
	u.UnbindCount = 0
	u.HasMiscast = false
}

// effectivePowerLevel returns the power level, defaulting to 1.
func (u *Unit) effectivePowerLevel() int {
	if u.PowerLevel <= 0 {
		return 1
	}
	return u.PowerLevel
}

// CanCast returns true if this Wizard has remaining spell/banish uses this phase.
// A Wizard(X) can use X spell/banish abilities per hero phase.
// Miscast prevents further casting.
func (u *Unit) CanCast() bool {
	if !u.HasKeyword(KeywordWizard) || len(u.Spells) == 0 {
		return false
	}
	if u.HasMiscast {
		return false
	}
	return u.CastCount < u.effectivePowerLevel()
}

// CanChant returns true if this Priest has remaining prayer uses this phase.
// A Priest(X) can use X prayer abilities per hero phase.
func (u *Unit) CanChant() bool {
	if !u.HasKeyword(KeywordPriest) || len(u.Prayers) == 0 {
		return false
	}
	return u.ChantCount < u.effectivePowerLevel()
}

// CanUnbind returns true if this Wizard has remaining unbind uses this phase.
// A Wizard(X) can unbind X times per phase.
func (u *Unit) CanUnbind() bool {
	if !u.HasKeyword(KeywordWizard) || u.IsDestroyed() {
		return false
	}
	return u.UnbindCount < u.effectivePowerLevel()
}

// AllocateDamage distributes damage across models in the unit.
// Damage spills over from one model to the next.
func (u *Unit) AllocateDamage(totalDamage int) {
	remaining := totalDamage
	for i := range u.Models {
		if remaining <= 0 {
			break
		}
		if u.Models[i].IsAlive {
			remaining = u.Models[i].TakeDamage(remaining)
		}
	}
}
