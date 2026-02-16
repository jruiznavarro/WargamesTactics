package core

// UnitID is a unique identifier for a unit.
type UnitID int

// StrikeOrder determines when a unit fights in the combat phase.
type StrikeOrder int

const (
	StrikeNormal StrikeOrder = 0 // Default: fights in the normal sub-phase
	StrikeFirst  StrikeOrder = 1 // Fights before normal units
	StrikeLast   StrikeOrder = -1 // Fights after normal units
)

// Unit represents a group of models fighting together.
type Unit struct {
	ID      UnitID
	Name    string
	Stats   Stats
	Models  []Model
	Weapons []Weapon
	OwnerID int // Player ID of the owner

	StrikeOrder StrikeOrder // Determines combat activation priority

	HasMoved   bool
	HasShot    bool
	HasFought  bool
	HasCharged bool
	HasPiledIn bool
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

// ResetPhaseFlags resets all per-turn action flags.
func (u *Unit) ResetPhaseFlags() {
	u.HasMoved = false
	u.HasShot = false
	u.HasFought = false
	u.HasCharged = false
	u.HasPiledIn = false
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
