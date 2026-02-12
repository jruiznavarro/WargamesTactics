package game

import (
	"errors"
	"fmt"

	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
	"github.com/jruiznavarro/wargamestactics/pkg/dice"
)

// Game holds the entire game state.
type Game struct {
	Board        *board.Board
	Units        map[core.UnitID]*core.Unit
	Players      []Player
	Roller       *dice.Roller
	BattleRound  int
	CurrentPhase phase.PhaseType
	ActivePlayer int // Index into Players slice
	NextUnitID   core.UnitID
	Log          []string
	IsOver       bool
	Winner       int // Player ID of winner, -1 if draw
}

// NewGame creates a new game with the given seed and board dimensions.
func NewGame(seed int64, boardWidth, boardHeight float64) *Game {
	return &Game{
		Board:      board.NewBoard(boardWidth, boardHeight),
		Units:      make(map[core.UnitID]*core.Unit),
		Roller:     dice.NewRoller(seed),
		NextUnitID: 1,
		Winner:     -1,
	}
}

// AddPlayer adds a player to the game. Returns the player index.
func (g *Game) AddPlayer(p Player) int {
	g.Players = append(g.Players, p)
	return len(g.Players) - 1
}

// CreateUnit creates a new unit with the given parameters and adds it to the game.
func (g *Game) CreateUnit(name string, ownerID int, stats core.Stats, weapons []core.Weapon, numModels int, position core.Position, baseSize float64) *core.Unit {
	id := g.NextUnitID
	g.NextUnitID++

	models := make([]core.Model, numModels)
	for i := range models {
		models[i] = core.Model{
			ID:            i,
			Position:      position,
			BaseSize:      baseSize,
			CurrentWounds: stats.Wounds,
			MaxWounds:     stats.Wounds,
			IsAlive:       true,
		}
	}

	unit := &core.Unit{
		ID:      id,
		Name:    name,
		Stats:   stats,
		Models:  models,
		Weapons: weapons,
		OwnerID: ownerID,
	}

	g.Units[id] = unit
	return unit
}

// GetUnit returns a unit by ID, or nil if not found.
func (g *Game) GetUnit(id core.UnitID) *core.Unit {
	return g.Units[id]
}

// UnitsForPlayer returns all non-destroyed units belonging to a player.
func (g *Game) UnitsForPlayer(playerID int) []*core.Unit {
	var units []*core.Unit
	for _, u := range g.Units {
		if u.OwnerID == playerID && !u.IsDestroyed() {
			units = append(units, u)
		}
	}
	return units
}

// View creates a read-only game view for the given player.
func (g *Game) View(playerID int) *GameView {
	unitsByOwner := make(map[int][]UnitView)

	for _, u := range g.Units {
		if u.IsDestroyed() {
			continue
		}
		pos := u.Position()
		totalWounds := 0
		maxWounds := 0
		for i := range u.Models {
			if u.Models[i].IsAlive {
				totalWounds += u.Models[i].CurrentWounds
				maxWounds += u.Models[i].MaxWounds
			}
		}

		var weaponViews []WeaponView
		for _, w := range u.Weapons {
			weaponViews = append(weaponViews, WeaponView{
				Name:    w.Name,
				Range:   w.Range,
				Attacks: w.Attacks,
				ToHit:   w.ToHit,
				ToWound: w.ToWound,
				Rend:    w.Rend,
				Damage:  w.Damage,
			})
		}

		view := UnitView{
			ID:            int(u.ID),
			Name:          u.Name,
			OwnerID:       u.OwnerID,
			Position:      [2]float64{pos.X, pos.Y},
			AliveModels:   u.AliveModels(),
			TotalModels:   len(u.Models),
			CurrentWounds: totalWounds,
			MaxWounds:     maxWounds,
			MoveSpeed:     u.Stats.Move,
			Save:          u.Stats.Save,
			Weapons:       weaponViews,
			HasMoved:      u.HasMoved,
			HasShot:       u.HasShot,
			HasFought:     u.HasFought,
			HasCharged:    u.HasCharged,
		}
		unitsByOwner[u.OwnerID] = append(unitsByOwner[u.OwnerID], view)
	}

	return &GameView{
		Units:        unitsByOwner,
		BoardWidth:   g.Board.Width,
		BoardHeight:  g.Board.Height,
		CurrentPhase: g.CurrentPhase,
		BattleRound:  g.BattleRound,
		ActivePlayer: g.ActivePlayer,
	}
}

// Logf adds a formatted log message.
func (g *Game) Logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	g.Log = append(g.Log, msg)
}

// ExecuteCommand validates and executes a command against the game state.
func (g *Game) ExecuteCommand(cmd interface{}) (command.Result, error) {
	switch c := cmd.(type) {
	case *command.MoveCommand:
		return g.executeMove(c)
	case *command.ShootCommand:
		return g.executeShoot(c)
	case *command.FightCommand:
		return g.executeFight(c)
	case *command.ChargeCommand:
		return g.executeCharge(c)
	case *command.EndPhaseCommand:
		return command.Result{Description: "Phase ended", Success: true}, nil
	default:
		return command.Result{}, errors.New("unknown command type")
	}
}

