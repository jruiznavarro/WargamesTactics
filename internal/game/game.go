package game

import (
	"errors"
	"fmt"
	"math"

	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
	"github.com/jruiznavarro/wargamestactics/pkg/dice"
)

// Game holds the entire game state.
type Game struct {
	Board          *board.Board
	Units          map[core.UnitID]*core.Unit
	Players        []Player
	Roller         *dice.Roller
	BattleRound    int
	CurrentPhase   phase.PhaseType
	ActivePlayer   int // Index into Players slice
	PriorityPlayer int // Index of the player who won the priority roll this round
	NextUnitID     core.UnitID
	Log            []string
	IsOver         bool
	Winner         int // Player ID of winner, -1 if draw
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
			StrikeOrder:   u.StrikeOrder,
			HasMoved:      u.HasMoved,
			HasShot:       u.HasShot,
			HasFought:     u.HasFought,
			HasCharged:    u.HasCharged,
			HasPiledIn:    u.HasPiledIn,
			IsEngaged:     g.isEngaged(u),
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

// logCombatResult logs a visual representation of a combat resolution.
func (g *Game) logCombatResult(attackerName, defenderName string, r CombatResult) {
	g.Logf("    %s -> %s [%s]", attackerName, defenderName, r.WeaponName)
	g.Logf("      %d attacks --(hit)--> %d --(wound)--> %d --(save)--> %d unsaved",
		r.TotalAttacks, r.Hits, r.Wounds, r.SavesFailed)
	if r.DamageDealt > 0 {
		slainStr := ""
		if r.ModelsSlain > 0 {
			slainStr = fmt.Sprintf("  (%d models slain)", r.ModelsSlain)
		}
		g.Logf("      => %d damage dealt%s", r.DamageDealt, slainStr)
	} else {
		g.Logf("      => No damage")
	}
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
	case *command.PileInCommand:
		return g.executePileIn(c)
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
		g.logCombatResult(shooter.Name, target.Name, r)
	}

	desc := fmt.Sprintf("%s shot at %s: %d damage, %d models slain", shooter.Name, target.Name, totalDamage, totalSlain)
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
		g.logCombatResult(attacker.Name, target.Name, r)
	}

	desc := fmt.Sprintf("%s fought %s: %d damage, %d models slain", attacker.Name, target.Name, totalDamage, totalSlain)
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

// rollOffPriority rolls a die for each player and returns the player indices
// in activation order: [first, second]. The winner of the roll-off chooses
// who goes first (we assume the winner always picks themselves).
// On a tie, re-roll.
func (g *Game) rollOffPriority() (first, second int) {
	for {
		roll0 := g.Roller.RollD6()
		roll1 := g.Roller.RollD6()
		g.Logf("Priority roll: %s rolled %d, %s rolled %d",
			g.Players[0].Name(), roll0, g.Players[1].Name(), roll1)

		if roll0 > roll1 {
			g.Logf("%s wins priority and chooses to go first", g.Players[0].Name())
			return 0, 1
		} else if roll1 > roll0 {
			g.Logf("%s wins priority and chooses to go first", g.Players[1].Name())
			return 1, 0
		}
		g.Logf("Tie! Re-rolling priority...")
	}
}

// runPlayerPhase runs a single non-alternating phase for one player.
// The player issues commands until they end the phase.
func (g *Game) runPlayerPhase(playerIdx int, p phase.Phase) {
	player := g.Players[playerIdx]
	g.ActivePlayer = playerIdx

	for {
		view := g.View(player.ID())
		cmd := player.GetNextCommand(view, p)

		if cmd == nil {
			break
		}

		if _, ok := cmd.(*command.EndPhaseCommand); ok {
			break
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			g.Logf("    Error: %s", err.Error())
			continue
		}
		g.Logf("    %s", result.String())

		g.CheckVictory()
		if g.IsOver {
			return
		}
	}
}

// engagedUnits returns all non-destroyed units for a player that are within
// 3" engagement range of at least one enemy unit and haven't fought yet.
func (g *Game) engagedUnits(playerID int, strikeOrder core.StrikeOrder) []*core.Unit {
	var result []*core.Unit
	for _, u := range g.Units {
		if u.OwnerID != playerID || u.IsDestroyed() || u.HasFought {
			continue
		}
		if u.StrikeOrder != strikeOrder {
			continue
		}
		if g.isEngaged(u) {
			result = append(result, u)
		}
	}
	return result
}

// isEngaged returns true if the unit is within 3" of any enemy unit.
func (g *Game) isEngaged(u *core.Unit) bool {
	for _, other := range g.Units {
		if other.OwnerID == u.OwnerID || other.IsDestroyed() {
			continue
		}
		if core.Distance(u.Position(), other.Position()) <= 3.0 {
			return true
		}
	}
	return false
}

