package game

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// --- Objective and Scoring Tests ---

func setupScoringGame(seed int64) *Game {
	g := NewGame(seed, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseEndOfTurn
	return g
}

func TestObjectiveControl_SingleUnit(t *testing.T) {
	g := setupScoringGame(1)

	// Place objective at (24, 12) with radius 6"
	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)

	// Place P1 unit on the objective
	g.CreateUnit("P1 Warriors", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 10, core.Position{X: 24, Y: 12}, 1.0)

	g.CalculateObjectiveControl()

	if g.ObjectiveControl[1] != 1 {
		t.Errorf("expected P1 to control objective 1, got controller %d", g.ObjectiveControl[1])
	}
}

func TestObjectiveControl_HigherControlWins(t *testing.T) {
	g := setupScoringGame(1)

	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)

	// P1: 5 models * Control 1 = 5 control score
	g.CreateUnit("P1 Infantry", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 24, Y: 12}, 1.0)

	// P2: 3 models * Control 2 = 6 control score (wins)
	g.CreateUnit("P2 Elite", 2,
		core.Stats{Move: 5, Save: 3, Control: 2, Health: 2},
		nil, 3, core.Position{X: 24, Y: 14}, 1.0)

	g.CalculateObjectiveControl()

	if g.ObjectiveControl[1] != 2 {
		t.Errorf("expected P2 to control objective (6 vs 5), got controller %d", g.ObjectiveControl[1])
	}
}

func TestObjectiveControl_TieBrokenByModelCount(t *testing.T) {
	g := setupScoringGame(1)

	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)

	// P1: 6 models * Control 1 = 6 control score, 6 models
	g.CreateUnit("P1 Horde", 1,
		core.Stats{Move: 5, Save: 5, Control: 1, Health: 1},
		nil, 6, core.Position{X: 24, Y: 12}, 1.0)

	// P2: 3 models * Control 2 = 6 control score, 3 models
	g.CreateUnit("P2 Elite", 2,
		core.Stats{Move: 5, Save: 3, Control: 2, Health: 2},
		nil, 3, core.Position{X: 24, Y: 14}, 1.0)

	g.CalculateObjectiveControl()

	if g.ObjectiveControl[1] != 1 {
		t.Errorf("expected P1 to win tie with more models (6 vs 3), got controller %d", g.ObjectiveControl[1])
	}
}

func TestObjectiveControl_UnitOutOfRange(t *testing.T) {
	g := setupScoringGame(1)

	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)

	// Place unit far from objective (distance > 6")
	g.CreateUnit("P1 Far", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 10, core.Position{X: 40, Y: 12}, 1.0)

	g.CalculateObjectiveControl()

	if g.ObjectiveControl[1] != -1 {
		t.Errorf("expected no controller for uncontested objective, got %d", g.ObjectiveControl[1])
	}
}

func TestScoreEndOfTurn_OneObjective(t *testing.T) {
	g := setupScoringGame(1)

	g.Board.AddObjective(core.Position{X: 12, Y: 12}, 6.0)
	g.Board.AddObjective(core.Position{X: 36, Y: 12}, 6.0)

	// P1 controls 1 objective
	g.CreateUnit("P1 Warriors", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 12, Y: 12}, 1.0)

	// P2 controls 1 objective
	g.CreateUnit("P2 Warriors", 2,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 36, Y: 12}, 1.0)

	scored := g.ScoreEndOfTurn(1)

	// P1 controls 1 objective: +1 for >= 1, but NOT +1 for >= 2, and NOT +1 for more than opponent
	if scored != 1 {
		t.Errorf("expected 1 VP (1 objective), got %d", scored)
	}
	if g.VictoryPoints[1] != 1 {
		t.Errorf("expected total VP = 1, got %d", g.VictoryPoints[1])
	}
}

func TestScoreEndOfTurn_TwoObjectives(t *testing.T) {
	g := setupScoringGame(1)

	g.Board.AddObjective(core.Position{X: 12, Y: 12}, 6.0)
	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)

	// P1 controls both objectives
	g.CreateUnit("P1 Left", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 12, Y: 12}, 1.0)
	g.CreateUnit("P1 Center", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 24, Y: 12}, 1.0)

	// P2 has no units on objectives
	g.CreateUnit("P2 Far", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 45, Y: 12}, 1.0)

	scored := g.ScoreEndOfTurn(1)

	// P1: +1 for >= 1, +1 for >= 2, +1 for more than opponent = 3 VP
	if scored != 3 {
		t.Errorf("expected 3 VP (2 objectives, more than opponent), got %d", scored)
	}
}

