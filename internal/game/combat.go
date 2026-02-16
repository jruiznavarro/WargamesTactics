package game

import (
	"fmt"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
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
// The rules engine is consulted at each step for modifiers.
func ResolveAttacks(roller *dice.Roller, engine *rules.Engine, attacker *core.Unit, defender *core.Unit, weapon *core.Weapon) CombatResult {
	aliveModelsBefore := defender.AliveModels()
	baseAttacks := attacker.AliveModels() * weapon.Attacks

	// Build base context for this weapon resolution
	baseCtx := &rules.Context{
		Attacker: attacker,
		Defender: defender,
		Weapon:   weapon,
	}

	// Step 0: Attack count modifiers
	engine.Evaluate(rules.BeforeAttackCount, baseCtx)
	totalAttacks := baseAttacks + baseCtx.Modifiers.AttacksMod
	if totalAttacks < 0 {
		totalAttacks = 0
	}

	result := CombatResult{
		AttackerID:   attacker.ID,
		DefenderID:   defender.ID,
		WeaponName:   weapon.Name,
		TotalAttacks: totalAttacks,
	}

	// Step 1: Hit rolls (evaluate modifiers)
	hitCtx := &rules.Context{Attacker: attacker, Defender: defender, Weapon: weapon}
	engine.Evaluate(rules.BeforeHitRoll, hitCtx)
	hits := rollHits(roller, totalAttacks, weapon.ToHit, hitCtx.Modifiers.HitMod)
	result.Hits = hits

	// Step 2: Wound rolls (evaluate modifiers)
	woundCtx := &rules.Context{Attacker: attacker, Defender: defender, Weapon: weapon}
	engine.Evaluate(rules.BeforeWoundRoll, woundCtx)
	wounds := rollWounds(roller, hits, weapon.ToWound, woundCtx.Modifiers.WoundMod)
	result.Wounds = wounds

	// Step 3: Save rolls (evaluate modifiers -- cover, etc.)
	saveCtx := &rules.Context{Attacker: attacker, Defender: defender, Weapon: weapon}
	engine.Evaluate(rules.BeforeSaveRoll, saveCtx)
	saveThreshold := defender.Stats.Save - weapon.Rend // Rend is negative, subtracting makes save harder
	saveThreshold -= saveCtx.Modifiers.SaveMod         // Positive SaveMod = easier save = lower threshold
	savesFailed := rollSaves(roller, wounds, saveThreshold)
	result.SavesFailed = savesFailed

	// Step 4: Damage (evaluate modifiers)
	dmgCtx := &rules.Context{Attacker: attacker, Defender: defender, Weapon: weapon}
	engine.Evaluate(rules.BeforeDamage, dmgCtx)
	damagePerWound := weapon.Damage + dmgCtx.Modifiers.DamageMod
	if damagePerWound < 1 {
		damagePerWound = 1
	}
	totalDamage := savesFailed * damagePerWound
	defender.AllocateDamage(totalDamage)
	result.DamageDealt = totalDamage

	// Apply mortal wounds from rules (bypasses saves)
	totalMortals := baseCtx.Modifiers.MortalWounds +
		hitCtx.Modifiers.MortalWounds +
		woundCtx.Modifiers.MortalWounds +
		saveCtx.Modifiers.MortalWounds +
		dmgCtx.Modifiers.MortalWounds
	if totalMortals > 0 {
		defender.AllocateDamage(totalMortals)
		result.DamageDealt += totalMortals
	}

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
func ResolveCombat(roller *dice.Roller, engine *rules.Engine, attacker *core.Unit, defender *core.Unit) []CombatResult {
	var results []CombatResult
	for _, idx := range attacker.MeleeWeapons() {
		if attacker.IsDestroyed() || defender.IsDestroyed() {
			break
		}
		result := ResolveAttacks(roller, engine, attacker, defender, &attacker.Weapons[idx])
		results = append(results, result)
	}
	return results
}

// ResolveShooting resolves all ranged weapon attacks from attacker against defender.
func ResolveShooting(roller *dice.Roller, engine *rules.Engine, attacker *core.Unit, defender *core.Unit) []CombatResult {
	var results []CombatResult
	for _, idx := range attacker.RangedWeapons() {
		if attacker.IsDestroyed() || defender.IsDestroyed() {
			break
		}
		result := ResolveAttacks(roller, engine, attacker, defender, &attacker.Weapons[idx])
		results = append(results, result)
	}
	return results
}

// rollHits rolls numAttacks D6s and counts successes (>= toHit).
// Natural 1 always fails. The modifier is added to each roll result.
func rollHits(roller *dice.Roller, numAttacks, toHit, modifier int) int {
	hits := 0
	for i := 0; i < numAttacks; i++ {
		roll := roller.RollD6()
		if roll == 1 {
			continue // Natural 1 always fails
		}
		if roll+modifier >= toHit {
			hits++
		}
	}
	return hits
}

// rollWounds rolls numHits D6s and counts successes (>= toWound).
// Natural 1 always fails. The modifier is added to each roll result.
func rollWounds(roller *dice.Roller, numHits, toWound, modifier int) int {
	wounds := 0
	for i := 0; i < numHits; i++ {
		roll := roller.RollD6()
		if roll == 1 {
			continue // Natural 1 always fails
		}
		if roll+modifier >= toWound {
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
			failed++
			continue
		}
		roll := roller.RollD6()
		if roll == 1 || roll < saveThreshold {
			failed++
		}
	}
	return failed
}