// nearestEnemyPos returns the position of the closest enemy unit to u.
func (g *Game) nearestEnemyPos(u *core.Unit) (core.Position, bool) {
	bestDist := math.MaxFloat64
	var bestPos core.Position
	found := false
	for _, other := range g.Units {
		if other.OwnerID == u.OwnerID || other.IsDestroyed() {
			continue
		}
		d := core.Distance(u.Position(), other.Position())
		if d < bestDist {
			bestDist = d
			bestPos = other.Position()
			found = true
		}
	}
	return bestPos, found
}

// executePileIn moves a unit up to 3" but it must end closer to the
// nearest enemy model than where it started.
func (g *Game) executePileIn(cmd *command.PileInCommand) (command.Result, error) {
	unit := g.GetUnit(cmd.UnitID)
	if unit == nil {
		return command.Result{}, fmt.Errorf("unit %d not found", cmd.UnitID)
	}
	if unit.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("unit %d does not belong to player %d", cmd.UnitID, cmd.OwnerID)
	}
	if unit.HasPiledIn {
		return command.Result{}, fmt.Errorf("unit %d has already piled in", cmd.UnitID)
	}

	enemyPos, found := g.nearestEnemyPos(unit)
	if !found {
		unit.HasPiledIn = true
		return command.Result{Description: fmt.Sprintf("%s: no enemy to pile in to", unit.Name), Success: true}, nil
	}

	origin := unit.Position()
	distBefore := core.Distance(origin, enemyPos)

	// Move up to 3" toward nearest enemy
	pileInDist := 3.0
	if distBefore <= pileInDist {
		// Already very close, move to within 0.5"
		pileInDist = distBefore - 0.5
		if pileInDist < 0 {
			pileInDist = 0
		}
	}

	if pileInDist <= 0 {
		unit.HasPiledIn = true
		return command.Result{Description: fmt.Sprintf("%s: already in base contact", unit.Name), Success: true}, nil
	}

	newPos := origin.Towards(enemyPos, pileInDist)
	distAfter := core.Distance(newPos, enemyPos)

	// Must end closer to the nearest enemy
	if distAfter >= distBefore {
		unit.HasPiledIn = true
		return command.Result{Description: fmt.Sprintf("%s: cannot pile in closer", unit.Name), Success: true}, nil
	}

	for i := range unit.Models {
		if unit.Models[i].IsAlive {
			unit.Models[i].Position = newPos
		}
	}
	unit.HasPiledIn = true

	moved := core.Distance(origin, newPos)
	desc := fmt.Sprintf("%s piled in %.1f\" toward enemy", unit.Name, moved)
	g.Logf("    %s", desc)
	return command.Result{Description: desc, Success: true}, nil
}

// runCombatSubPhase runs one sub-phase of combat (strike-first, normal, or strike-last).
// Both players alternate activations. The priority player picks first.
// Each activation: pile-in -> fight. Units in engagement range must fight.
func (g *Game) runCombatSubPhase(turnPlayerIdx int, order core.StrikeOrder, p phase.Phase) {
	otherIdx := 1 - turnPlayerIdx
	playerOrder := [2]int{turnPlayerIdx, otherIdx}

	for {
		if g.IsOver {
			return
		}

		// Check which players still have eligible units
		anyActivated := false

		for i := 0; i < 2; i++ {
			if g.IsOver {
				return
			}

			playerIdx := playerOrder[i]
			player := g.Players[playerIdx]
			eligible := g.engagedUnits(player.ID(), order)

			if len(eligible) == 0 {
				continue
			}

			g.ActivePlayer = playerIdx

			// Let the player choose which unit to activate
			view := g.View(player.ID())
			cmd := player.GetNextCommand(view, p)

			if cmd == nil {
				continue
			}

			if _, ok := cmd.(*command.EndPhaseCommand); ok {
				// Player passes this activation, but if they have engaged units
				// they'll need to fight eventually. For now, auto-fight the first eligible.
				g.autoFightUnit(eligible[0])
				anyActivated = true
				continue
			}

			// Handle pile-in command
			if pileCmd, ok := cmd.(*command.PileInCommand); ok {
				result, err := g.executePileIn(pileCmd)
				if err != nil {
					g.Logf("    Error: %s", err.Error())
				} else {
					g.Logf("    %s", result.String())
				}
				// After pile-in the player still needs to fight - auto-fight
				unit := g.GetUnit(pileCmd.UnitID)
				if unit != nil && !unit.HasFought && !unit.IsDestroyed() {
					g.autoFightUnit(unit)
				}
				anyActivated = true
				continue
			}

			// Handle fight command (player may skip pile-in)
			if fightCmd, ok := cmd.(*command.FightCommand); ok {
				// Auto pile-in first if not done
				unit := g.GetUnit(fightCmd.AttackerID)
				if unit != nil && !unit.HasPiledIn {
					g.executePileIn(&command.PileInCommand{
						OwnerID: fightCmd.OwnerID,
						UnitID:  fightCmd.AttackerID,
					})
				}
				result, err := g.ExecuteCommand(cmd)
				if err != nil {
					g.Logf("    Error: %s", err.Error())
				} else {
					g.Logf("    %s", result.String())
				}
				anyActivated = true

				g.CheckVictory()
				continue
			}

			// Any other command in combat phase
			result, err := g.ExecuteCommand(cmd)
			if err != nil {
				g.Logf("    Error: %s", err.Error())
			} else {
				g.Logf("    %s", result.String())
			}
			anyActivated = true

			g.CheckVictory()
		}

		if !anyActivated {
			break
		}
	}
}

