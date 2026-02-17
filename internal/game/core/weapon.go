package core

// WeaponAbility represents a special ability on a weapon (AoS4 Rule 20.0).
type WeaponAbility int

const (
	AbilityNone          WeaponAbility = 0
	AbilityAntiInfantry  WeaponAbility = 1 << iota // Anti-Infantry (+1 Rend)
	AbilityAntiCavalry                              // Anti-Cavalry (+1 Rend)
	AbilityAntiHero                                 // Anti-Hero (+1 Rend)
	AbilityAntiMonster                              // Anti-Monster (+1 Rend)
	AbilityAntiCharge                               // Anti-charge (+1 Rend)
	AbilityCharge                                   // Charge (+1 Damage)
	AbilityCrit2Hits                                // Crit (2 Hits)
	AbilityCritAutoWound                            // Crit (Auto-wound)
	AbilityCritMortal                               // Crit (Mortal)
	AbilityCompanion                                // Companion (not affected by friendly buffs)
	AbilityShootInCombat                            // Shoot in Combat
)

// Weapon represents a weapon profile (AoS4 Rule 4.0/16.0).
type Weapon struct {
	Name      string
	Range     int           // Range in inches (0 = melee)
	Attacks   int           // Number of attacks
	ToHit     int           // Roll needed to hit (e.g. 3+ = 3)
	ToWound   int           // Roll needed to wound (e.g. 4+ = 4)
	Rend      int           // Rend characteristic (positive value, e.g. 1 = -1 to save)
	Damage    int           // Damage per successful attack
	Abilities WeaponAbility // Bitmask of weapon abilities
}

// IsMelee returns true if the weapon is a melee weapon.
func (w *Weapon) IsMelee() bool {
	return w.Range == 0
}

// IsRanged returns true if the weapon is a ranged weapon.
func (w *Weapon) IsRanged() bool {
	return w.Range > 0
}

// HasAbility returns true if the weapon has the given ability.
func (w *Weapon) HasAbility(a WeaponAbility) bool {
	return w.Abilities&a != 0
}