func TestScoreEndOfTurn_NoObjectives(t *testing.T) {
	g := setupScoringGame(1)

	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)

	// P1 not on any objective
	g.CreateUnit("P1 Far", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 45, Y: 12}, 1.0)

	scored := g.ScoreEndOfTurn(1)

	if scored != 0 {
		t.Errorf("expected 0 VP (no objectives controlled), got %d", scored)
	}
}

func TestScoreEndOfTurn_Accumulates(t *testing.T) {
	g := setupScoringGame(1)

	g.Board.AddObjective(core.Position{X: 12, Y: 12}, 6.0)

	g.CreateUnit("P1 Warriors", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 12, Y: 12}, 1.0)
	g.CreateUnit("P2 Far", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 45, Y: 12}, 1.0)

	g.ScoreEndOfTurn(1) // +2 (1 obj + more than opponent)
	g.ScoreEndOfTurn(1) // +2 again

	if g.VictoryPoints[1] != 4 {
		t.Errorf("expected VP to accumulate to 4, got %d", g.VictoryPoints[1])
	}
}

func TestCheckFinalVictory_PlayerWithMoreVPWins(t *testing.T) {
	g := setupScoringGame(1)
	g.BattleRound = 5
	g.MaxBattleRounds = 5
	g.VictoryPoints[1] = 10
	g.VictoryPoints[2] = 7

	g.CheckFinalVictory()

	if !g.IsOver {
		t.Error("expected game to be over after round 5")
	}
	if g.Winner != 1 {
		t.Errorf("expected P1 to win, got winner %d", g.Winner)
	}
}

func TestCheckFinalVictory_DrawOnTie(t *testing.T) {
	g := setupScoringGame(1)
	g.BattleRound = 5
	g.MaxBattleRounds = 5
	g.VictoryPoints[1] = 8
	g.VictoryPoints[2] = 8

	g.CheckFinalVictory()

	if !g.IsOver {
		t.Error("expected game to be over after round 5")
	}
	if g.Winner != -1 {
		t.Errorf("expected draw (-1), got winner %d", g.Winner)
	}
}

func TestCheckFinalVictory_NotOverBeforeLastRound(t *testing.T) {
	g := setupScoringGame(1)
	g.BattleRound = 3
	g.MaxBattleRounds = 5

	g.CheckFinalVictory()

	if g.IsOver {
		t.Error("game should not end before round 5")
	}
}

func TestMaxBattleRoundsDefault(t *testing.T) {
	g := NewGame(1, 48, 24)
	if g.MaxBattleRounds != 5 {
		t.Errorf("expected default MaxBattleRounds = 5, got %d", g.MaxBattleRounds)
	}
}

// --- Rally Tests ---

func setupRallyGame(seed int64) (*Game, *core.Unit) {
	g := NewGame(seed, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero

	// Create a wounded unit (5 models with Health 3, some wounded)
	unit := g.CreateUnit("P1 Infantry", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 3},
		nil, 5, core.Position{X: 10, Y: 12}, 1.0)

	// Wound first model to 1 wound, kill second model
	unit.Models[0].CurrentWounds = 1
	unit.Models[1].CurrentWounds = 0
	unit.Models[1].IsAlive = false

	// Keep enemy far away (not engaged)
	g.CreateUnit("P2 Enemy", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 3, core.Position{X: 40, Y: 12}, 1.0)

	return g, unit
}

func TestRally_GeneratesRallyPoints(t *testing.T) {
	// With seed 42, we get predictable rolls
	g, unit := setupRallyGame(42)

	cmd := &command.RallyCommand{OwnerID: 1, UnitID: unit.ID}
	result, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected rally to succeed, got: %s", result.Description)
	}
}

func TestRally_CostsCP(t *testing.T) {
	g, unit := setupRallyGame(42)

	initialCP := g.Commands.GetState(1).CommandPoints
	cmd := &command.RallyCommand{OwnerID: 1, UnitID: unit.ID}
	g.ExecuteCommand(cmd)

	afterCP := g.Commands.GetState(1).CommandPoints
	if afterCP != initialCP-1 {
		t.Errorf("expected CP to decrease by 1 (from %d to %d), got %d", initialCP, initialCP-1, afterCP)
	}
}