func (g *Game) executeMove(cmd *command.MoveCommand) (command.Result, error) {
	unit := g.GetUnit(cmd.UnitID)
	if unit == nil {
		return command.Result{}, fmt.Errorf("unit %d not found", cmd.UnitID)
	}
	if unit.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("unit %d does not belong to player %d", cmd.UnitID, cmd.OwnerID)
	}
	if unit.HasMoved {
		return command.Result{}, fmt.Errorf("unit %d has already moved this turn", cmd.UnitID)
	}
	if !g.Board.IsInBounds(cmd.Destination) {
		return command.Result{}, fmt.Errorf("destination (%.1f, %.1f) is out of bounds", cmd.Destination.X, cmd.Destination.Y)
	}

	origin := unit.Position()
	dist := core.Distance(origin, cmd.Destination)
	maxMove := float64(unit.Stats.Move)

	if !board.MoveDistanceValid(origin, cmd.Destination, maxMove) {
		return command.Result{}, fmt.Errorf("move distance %.1f exceeds maximum %d", dist, unit.Stats.Move)
	}

	// Move all alive models to destination (simplified: whole unit moves as one)
	for i := range unit.Models {
		if unit.Models[i].IsAlive {
			unit.Models[i].Position = cmd.Destination
		}
	}
	unit.HasMoved = true

	desc := fmt.Sprintf("%s moved %.1f\" to (%.1f, %.1f)", unit.Name, dist, cmd.Destination.X, cmd.Destination.Y)
	g.Logf("%s", desc)
	return command.Result{Description: desc, Success: true}, nil
}

func (g *Game) executeShoot(cmd *command.ShootCommand) (command.Result, error) {
	shooter := g.GetUnit(cmd.ShooterID)
	if shooter == nil {
		return command.Result{}, fmt.Errorf("shooter unit %d not found", cmd.ShooterID)
	}
	target := g.GetUnit(cmd.TargetID)
	if target == nil {
		return command.Result{}, fmt.Errorf("target unit %d not found", cmd.TargetID)
	}
	if shooter.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("shooter unit %d does not belong to player %d", cmd.ShooterID, cmd.OwnerID)
	}
	if shooter.OwnerID == target.OwnerID {
		return command.Result{}, errors.New("cannot shoot friendly units")
	}
	if shooter.HasShot {
		return command.Result{}, fmt.Errorf("unit %d has already shot this turn", cmd.ShooterID)
	}
	if len(shooter.RangedWeapons()) == 0 {
		return command.Result{}, fmt.Errorf("unit %d has no ranged weapons", cmd.ShooterID)
	}

	// Check range for each ranged weapon
	dist := core.Distance(shooter.Position(), target.Position())
	for _, idx := range shooter.RangedWeapons() {
		if dist > float64(shooter.Weapons[idx].Range) {
			return command.Result{}, fmt.Errorf("target is out of range (%.1f\" > %d\")", dist, shooter.Weapons[idx].Range)
		}
	}

	results := ResolveShooting(g.Roller, shooter, target)
	shooter.HasShot = true

	totalDamage := 0
	totalSlain := 0
	for _, r := range results {
		totalDamage += r.DamageDealt
		totalSlain += r.ModelsSlain
		g.Logf("  %s", r.String())
	}

	desc := fmt.Sprintf("%s shot at %s: %d damage, %d models slain", shooter.Name, target.Name, totalDamage, totalSlain)
	g.Logf("%s", desc)
	return command.Result{Description: desc, Success: true}, nil
}

func (g *Game) executeFight(cmd *command.FightCommand) (command.Result, error) {
	attacker := g.GetUnit(cmd.AttackerID)
	if attacker == nil {
		return command.Result{}, fmt.Errorf("attacker unit %d not found", cmd.AttackerID)
	}
	target := g.GetUnit(cmd.TargetID)
	if target == nil {
		return command.Result{}, fmt.Errorf("target unit %d not found", cmd.TargetID)
	}
	if attacker.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("attacker unit %d does not belong to player %d", cmd.AttackerID, cmd.OwnerID)
	}
	if attacker.OwnerID == target.OwnerID {
		return command.Result{}, errors.New("cannot fight friendly units")
	}
	if attacker.HasFought {
		return command.Result{}, fmt.Errorf("unit %d has already fought this turn", cmd.AttackerID)
	}
	if len(attacker.MeleeWeapons()) == 0 {
		return command.Result{}, fmt.Errorf("unit %d has no melee weapons", cmd.AttackerID)
	}

	// Check melee range (3" engagement range in AoS)
	dist := core.Distance(attacker.Position(), target.Position())
	if dist > 3.0 {
		return command.Result{}, fmt.Errorf("target is out of melee range (%.1f\" > 3\")", dist)
	}

	results := ResolveCombat(g.Roller, attacker, target)
	attacker.HasFought = true

	totalDamage := 0
	totalSlain := 0
	for _, r := range results {
		totalDamage += r.DamageDealt
		totalSlain += r.ModelsSlain
		g.Logf("  %s", r.String())
	}

	desc := fmt.Sprintf("%s fought %s: %d damage, %d models slain", attacker.Name, target.Name, totalDamage, totalSlain)
	g.Logf("%s", desc)
	return command.Result{Description: desc, Success: true}, nil
}

