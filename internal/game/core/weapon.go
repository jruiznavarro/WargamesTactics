package core

// Weapon represents a weapon profile.
type Weapon struct {
	Name    string
	Range   int // Range in inches (0 = melee)
	Attacks int // Number of attacks
	ToHit   int // Roll needed to hit (e.g. 3+ = 3)
	ToWound int // Roll needed to wound (e.g. 4+ = 4)
	Rend    int // Armour penetration (negative modifier to enemy save)
	Damage  int // Damage per successful attack
}

// IsMelee returns true if the weapon is a melee weapon.
func (w *Weapon) IsMelee() bool {
	return w.Range == 0
}

// IsRanged returns true if the weapon is a ranged weapon.
func (w *Weapon) IsRanged() bool {
	return w.Range > 0
}