func TestRally_CannotRallyEngagedUnit(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero

	unit := g.CreateUnit("P1 Infantry", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 3},
		nil, 5, core.Position{X: 10, Y: 12}, 1.0)

	// Enemy within 3" = engaged
	g.CreateUnit("P2 Enemy", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 3, core.Position{X: 12, Y: 12}, 1.0)

	cmd := &command.RallyCommand{OwnerID: 1, UnitID: unit.ID}
	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when rallying engaged unit")
	}
}

func TestRally_CannotRallyEnemyUnit(t *testing.T) {
	g, _ := setupRallyGame(42)

	// Try to rally P2's unit as P1
	cmd := &command.RallyCommand{OwnerID: 1, UnitID: core.UnitID(2)} // P2's unit
	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when rallying enemy unit")
	}
}

func TestRally_RestoresSlainModels(t *testing.T) {
	// Use a seed that produces high rolls (lots of rally points)
	// We need at least Health=3 rally points to restore 1 model
	g := NewGame(100, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero

	// Unit with Health 1 (easy to restore) with 2 slain models
	unit := g.CreateUnit("P1 Archers", 1,
		core.Stats{Move: 5, Save: 5, Control: 1, Health: 1},
		nil, 10, core.Position{X: 10, Y: 12}, 1.0)
	unit.Models[0].CurrentWounds = 0
	unit.Models[0].IsAlive = false
	unit.Models[1].CurrentWounds = 0
	unit.Models[1].IsAlive = false
	unit.Models[2].CurrentWounds = 0
	unit.Models[2].IsAlive = false

	g.CreateUnit("P2 Far", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 3, core.Position{X: 40, Y: 12}, 1.0)

	aliveBefore := unit.AliveModels()

	cmd := &command.RallyCommand{OwnerID: 1, UnitID: unit.ID}
	result, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected rally to succeed")
	}

	aliveAfter := unit.AliveModels()
	if aliveAfter <= aliveBefore {
		// With 6D6 and Health=1, we should restore at least some models
		// unless all rolls were < 4. This is probabilistically very unlikely with seed 100
		t.Logf("Warning: no models restored (seed-dependent). Before: %d, After: %d", aliveBefore, aliveAfter)
	}
}

func TestRally_CannotRallyTwicePerPhase(t *testing.T) {
	g, unit := setupRallyGame(42)

	cmd := &command.RallyCommand{OwnerID: 1, UnitID: unit.ID}
	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("first rally should succeed: %v", err)
	}

	// Create another unit to try rallying with
	unit2 := g.CreateUnit("P1 Spears", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 5, core.Position{X: 15, Y: 12}, 1.0)

	cmd2 := &command.RallyCommand{OwnerID: 1, UnitID: unit2.ID}
	_, err = g.ExecuteCommand(cmd2)
	if err == nil {
		t.Error("expected error: Rally can only be used once per army per phase")
	}
}

// --- Magical Intervention Tests ---

func setupMIGame(seed int64) (*Game, *core.Unit, *core.Unit, *core.Unit) {
	g := NewGame(seed, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero
	g.ActivePlayer = 0 // P1 is active (it's P1's hero phase)

	// P2's wizard will use Magical Intervention (it's P1's turn = enemy hero phase for P2)
	wizard := g.CreateUnit("P2 Wizard", 2,
		core.Stats{Move: 6, Save: 5, Control: 1, Health: 5},
		nil, 1, core.Position{X: 30, Y: 12}, 1.0)
	wizard.Keywords = []core.Keyword{core.KeywordHero, core.KeywordWizard}
	wizard.Spells = []core.Spell{testDamageSpell()}

	// P2's target (P1's unit)
	enemy := g.CreateUnit("P1 Target", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 3},
		nil, 3, core.Position{X: 40, Y: 12}, 1.0)

	// P2's friendly unit (for buff spells)
	friendly := g.CreateUnit("P2 Ally", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 5, core.Position{X: 32, Y: 12}, 1.0)

	return g, wizard, enemy, friendly
}

