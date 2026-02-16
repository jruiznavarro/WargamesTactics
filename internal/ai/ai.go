package ai

import (
	"math"

	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// AIPlayer implements the Player interface with basic heuristic decisions.
type AIPlayer struct {
	id   int
	name string
}

// NewAIPlayer creates a new AI player.
func NewAIPlayer(id int, name string) *AIPlayer {
	return &AIPlayer{id: id, name: name}
}

func (a *AIPlayer) ID() int      { return a.id }
func (a *AIPlayer) Name() string { return a.name }

func (a *AIPlayer) GetNextCommand(view *game.GameView, currentPhase phase.Phase) interface{} {
	switch currentPhase.Type {
	case phase.PhaseMovement:
		return a.decideMovement(view)
	case phase.PhaseShooting:
		return a.decideShooting(view)
	case phase.PhaseCharging:
		return a.decideCharge(view)
	case phase.PhaseCombat:
		return a.decideFight(view)
	default:
		return &command.EndPhaseCommand{OwnerID: a.id}
	}
}

func (a *AIPlayer) decideMovement(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	if len(enemies) == 0 {
		return &command.EndPhaseCommand{OwnerID: a.id}
	}

	// Find the first unmoved unit and move it toward the nearest enemy
	for _, u := range myUnits {
		if u.HasMoved {
			continue
		}

		nearest := a.findNearestEnemy(u, enemies)
		if nearest == nil {
			continue
		}

		origin := core.Position{X: u.Position[0], Y: u.Position[1]}
		target := core.Position{X: nearest.Position[0], Y: nearest.Position[1]}
		dest := origin.Towards(target, float64(u.MoveSpeed))

		return &command.MoveCommand{
			OwnerID:     a.id,
			UnitID:      core.UnitID(u.ID),
			Destination: dest,
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
}

func (a *AIPlayer) decideShooting(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	if len(enemies) == 0 {
		return &command.EndPhaseCommand{OwnerID: a.id}
	}

	for _, u := range myUnits {
		if u.HasShot {
			continue
		}
		// Check if unit has ranged weapons
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

		// Find nearest enemy within range
		for _, enemy := range enemies {
			dist := a.distBetween(u, *enemy)
			if dist <= float64(maxRange) {
				return &command.ShootCommand{
					OwnerID:   a.id,
					ShooterID: core.UnitID(u.ID),
					TargetID:  core.UnitID(enemy.ID),
				}
			}
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
}

func (a *AIPlayer) decideCharge(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	if len(enemies) == 0 {
		return &command.EndPhaseCommand{OwnerID: a.id}
	}

	for _, u := range myUnits {
		if u.HasCharged {
			continue
		}

		// Only charge with melee-focused units or if within 12"
		for _, enemy := range enemies {
			dist := a.distBetween(u, *enemy)
			if dist <= 12.0 && dist > 3.0 {
				return &command.ChargeCommand{
					OwnerID:   a.id,
					ChargerID: core.UnitID(u.ID),
					TargetID:  core.UnitID(enemy.ID),
				}
			}
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
}

func (a *AIPlayer) decideFight(view *game.GameView) interface{} {
	myUnits := view.Units[a.id]
	enemies := a.getEnemyUnits(view)

	if len(enemies) == 0 {
		return &command.EndPhaseCommand{OwnerID: a.id}
	}

	for _, u := range myUnits {
		if u.HasFought {
			continue
		}
		if !u.IsEngaged {
			continue
		}

		// Find nearest enemy within 3" to fight
		for _, enemy := range enemies {
			dist := a.distBetween(u, *enemy)
			if dist <= 3.0 {
				return &command.FightCommand{
					OwnerID:    a.id,
					AttackerID: core.UnitID(u.ID),
					TargetID:   core.UnitID(enemy.ID),
				}
			}
		}
	}

	return &command.EndPhaseCommand{OwnerID: a.id}
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
