package game

import (
	"fmt"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/pkg/dice"
)

// CombatResult records the outcome of a combat resolution.
type CombatResult struct {
	AttackerID   core.UnitID
	DefenderID   core.UnitID
	WeaponName   string
	TotalAttacks int
	Hits         int
	Wounds       int
	SavesFailed  int
	DamageDealt  int
	ModelsSlain  int
}

func (r CombatResult) String() string {
	return fmt.Sprintf(
		"%s: %d attacks -> %d hits -> %d wounds -> %d unsaved -> %d damage (%d models slain)",
		r.WeaponName, r.TotalAttacks, r.Hits, r.Wounds, r.SavesFailed, r.DamageDealt, r.ModelsSlain,
	)
}

// ResolveAttacks resolves the full attack sequence for one weapon profile:
// hit rolls -> wound rolls -> save rolls -> damage allocation.
func ResolveAttacks(roller *dice.Roller, attacker *core.Unit, defender *core.Unit, weapon *core.Weapon) CombatResult {
	aliveModelsBefore := defender.AliveModels()
	totalAttacks := attacker.AliveModels() * weapon.Attacks

	result := CombatResult{
		AttackerID:   attacker.ID,
		DefenderID:   defender.ID,
		WeaponName:   weapon.Name,
		TotalAttacks: totalAttacks,
	}

	// Step 1: Hit rolls
	hits := rollHits(roller, totalAttacks, weapon.ToHit)
	result.Hits = hits

	// Step 2: Wound rolls
	wounds := rollWounds(roller, hits, weapon.ToWound)
	result.Wounds = wounds

	// Step 3: Save rolls
	saveThreshold := defender.Stats.Save - weapon.Rend // Rend is negative, subtracting makes save harder
	savesFailed := rollSaves(roller, wounds, saveThreshold)
	result.SavesFailed = savesFailed

	// Step 4: Damage allocation
	totalDamage := savesFailed * weapon.Damage
	defender.AllocateDamage(totalDamage)
	result.DamageDealt = totalDamage

	result.ModelsSlain = aliveModelsBefore - defender.AliveModels()

	return result
}

// ResolveMortalWounds applies mortal wounds directly, bypassing saves.
func ResolveMortalWounds(defender *core.Unit, mortalWounds int) int {
	aliveModelsBefore := defender.AliveModels()
	defender.AllocateDamage(mortalWounds)
	return aliveModelsBefore - defender.AliveModels()
}

// ResolveCombat resolves all melee weapon attacks from attacker against defender.
func ResolveCombat(roller *dice.Roller, attacker *core.Unit, defender *core.Unit) []CombatResult {
	var results []CombatResult
	for _, idx := range attacker.MeleeWeapons() {
		if attacker.IsDestroyed() || defender.IsDestroyed() {
			break
		}
		result := ResolveAttacks(roller, attacker, defender, &attacker.Weapons[idx])
		results = append(results, result)
	}
	return results
}

// ResolveShooting resolves all ranged weapon attacks from attacker against defender.
func ResolveShooting(roller *dice.Roller, attacker *core.Unit, defender *core.Unit) []CombatResult {
	var results []CombatResult
	for _, idx := range attacker.RangedWeapons() {
		if attacker.IsDestroyed() || defender.IsDestroyed() {
			break
		}
		result := ResolveAttacks(roller, attacker, defender, &attacker.Weapons[idx])
		results = append(results, result)
	}
	return results
}

// rollHits rolls numAttacks D6s and counts successes (>= toHit).
// Natural 1 always fails.
func rollHits(roller *dice.Roller, numAttacks, toHit int) int {
	hits := 0
	for i := 0; i < numAttacks; i++ {
		_, success := roller.RollWithThreshold(toHit)
		if success {
			hits++
		}
	}
	return hits
}

// rollWounds rolls numHits D6s and counts successes (>= toWound).
// Natural 1 always fails.
func rollWounds(roller *dice.Roller, numHits, toWound int) int {
	wounds := 0
	for i := 0; i < numHits; i++ {
		_, success := roller.RollWithThreshold(toWound)
		if success {
			wounds++
		}
	}
	return wounds
}

// rollSaves rolls numWounds D6s for the defender. Each roll that fails
// (< saveThreshold) counts as a failed save. Natural 1 always fails.
// A save threshold > 6 means saves are impossible.
func rollSaves(roller *dice.Roller, numWounds, saveThreshold int) int {
	failed := 0
	for i := 0; i < numWounds; i++ {
		if saveThreshold > 6 {
			// No possible save
			failed++
			continue
		}
		_, success := roller.RollWithThreshold(saveThreshold)
		if !success {
			failed++
		}
	}
	return failed
}