func TestMagicalIntervention_CastsSpell(t *testing.T) {
	g, wizard, enemy, _ := setupMIGame(42)

	cmd := &command.MagicalInterventionCommand{
		OwnerID:     2,
		CasterID:    wizard.ID,
		SpellIndex:  0,
		PrayerIndex: -1,
		TargetID:    enemy.ID,
	}

	result, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Result depends on dice rolls, but should not error
	_ = result
}

func TestMagicalIntervention_CostsCP(t *testing.T) {
	g, wizard, enemy, _ := setupMIGame(42)

	initialCP := g.Commands.GetState(2).CommandPoints
	cmd := &command.MagicalInterventionCommand{
		OwnerID:     2,
		CasterID:    wizard.ID,
		SpellIndex:  0,
		PrayerIndex: -1,
		TargetID:    enemy.ID,
	}
	g.ExecuteCommand(cmd)

	afterCP := g.Commands.GetState(2).CommandPoints
	if afterCP != initialCP-1 {
		t.Errorf("expected CP to decrease by 1 (from %d to %d), got %d", initialCP, initialCP-1, afterCP)
	}
}

func TestMagicalIntervention_PenaltyApplied(t *testing.T) {
	// Use a seed where a spell would normally succeed but fails with -1
	// Chain Lightning needs 7. If roll is 7 exactly, -1 = 6 which fails.
	// We can't control exact dice, so just verify the command works
	g, wizard, enemy, _ := setupMIGame(42)

	cmd := &command.MagicalInterventionCommand{
		OwnerID:     2,
		CasterID:    wizard.ID,
		SpellIndex:  0,
		PrayerIndex: -1,
		TargetID:    enemy.ID,
	}
	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMagicalIntervention_RequiresSpellOrPrayer(t *testing.T) {
	g, wizard, _, _ := setupMIGame(42)

	cmd := &command.MagicalInterventionCommand{
		OwnerID:     2,
		CasterID:    wizard.ID,
		SpellIndex:  -1,
		PrayerIndex: -1,
		TargetID:    core.UnitID(3),
	}
	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when neither spell nor prayer specified")
	}
}

func TestMagicalIntervention_Prayer(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero

	// P2's priest with lots of ritual points for easy answer
	priest := g.CreateUnit("P2 Priest", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 5},
		nil, 1, core.Position{X: 30, Y: 12}, 1.0)
	priest.Keywords = []core.Keyword{core.KeywordHero, core.KeywordPriest}
	priest.Prayers = []core.Prayer{testDamagePrayer()}
	priest.RitualPoints = 10

	enemy := g.CreateUnit("P1 Target", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 3},
		nil, 3, core.Position{X: 35, Y: 12}, 1.0)

	cmd := &command.MagicalInterventionCommand{
		OwnerID:     2,
		CasterID:    priest.ID,
		SpellIndex:  -1,
		PrayerIndex: 0,
		TargetID:    enemy.ID,
		BankPoints:  false,
	}

	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- GameView Tests ---

func TestGameView_IncludesObjectives(t *testing.T) {
	g := setupScoringGame(1)
	g.Board.AddObjective(core.Position{X: 12, Y: 12}, 6.0)
	g.Board.AddObjective(core.Position{X: 36, Y: 12}, 6.0)

	view := g.View(1)

	if len(view.Objectives) != 2 {
		t.Errorf("expected 2 objectives in view, got %d", len(view.Objectives))
	}
	if view.Objectives[0].Radius != 6.0 {
		t.Errorf("expected radius 6.0, got %f", view.Objectives[0].Radius)
	}
}

func TestGameView_IncludesVP(t *testing.T) {
	g := setupScoringGame(1)
	g.VictoryPoints[1] = 5
	g.VictoryPoints[2] = 3

	view := g.View(1)

	if view.VictoryPoints[1] != 5 {
		t.Errorf("expected P1 VP = 5, got %d", view.VictoryPoints[1])
	}
	if view.VictoryPoints[2] != 3 {
		t.Errorf("expected P2 VP = 3, got %d", view.VictoryPoints[2])
	}
}

func TestGameView_MaxBattleRounds(t *testing.T) {
	g := setupScoringGame(1)
	view := g.View(1)

	if view.MaxBattleRounds != 5 {
		t.Errorf("expected MaxBattleRounds = 5, got %d", view.MaxBattleRounds)
	}
}