func (g *Game) executeCharge(cmd *command.ChargeCommand) (command.Result, error) {
	charger := g.GetUnit(cmd.ChargerID)
	if charger == nil {
		return command.Result{}, fmt.Errorf("charger unit %d not found", cmd.ChargerID)
	}
	target := g.GetUnit(cmd.TargetID)
	if target == nil {
		return command.Result{}, fmt.Errorf("target unit %d not found", cmd.TargetID)
	}
	if charger.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("charger unit %d does not belong to player %d", cmd.ChargerID, cmd.OwnerID)
	}
	if charger.OwnerID == target.OwnerID {
		return command.Result{}, errors.New("cannot charge friendly units")
	}
	if charger.HasCharged {
		return command.Result{}, fmt.Errorf("unit %d has already charged this turn", cmd.ChargerID)
	}

	dist := core.Distance(charger.Position(), target.Position())

	// Must be within 12" to declare a charge
	if dist > 12.0 {
		return command.Result{}, fmt.Errorf("target is too far to charge (%.1f\" > 12\")", dist)
	}

	// Roll 2D6 for charge distance
	chargeRoll := g.Roller.Roll2D6()
	g.Logf("Charge roll: %d", chargeRoll)

	if float64(chargeRoll) < dist {
		charger.HasCharged = true
		desc := fmt.Sprintf("%s failed charge against %s (rolled %d, needed %.1f\")", charger.Name, target.Name, chargeRoll, dist)
		g.Logf("%s", desc)
		return command.Result{Description: desc, Success: false}, nil
	}

	// Move charger into engagement range (within 0.5" of target)
	newPos := charger.Position().Towards(target.Position(), dist-0.5)
	for i := range charger.Models {
		if charger.Models[i].IsAlive {
			charger.Models[i].Position = newPos
		}
	}
	charger.HasCharged = true

	desc := fmt.Sprintf("%s charged %s (rolled %d, needed %.1f\")", charger.Name, target.Name, chargeRoll, dist)
	g.Logf("%s", desc)
	return command.Result{Description: desc, Success: true}, nil
}

// ResetTurnFlags resets all unit action flags for a new battle round.
func (g *Game) ResetTurnFlags() {
	for _, u := range g.Units {
		u.ResetPhaseFlags()
	}
}

// CheckVictory checks if a player has won (opponent has no units left).
func (g *Game) CheckVictory() {
	if len(g.Players) < 2 {
		return
	}
	for _, p := range g.Players {
		units := g.UnitsForPlayer(p.ID())
		if len(units) == 0 {
			g.IsOver = true
			// The other player wins
			for _, other := range g.Players {
				if other.ID() != p.ID() {
					g.Winner = other.ID()
					g.Logf("Player %d (%s) wins! Player %d has no units remaining.", other.ID(), other.Name(), p.ID())
					return
				}
			}
		}
	}
}

// RunGame executes the main game loop for a given number of battle rounds.
func (g *Game) RunGame(maxRounds int) {
	if len(g.Players) < 2 {
		g.Logf("Need at least 2 players to start a game")
		return
	}

	phases := phase.StandardTurnSequence()

	for round := 1; round <= maxRounds; round++ {
		g.BattleRound = round
		g.ResetTurnFlags()
		g.Logf("=== BATTLE ROUND %d ===", round)

		for _, p := range phases {
			if g.IsOver {
				return
			}

			g.CurrentPhase = p.Type
			g.Logf("--- %s ---", p.Type)

			// Each player acts in each phase
			for playerIdx := range g.Players {
				if g.IsOver {
					return
				}

				g.ActivePlayer = playerIdx
				player := g.Players[playerIdx]
				g.Logf("Player %d (%s) turn", player.ID(), player.Name())

				// Get commands from player until they end the phase
				for {
					view := g.View(player.ID())
					cmd := player.GetNextCommand(view, p)

					if cmd == nil {
						break
					}

					// Check for end phase command
					if _, ok := cmd.(*command.EndPhaseCommand); ok {
						break
					}

					result, err := g.ExecuteCommand(cmd)
					if err != nil {
						g.Logf("Error: %s", err.Error())
						continue
					}
					g.Logf("%s", result.String())

					g.CheckVictory()
					if g.IsOver {
						return
					}
				}
			}
		}
	}

	if !g.IsOver {
		g.Logf("Game ended after %d battle rounds", maxRounds)
		g.IsOver = true
	}
}
