package game

import (
	"errors"
	"fmt"
	"math"

	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/commands"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
	"github.com/jruiznavarro/wargamestactics/pkg/dice"
)

// Game holds the entire game state.
type Game struct {
	Board          *board.Board
	Units          map[core.UnitID]*core.Unit
	Players        []Player
	Roller         *dice.Roller
	Rules          *rules.Engine
	Commands       *commands.CommandTracker
	BattleRound    int
	CurrentPhase   phase.PhaseType
	ActivePlayer   int // Index into Players slice
	PriorityPlayer int // Index of the player who won the priority roll this round
	NextUnitID     core.UnitID
	Log            []string
	IsOver         bool
	Winner         int // Player ID of winner, -1 if draw
	MaxBattleRounds int // Total battle rounds (default 5)

	// Scoring: VP tracked per player
	VictoryPoints map[int]int // playerID -> VP
	// Objective control: which player controls each objective (objectiveID -> playerID, -1 = uncontrolled)
	ObjectiveControl map[int]int

	// Per-turn magic tracking: spell names cast this turn per player (same spell once per turn unless Unlimited)
	SpellsCastThisTurn map[int]map[string]bool // playerID -> spell name -> cast
}

// NewGame creates a new game with the given seed and board dimensions.
func NewGame(seed int64, boardWidth, boardHeight float64) *Game {
	return &Game{
		Board:              board.NewBoard(boardWidth, boardHeight),
		Units:              make(map[core.UnitID]*core.Unit),
		Roller:             dice.NewRoller(seed),
		Rules:              rules.NewEngine(),
		Commands:           commands.NewCommandTracker(),
		NextUnitID:         1,
		Winner:             -1,
		MaxBattleRounds:    5,
		VictoryPoints:      make(map[int]int),
		ObjectiveControl:   make(map[int]int),
		SpellsCastThisTurn: make(map[int]map[string]bool),
	}
}

// AddPlayer adds a player to the game. Returns the player index.
func (g *Game) AddPlayer(p Player) int {
	g.Players = append(g.Players, p)
	return len(g.Players) - 1
}

// RegisterTerrainRules generates rules from all terrain on the board.
func (g *Game) RegisterTerrainRules() {
	for _, r := range board.TerrainRules(g.Board) {
		g.Rules.AddRule(r)
	}
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
			CurrentWounds: stats.Health,
			MaxWounds:     stats.Health,
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
				Name:      w.Name,
				Range:     w.Range,
				Attacks:   w.Attacks,
				ToHit:     w.ToHit,
				ToWound:   w.ToWound,
				Rend:      w.Rend,
				Damage:    w.Damage,
				Abilities: w.Abilities,
			})
		}

		var spellViews []SpellView
		for _, s := range u.Spells {
			spellViews = append(spellViews, SpellView{
				Name:         s.Name,
				CastingValue: s.CastingValue,
				Range:        s.Range,
			})
		}
		var prayerViews []PrayerView
		for _, p := range u.Prayers {
			prayerViews = append(prayerViews, PrayerView{
				Name:          p.Name,
				ChantingValue: p.ChantingValue,
				Range:         p.Range,
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
			WardSave:      u.WardSave,
			Weapons:       weaponViews,
			StrikeOrder:   u.StrikeOrder,
			HasMoved:      u.HasMoved,
			HasRun:        u.HasRun,
			HasRetreated:  u.HasRetreated,
			HasShot:       u.HasShot,
			HasFought:     u.HasFought,
			HasCharged:    u.HasCharged,
			HasPiledIn:    u.HasPiledIn,
			IsEngaged:     g.isEngaged(u),
			Spells:        spellViews,
			Prayers:       prayerViews,
			CanCast:       u.CanCast(),
			CanChant:      u.CanChant(),
		}
		unitsByOwner[u.OwnerID] = append(unitsByOwner[u.OwnerID], view)
	}

	var terrainViews []TerrainView
	for _, t := range g.Board.Terrain {
		terrainViews = append(terrainViews, TerrainView{
			Name:   t.Name,
			Type:   t.Type.String(),
			Symbol: t.Symbol(),
			Pos:    [2]float64{t.Pos.X, t.Pos.Y},
			Width:  t.Width,
			Height: t.Height,
		})
	}

	var objectiveViews []ObjectiveView
	for _, o := range g.Board.Objectives {
		controlledBy := -1
		if pid, ok := g.ObjectiveControl[o.ID]; ok {
			controlledBy = pid
		}
		objectiveViews = append(objectiveViews, ObjectiveView{
			ID:           o.ID,
			Position:     [2]float64{o.Position.X, o.Position.Y},
			Radius:       o.Radius,
			ControlledBy: controlledBy,
		})
	}

	cpMap := make(map[int]int)
	vpMap := make(map[int]int)
	for _, p := range g.Players {
		if state := g.Commands.GetState(p.ID()); state != nil {
			cpMap[p.ID()] = state.CommandPoints
		}
		vpMap[p.ID()] = g.VictoryPoints[p.ID()]
	}

	return &GameView{
		Units:           unitsByOwner,
		Terrain:         terrainViews,
		Objectives:      objectiveViews,
		BoardWidth:      g.Board.Width,
		BoardHeight:     g.Board.Height,
		CurrentPhase:    g.CurrentPhase,
		BattleRound:     g.BattleRound,
		MaxBattleRounds: g.MaxBattleRounds,
		ActivePlayer:    g.ActivePlayer,
		CommandPoints:   cpMap,
		VictoryPoints:   vpMap,
	}
}

// Logf adds a formatted log message.
func (g *Game) Logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	g.Log = append(g.Log, msg)
}

func (g *Game) logCombatResult(attackerName, defenderName string, r CombatResult) {
	g.Logf("    %s -> %s [%s]", attackerName, defenderName, r.WeaponName)
	g.Logf("      %d attacks --(hit)--> %d (%d crit) --(wound)--> %d --(save)--> %d unsaved",
		r.TotalAttacks, r.Hits, r.CriticalHits, r.Wounds, r.SavesFailed)
	if r.DamageDealt > 0 || r.MortalDealt > 0 {
		slainStr := ""
		if r.ModelsSlain > 0 {
			slainStr = fmt.Sprintf("  (%d models slain)", r.ModelsSlain)
		}
		wardStr := ""
		if r.WardSaved > 0 {
			wardStr = fmt.Sprintf(", %d warded", r.WardSaved)
		}
		mortalStr := ""
		if r.MortalDealt > 0 {
			mortalStr = fmt.Sprintf(", %d mortal", r.MortalDealt)
		}
		g.Logf("      => %d damage dealt%s%s%s", r.DamageDealt, mortalStr, wardStr, slainStr)
	} else {
		g.Logf("      => No damage")
	}
}

