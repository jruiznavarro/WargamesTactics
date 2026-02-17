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
	CriticalHits int
	Wounds       int
	SavesFailed  int
	DamageDealt  int
	MortalDealt  int
	WardSaved    int
	ModelsSlain  int
}

func (r CombatResult) String() string {
	return fmt.Sprintf(
		"%s: %d attacks -> %d hits (%d crit) -> %d wounds -> %d unsaved -> %d damage (%d mortal, %d warded, %d slain)",
		r.WeaponName, r.TotalAttacks, r.Hits, r.CriticalHits, r.Wounds,
		r.SavesFailed, r.DamageDealt, r.MortalDealt, r.WardSaved, r.ModelsSlain,
	)
}

// clampHitWoundMod caps hit/wound modifiers. AoS4 Rule 17.1:
// Hit and wound roll modifiers are capped at +1/-1.
func clampHitWoundMod(mod int) int {
	if mod > 1 {
		return 1
	}
	if mod < -1 {
		return -1
	}
	return mod
}

// clampSaveMod caps save modifiers. AoS4 Rule 17.1:
// Save roll positive modifiers capped at +1, no cap on negative.
func clampSaveMod(mod int) int {
	if mod > 1 {
		return 1
	}
	return mod
}

// ResolveAttacks resolves the full attack sequence for one weapon profile.
// AoS4 Rules 17.0: hit -> wound -> save -> damage, with modifier caps,
// critical hits, weapon abilities, and ward saves.
func ResolveAttacks(roller *dice.Roller, engine *rules.Engine, attacker *core.Unit, defender *core.Unit, weapon *core.Weapon, isShooting bool) CombatResult {
	aliveModelsBefore := defender.AliveModels()
	baseAttacks := attacker.AliveModels() * weapon.Attacks

	// Build base context
	baseCtx := &rules.Context{
		Attacker:   attacker,
		Defender:   defender,
		Weapon:     weapon,
		IsShooting: isShooting,
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

	// Companion weapons ignore positive hit/wound modifiers from abilities (Rule 20.0)
	isCompanion := weapon.HasAbility(core.AbilityCompanion)

	// Step 1: Hit rolls with modifier caps (Rule 17.1)
	hitCtx := &rules.Context{Attacker: attacker, Defender: defender, Weapon: weapon, IsShooting: isShooting}
	engine.Evaluate(rules.BeforeHitRoll, hitCtx)
	hitMod := clampHitWoundMod(hitCtx.Modifiers.HitMod)
	if isCompanion && hitMod > 0 {
		hitMod = 0
	}

	hr := rollHits(roller, totalAttacks, weapon.ToHit, hitMod, weapon)
	result.Hits = hr.Hits + hr.AutoWounds // Total hits for display
	result.CriticalHits = hr.Crits

	// Step 2: Wound rolls with modifier caps (Rule 17.1)
	// Only normal hits go through wound rolls; auto-wounds skip this step
	woundCtx := &rules.Context{Attacker: attacker, Defender: defender, Weapon: weapon, IsShooting: isShooting}
	engine.Evaluate(rules.BeforeWoundRoll, woundCtx)
	woundMod := clampHitWoundMod(woundCtx.Modifiers.WoundMod)
	if isCompanion && woundMod > 0 {
		woundMod = 0
	}
	wounds := rollWounds(roller, hr.Hits, weapon.ToWound, woundMod) + hr.AutoWounds
	result.Wounds = wounds

	// Step 3: Save rolls with modifier caps (Rule 17.1)
	saveCtx := &rules.Context{Attacker: attacker, Defender: defender, Weapon: weapon, IsShooting: isShooting}
	engine.Evaluate(rules.BeforeSaveRoll, saveCtx)
	saveMod := clampSaveMod(saveCtx.Modifiers.SaveMod)

	// Calculate effective Rend: base + Anti-X bonuses + rule modifiers
	effectiveRend := weapon.Rend + saveCtx.Modifiers.RendMod + applyAntiRend(weapon, defender)

	// Save threshold = Save + Rend - SaveMod
	// Rend is stored as positive (e.g. 1), making save harder
	saveThreshold := defender.Stats.Save + effectiveRend - saveMod
	savesFailed := rollSaves(roller, wounds, saveThreshold)
	result.SavesFailed = savesFailed

	// Step 4: Damage (with Charge weapon ability, Rule 20.0)
	dmgCtx := &rules.Context{Attacker: attacker, Defender: defender, Weapon: weapon, IsShooting: isShooting}
	engine.Evaluate(rules.BeforeDamage, dmgCtx)
	damagePerWound := weapon.Damage + dmgCtx.Modifiers.DamageMod
	if weapon.HasAbility(core.AbilityCharge) && attacker.HasCharged {
		damagePerWound++
	}
	if damagePerWound < 1 {
		damagePerWound = 1
	}

	// Build damage pool (Rule 18.0)
	damagePool := savesFailed * damagePerWound

	// Add mortal wounds from Crit (Mortal) and rules
	totalMortals := hr.CritMortals +
		baseCtx.Modifiers.MortalWounds +
		hitCtx.Modifiers.MortalWounds +
		woundCtx.Modifiers.MortalWounds +
		saveCtx.Modifiers.MortalWounds +
		dmgCtx.Modifiers.MortalWounds
	damagePool += totalMortals
	result.MortalDealt = totalMortals

	// Step 5: Ward saves (Rule 18.1)
	wardSaved := 0
	if defender.WardSave > 0 && damagePool > 0 {
		wardSaved = rollWards(roller, damagePool, defender.WardSave)
		damagePool -= wardSaved
	}
	result.WardSaved = wardSaved

	// Step 6: Allocate damage (Rule 18.2)
	if damagePool > 0 {
		defender.AllocateDamage(damagePool)
	}
	result.DamageDealt = damagePool
	result.ModelsSlain = aliveModelsBefore - defender.AliveModels()

	return result
}

// ResolveMortalWounds applies mortal wounds directly, bypassing saves.
// Ward saves still apply (Rule 18.1).
func ResolveMortalWounds(roller *dice.Roller, defender *core.Unit, mortalWounds int) (damage int, slain int) {
	aliveModelsBefore := defender.AliveModels()
	pool := mortalWounds

	if defender.WardSave > 0 && pool > 0 {
		warded := rollWards(roller, pool, defender.WardSave)
		pool -= warded
	}

	if pool > 0 {
		defender.AllocateDamage(pool)
	}
	return pool, aliveModelsBefore - defender.AliveModels()
}

// ResolveCombat resolves all melee weapon attacks from attacker against defender.
func ResolveCombat(roller *dice.Roller, engine *rules.Engine, attacker *core.Unit, defender *core.Unit) []CombatResult {
	var results []CombatResult
	defenderAliveStart := defender.AliveModels()
	for _, idx := range attacker.MeleeWeapons() {
		if attacker.IsDestroyed() || defender.IsDestroyed() {
			break
		}
		result := ResolveAttacks(roller, engine, attacker, defender, &attacker.Weapons[idx], false)
		results = append(results, result)
	}
	fireCombatTriggers(engine, attacker, defender, results, defenderAliveStart, false)
	return results
}

// ResolveShooting resolves all ranged weapon attacks from attacker against defender.
func ResolveShooting(roller *dice.Roller, engine *rules.Engine, attacker *core.Unit, defender *core.Unit) []CombatResult {
	var results []CombatResult
	defenderAliveStart := defender.AliveModels()
	for _, idx := range attacker.RangedWeapons() {
		if attacker.IsDestroyed() || defender.IsDestroyed() {
			break
		}
		result := ResolveAttacks(roller, engine, attacker, defender, &attacker.Weapons[idx], true)
		results = append(results, result)
	}
	fireCombatTriggers(engine, attacker, defender, results, defenderAliveStart, true)
	return results
}

// fireCombatTriggers fires post-combat events via the rules engine.
func fireCombatTriggers(engine *rules.Engine, attacker *core.Unit, defender *core.Unit, results []CombatResult, defenderAliveStart int, isShooting bool) {
	totalSlain := defenderAliveStart - defender.AliveModels()

	// AfterCombatResolve — always fires after a combat resolution
	totalDamage := 0
	for _, r := range results {
		totalDamage += r.DamageDealt
	}
	afterCtx := &rules.Context{
		Attacker:   attacker,
		Defender:   defender,
		IsShooting: isShooting,
	}
	afterCtx.Modifiers.DamageMod = totalDamage // Reuse as "total damage dealt" info
	engine.Evaluate(rules.AfterCombatResolve, afterCtx)

	// OnModelSlain — fires once per model slain
	if totalSlain > 0 {
		slainCtx := &rules.Context{
			Attacker:   attacker,
			Defender:   defender,
			IsShooting: isShooting,
		}
		slainCtx.Modifiers.AttacksMod = totalSlain // Reuse as "models slain count" info
		engine.Evaluate(rules.OnModelSlain, slainCtx)
	}

	// OnUnitDestroyed — fires if the defender was fully destroyed
	if defender.IsDestroyed() {
		destroyedCtx := &rules.Context{
			Attacker:   attacker,
			Defender:   defender,
			IsShooting: isShooting,
		}
		engine.Evaluate(rules.OnUnitDestroyed, destroyedCtx)
	}
}

// applyAntiRend calculates extra Rend from Anti-X weapon abilities (Rule 20.0).
func applyAntiRend(weapon *core.Weapon, defender *core.Unit) int {
	extra := 0
	if weapon.HasAbility(core.AbilityAntiInfantry) && defender.HasKeyword(core.KeywordInfantry) {
		extra++
	}
	if weapon.HasAbility(core.AbilityAntiCavalry) && defender.HasKeyword(core.KeywordCavalry) {
		extra++
	}
	if weapon.HasAbility(core.AbilityAntiHero) && defender.HasKeyword(core.KeywordHero) {
		extra++
	}
	if weapon.HasAbility(core.AbilityAntiMonster) && defender.HasKeyword(core.KeywordMonster) {
		extra++
	}
	if weapon.HasAbility(core.AbilityAntiCharge) && defender.HasCharged {
		extra++
	}
	return extra
}

// hitResult holds the output of rollHits.
type hitResult struct {
	Hits        int // Normal hits that proceed to wound roll
	Crits       int // Critical hit count (for tracking)
	CritMortals int // Mortal wounds from Crit(Mortal)
	AutoWounds  int // Auto-wounds from Crit(Auto-wound), skip wound roll -> go to save
}

// rollHits rolls D6s for hits. AoS4 Rule 17.0:
// Natural 1 always fails. Unmodified 6 = critical hit.
// Processes Crit weapon abilities.
func rollHits(roller *dice.Roller, numAttacks, toHit, modifier int, weapon *core.Weapon) hitResult {
	var r hitResult
	for i := 0; i < numAttacks; i++ {
		roll := roller.RollD6()

		if roll == 1 {
			continue
		}

		isCrit := roll == 6

		// Crit (Mortal): inflict mortal damage = weapon Damage, attack sequence ends
		if isCrit && weapon.HasAbility(core.AbilityCritMortal) {
			r.CritMortals += weapon.Damage
			r.Crits++
			continue
		}

		// Crit (Auto-wound): automatically wounds, skips wound roll -> goes to save
		if isCrit && weapon.HasAbility(core.AbilityCritAutoWound) {
			r.AutoWounds++
			r.Crits++
			continue
		}

		modifiedRoll := roll + modifier

		// Crit (2 Hits): scores 2 hits instead of 1
		if isCrit && weapon.HasAbility(core.AbilityCrit2Hits) {
			if modifiedRoll >= toHit {
				r.Hits += 2
				r.Crits++
			}
			continue
		}

		if modifiedRoll >= toHit {
			r.Hits++
			if isCrit {
				r.Crits++
			}
		}
	}
	return r
}

// rollWounds rolls D6s for wounds. Natural 1 always fails.
func rollWounds(roller *dice.Roller, numHits, toWound, modifier int) int {
	wounds := 0
	for i := 0; i < numHits; i++ {
		roll := roller.RollD6()
		if roll == 1 {
			continue
		}
		if roll+modifier >= toWound {
			wounds++
		}
	}
	return wounds
}

// rollSaves rolls D6s for saves. Natural 1 always fails.
// saveThreshold > 6 means saves are impossible.
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

// rollWards rolls D6 per damage point. AoS4 Rule 18.1.
func rollWards(roller *dice.Roller, damagePool, wardValue int) int {
	saved := 0
	for i := 0; i < damagePool; i++ {
		roll := roller.RollD6()
		if roll >= wardValue {
			saved++
		}
	}
	return saved
}