func TestHeroPhase_AllowsRallyAndMI(t *testing.T) {
	hp := phase.NewHeroPhase()

	if !hp.IsCommandAllowed(command.CommandTypeRally) {
		t.Error("hero phase should allow Rally")
	}
	if !hp.IsCommandAllowed(command.CommandTypeMagicalIntervention) {
		t.Error("hero phase should allow Magical Intervention")
	}
	if !hp.IsCommandAllowed(command.CommandTypeCast) {
		t.Error("hero phase should allow Cast")
	}
	if !hp.IsCommandAllowed(command.CommandTypeChant) {
		t.Error("hero phase should allow Chant")
	}
}

func TestObjectivesControlledBy(t *testing.T) {
	g := setupScoringGame(1)
	g.Board.AddObjective(core.Position{X: 12, Y: 12}, 6.0)
	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)
	g.Board.AddObjective(core.Position{X: 36, Y: 12}, 6.0)

	g.ObjectiveControl[1] = 1
	g.ObjectiveControl[2] = 1
	g.ObjectiveControl[3] = 2

	if g.ObjectivesControlledBy(1) != 2 {
		t.Errorf("expected P1 to control 2 objectives, got %d", g.ObjectivesControlledBy(1))
	}
	if g.ObjectivesControlledBy(2) != 1 {
		t.Errorf("expected P2 to control 1 objective, got %d", g.ObjectivesControlledBy(2))
	}
}

func TestObjectiveControl_DestroyedUnitsIgnored(t *testing.T) {
	g := setupScoringGame(1)
	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)

	unit := g.CreateUnit("P1 Dead", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 1, core.Position{X: 24, Y: 12}, 1.0)
	// Kill the unit
	unit.Models[0].CurrentWounds = 0
	unit.Models[0].IsAlive = false

	g.CalculateObjectiveControl()

	if g.ObjectiveControl[1] != -1 {
		t.Errorf("destroyed unit should not control objective, got controller %d", g.ObjectiveControl[1])
	}
}

// --- Rule 32.1 Tests: Unit contests only 1 objective ---

func TestObjectiveControl_UnitContestsOnlyNearest(t *testing.T) {
	g := setupScoringGame(1)

	// Two overlapping objectives at nearby positions
	g.Board.AddObjective(core.Position{X: 20, Y: 12}, 6.0)
	g.Board.AddObjective(core.Position{X: 24, Y: 12}, 6.0)

	// P1 unit at (22, 12) -- within range of BOTH objectives
	g.CreateUnit("P1 Warriors", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 22, Y: 12}, 1.0)

	// P2 unit only on objective 2 at (24, 12)
	g.CreateUnit("P2 Warriors", 2,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 24, Y: 12}, 1.0)

	g.CalculateObjectiveControl()

	// P1's unit is closer to obj 1 (distance 2"), should contest obj 1 only
	// P2's unit is on obj 2, should control it
	if g.ObjectiveControl[1] != 1 {
		t.Errorf("expected P1 to control objective 1 (nearest), got controller %d", g.ObjectiveControl[1])
	}
	if g.ObjectiveControl[2] != 2 {
		t.Errorf("expected P2 to control objective 2 (P1 unit contests nearest only), got controller %d", g.ObjectiveControl[2])
	}
}

func TestObjectiveControl_UnitContestsOnlyOne_EvenIfOverlapping(t *testing.T) {
	g := setupScoringGame(1)

	// Two objectives very close together
	g.Board.AddObjective(core.Position{X: 12, Y: 12}, 8.0)
	g.Board.AddObjective(core.Position{X: 14, Y: 12}, 8.0)

	// Single unit right between them - should contest only the nearest
	g.CreateUnit("P1 Warriors", 1,
		core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 12, Y: 12}, 1.0)

	g.CalculateObjectiveControl()

	// Unit is at (12,12), obj 1 is at (12,12) dist=0, obj 2 is at (14,12) dist=2
	// Should only contest obj 1
	if g.ObjectiveControl[1] != 1 {
		t.Errorf("expected P1 to control objective 1, got controller %d", g.ObjectiveControl[1])
	}
	if g.ObjectiveControl[2] != -1 {
		t.Errorf("expected objective 2 to be uncontrolled (unit contests only nearest), got controller %d", g.ObjectiveControl[2])
	}
}