// ExecuteCommand validates and executes a command.
func (g *Game) ExecuteCommand(cmd interface{}) (command.Result, error) {
	switch c := cmd.(type) {
	case *command.MoveCommand:
		return g.executeMove(c)
	case *command.RunCommand:
		return g.executeRun(c)
	case *command.RetreatCommand:
		return g.executeRetreat(c)
	case *command.ShootCommand:
		return g.executeShoot(c)
	case *command.FightCommand:
		return g.executeFight(c)
	case *command.ChargeCommand:
		return g.executeCharge(c)
	case *command.PileInCommand:
		return g.executePileIn(c)
	case *command.CastCommand:
		return g.executeCast(c)
	case *command.ChantCommand:
		return g.executeChant(c)
	case *command.RallyCommand:
		return g.executeRally(c)
	case *command.MagicalInterventionCommand:
		return g.executeMagicalIntervention(c)
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
	// Cannot normal move if engaged -- must retreat (Rule 14.1)
	if g.isEngaged(unit) {
		return command.Result{}, fmt.Errorf("unit %d is engaged, must retreat to leave combat", cmd.UnitID)
	}
	if !g.Board.IsInBounds(cmd.Destination) {
		return command.Result{}, fmt.Errorf("destination (%.1f, %.1f) is out of bounds", cmd.Destination.X, cmd.Destination.Y)
	}

	origin := unit.Position()
	dist := core.Distance(origin, cmd.Destination)

	moveCtx := &rules.Context{
		Attacker:    unit,
		Origin:      origin,
		Destination: cmd.Destination,
		Distance:    dist,
	}
	g.Rules.Evaluate(rules.BeforeMove, moveCtx)

	if moveCtx.Blocked {
		return command.Result{}, fmt.Errorf("move blocked: %s", moveCtx.BlockMessage)
	}

	maxMove := float64(unit.Stats.Move + moveCtx.Modifiers.MoveMod)
	if maxMove < 0 {
		maxMove = 0
	}

	if !board.MoveDistanceValid(origin, cmd.Destination, maxMove) {
		return command.Result{}, fmt.Errorf("move distance %.1f exceeds maximum %.0f", dist, maxMove)
	}

	// Cannot end normal move within 3" of enemy (Rule 14.1)
	if g.wouldEngageEnemy(unit, cmd.Destination) {
		return command.Result{}, fmt.Errorf("cannot end normal move within 3\" of enemy unit")
	}

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

// executeRun: AoS4 Rule 14.1. Move + D6" extra. Cannot shoot or charge after.
func (g *Game) executeRun(cmd *command.RunCommand) (command.Result, error) {
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
	if g.isEngaged(unit) {
		return command.Result{}, fmt.Errorf("unit %d is engaged, cannot run", cmd.UnitID)
	}
	if !g.Board.IsInBounds(cmd.Destination) {
		return command.Result{}, fmt.Errorf("destination (%.1f, %.1f) is out of bounds", cmd.Destination.X, cmd.Destination.Y)
	}

	origin := unit.Position()
	dist := core.Distance(origin, cmd.Destination)

	moveCtx := &rules.Context{
		Attacker:    unit,
		Origin:      origin,
		Destination: cmd.Destination,
		Distance:    dist,
	}
	g.Rules.Evaluate(rules.BeforeMove, moveCtx)

	if moveCtx.Blocked {
		return command.Result{}, fmt.Errorf("run blocked: %s", moveCtx.BlockMessage)
	}

	runRoll := g.Roller.RollD6()
	maxMove := float64(unit.Stats.Move+moveCtx.Modifiers.MoveMod) + float64(runRoll)
	if maxMove < 0 {
		maxMove = 0
	}

	if !board.MoveDistanceValid(origin, cmd.Destination, maxMove) {
		return command.Result{}, fmt.Errorf("run distance %.1f exceeds maximum %.0f (Move %d + D6 roll %d)", dist, maxMove, unit.Stats.Move, runRoll)
	}

	if g.wouldEngageEnemy(unit, cmd.Destination) {
		return command.Result{}, fmt.Errorf("cannot end run within 3\" of enemy unit")
	}

	for i := range unit.Models {
		if unit.Models[i].IsAlive {
			unit.Models[i].Position = cmd.Destination
		}
	}
	unit.HasMoved = true
	unit.HasRun = true

	desc := fmt.Sprintf("%s ran %.1f\" to (%.1f, %.1f) (roll: %d)", unit.Name, dist, cmd.Destination.X, cmd.Destination.Y, runRoll)
	g.Logf("%s", desc)
	return command.Result{Description: desc, Success: true}, nil
}

// executeRetreat: AoS4 Rule 14.1. Move from combat, D3 mortal damage.
func (g *Game) executeRetreat(cmd *command.RetreatCommand) (command.Result, error) {
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
	if !g.isEngaged(unit) {
		return command.Result{}, fmt.Errorf("unit %d is not engaged, use normal move", cmd.UnitID)
	}
	if !g.Board.IsInBounds(cmd.Destination) {
		return command.Result{}, fmt.Errorf("destination (%.1f, %.1f) is out of bounds", cmd.Destination.X, cmd.Destination.Y)
	}

	origin := unit.Position()
	dist := core.Distance(origin, cmd.Destination)

	maxMove := float64(unit.Stats.Move)
	if !board.MoveDistanceValid(origin, cmd.Destination, maxMove) {
		return command.Result{}, fmt.Errorf("retreat distance %.1f exceeds maximum %.0f", dist, maxMove)
	}

	if g.wouldEngageEnemy(unit, cmd.Destination) {
		return command.Result{}, fmt.Errorf("cannot end retreat within 3\" of enemy unit")
	}

	// D3 mortal damage for retreating
	mortalDmg := g.Roller.RollD3()
	g.Logf("    %s retreats and suffers %d mortal damage", unit.Name, mortalDmg)
	ResolveMortalWounds(g.Roller, unit, mortalDmg)

	if unit.IsDestroyed() {
		desc := fmt.Sprintf("%s was destroyed while retreating!", unit.Name)
		g.Logf("    %s", desc)
		g.CheckVictory()
		return command.Result{Description: desc, Success: true}, nil
	}

	for i := range unit.Models {
		if unit.Models[i].IsAlive {
			unit.Models[i].Position = cmd.Destination
		}
	}
	unit.HasMoved = true
	unit.HasRetreated = true

	desc := fmt.Sprintf("%s retreated %.1f\" to (%.1f, %.1f)", unit.Name, dist, cmd.Destination.X, cmd.Destination.Y)
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
	if shooter.HasRun {
		return command.Result{}, fmt.Errorf("unit %d ran this turn and cannot shoot", cmd.ShooterID)
	}
	if shooter.HasRetreated {
		return command.Result{}, fmt.Errorf("unit %d retreated this turn and cannot shoot", cmd.ShooterID)
	}
	if len(shooter.RangedWeapons()) == 0 {
		return command.Result{}, fmt.Errorf("unit %d has no ranged weapons", cmd.ShooterID)
	}
	// Cannot shoot while engaged unless Shoot in Combat ability (Rule 15.0)
	if g.isEngaged(shooter) {
		hasShootInCombat := false
		for _, idx := range shooter.RangedWeapons() {
			if shooter.Weapons[idx].HasAbility(core.AbilityShootInCombat) {
				hasShootInCombat = true
				break
			}
		}
		if !hasShootInCombat {
			return command.Result{}, fmt.Errorf("unit %d is engaged and cannot shoot", cmd.ShooterID)
		}
	}

	dist := core.Distance(shooter.Position(), target.Position())
	shootCtx := &rules.Context{
		Attacker:   shooter,
		Defender:   target,
		Distance:   dist,
		IsShooting: true,
	}
	g.Rules.Evaluate(rules.BeforeShoot, shootCtx)

	if shootCtx.Blocked {
		return command.Result{}, fmt.Errorf("shooting blocked: %s", shootCtx.BlockMessage)
	}

	for _, idx := range shooter.RangedWeapons() {
		if dist > float64(shooter.Weapons[idx].Range) {
			return command.Result{}, fmt.Errorf("target is out of range (%.1f\" > %d\")", dist, shooter.Weapons[idx].Range)
		}
	}

	results := ResolveShooting(g.Roller, g.Rules, shooter, target)
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

	dist := core.Distance(attacker.Position(), target.Position())
	if dist > 3.0 {
		return command.Result{}, fmt.Errorf("target is out of melee range (%.1f\" > 3\")", dist)
	}

	results := ResolveCombat(g.Roller, g.Rules, attacker, target)
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
	if charger.HasRun {
		return command.Result{}, fmt.Errorf("unit %d ran this turn and cannot charge", cmd.ChargerID)
	}
	if charger.HasRetreated {
		return command.Result{}, fmt.Errorf("unit %d retreated this turn and cannot charge", cmd.ChargerID)
	}

	dist := core.Distance(charger.Position(), target.Position())
	if dist > 12.0 {
		return command.Result{}, fmt.Errorf("target is too far to charge (%.1f\" > 12\")", dist)
	}

	chargeCtx := &rules.Context{
		Attacker: charger,
		Defender: target,
		Origin:   charger.Position(),
		Distance: dist,
	}
	g.Rules.Evaluate(rules.BeforeCharge, chargeCtx)

	if chargeCtx.Blocked {
		charger.HasCharged = true
		return command.Result{}, fmt.Errorf("charge blocked: %s", chargeCtx.BlockMessage)
	}

	chargeRoll := g.Roller.Roll2D6() + chargeCtx.Modifiers.ChargeMod
	g.Logf("Charge roll: %d", chargeRoll)

	if float64(chargeRoll) < dist {
		charger.HasCharged = true
		desc := fmt.Sprintf("%s failed charge against %s (rolled %d, needed %.1f\")", charger.Name, target.Name, chargeRoll, dist)
		g.Logf("%s", desc)
		return command.Result{Description: desc, Success: false}, nil
	}

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

// ResetTurnFlags resets all unit action flags and per-turn spell tracking.
func (g *Game) ResetTurnFlags() {
	for _, u := range g.Units {
		u.ResetPhaseFlags()
	}
	g.SpellsCastThisTurn = make(map[int]map[string]bool)
}

// CheckVictory checks if a player has lost all units (immediate loss).
func (g *Game) CheckVictory() {
	if len(g.Players) < 2 {
		return
	}
	for _, p := range g.Players {
		units := g.UnitsForPlayer(p.ID())
		if len(units) == 0 {
			g.IsOver = true
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

// CalculateObjectiveControl updates which player controls each objective.
// Control score = sum of Control characteristics of all alive models in units contesting the objective.
// Ties broken by number of models contesting.
func (g *Game) CalculateObjectiveControl() {
	for _, obj := range g.Board.Objectives {
		type playerScore struct {
			controlScore int
			modelCount   int
		}
		scores := make(map[int]*playerScore) // playerID -> score

		for _, u := range g.Units {
			if u.IsDestroyed() {
				continue
			}
			if obj.IsContested(u.Position()) {
				if scores[u.OwnerID] == nil {
					scores[u.OwnerID] = &playerScore{}
				}
				alive := u.AliveModels()
				scores[u.OwnerID].controlScore += alive * u.Stats.Control
				scores[u.OwnerID].modelCount += alive
			}
		}

		bestPlayer := -1
		bestControl := 0
		bestModels := 0
		for pid, s := range scores {
			if s.controlScore > bestControl ||
				(s.controlScore == bestControl && s.modelCount > bestModels) {
				bestPlayer = pid
				bestControl = s.controlScore
				bestModels = s.modelCount
			}
		}
		g.ObjectiveControl[obj.ID] = bestPlayer
	}
}

// ObjectivesControlledBy returns the number of objectives controlled by a player.
func (g *Game) ObjectivesControlledBy(playerID int) int {
	count := 0
	for _, controllerID := range g.ObjectiveControl {
		if controllerID == playerID {
			count++
		}
	}
	return count
}

// ScoreEndOfTurn awards victory points for the active player at end of their turn.
// AoS4 scoring: 1 VP for controlling >= 1 objective, 1 VP for >= 2, 1 VP for controlling more than opponent.
func (g *Game) ScoreEndOfTurn(playerID int) int {
	g.CalculateObjectiveControl()

	scored := 0
	myObjectives := g.ObjectivesControlledBy(playerID)

	// 1 VP for controlling at least 1 objective
	if myObjectives >= 1 {
		scored++
		g.Logf("  +1 VP: %s controls at least 1 objective", g.playerName(playerID))
	}

	// 1 VP for controlling 2 or more objectives
	if myObjectives >= 2 {
		scored++
		g.Logf("  +1 VP: %s controls 2+ objectives", g.playerName(playerID))
	}

	// 1 VP for controlling more objectives than opponent
	for _, p := range g.Players {
		if p.ID() != playerID {
			opponentObjectives := g.ObjectivesControlledBy(p.ID())
			if myObjectives > opponentObjectives {
				scored++
				g.Logf("  +1 VP: %s controls more objectives than %s (%d vs %d)",
					g.playerName(playerID), g.playerName(p.ID()), myObjectives, opponentObjectives)
			}
			break // 2-player game
		}
	}

	g.VictoryPoints[playerID] += scored
	g.Logf("  %s scored %d VP this turn (total: %d)", g.playerName(playerID), scored, g.VictoryPoints[playerID])
	return scored
}

// CheckFinalVictory determines the winner after all battle rounds are complete.
func (g *Game) CheckFinalVictory() {
	if g.BattleRound < g.MaxBattleRounds {
		return
	}

	g.IsOver = true
	bestScore := -1
	bestPlayer := -1
	tied := false

	for _, p := range g.Players {
		vp := g.VictoryPoints[p.ID()]
		if vp > bestScore {
			bestScore = vp
			bestPlayer = p.ID()
			tied = false
		} else if vp == bestScore {
			tied = true
		}
	}

	if tied {
		g.Winner = -1 // Draw
		g.Logf("Game ends in a draw! Both players have %d VP.", bestScore)
	} else {
		g.Winner = bestPlayer
		g.Logf("%s wins with %d VP!", g.playerName(bestPlayer), bestScore)
	}
}

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

// wouldEngageEnemy returns true if moving to pos would bring the unit within 3" of any enemy.
func (g *Game) wouldEngageEnemy(u *core.Unit, pos core.Position) bool {
	for _, other := range g.Units {
		if other.OwnerID == u.OwnerID || other.IsDestroyed() {
			continue
		}
		if core.Distance(pos, other.Position()) <= 3.0 {
			return true
		}
	}
	return false
}

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

	pileInCtx := &rules.Context{
		Attacker:    unit,
		Origin:      origin,
		Destination: enemyPos,
		Distance:    distBefore,
	}
	g.Rules.Evaluate(rules.BeforePileIn, pileInCtx)

	if pileInCtx.Blocked {
		unit.HasPiledIn = true
		return command.Result{Description: fmt.Sprintf("%s: pile-in blocked: %s", unit.Name, pileInCtx.BlockMessage), Success: false}, nil
	}

	pileInDist := 3.0 + float64(pileInCtx.Modifiers.PileInMod)
	if distBefore <= pileInDist {
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

func (g *Game) runCombatSubPhase(turnPlayerIdx int, order core.StrikeOrder, p phase.Phase) {
	otherIdx := 1 - turnPlayerIdx
	playerOrder := [2]int{turnPlayerIdx, otherIdx}

	for {
		if g.IsOver {
			return
		}

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
			view := g.View(player.ID())
			cmd := player.GetNextCommand(view, p)

			if cmd == nil {
				continue
			}

			if _, ok := cmd.(*command.EndPhaseCommand); ok {
				g.autoFightUnit(eligible[0])
				anyActivated = true
				continue
			}

			if pileCmd, ok := cmd.(*command.PileInCommand); ok {
				result, err := g.executePileIn(pileCmd)
				if err != nil {
					g.Logf("    Error: %s", err.Error())
				} else {
					g.Logf("    %s", result.String())
				}
				unit := g.GetUnit(pileCmd.UnitID)
				if unit != nil && !unit.HasFought && !unit.IsDestroyed() {
					g.autoFightUnit(unit)
				}
				anyActivated = true
				continue
			}

			if fightCmd, ok := cmd.(*command.FightCommand); ok {
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

func (g *Game) autoFightUnit(unit *core.Unit) {
	if unit.IsDestroyed() || unit.HasFought {
		return
	}

	if !unit.HasPiledIn {
		g.executePileIn(&command.PileInCommand{
			OwnerID: unit.OwnerID,
			UnitID:  unit.ID,
		})
	}

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

func (g *Game) runPlayerTurn(playerIdx int) {
	phases := phase.StandardTurnSequence()
	player := g.Players[playerIdx]

	g.Logf("--- %s's Turn ---", player.Name())

	for _, p := range phases {
		if g.IsOver {
			return
		}

		g.CurrentPhase = p.Type
		g.Commands.ResetPhase()
		g.Logf("  -- %s --", p.Type)

		if p.Alternating {
			g.runAlternatingPhase(playerIdx, p)
		} else {
			g.runPlayerPhase(playerIdx, p)
		}

		// Clean up temporary rules from commands (All-out Attack/Defence)
		g.CleanupPhaseRules()
	}
}

// RunGame executes the main game loop.
func (g *Game) RunGame(maxRounds int) {
	if len(g.Players) < 2 {
		g.Logf("Need at least 2 players to start a game")
		return
	}

	for round := 1; round <= maxRounds; round++ {
		g.BattleRound = round
		g.Logf("=== BATTLE ROUND %d ===", round)

		first, second := g.rollOffPriority()
		g.PriorityPlayer = first

		// Initialize command points: 4 CP each, underdog gets +1
		// Underdog = player with fewer total wounds remaining (-1 = no underdog)
		underdogID := g.determineUnderdog()
		playerIDs := make([]int, len(g.Players))
		for i, p := range g.Players {
			playerIDs[i] = p.ID()
		}
		g.Commands.InitRound(playerIDs, 4, underdogID)

		if underdogID >= 0 {
			for _, p := range g.Players {
				if p.ID() == underdogID {
					g.Logf("  %s is the underdog (+1 CP)", p.Name())
				}
			}
		}
		for _, p := range g.Players {
			state := g.Commands.GetState(p.ID())
			g.Logf("  %s: %d CP", p.Name(), state.CommandPoints)
		}

		g.ResetTurnFlags()
		g.runPlayerTurn(first)
		if g.IsOver {
			return
		}

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

// determineUnderdog returns the player ID with fewer total wounds, or -1 if tied.
func (g *Game) determineUnderdog() int {
	if len(g.Players) < 2 {
		return -1
	}
	woundsPerPlayer := make(map[int]int)
	for _, u := range g.Units {
		if !u.IsDestroyed() {
			woundsPerPlayer[u.OwnerID] += u.TotalCurrentWounds()
		}
	}

	p0 := g.Players[0].ID()
	p1 := g.Players[1].ID()
	w0 := woundsPerPlayer[p0]
	w1 := woundsPerPlayer[p1]

	if w0 < w1 {
		return p0
	} else if w1 < w0 {
		return p1
	}
	return -1
}

// UseCommand validates and executes a command ability for a player.
func (g *Game) UseCommand(playerID int, cmdID commands.CommandID, unitID core.UnitID) error {
	state := g.Commands.GetState(playerID)
	if state == nil {
		return fmt.Errorf("no command state for player %d", playerID)
	}

	unit := g.GetUnit(unitID)
	if unit == nil {
		return fmt.Errorf("unit %d not found", unitID)
	}
	if unit.OwnerID != playerID {
		return fmt.Errorf("unit %d does not belong to player %d", unitID, playerID)
	}

	if err := state.Spend(cmdID, unitID); err != nil {
		return err
	}

	def := commands.Registry[cmdID]
	g.Logf("  [CMD] %s uses %s on %s (%d CP remaining)",
		g.playerName(playerID), def.Name, unit.Name, state.CommandPoints)

	return nil
}

// ExecuteRally performs the Rally command: 6D6, each 4+ = 1 rally point.
// Rally points can heal 1 wound (cost 1) or return a slain model (cost = Health stat).
func (g *Game) ExecuteRally(playerID int, unitID core.UnitID) (int, error) {
	unit := g.GetUnit(unitID)
	if unit == nil {
		return 0, fmt.Errorf("unit %d not found", unitID)
	}
	if g.isEngaged(unit) {
		return 0, fmt.Errorf("cannot rally a unit in combat")
	}

	if err := g.UseCommand(playerID, commands.CmdRally, unitID); err != nil {
		return 0, err
	}

	rallyPoints := 0
	for i := 0; i < 6; i++ {
		roll := g.Roller.RollD6()
		if roll >= 4 {
			rallyPoints++
		}
	}

	g.Logf("    Rally: %d points earned", rallyPoints)

	healed := 0
	remaining := rallyPoints

	// First try to return slain models
	for i := range unit.Models {
		if remaining <= 0 {
			break
		}
		if !unit.Models[i].IsAlive && remaining >= unit.Stats.Health {
			unit.Models[i].IsAlive = true
			unit.Models[i].CurrentWounds = unit.Stats.Health
			remaining -= unit.Stats.Health
			healed += unit.Stats.Health
			g.Logf("    Returned slain model (cost %d points)", unit.Stats.Health)
		}
	}

	// Then heal wounded models
	for i := range unit.Models {
		if remaining <= 0 {
			break
		}
		if unit.Models[i].IsAlive && unit.Models[i].CurrentWounds < unit.Models[i].MaxWounds {
			unit.Models[i].CurrentWounds++
			remaining--
			healed++
			g.Logf("    Healed 1 wound")
		}
	}

	return healed, nil
}

// ApplyAllOutAttack applies +1 hit modifier via the rules engine for one attack.
func (g *Game) ApplyAllOutAttack(playerID int, unitID core.UnitID) error {
	if err := g.UseCommand(playerID, commands.CmdAllOutAttack, unitID); err != nil {
		return err
	}

	unit := g.GetUnit(unitID)
	g.Rules.AddRule(rules.Rule{
		Name:    fmt.Sprintf("AllOutAttack_%d", unitID),
		Trigger: rules.BeforeHitRoll,
		Source:  rules.SourceGlobal,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Attacker != nil && ctx.Attacker.ID == unitID
		},
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.HitMod += 1
		},
	})

	g.Logf("    %s gains +1 to hit rolls this phase", unit.Name)
	return nil
}

// ApplyAllOutDefence applies +1 save modifier via the rules engine for one phase.
func (g *Game) ApplyAllOutDefence(playerID int, unitID core.UnitID) error {
	if err := g.UseCommand(playerID, commands.CmdAllOutDefence, unitID); err != nil {
		return err
	}

	unit := g.GetUnit(unitID)
	g.Rules.AddRule(rules.Rule{
		Name:    fmt.Sprintf("AllOutDefence_%d", unitID),
		Trigger: rules.BeforeSaveRoll,
		Source:  rules.SourceGlobal,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Defender != nil && ctx.Defender.ID == unitID
		},
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.SaveMod += 1
		},
	})

	g.Logf("    %s gains +1 to save rolls this phase", unit.Name)
	return nil
}

// ExecuteRedeploy performs the Redeploy command: unit moves up to D6" in enemy movement phase.
func (g *Game) ExecuteRedeploy(playerID int, unitID core.UnitID, destination core.Position) error {
	unit := g.GetUnit(unitID)
	if unit == nil {
		return fmt.Errorf("unit %d not found", unitID)
	}
	if g.isEngaged(unit) {
		return fmt.Errorf("cannot redeploy a unit in combat")
	}
	if !g.Board.IsInBounds(destination) {
		return fmt.Errorf("destination out of bounds")
	}

	if err := g.UseCommand(playerID, commands.CmdRedeploy, unitID); err != nil {
		return err
	}

	redeployDist := float64(g.Roller.RollD6())
	origin := unit.Position()
	dist := core.Distance(origin, destination)

	if dist > redeployDist {
		return fmt.Errorf("redeploy distance %.1f exceeds D6 roll %.0f", dist, redeployDist)
	}

	if g.wouldEngageEnemy(unit, destination) {
		return fmt.Errorf("cannot end redeploy within 3\" of enemy")
	}

	for i := range unit.Models {
		if unit.Models[i].IsAlive {
			unit.Models[i].Position = destination
		}
	}

	g.Logf("    %s redeployed %.1f\" (max %.0f\")", unit.Name, dist, redeployDist)
	return nil
}

// ExecuteForwardToVictory re-rolls a charge. Returns new charge roll.
func (g *Game) ExecuteForwardToVictory(playerID int, unitID core.UnitID) (int, error) {
	if err := g.UseCommand(playerID, commands.CmdForwardToVictory, unitID); err != nil {
		return 0, err
	}

	newRoll := g.Roller.Roll2D6()
	unit := g.GetUnit(unitID)
	g.Logf("    %s re-rolls charge: %d", unit.Name, newRoll)
	return newRoll, nil
}

// ExecutePowerThrough performs the Power Through command.
func (g *Game) ExecutePowerThrough(playerID int, unitID core.UnitID, targetID core.UnitID) error {
	unit := g.GetUnit(unitID)
	if unit == nil {
		return fmt.Errorf("unit %d not found", unitID)
	}
	if !unit.HasCharged {
		return fmt.Errorf("unit must have charged this turn")
	}

	target := g.GetUnit(targetID)
	if target == nil {
		return fmt.Errorf("target %d not found", targetID)
	}
	if target.Stats.Health >= unit.Stats.Health {
		return fmt.Errorf("target Health (%d) must be lower than unit Health (%d)",
			target.Stats.Health, unit.Stats.Health)
	}

	dist := core.Distance(unit.Position(), target.Position())
	if dist > 3.0 {
		return fmt.Errorf("target not in combat range")
	}

	if err := g.UseCommand(playerID, commands.CmdPowerThrough, unitID); err != nil {
		return err
	}

	mortalDmg := g.Roller.RollD3()
	g.Logf("    Power Through: %s deals %d mortal damage to %s", unit.Name, mortalDmg, target.Name)
	ResolveMortalWounds(g.Roller, target, mortalDmg)

	g.CheckVictory()
	return nil
}

// CleanupPhaseRules removes temporary rules added by commands (All-out Attack/Defence).
func (g *Game) CleanupPhaseRules() {
	g.Rules.RemoveRulesBySource(rules.SourceGlobal, "")
}

func (g *Game) playerName(playerID int) string {
	for _, p := range g.Players {
		if p.ID() == playerID {
			return p.Name()
		}
	}
	return fmt.Sprintf("Player %d", playerID)
}

// executeCast: AoS4 Rule 19.0. Wizard rolls 2D6 >= casting value.
// Miscast on double 1s: fail + D3 mortal + no more spells this phase.
// Same spell can only be cast once per turn per army (unless Unlimited).
// Unbind: enemy wizard within 30" rolls 2D6 > casting roll.
func (g *Game) executeCast(cmd *command.CastCommand) (command.Result, error) {
	caster := g.GetUnit(cmd.CasterID)
	if caster == nil {
		return command.Result{}, fmt.Errorf("caster unit %d not found", cmd.CasterID)
	}
	if caster.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("unit %d does not belong to player %d", cmd.CasterID, cmd.OwnerID)
	}
	if !caster.CanCast() {
		return command.Result{}, fmt.Errorf("unit %s cannot cast (not a wizard, miscast, or no casts remaining)", caster.Name)
	}
	if cmd.SpellIndex < 0 || cmd.SpellIndex >= len(caster.Spells) {
		return command.Result{}, fmt.Errorf("invalid spell index %d", cmd.SpellIndex)
	}

	spell := caster.Spells[cmd.SpellIndex]

	// Same spell once per turn per army (unless Unlimited)
	if !spell.Unlimited {
		if playerSpells, ok := g.SpellsCastThisTurn[caster.OwnerID]; ok {
			if playerSpells[spell.Name] {
				return command.Result{}, fmt.Errorf("%s has already been cast this turn (not Unlimited)", spell.Name)
			}
		}
	}

	target := g.GetUnit(cmd.TargetID)
	if target == nil {
		return command.Result{}, fmt.Errorf("target unit %d not found", cmd.TargetID)
	}

	// Validate target ownership vs spell type
	if spell.TargetFriendly && target.OwnerID != caster.OwnerID {
		return command.Result{}, fmt.Errorf("%s targets friendly units, but target belongs to enemy", spell.Name)
	}
	if !spell.TargetFriendly && target.OwnerID == caster.OwnerID {
		return command.Result{}, fmt.Errorf("%s targets enemy units, but target is friendly", spell.Name)
	}

	// Range check
	dist := core.Distance(caster.Position(), target.Position())
	if dist > float64(spell.Range) {
		return command.Result{}, fmt.Errorf("target is out of spell range (%.1f\" > %d\")", dist, spell.Range)
	}

	caster.CastCount++

	// Roll 2D6
	die1 := g.Roller.RollD6()
	die2 := g.Roller.RollD6()
	castingRoll := die1 + die2

	g.Logf("    %s casts %s: rolled %d+%d = %d (needs %d)",
		caster.Name, spell.Name, die1, die2, castingRoll, spell.CastingValue)

	// Miscast: double 1s = fail + D3 mortal + no more spells this phase
	if die1 == 1 && die2 == 1 {
		caster.HasMiscast = true
		mortalDmg := g.Roller.RollD3()
		g.Logf("    MISCAST! %s suffers %d mortal damage and cannot cast again this phase",
			caster.Name, mortalDmg)
		ResolveMortalWounds(g.Roller, caster, mortalDmg)
		g.CheckVictory()
		desc := fmt.Sprintf("%s miscast %s! %d mortal damage", caster.Name, spell.Name, mortalDmg)
		return command.Result{Description: desc, Success: false}, nil
	}

	if castingRoll < spell.CastingValue {
		desc := fmt.Sprintf("%s failed to cast %s (rolled %d, needed %d)", caster.Name, spell.Name, castingRoll, spell.CastingValue)
		g.Logf("    %s", desc)
		return command.Result{Description: desc, Success: false}, nil
	}

	// Unbind attempt: closest enemy wizard within 30" that hasn't used all unbinds
	if unbound := g.attemptUnbind(caster, castingRoll); unbound {
		desc := fmt.Sprintf("%s was unbound!", spell.Name)
		return command.Result{Description: desc, Success: false}, nil
	}

	// Track same-spell restriction
	if g.SpellsCastThisTurn[caster.OwnerID] == nil {
		g.SpellsCastThisTurn[caster.OwnerID] = make(map[string]bool)
	}
	g.SpellsCastThisTurn[caster.OwnerID][spell.Name] = true

	// Spell succeeds - apply effect
	return g.applySpellEffect(caster, target, &spell)
}

// attemptUnbind finds the closest enemy wizard within 30" and tries to unbind.
// A Wizard(X) can unbind X times per phase. Returns true if spell was unbound.
func (g *Game) attemptUnbind(caster *core.Unit, castingRoll int) bool {
	var bestWizard *core.Unit
	bestDist := math.MaxFloat64

	for _, u := range g.Units {
		if u.OwnerID == caster.OwnerID || u.IsDestroyed() {
			continue
		}
		if !u.CanUnbind() {
			continue
		}
		d := core.Distance(caster.Position(), u.Position())
		if d <= 30.0 && d < bestDist {
			bestDist = d
			bestWizard = u
		}
	}

	if bestWizard == nil {
		return false
	}

	bestWizard.UnbindCount++
	unbindRoll := g.Roller.Roll2D6()
	g.Logf("    %s attempts to unbind: rolled %d (needs > %d)",
		bestWizard.Name, unbindRoll, castingRoll)

	if unbindRoll > castingRoll {
		g.Logf("    Spell unbound by %s!", bestWizard.Name)
		return true
	}
	g.Logf("    Unbind failed")
	return false
}

// applySpellEffect resolves the effect of a successfully cast spell.
func (g *Game) applySpellEffect(caster *core.Unit, target *core.Unit, spell *core.Spell) (command.Result, error) {
	switch spell.Effect {
	case core.SpellEffectDamage:
		mortalDmg := g.Roller.RollD3()
		g.Logf("    %s deals %d mortal wounds to %s", spell.Name, mortalDmg, target.Name)
		ResolveMortalWounds(g.Roller, target, mortalDmg)
		g.CheckVictory()
		desc := fmt.Sprintf("%s cast %s on %s: %d mortal wounds", caster.Name, spell.Name, target.Name, mortalDmg)
		return command.Result{Description: desc, Success: true}, nil

	case core.SpellEffectHeal:
		healAmount := g.Roller.RollD3()
		healed := g.healUnit(target, healAmount)
		g.Logf("    %s heals %d wounds on %s", spell.Name, healed, target.Name)
		desc := fmt.Sprintf("%s cast %s on %s: healed %d wounds", caster.Name, spell.Name, target.Name, healed)
		return command.Result{Description: desc, Success: true}, nil

	case core.SpellEffectBuff:
		g.Rules.AddRule(rules.Rule{
			Name:    fmt.Sprintf("SpellBuff_%s_%d", spell.Name, target.ID),
			Trigger: rules.BeforeSaveRoll,
			Source:  rules.SourceGlobal,
			Condition: func(ctx *rules.Context) bool {
				return ctx.Defender != nil && ctx.Defender.ID == target.ID
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.SaveMod += spell.EffectValue
			},
		})
		g.Logf("    %s gains +%d to save rolls (%s)", target.Name, spell.EffectValue, spell.Name)
		desc := fmt.Sprintf("%s cast %s on %s: +%d save", caster.Name, spell.Name, target.Name, spell.EffectValue)
		return command.Result{Description: desc, Success: true}, nil

	default:
		return command.Result{}, fmt.Errorf("unknown spell effect type")
	}
}

// executeChant: AoS4 Rule 19.2. Priest rolls D6 for ritual points.
// Roll of 1: fail, remove D3 ritual points.
// Roll of 2+: choose to bank points (= roll) or spend (ritual points + roll vs ChantingValue).
func (g *Game) executeChant(cmd *command.ChantCommand) (command.Result, error) {
	chanter := g.GetUnit(cmd.ChanterID)
	if chanter == nil {
		return command.Result{}, fmt.Errorf("chanter unit %d not found", cmd.ChanterID)
	}
	if chanter.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("unit %d does not belong to player %d", cmd.ChanterID, cmd.OwnerID)
	}
	if !chanter.CanChant() {
		return command.Result{}, fmt.Errorf("unit %s cannot chant (not a priest or no chants remaining)", chanter.Name)
	}
	if cmd.PrayerIndex < 0 || cmd.PrayerIndex >= len(chanter.Prayers) {
		return command.Result{}, fmt.Errorf("invalid prayer index %d", cmd.PrayerIndex)
	}

	prayer := chanter.Prayers[cmd.PrayerIndex]
	chanter.ChantCount++

	// Roll D6
	chantRoll := g.Roller.RollD6()
	g.Logf("    %s chants %s: rolled %d (ritual points: %d)",
		chanter.Name, prayer.Name, chantRoll, chanter.RitualPoints)

	// Unmodified 1: fail and remove D3 ritual points
	if chantRoll == 1 {
		lost := g.Roller.RollD3()
		chanter.RitualPoints -= lost
		if chanter.RitualPoints < 0 {
			chanter.RitualPoints = 0
		}
		desc := fmt.Sprintf("%s failed chanting (rolled 1), lost %d ritual points (now %d)",
			chanter.Name, lost, chanter.RitualPoints)
		g.Logf("    %s", desc)
		return command.Result{Description: desc, Success: false}, nil
	}

	// Roll of 2+: bank or spend
	if cmd.BankPoints {
		// Bank: gain ritual points equal to roll
		chanter.RitualPoints += chantRoll
		desc := fmt.Sprintf("%s banks %d ritual points (now %d)",
			chanter.Name, chantRoll, chanter.RitualPoints)
		g.Logf("    %s", desc)
		return command.Result{Description: desc, Success: true}, nil
	}

	// Spend: add ritual points to roll and check vs ChantingValue
	total := chantRoll + chanter.RitualPoints
	g.Logf("    %s spends ritual points: %d (roll) + %d (ritual) = %d (needs %d)",
		chanter.Name, chantRoll, chanter.RitualPoints, total, prayer.ChantingValue)

	if total < prayer.ChantingValue {
		desc := fmt.Sprintf("%s failed to answer %s (total %d, needed %d)",
			chanter.Name, prayer.Name, total, prayer.ChantingValue)
		g.Logf("    %s", desc)
		// Ritual points are NOT consumed on failure - player chose to spend but didn't meet threshold
		// Actually per rules: "add the Priest's ritual points to the chanting roll" and if < CV, fail
		// The points are still spent (reset to 0) regardless
		chanter.RitualPoints = 0
		return command.Result{Description: desc, Success: false}, nil
	}

	// Prayer answered - reset ritual points
	chanter.RitualPoints = 0

	// Validate target for prayer effect
	target := g.GetUnit(cmd.TargetID)
	if target == nil {
		return command.Result{}, fmt.Errorf("target unit %d not found", cmd.TargetID)
	}
	if prayer.TargetFriendly && target.OwnerID != chanter.OwnerID {
		return command.Result{}, fmt.Errorf("%s targets friendly units, but target belongs to enemy", prayer.Name)
	}
	if !prayer.TargetFriendly && target.OwnerID == chanter.OwnerID {
		return command.Result{}, fmt.Errorf("%s targets enemy units, but target is friendly", prayer.Name)
	}

	dist := core.Distance(chanter.Position(), target.Position())
	if dist > float64(prayer.Range) {
		return command.Result{}, fmt.Errorf("target is out of prayer range (%.1f\" > %d\")", dist, prayer.Range)
	}

	g.Logf("    Prayer %s answered!", prayer.Name)
	return g.applyPrayerEffect(chanter, target, &prayer)
}

// applyPrayerEffect resolves the effect of a successfully answered prayer.
func (g *Game) applyPrayerEffect(chanter *core.Unit, target *core.Unit, prayer *core.Prayer) (command.Result, error) {
	switch prayer.Effect {
	case core.SpellEffectDamage:
		mortalDmg := g.Roller.RollD3()
		g.Logf("    %s deals %d mortal wounds to %s", prayer.Name, mortalDmg, target.Name)
		ResolveMortalWounds(g.Roller, target, mortalDmg)
		g.CheckVictory()
		desc := fmt.Sprintf("%s answered %s on %s: %d mortal wounds", chanter.Name, prayer.Name, target.Name, mortalDmg)
		return command.Result{Description: desc, Success: true}, nil

	case core.SpellEffectHeal:
		healAmount := g.Roller.RollD3()
		healed := g.healUnit(target, healAmount)
		g.Logf("    %s heals %d wounds on %s", prayer.Name, healed, target.Name)
		desc := fmt.Sprintf("%s answered %s on %s: healed %d wounds", chanter.Name, prayer.Name, target.Name, healed)
		return command.Result{Description: desc, Success: true}, nil

	case core.SpellEffectBuff:
		g.Rules.AddRule(rules.Rule{
			Name:    fmt.Sprintf("PrayerBuff_%s_%d", prayer.Name, target.ID),
			Trigger: rules.BeforeSaveRoll,
			Source:  rules.SourceGlobal,
			Condition: func(ctx *rules.Context) bool {
				return ctx.Defender != nil && ctx.Defender.ID == target.ID
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.SaveMod += prayer.EffectValue
			},
		})
		g.Logf("    %s gains +%d to save rolls (%s)", target.Name, prayer.EffectValue, prayer.Name)
		desc := fmt.Sprintf("%s answered %s on %s: +%d save", chanter.Name, prayer.Name, target.Name, prayer.EffectValue)
		return command.Result{Description: desc, Success: true}, nil

	default:
		return command.Result{}, fmt.Errorf("unknown prayer effect type")
	}
}

// healUnit applies healing to a unit, distributing across wounded models. Returns total healed.
func (g *Game) healUnit(target *core.Unit, amount int) int {
	healed := 0
	remaining := amount
	for i := range target.Models {
		if remaining <= 0 {
			break
		}
		if target.Models[i].IsAlive && target.Models[i].CurrentWounds < target.Models[i].MaxWounds {
			canHeal := target.Models[i].MaxWounds - target.Models[i].CurrentWounds
			if canHeal > remaining {
				canHeal = remaining
			}
			target.Models[i].CurrentWounds += canHeal
			remaining -= canHeal
			healed += canHeal
		}
	}
	return healed
}

// restoreModel returns a slain model to life with full wounds.
// Returns true if a model was restored, false if none could be.
func (g *Game) restoreModel(unit *core.Unit) bool {
	for i := range unit.Models {
		if !unit.Models[i].IsAlive {
			unit.Models[i].IsAlive = true
			unit.Models[i].CurrentWounds = unit.Models[i].MaxWounds
			return true
		}
	}
	return false
}

// executeRally: AoS4 Rule 20.0. Costs 1 CP.
// Pick a friendly unit not in combat. Roll 6D6: each 4+ = 1 rally point.
// Rally points heal 1 wound each, or can be accumulated to return a slain model
// (costs the model's Health characteristic worth of rally points).
func (g *Game) executeRally(cmd *command.RallyCommand) (command.Result, error) {
	unit := g.GetUnit(cmd.UnitID)
	if unit == nil {
		return command.Result{}, fmt.Errorf("unit %d not found", cmd.UnitID)
	}
	if unit.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("unit %d does not belong to player %d", cmd.UnitID, cmd.OwnerID)
	}
	if unit.IsDestroyed() {
		return command.Result{}, fmt.Errorf("unit %s is destroyed", unit.Name)
	}
	if g.isEngaged(unit) {
		return command.Result{}, fmt.Errorf("unit %s is in combat, cannot rally", unit.Name)
	}

	// Spend 1 CP
	state := g.Commands.GetState(cmd.OwnerID)
	if state == nil {
		return command.Result{}, fmt.Errorf("no command state for player %d", cmd.OwnerID)
	}
	if err := state.Spend(commands.CmdRally, cmd.UnitID); err != nil {
		return command.Result{}, fmt.Errorf("cannot use Rally: %w", err)
	}

	// Roll 6D6, count 4+
	rallyPoints := 0
	for i := 0; i < 6; i++ {
		roll := g.Roller.RollD6()
		if roll >= 4 {
			rallyPoints++
		}
	}

	g.Logf("    %s rallies: %d rally points", unit.Name, rallyPoints)

	// Spend rally points: try to return slain models first, then heal
	restored := 0
	healed := 0
	remaining := rallyPoints

	// Return slain models (costs Health characteristic per model)
	for remaining >= unit.Stats.Health && unit.AliveModels() < len(unit.Models) {
		if g.restoreModel(unit) {
			remaining -= unit.Stats.Health
			restored++
		} else {
			break
		}
	}

	// Heal remaining points
	if remaining > 0 {
		healed = g.healUnit(unit, remaining)
	}

	desc := fmt.Sprintf("%s rallied: %d points, %d models restored, %d wounds healed",
		unit.Name, rallyPoints, restored, healed)
	g.Logf("    %s", desc)
	return command.Result{Description: desc, Success: true}, nil
}

// executeMagicalIntervention: AoS4 Command. Costs 1 CP.
// A friendly Wizard or Priest uses a spell/prayer in the enemy's hero phase
// with -1 to the casting/chanting roll.
func (g *Game) executeMagicalIntervention(cmd *command.MagicalInterventionCommand) (command.Result, error) {
	caster := g.GetUnit(cmd.CasterID)
	if caster == nil {
		return command.Result{}, fmt.Errorf("caster unit %d not found", cmd.CasterID)
	}
	if caster.OwnerID != cmd.OwnerID {
		return command.Result{}, fmt.Errorf("unit %d does not belong to player %d", cmd.CasterID, cmd.OwnerID)
	}

	// Spend 1 CP
	state := g.Commands.GetState(cmd.OwnerID)
	if state == nil {
		return command.Result{}, fmt.Errorf("no command state for player %d", cmd.OwnerID)
	}
	if err := state.Spend(commands.CmdMagicalIntervention, cmd.CasterID); err != nil {
		return command.Result{}, fmt.Errorf("cannot use Magical Intervention: %w", err)
	}

	// Determine if using a spell or prayer
	if cmd.SpellIndex >= 0 {
		return g.executeMagicalInterventionSpell(cmd, caster)
	}
	if cmd.PrayerIndex >= 0 {
		return g.executeMagicalInterventionPrayer(cmd, caster)
	}
	return command.Result{}, fmt.Errorf("must specify either a spell or prayer index")
}

// executeMagicalInterventionSpell handles casting a spell via Magical Intervention (-1 penalty).
func (g *Game) executeMagicalInterventionSpell(cmd *command.MagicalInterventionCommand, caster *core.Unit) (command.Result, error) {
	if !caster.CanCast() {
		return command.Result{}, fmt.Errorf("unit %s cannot cast", caster.Name)
	}
	if cmd.SpellIndex < 0 || cmd.SpellIndex >= len(caster.Spells) {
		return command.Result{}, fmt.Errorf("invalid spell index %d", cmd.SpellIndex)
	}

	spell := caster.Spells[cmd.SpellIndex]

	// Same spell once per turn restriction
	if !spell.Unlimited {
		if playerSpells, ok := g.SpellsCastThisTurn[caster.OwnerID]; ok {
			if playerSpells[spell.Name] {
				return command.Result{}, fmt.Errorf("%s has already been cast this turn", spell.Name)
			}
		}
	}

	target := g.GetUnit(cmd.TargetID)
	if target == nil {
		return command.Result{}, fmt.Errorf("target unit %d not found", cmd.TargetID)
	}
	if spell.TargetFriendly && target.OwnerID != caster.OwnerID {
		return command.Result{}, fmt.Errorf("%s targets friendly units, but target belongs to enemy", spell.Name)
	}
	if !spell.TargetFriendly && target.OwnerID == caster.OwnerID {
		return command.Result{}, fmt.Errorf("%s targets enemy units, but target is friendly", spell.Name)
	}

	dist := core.Distance(caster.Position(), target.Position())
	if dist > float64(spell.Range) {
		return command.Result{}, fmt.Errorf("target out of range (%.1f\" > %d\")", dist, spell.Range)
	}

	caster.CastCount++

	// Roll 2D6 with -1 penalty for magical intervention
	die1 := g.Roller.RollD6()
	die2 := g.Roller.RollD6()
	castingRoll := die1 + die2 - 1 // -1 for magical intervention

	g.Logf("    %s (Magical Intervention) casts %s: rolled %d+%d-1 = %d (needs %d)",
		caster.Name, spell.Name, die1, die2, castingRoll, spell.CastingValue)

	// Miscast on natural double 1s (before modifier)
	if die1 == 1 && die2 == 1 {
		caster.HasMiscast = true
		mortalDmg := g.Roller.RollD3()
		g.Logf("    MISCAST! %s suffers %d mortal damage", caster.Name, mortalDmg)
		ResolveMortalWounds(g.Roller, caster, mortalDmg)
		g.CheckVictory()
		desc := fmt.Sprintf("%s miscast %s via Magical Intervention! %d mortal damage", caster.Name, spell.Name, mortalDmg)
		return command.Result{Description: desc, Success: false}, nil
	}

	if castingRoll < spell.CastingValue {
		desc := fmt.Sprintf("%s failed to cast %s via Magical Intervention (rolled %d, needed %d)",
			caster.Name, spell.Name, castingRoll, spell.CastingValue)
		return command.Result{Description: desc, Success: false}, nil
	}

	// Unbind attempt
	if unbound := g.attemptUnbind(caster, castingRoll); unbound {
		desc := fmt.Sprintf("%s was unbound!", spell.Name)
		return command.Result{Description: desc, Success: false}, nil
	}

	// Track same-spell restriction
	if g.SpellsCastThisTurn[caster.OwnerID] == nil {
		g.SpellsCastThisTurn[caster.OwnerID] = make(map[string]bool)
	}
	g.SpellsCastThisTurn[caster.OwnerID][spell.Name] = true

	return g.applySpellEffect(caster, target, &spell)
}

// executeMagicalInterventionPrayer handles chanting a prayer via Magical Intervention (-1 penalty).
func (g *Game) executeMagicalInterventionPrayer(cmd *command.MagicalInterventionCommand, chanter *core.Unit) (command.Result, error) {
	if !chanter.CanChant() {
		return command.Result{}, fmt.Errorf("unit %s cannot chant", chanter.Name)
	}
	if cmd.PrayerIndex < 0 || cmd.PrayerIndex >= len(chanter.Prayers) {
		return command.Result{}, fmt.Errorf("invalid prayer index %d", cmd.PrayerIndex)
	}

	prayer := chanter.Prayers[cmd.PrayerIndex]
	chanter.ChantCount++

	// Roll D6 with -1 penalty
	rawRoll := g.Roller.RollD6()
	chantRoll := rawRoll - 1 // -1 for magical intervention

	g.Logf("    %s (Magical Intervention) chants %s: rolled %d-1 = %d (ritual points: %d)",
		chanter.Name, prayer.Name, rawRoll, chantRoll, chanter.RitualPoints)

	// Natural 1 still fails (check raw roll)
	if rawRoll == 1 {
		lost := g.Roller.RollD3()
		chanter.RitualPoints -= lost
		if chanter.RitualPoints < 0 {
			chanter.RitualPoints = 0
		}
		desc := fmt.Sprintf("%s failed chanting via MI (rolled 1), lost %d ritual points (now %d)",
			chanter.Name, lost, chanter.RitualPoints)
		return command.Result{Description: desc, Success: false}, nil
	}

	if cmd.BankPoints {
		// Bank: gain ritual points equal to modified roll (min 1)
		banked := chantRoll
		if banked < 1 {
			banked = 1
		}
		chanter.RitualPoints += banked
		desc := fmt.Sprintf("%s banks %d ritual points via MI (now %d)",
			chanter.Name, banked, chanter.RitualPoints)
		return command.Result{Description: desc, Success: true}, nil
	}

	// Spend: ritual points + modified roll vs ChantingValue
	total := chantRoll + chanter.RitualPoints
	if total < prayer.ChantingValue {
		desc := fmt.Sprintf("%s failed %s via MI (total %d, needed %d)",
			chanter.Name, prayer.Name, total, prayer.ChantingValue)
		chanter.RitualPoints = 0
		return command.Result{Description: desc, Success: false}, nil
	}

	chanter.RitualPoints = 0

	target := g.GetUnit(cmd.TargetID)
	if target == nil {
		return command.Result{}, fmt.Errorf("target unit %d not found", cmd.TargetID)
	}
	if prayer.TargetFriendly && target.OwnerID != chanter.OwnerID {
		return command.Result{}, fmt.Errorf("%s targets friendly units, but target belongs to enemy", prayer.Name)
	}
	if !prayer.TargetFriendly && target.OwnerID == chanter.OwnerID {
		return command.Result{}, fmt.Errorf("%s targets enemy units, but target is friendly", prayer.Name)
	}

	dist := core.Distance(chanter.Position(), target.Position())
	if dist > float64(prayer.Range) {
		return command.Result{}, fmt.Errorf("target out of prayer range (%.1f\" > %d\")", dist, prayer.Range)
	}

	g.Logf("    Prayer %s answered via Magical Intervention!", prayer.Name)
	return g.applyPrayerEffect(chanter, target, &prayer)
}