// autoFightUnit performs a pile-in and fight for a unit automatically,
// targeting the nearest enemy in melee range.
func (g *Game) autoFightUnit(unit *core.Unit) {
	if unit.IsDestroyed() || unit.HasFought {
		return
	}

	// Pile-in if not done yet
	if !unit.HasPiledIn {
		g.executePileIn(&command.PileInCommand{
			OwnerID: unit.OwnerID,
			UnitID:  unit.ID,
		})
	}

	// Find nearest enemy within 3"
	var bestTarget *core.Unit
	bestDist := math.MaxFloat64
	for _, other := range g.Units {
		if other.OwnerID == unit.OwnerID || other.IsDestroyed() {
			continue
		}
		d := core.Distance(unit.Position(), other.Position())
		if d <= 3.0 && d < bestDist {
			bestDist = d
			bestTarget = other
		}
	}

	if bestTarget == nil || len(unit.MeleeWeapons()) == 0 {
		unit.HasFought = true
		return
	}

	fightCmd := &command.FightCommand{
		OwnerID:    unit.OwnerID,
		AttackerID: unit.ID,
		TargetID:   bestTarget.ID,
	}
	result, err := g.ExecuteCommand(fightCmd)
	if err != nil {
		g.Logf("    Error: %s", err.Error())
	} else {
		g.Logf("    %s", result.String())
	}
	g.CheckVictory()
}

// runAlternatingPhase runs the combat phase with 3 sub-phases:
// 1. Strike-first units (both players alternate, priority player first)
// 2. Normal units (both players alternate, priority player first)
// 3. Strike-last units (both players alternate, priority player first)
// Units in engagement range (3") must fight.
func (g *Game) runAlternatingPhase(turnPlayerIdx int, p phase.Phase) {
	subPhases := []struct {
		order core.StrikeOrder
		label string
	}{
		{core.StrikeFirst, "Strike First"},
		{core.StrikeNormal, ""},
		{core.StrikeLast, "Strike Last"},
	}

	for _, sp := range subPhases {
		if g.IsOver {
			return
		}

		// Check if any units exist for this sub-phase
		hasUnits := false
		for _, player := range g.Players {
			if len(g.engagedUnits(player.ID(), sp.order)) > 0 {
				hasUnits = true
				break
			}
		}
		if !hasUnits {
			continue
		}

		if sp.label != "" {
			g.Logf("    [%s]", sp.label)
		}

		g.runCombatSubPhase(turnPlayerIdx, sp.order, p)
	}
}

// runPlayerTurn executes all 6 phases for a single player's turn.
// Most phases are exclusive to the active player. Alternating phases
// (Combat) have both players taking turns picking units.
func (g *Game) runPlayerTurn(playerIdx int) {
	phases := phase.StandardTurnSequence()
	player := g.Players[playerIdx]

	g.Logf("--- %s's Turn ---", player.Name())

	for _, p := range phases {
		if g.IsOver {
			return
		}

		g.CurrentPhase = p.Type
		g.Logf("  -- %s --", p.Type)

		if p.Alternating {
			g.runAlternatingPhase(playerIdx, p)
		} else {
			g.runPlayerPhase(playerIdx, p)
		}
	}
}

// RunGame executes the main game loop for a given number of battle rounds.
// Each round: roll-off for priority, then first player does all 6 phases
// (with Combat being alternating), then second player does the same.
func (g *Game) RunGame(maxRounds int) {
	if len(g.Players) < 2 {
		g.Logf("Need at least 2 players to start a game")
		return
	}

	for round := 1; round <= maxRounds; round++ {
		g.BattleRound = round
		g.Logf("=== BATTLE ROUND %d ===", round)

		// Priority roll-off: winner chooses who goes first
		first, second := g.rollOffPriority()
		g.PriorityPlayer = first

		// First player's turn: all 6 phases
		g.ResetTurnFlags()
		g.runPlayerTurn(first)
		if g.IsOver {
			return
		}

		// Second player's turn: all 6 phases
		g.ResetTurnFlags()
		g.runPlayerTurn(second)
		if g.IsOver {
			return
		}
	}

	if !g.IsOver {
		g.Logf("Game ended after %d battle rounds", maxRounds)
		g.IsOver = true
	}
}
