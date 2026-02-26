package ai

import (
	"math"
	"sort"

	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// AIPlayer implements the Player interface with heuristic decisions.
type AIPlayer struct {
	id   int
	name string

	// Per-phase state to track commands already issued this call cycle
	boostedAttackUnit core.UnitID // unit that got All-out Attack this phase
	boostedDefenceUnit core.UnitID // unit that got All-out Defence this phase
	tacticSelected     bool        // whether we selected a tactic this turn
}

// NewAIPlayer creates a new AI player.
func NewAIPlayer(id int, name string) *AIPlayer {
	return &AIPlayer{id: id, name: name}
}

func (a *AIPlayer) ID() int      { return a.id }
func (a *AIPlayer) Name() string { return a.name }

// cpAvailable returns the AI's remaining command points.
func (a *AIPlayer) cpAvailable(view *game.GameView) int {
	return view.CommandPoints[a.id]
}

// reserveCP is the minimum CP we try to keep for defensive reactions.
const reserveCP = 1

// canSpendCP returns true if spending 'cost' CP would still leave reserveCP available,
// unless force is true (for high-priority actions).
func (a *AIPlayer) canSpendCP(view *game.GameView, cost int, force bool) bool {
	cp := a.cpAvailable(view)
	if force {
		return cp >= cost
	}
	return cp >= cost+reserveCP
}

func (a *AIPlayer) GetNextCommand(view *game.GameView, currentPhase phase.Phase) interface{} {
	switch currentPhase.Type {
	case phase.PhaseHero:
		return a.decideHeroPhase(view)
	case phase.PhaseMovement:
		return a.decideMovement(view)
	case phase.PhaseShooting:
		return a.decideShooting(view)
	case phase.PhaseCharging:
		return a.decideCharge(view)
	case phase.PhaseCombat:
		return a.decideCombat(view)
	case phase.PhaseEndOfTurn:
		return a.decideEndOfTurn(view)
	default:
		return &command.EndPhaseCommand{OwnerID: a.id}
	}
}

// =============================================================================
// Hero Phase: Cast spells, Chant prayers, Rally wounded units
// =============================================================================

func (a *AIPlayer) decideHeroPhase(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	// 1. Try to cast damage spells on enemies
	for _, u := range myUnits {
		if !u.CanCast || len(u.Spells) == 0 {
			continue
		}
		for spellIdx, spell := range u.Spells {
			if len(enemies) == 0 {
				continue
			}
			// Find best target within spell range
			target := a.findBestSpellTarget(u, spell, enemies)
			if target != nil {
				return &command.CastCommand{
					OwnerID:    a.id,
					CasterID:   core.UnitID(u.ID),
					SpellIndex: spellIdx,
					TargetID:   core.UnitID(target.ID),
				}
			}
		}
	}

	// 2. Try to chant prayers
	for _, u := range myUnits {
		if !u.CanChant || len(u.Prayers) == 0 {
			continue
		}
		for prayerIdx, prayer := range u.Prayers {
			target := a.findBestPrayerTarget(u, prayer, myUnits, enemies)
			if target != nil {
				return &command.ChantCommand{
					OwnerID:     a.id,
					ChanterID:   core.UnitID(u.ID),
					PrayerIndex: prayerIdx,
					TargetID:    core.UnitID(target.ID),
					BankPoints:  false,
				}
			}
		}
	}

	// 3. Rally wounded units (costs 1 CP)
	if a.canSpendCP(view, 1, false) {
		for _, u := range myUnits {
			if u.IsEngaged {
				continue
			}
			// Rally if unit has lost models or wounds
			if u.CurrentWounds < u.MaxWounds && u.AliveModels > 0 {
				return &command.RallyCommand{
					OwnerID: a.id,
					UnitID:  core.UnitID(u.ID),
				}
			}
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
}

func (a *AIPlayer) findBestSpellTarget(caster game.UnitView, spell game.SpellView, enemies []*game.UnitView) *game.UnitView {
	casterPos := core.Position{X: caster.Position[0], Y: caster.Position[1]}
	var best *game.UnitView
	bestScore := -1.0

	for _, enemy := range enemies {
		enemyPos := core.Position{X: enemy.Position[0], Y: enemy.Position[1]}
		dist := core.Distance(casterPos, enemyPos)
		if spell.Range > 0 && dist > float64(spell.Range) {
			continue
		}
		// Prefer enemies with fewer wounds remaining (easier to finish off)
		score := float64(enemy.MaxWounds-enemy.CurrentWounds) + 1.0/dist
		if score > bestScore {
			bestScore = score
			best = enemy
		}
	}
	return best
}

func (a *AIPlayer) findBestPrayerTarget(chanter game.UnitView, prayer game.PrayerView, allies []game.UnitView, enemies []*game.UnitView) *game.UnitView {
	chanterPos := core.Position{X: chanter.Position[0], Y: chanter.Position[1]}

	// For prayers, try targeting the nearest enemy in range
	for _, enemy := range enemies {
		enemyPos := core.Position{X: enemy.Position[0], Y: enemy.Position[1]}
		dist := core.Distance(chanterPos, enemyPos)
		if prayer.Range > 0 && dist <= float64(prayer.Range) {
			return enemy
		}
	}
	return nil
}

// =============================================================================
// Movement Phase: Objective-aware movement, Run when beneficial
// =============================================================================

func (a *AIPlayer) decideMovement(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	for _, u := range myUnits {
		if u.HasMoved {
			continue
		}

		origin := core.Position{X: u.Position[0], Y: u.Position[1]}
		moveAllowance := float64(u.MoveSpeed)

		// Determine best destination: objective or enemy
		dest, shouldRun := a.bestMoveDestination(u, origin, moveAllowance, view, enemies)
		if dest == nil {
			continue
		}

		target := *dest
		dist := core.Distance(origin, target)

		if dist < 0.5 {
			// Already at destination
			continue
		}

		if shouldRun && !u.IsEngaged {
			// Run: move + D6" (average 3.5"), but cannot shoot or charge after
			// We approximate run distance as move + 3" for destination calc
			runAllowance := moveAllowance + 3.0
			runDest := origin.Towards(target, math.Min(runAllowance, dist))

			// Make sure we don't run into engagement range
			if !a.wouldEngageEnemy(runDest, enemies) {
				return &command.RunCommand{
					OwnerID:     a.id,
					UnitID:      core.UnitID(u.ID),
					Destination: runDest,
				}
			}
		}

		if u.IsEngaged {
			// Engaged units can't normal move
			continue
		}

		// Normal move: respect 3" engagement rule
		const safeDistance = 3.1
		if dist <= safeDistance {
			continue
		}
		actualMove := math.Min(moveAllowance, dist-safeDistance)
		if actualMove <= 0 {
			continue
		}

		moveDest := origin.Towards(target, actualMove)

		// Double check we won't engage
		if a.wouldEngageEnemy(moveDest, enemies) {
			// Clamp to safe distance from nearest enemy
			nearestEnemy := a.findNearestEnemy(u, enemies)
			if nearestEnemy != nil {
				enemyPos := core.Position{X: nearestEnemy.Position[0], Y: nearestEnemy.Position[1]}
				enemyDist := core.Distance(origin, enemyPos)
				if enemyDist > safeDistance {
					actualMove = math.Min(moveAllowance, enemyDist-safeDistance)
					moveDest = origin.Towards(target, actualMove)
				} else {
					continue
				}
			}
		}

		return &command.MoveCommand{
			OwnerID:     a.id,
			UnitID:      core.UnitID(u.ID),
			Destination: moveDest,
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
}

// bestMoveDestination decides where a unit should move.
// Returns the target position and whether the unit should run.
func (a *AIPlayer) bestMoveDestination(u game.UnitView, origin core.Position, moveAllowance float64, view *game.GameView, enemies []*game.UnitView) (*core.Position, bool) {
	// Score potential targets: uncontrolled/enemy objectives vs nearest enemy
	var bestTarget *core.Position
	bestScore := -math.MaxFloat64
	shouldRun := false

	// Evaluate objectives
	for _, obj := range view.Objectives {
		if obj.ControlledBy == a.id {
			continue // Already ours
		}
		objPos := core.Position{X: obj.Position[0], Y: obj.Position[1]}
		dist := core.Distance(origin, objPos)

		// Score: closer objectives are better, contested/enemy objectives are high priority
		score := 100.0 - dist
		if obj.ControlledBy >= 0 && obj.ControlledBy != a.id {
			score += 50.0 // Enemy-controlled: higher priority
		}
		// Penalize if too far to reach this turn
		if dist > moveAllowance*2 {
			score -= 30.0
		}

		if score > bestScore {
			bestScore = score
			pos := objPos
			bestTarget = &pos
			// Run toward distant objectives if no enemies within charge range
			nearestEnemyDist := a.nearestEnemyDistance(u, enemies)
			shouldRun = dist > moveAllowance && nearestEnemyDist > 15.0
		}
	}

	// Evaluate moving toward nearest enemy (default behavior)
	if len(enemies) > 0 {
		nearest := a.findNearestEnemy(u, enemies)
		if nearest != nil {
			enemyPos := core.Position{X: nearest.Position[0], Y: nearest.Position[1]}
			dist := core.Distance(origin, enemyPos)

			// Base score for enemy approach
			score := 80.0 - dist

			// Bonus if unit has melee weapons (wants to engage)
			hasMelee := false
			for _, w := range u.Weapons {
				if w.Range == 0 {
					hasMelee = true
					break
				}
			}
			if hasMelee {
				score += 20.0
			}

			if score > bestScore {
				bestScore = score
				bestTarget = &enemyPos
				shouldRun = false // Don't run toward enemies (need to charge)
			}
		}
	}

	return bestTarget, shouldRun
}

func (a *AIPlayer) nearestEnemyDistance(u game.UnitView, enemies []*game.UnitView) float64 {
	minDist := math.MaxFloat64
	origin := core.Position{X: u.Position[0], Y: u.Position[1]}
	for _, enemy := range enemies {
		pos := core.Position{X: enemy.Position[0], Y: enemy.Position[1]}
		d := core.Distance(origin, pos)
		if d < minDist {
			minDist = d
		}
	}
	return minDist
}

func (a *AIPlayer) wouldEngageEnemy(pos core.Position, enemies []*game.UnitView) bool {
	for _, enemy := range enemies {
		enemyPos := core.Position{X: enemy.Position[0], Y: enemy.Position[1]}
		if core.Distance(pos, enemyPos) <= 3.0 {
			return true
		}
	}
	return false
}

// =============================================================================
// Shooting Phase: Target prioritization
// =============================================================================

func (a *AIPlayer) decideShooting(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	if len(enemies) == 0 {
		return &command.EndPhaseCommand{OwnerID: a.id}
	}

	for _, u := range myUnits {
		if u.HasShot || u.HasRun {
			continue
		}

		// Find ranged weapons and max range
		hasRanged := false
		maxRange := 0
		for _, w := range u.Weapons {
			if w.Range > 0 {
				hasRanged = true
				if w.Range > maxRange {
					maxRange = w.Range
				}
			}
		}
		if !hasRanged {
			continue
		}

		// Find best target in range (prioritize wounded enemies and objective contesters)
		target := a.findBestShootTarget(u, enemies, maxRange, view)
		if target != nil {
			return &command.ShootCommand{
				OwnerID:   a.id,
				ShooterID: core.UnitID(u.ID),
				TargetID:  core.UnitID(target.ID),
			}
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
}

func (a *AIPlayer) findBestShootTarget(shooter game.UnitView, enemies []*game.UnitView, maxRange int, view *game.GameView) *game.UnitView {
	shooterPos := core.Position{X: shooter.Position[0], Y: shooter.Position[1]}

	type scoredTarget struct {
		enemy *game.UnitView
		score float64
	}

	var candidates []scoredTarget
	for _, enemy := range enemies {
		enemyPos := core.Position{X: enemy.Position[0], Y: enemy.Position[1]}
		dist := core.Distance(shooterPos, enemyPos)
		if dist > float64(maxRange) {
			continue
		}

		score := 0.0
		// Prefer wounded enemies (easier to finish off)
		woundRatio := 1.0 - float64(enemy.CurrentWounds)/float64(enemy.MaxWounds)
		score += woundRatio * 30.0

		// Prefer enemies contesting our objectives
		for _, obj := range view.Objectives {
			if obj.ControlledBy == a.id {
				objPos := core.Position{X: obj.Position[0], Y: obj.Position[1]}
				if core.Distance(enemyPos, objPos) <= 6.0 {
					score += 20.0
					break
				}
			}
		}

		// Prefer closer targets
		score += (float64(maxRange) - dist)

		candidates = append(candidates, scoredTarget{enemy, score})
	}

	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	return candidates[0].enemy
}

// =============================================================================
// Charge Phase: Prioritized charging
// =============================================================================

func (a *AIPlayer) decideCharge(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	if len(enemies) == 0 {
		return &command.EndPhaseCommand{OwnerID: a.id}
	}

	for _, u := range myUnits {
		if u.HasCharged || u.HasRun || u.HasRetreated {
			continue
		}

		// Find best charge target within 12"
		target := a.findBestChargeTarget(u, enemies, view)
		if target != nil {
			return &command.ChargeCommand{
				OwnerID:   a.id,
				ChargerID: core.UnitID(u.ID),
				TargetID:  core.UnitID(target.ID),
			}
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
}

func (a *AIPlayer) findBestChargeTarget(charger game.UnitView, enemies []*game.UnitView, view *game.GameView) *game.UnitView {
	chargerPos := core.Position{X: charger.Position[0], Y: charger.Position[1]}
	var best *game.UnitView
	bestScore := -math.MaxFloat64

	for _, enemy := range enemies {
		enemyPos := core.Position{X: enemy.Position[0], Y: enemy.Position[1]}
		dist := core.Distance(chargerPos, enemyPos)
		if dist > 12.0 || dist <= 3.0 {
			continue
		}

		score := 0.0
		// Prefer closer targets (easier to charge)
		score += (12.0 - dist) * 5.0

		// Prefer wounded enemies
		if enemy.MaxWounds > 0 {
			woundRatio := 1.0 - float64(enemy.CurrentWounds)/float64(enemy.MaxWounds)
			score += woundRatio * 20.0
		}

		// Prefer enemies near objectives
		for _, obj := range view.Objectives {
			objPos := core.Position{X: obj.Position[0], Y: obj.Position[1]}
			if core.Distance(enemyPos, objPos) <= 6.0 {
				score += 15.0
				break
			}
		}

		if score > bestScore {
			bestScore = score
			best = enemy
		}
	}
	return best
}

// =============================================================================
// Combat Phase: Fight with All-out Attack/Defence
// =============================================================================

func (a *AIPlayer) decideCombat(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	if len(enemies) == 0 {
		return &command.EndPhaseCommand{OwnerID: a.id}
	}

	for _, u := range myUnits {
		if u.HasFought || !u.IsEngaged {
			continue
		}

		// Find nearest enemy within 3" to fight
		var target *game.UnitView
		for _, enemy := range enemies {
			dist := a.distBetween(u, *enemy)
			if dist <= 3.0 {
				target = enemy
				break
			}
		}

		if target == nil {
			continue
		}

		// Use All-out Attack for strong attackers (high damage output)
		if a.boostedAttackUnit != core.UnitID(u.ID) && a.canSpendCP(view, 1, false) {
			totalAttacks := 0
			for _, w := range u.Weapons {
				totalAttacks += w.Attacks
			}
			// Boost units with 4+ attacks for maximum value
			if totalAttacks >= 4 {
				a.boostedAttackUnit = core.UnitID(u.ID)
				return &command.AllOutAttackCommand{
					OwnerID: a.id,
					UnitID:  core.UnitID(u.ID),
				}
			}
		}

		// Use All-out Defence for wounded or valuable units
		if a.boostedDefenceUnit != core.UnitID(u.ID) && a.canSpendCP(view, 1, false) {
			if u.CurrentWounds < u.MaxWounds/2 && u.AliveModels > 0 {
				a.boostedDefenceUnit = core.UnitID(u.ID)
				return &command.AllOutDefenceCommand{
					OwnerID: a.id,
					UnitID:  core.UnitID(u.ID),
				}
			}
		}

		return &command.FightCommand{
			OwnerID:    a.id,
			AttackerID: core.UnitID(u.ID),
			TargetID:   core.UnitID(target.ID),
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
}

// =============================================================================
// End of Turn: Battle Tactic selection
// =============================================================================

func (a *AIPlayer) decideEndOfTurn(view *game.GameView) interface{} {
	// Select a battle tactic if available and not yet selected
	if !a.tacticSelected {
		if tactics, ok := view.BattleTactics[a.id]; ok && tactics != nil {
			if tactics.ActiveTactic == nil && len(tactics.AvailableCards) > 0 {
				cardID, tier := a.chooseBestTactic(view, tactics)
				if cardID > 0 {
					a.tacticSelected = true
					return &command.SelectBattleTacticCommand{
						OwnerID: a.id,
						CardID:  cardID,
						Tier:    tier,
					}
				}
			}
		}
	}

	// Reset per-turn state
	a.boostedAttackUnit = 0
	a.boostedDefenceUnit = 0
	a.tacticSelected = false

	return &command.EndPhaseCommand{OwnerID: a.id}
}

// chooseBestTactic evaluates available tactics and picks the most achievable one.
// Returns cardID and tier as ints.
func (a *AIPlayer) chooseBestTactic(view *game.GameView, tactics *game.BattleTacticsView) (int, int) {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	bestCardID := 0
	bestTier := 0
	bestScore := -1.0

	for _, card := range tactics.AvailableCards {
		for tierIdx := 0; tierIdx < 3; tierIdx++ {
			score := a.scoreTactic(card.CardID, tierIdx, myUnits, enemies, view)
			if score > bestScore {
				bestScore = score
				bestCardID = card.CardID
				bestTier = tierIdx
			}
		}
	}

	// Only select if we have a reasonable chance (score > 0)
	if bestScore <= 0 {
		// Fall back to Broken Ranks Affray (easiest - just destroy 1 unit)
		for _, card := range tactics.AvailableCards {
			if card.CardID == 2 { // CardBrokenRanks
				return card.CardID, 0 // TierAffray
			}
		}
		// If no Broken Ranks, pick first available at Affray
		if len(tactics.AvailableCards) > 0 {
			return tactics.AvailableCards[0].CardID, 0
		}
	}

	return bestCardID, bestTier
}

// scoreTactic evaluates how likely we are to complete a tactic.
func (a *AIPlayer) scoreTactic(cardID int, tier int, myUnits []game.UnitView, enemies []*game.UnitView, view *game.GameView) float64 {
	switch cardID {
	case 1: // Savage Spearhead - objective control
		return a.scoreSavageSpearhead(tier, view)
	case 2: // Broken Ranks - destroy enemy units
		return a.scoreBrokenRanks(tier, enemies)
	case 3: // Conquer and Hold - units in enemy territory
		return a.scoreConquerAndHold(tier, myUnits, view)
	case 4: // Ferocious Advance - charge/run/fight
		return a.scoreFerocousAdvance(tier, myUnits)
	case 5: // Scouting Force - non-Heroes outside territory
		return a.scoreScoutingForce(tier, myUnits, view)
	case 6: // Attuned to Ghyran - centre control
		return a.scoreAttunedToGhyran(tier, myUnits, view)
	}
	return 0
}

func (a *AIPlayer) scoreSavageSpearhead(tier int, view *game.GameView) float64 {
	myObjectives := 0
	enemyObjectives := 0
	for _, obj := range view.Objectives {
		if obj.ControlledBy == a.id {
			myObjectives++
		} else if obj.ControlledBy >= 0 {
			enemyObjectives++
		}
	}
	switch tier {
	case 0: // Control > enemy objectives
		if myObjectives > enemyObjectives {
			return 8.0
		}
		return 3.0
	case 1: // Control > objectives + 1 near center
		return 2.0
	case 2: // Control ALL objectives
		if len(view.Objectives) <= 2 {
			return 1.0
		}
		return 0.5
	}
	return 0
}

func (a *AIPlayer) scoreBrokenRanks(tier int, enemies []*game.UnitView) float64 {
	// Count enemies that are nearly dead
	nearlyDead := 0
	for _, e := range enemies {
		if e.MaxWounds > 0 && float64(e.CurrentWounds)/float64(e.MaxWounds) <= 0.3 {
			nearlyDead++
		}
	}

	switch tier {
	case 0: // Destroy ≥1 enemy
		if nearlyDead >= 1 {
			return 9.0
		}
		if len(enemies) > 0 {
			return 5.0
		}
		return 0
	case 1: // Destroy ≥2 enemies
		if nearlyDead >= 2 {
			return 7.0
		}
		return 2.0
	case 2: // Destroy ≥3 enemies
		if nearlyDead >= 3 {
			return 6.0
		}
		return 1.0
	}
	return 0
}

func (a *AIPlayer) scoreConquerAndHold(tier int, myUnits []game.UnitView, view *game.GameView) float64 {
	if view.Territories[0].Name == "" {
		return 0 // No battleplan/territories
	}

	// Count units that can reach enemy territory
	enemyTerritory := a.getEnemyTerritory(view)
	unitsInEnemyTerritory := 0
	for _, u := range myUnits {
		pos := core.Position{X: u.Position[0], Y: u.Position[1]}
		if a.isInTerritory(pos, enemyTerritory) {
			unitsInEnemyTerritory++
		}
	}

	switch tier {
	case 0: // ≥2 in enemy territory
		if unitsInEnemyTerritory >= 2 {
			return 8.0
		}
		return 2.0
	case 1: // ≥3 in enemy territory
		if unitsInEnemyTerritory >= 3 {
			return 6.0
		}
		return 1.0
	case 2:
		return 0.5
	}
	return 0
}

func (a *AIPlayer) scoreFerocousAdvance(tier int, myUnits []game.UnitView) float64 {
	switch tier {
	case 0: // ≥3 units ran or charged
		mobileUnits := 0
		for _, u := range myUnits {
			if u.MoveSpeed >= 5 {
				mobileUnits++
			}
		}
		if mobileUnits >= 3 {
			return 6.0
		}
		return 2.0
	case 1: // Underdog + ≥2 units fought
		return 1.0 // Hard to predict underdog status
	case 2:
		return 0.5
	}
	return 0
}

func (a *AIPlayer) scoreScoutingForce(tier int, myUnits []game.UnitView, view *game.GameView) float64 {
	if view.Territories[0].Name == "" {
		return 0
	}

	switch tier {
	case 0:
		nonHeroes := 0
		for _, u := range myUnits {
			if u.AliveModels > 0 {
				nonHeroes++
			}
		}
		if nonHeroes >= 3 {
			return 4.0
		}
		return 1.0
	case 1:
		return 1.5
	case 2:
		return 0.5
	}
	return 0
}

func (a *AIPlayer) scoreAttunedToGhyran(tier int, myUnits []game.UnitView, view *game.GameView) float64 {
	centreX := view.BoardWidth / 2
	centreY := view.BoardHeight / 2
	centre := core.Position{X: centreX, Y: centreY}

	switch tier {
	case 0: // ≥2 units within 12" of centre, not in combat
		nearCentre := 0
		for _, u := range myUnits {
			pos := core.Position{X: u.Position[0], Y: u.Position[1]}
			if core.Distance(pos, centre) <= 12.0 && !u.IsEngaged {
				nearCentre++
			}
		}
		if nearCentre >= 2 {
			return 7.0
		}
		return 3.0
	case 1:
		return 1.0
	case 2:
		return 0.5
	}
	return 0
}

// =============================================================================
// Helpers
// =============================================================================

func (a *AIPlayer) getEnemyTerritory(view *game.GameView) game.TerritoryView {
	// Player index 0 gets territory 0, player index 1 gets territory 1
	// The enemy territory is the opposite
	if a.id == 1 {
		return view.Territories[1]
	}
	return view.Territories[0]
}

func (a *AIPlayer) isInTerritory(pos core.Position, territory game.TerritoryView) bool {
	return pos.X >= territory.MinPos[0] && pos.X <= territory.MaxPos[0] &&
		pos.Y >= territory.MinPos[1] && pos.Y <= territory.MaxPos[1]
}

func (a *AIPlayer) getEnemyUnits(view *game.GameView) []*game.UnitView {
	var enemies []*game.UnitView
	for ownerID, units := range view.Units {
		if ownerID == a.id {
			continue
		}
		for i := range units {
			enemies = append(enemies, &units[i])
		}
	}
	return enemies
}

func (a *AIPlayer) findNearestEnemy(unit game.UnitView, enemies []*game.UnitView) *game.UnitView {
	var nearest *game.UnitView
	minDist := math.MaxFloat64

	for _, enemy := range enemies {
		dist := a.distBetween(unit, *enemy)
		if dist < minDist {
			minDist = dist
			nearest = enemy
		}
	}
	return nearest
}

func (a *AIPlayer) distBetween(u1, u2 game.UnitView) float64 {
	dx := u1.Position[0] - u2.Position[0]
	dy := u1.Position[1] - u2.Position[1]
	return math.Sqrt(dx*dx + dy*dy)
}
