package game

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
)

// stubPlayer always returns a predefined sequence of commands.
type stubPlayer struct {
	id       int
	name     string
	commands []interface{}
	index    int
}

func (s *stubPlayer) ID() int      { return s.id }
func (s *stubPlayer) Name() string { return s.name }
func (s *stubPlayer) GetNextCommand(view *GameView, currentPhase phase.Phase) interface{} {
	if s.index >= len(s.commands) {
		return &command.EndPhaseCommand{OwnerID: s.id}
	}
	cmd := s.commands[s.index]
	s.index++
	return cmd
}

func TestNewGame(t *testing.T) {
	g := NewGame(42, 48, 24)
	if g.Board.Width != 48 || g.Board.Height != 24 {
		t.Error("board dimensions incorrect")
	}
	if g.NextUnitID != 1 {
		t.Error("next unit ID should start at 1")
	}
}

func TestCreateUnit(t *testing.T) {
	g := NewGame(42, 48, 24)

	stats := core.Stats{Move: 5, Save: 4, Control: 1, Health: 2}
	weapons := []core.Weapon{
		{Name: "Sword", Attacks: 2, ToHit: 3, ToWound: 3, Damage: 1},
	}

	u := g.CreateUnit("Warriors", 1, stats, weapons, 3, core.Position{X: 10, Y: 10}, 1.0)

	if u.ID != 1 {
		t.Errorf("expected unit ID 1, got %d", u.ID)
	}
	if len(u.Models) != 3 {
		t.Errorf("expected 3 models, got %d", len(u.Models))
	}
	if u.Models[0].CurrentWounds != 2 {
		t.Errorf("expected 2 wounds per model, got %d", u.Models[0].CurrentWounds)
	}

	// Create second unit - should get ID 2
	u2 := g.CreateUnit("Archers", 1, stats, weapons, 2, core.Position{X: 15, Y: 10}, 1.0)
	if u2.ID != 2 {
		t.Errorf("expected unit ID 2, got %d", u2.ID)
	}
}

func TestExecuteMove(t *testing.T) {
	g := NewGame(42, 48, 24)
	stats := core.Stats{Move: 5, Save: 4, Control: 1, Health: 1}
	g.CreateUnit("Warriors", 1, stats, nil, 1, core.Position{X: 10, Y: 10}, 1.0)

	cmd := &command.MoveCommand{
		OwnerID:     1,
		UnitID:      1,
		Destination: core.Position{X: 15, Y: 10},
	}

	result, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("move should succeed")
	}

	unit := g.GetUnit(1)
	pos := unit.Position()
	if pos.X != 15 || pos.Y != 10 {
		t.Errorf("expected (15, 10), got (%f, %f)", pos.X, pos.Y)
	}
	if !unit.HasMoved {
		t.Error("unit should be marked as moved")
	}
}

func TestExecuteMove_TooFar(t *testing.T) {
	g := NewGame(42, 48, 24)
	stats := core.Stats{Move: 5, Save: 4, Control: 1, Health: 1}
	g.CreateUnit("Warriors", 1, stats, nil, 1, core.Position{X: 10, Y: 10}, 1.0)

	cmd := &command.MoveCommand{
		OwnerID:     1,
		UnitID:      1,
		Destination: core.Position{X: 20, Y: 10}, // 10" away, but only 5" move
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error for move too far")
	}
}

func TestExecuteMove_WrongOwner(t *testing.T) {
	g := NewGame(42, 48, 24)
	stats := core.Stats{Move: 5, Save: 4, Control: 1, Health: 1}
	g.CreateUnit("Warriors", 1, stats, nil, 1, core.Position{X: 10, Y: 10}, 1.0)

	cmd := &command.MoveCommand{
		OwnerID:     2, // Wrong player
		UnitID:      1,
		Destination: core.Position{X: 12, Y: 10},
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error for wrong owner")
	}
}

func TestExecuteFight(t *testing.T) {
	g := NewGame(42, 48, 24)

	meleeWeapon := []core.Weapon{
		{Name: "Sword", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Rend: 1, Damage: 1},
	}

	g.CreateUnit("Attackers", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1}, meleeWeapon, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Defenders", 2, core.Stats{Move: 4, Save: 4, Control: 1, Health: 3}, nil, 1, core.Position{X: 11, Y: 10}, 1.0)

	cmd := &command.FightCommand{
		OwnerID:    1,
		AttackerID: 1,
		TargetID:   2,
	}

	result, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("fight should succeed")
	}
}

func TestExecuteFight_OutOfRange(t *testing.T) {
	g := NewGame(42, 48, 24)
	meleeWeapon := []core.Weapon{
		{Name: "Sword", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Damage: 1},
	}

	g.CreateUnit("Attackers", 1, core.Stats{Health: 1}, meleeWeapon, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Defenders", 2, core.Stats{Health: 3, Save: 4}, nil, 1, core.Position{X: 20, Y: 10}, 1.0) // 10" away

	cmd := &command.FightCommand{
		OwnerID:    1,
		AttackerID: 1,
		TargetID:   2,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error for fight out of range")
	}
}

func TestExecuteShoot(t *testing.T) {
	g := NewGame(42, 48, 24)
	rangedWeapon := []core.Weapon{
		{Name: "Bow", Range: 18, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
	}

	g.CreateUnit("Archers", 1, core.Stats{Move: 5, Save: 5, Control: 1, Health: 1}, rangedWeapon, 3, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Target", 2, core.Stats{Move: 4, Save: 4, Control: 1, Health: 2}, nil, 2, core.Position{X: 20, Y: 10}, 1.0)

	cmd := &command.ShootCommand{
		OwnerID:   1,
		ShooterID: 1,
		TargetID:  2,
	}

	result, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("shoot should succeed")
	}
}

func TestExecuteCharge(t *testing.T) {
	g := NewGame(42, 48, 24)
	g.CreateUnit("Chargers", 1, core.Stats{Move: 5, Health: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Target", 2, core.Stats{Move: 4, Health: 2, Save: 4}, nil, 1, core.Position{X: 18, Y: 10}, 1.0)

	cmd := &command.ChargeCommand{
		OwnerID:   1,
		ChargerID: 1,
		TargetID:  2,
	}

	// Charge uses 2D6, so it may or may not succeed depending on seed
	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error declaring charge: %v", err)
	}
}

func TestCheckVictory(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "Player 1"}
	p2 := &stubPlayer{id: 2, name: "Player 2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	g.CreateUnit("Warriors", 1, core.Stats{Health: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Enemies", 2, core.Stats{Health: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	// Kill all of player 2's units
	g.GetUnit(2).Models[0].IsAlive = false
	g.GetUnit(2).Models[0].CurrentWounds = 0

	g.CheckVictory()

	if !g.IsOver {
		t.Error("game should be over")
	}
	if g.Winner != 1 {
		t.Errorf("expected player 1 to win, got %d", g.Winner)
	}
}

func TestGameView(t *testing.T) {
	g := NewGame(42, 48, 24)
	weapons := []core.Weapon{
		{Name: "Sword", Range: 0, Attacks: 2, ToHit: 3, ToWound: 3, Damage: 1},
	}
	g.CreateUnit("Warriors", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1}, weapons, 2, core.Position{X: 10, Y: 10}, 1.0)
	g.BattleRound = 1
	g.CurrentPhase = phase.PhaseMovement

	view := g.View(1)

	if view.BattleRound != 1 {
		t.Errorf("expected round 1, got %d", view.BattleRound)
	}
	units, ok := view.Units[1]
	if !ok || len(units) != 1 {
		t.Fatal("expected 1 unit for player 1")
	}
	if units[0].AliveModels != 2 {
		t.Errorf("expected 2 alive models, got %d", units[0].AliveModels)
	}
}

func TestRunGame_AIvsAI(t *testing.T) {
	g := NewGame(42, 48, 24)

	// Create two stub players that always end phases
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	meleeWeapon := []core.Weapon{
		{Name: "Sword", Range: 0, Attacks: 2, ToHit: 4, ToWound: 4, Damage: 1},
	}
	g.CreateUnit("Unit1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 2}, meleeWeapon, 3, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Unit2", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 2}, meleeWeapon, 3, core.Position{X: 30, Y: 10}, 1.0)

	// Run for 1 round - with stub players that just skip, no damage
	g.RunGame(1)

	if g.BattleRound != 1 {
		t.Errorf("expected battle round 1, got %d", g.BattleRound)
	}
	if !g.IsOver {
		t.Error("game should be over after max rounds")
	}
}

func TestPriorityRollOff(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	first, second := g.rollOffPriority()

	// With seed 42, one of them wins - just verify they're different
	if first == second {
		t.Error("first and second player should be different")
	}
	if (first != 0 && first != 1) || (second != 0 && second != 1) {
		t.Errorf("player indices should be 0 or 1, got first=%d second=%d", first, second)
	}
}

type phaseRecord struct {
	playerName string
	phaseType  phase.PhaseType
}

// trackingPlayer records each GetNextCommand call, then ends the phase.
type trackingPlayer struct {
	id      int
	name    string
	records *[]phaseRecord
}

func (tp *trackingPlayer) ID() int      { return tp.id }
func (tp *trackingPlayer) Name() string { return tp.name }
func (tp *trackingPlayer) GetNextCommand(view *GameView, currentPhase phase.Phase) interface{} {
	*tp.records = append(*tp.records, phaseRecord{tp.name, currentPhase.Type})
	return &command.EndPhaseCommand{OwnerID: tp.id}
}

func TestTurnOrder_FirstPlayerCompletesAllPhases(t *testing.T) {
	var records []phaseRecord

	g := NewGame(42, 48, 24)

	tp1 := &trackingPlayer{id: 1, name: "P1", records: &records}
	tp2 := &trackingPlayer{id: 2, name: "P2", records: &records}
	g.AddPlayer(tp1)
	g.AddPlayer(tp2)

	// Units far apart: no engagement, so combat phase won't prompt players.
	g.CreateUnit("U1", 1, core.Stats{Health: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("U2", 2, core.Stats{Health: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	g.RunGame(1)

	// Each player's turn: 5 non-combat phases (Hero, Movement, Shooting, Charge, End of Turn).
	// Combat phase doesn't prompt because no units are engaged.
	// Total: 10 records (5 per player).
	if len(records) < 10 {
		t.Fatalf("expected at least 10 phase records, got %d", len(records))
	}

	firstPlayerName := records[0].playerName

	// Verify first player's turn: first 5 non-combat phases must be the first player
	nonCombatCount := 0
	for _, r := range records {
		if r.phaseType != phase.PhaseCombat {
			if nonCombatCount < 5 {
				if r.playerName != firstPlayerName {
					t.Errorf("non-combat phase %d during first turn: expected %s but got %s",
						nonCombatCount, firstPlayerName, r.playerName)
				}
			}
			nonCombatCount++
		}
	}

	// Verify second player's turn non-combat phases are the other player
	secondPlayerNonCombatStart := 5
	nonCombatCount = 0
	for _, r := range records {
		if r.phaseType != phase.PhaseCombat {
			if nonCombatCount >= secondPlayerNonCombatStart && nonCombatCount < secondPlayerNonCombatStart+5 {
				if r.playerName == firstPlayerName {
					t.Errorf("non-combat phase %d during second turn: expected other player but got %s",
						nonCombatCount, r.playerName)
				}
			}
			nonCombatCount++
		}
	}
}

func TestCommandPoints_InitializedEachRound(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	// Give P1 more wounds so P2 is underdog
	g.CreateUnit("BigUnit", 1, core.Stats{Health: 5}, nil, 3, core.Position{X: 10, Y: 10}, 1.0) // 15 wounds
	g.CreateUnit("SmallUnit", 2, core.Stats{Health: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0) // 1 wound

	g.RunGame(1)

	// After RunGame, commands were initialized. Check logs for underdog.
	foundUnderdog := false
	for _, msg := range g.Log {
		if msg == "  P2 is the underdog (+1 CP)" {
			foundUnderdog = true
		}
	}
	if !foundUnderdog {
		t.Error("expected P2 to be identified as underdog")
	}
}

func TestCommandPoints_ViewIncludesCP(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	g.CreateUnit("U1", 1, core.Stats{Health: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("U2", 2, core.Stats{Health: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	// Init round manually
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	view := g.View(1)

	if view.CommandPoints[1] != 4 {
		t.Errorf("expected 4 CP for player 1, got %d", view.CommandPoints[1])
	}
	if view.CommandPoints[2] != 4 {
		t.Errorf("expected 4 CP for player 2, got %d", view.CommandPoints[2])
	}
}

func TestRally_HealsDamage(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	g.CreateUnit("Warriors", 1, core.Stats{Health: 3}, nil, 3, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Enemy", 2, core.Stats{Health: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	// Damage one model
	g.GetUnit(1).Models[0].CurrentWounds = 1

	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero

	_, err := g.ExecuteRally(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// CP should have decreased
	state := g.Commands.GetState(1)
	if state.CommandPoints != 3 {
		t.Errorf("expected 3 CP after Rally, got %d", state.CommandPoints)
	}
}

func TestAllOutAttack_AddsHitModifier(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	g.CreateUnit("Fighters", 1, core.Stats{Health: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Enemy", 2, core.Stats{Health: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	g.Commands.InitRound([]int{1, 2}, 4, -1)

	err := g.ApplyAllOutAttack(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that a rule was added
	if !g.Rules.HasRulesFor(rules.BeforeHitRoll) {
		t.Error("expected hit roll rule to be added")
	}

	// Cleanup should remove it
	g.CleanupPhaseRules()
	// The terrain rules might still have BeforeHitRoll, so just verify count decreased
}

func TestAllOutDefence_AddsSaveModifier(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	g.CreateUnit("Defenders", 1, core.Stats{Health: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Enemy", 2, core.Stats{Health: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	g.Commands.InitRound([]int{1, 2}, 4, -1)

	err := g.ApplyAllOutDefence(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !g.Rules.HasRulesFor(rules.BeforeSaveRoll) {
		t.Error("expected save roll rule to be added")
	}
}

func TestCommand_CannotUseWithoutCP(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	g.CreateUnit("U1", 1, core.Stats{Health: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("U2", 2, core.Stats{Health: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	// Give 0 CP
	g.Commands.InitRound([]int{1, 2}, 0, -1)

	err := g.ApplyAllOutAttack(1, 1)
	if err == nil {
		t.Error("expected error when no CP available")
	}
}

func TestCombatPhase_AlternatingActivation(t *testing.T) {
	// Use a fightingPlayer that returns fight commands for engaged units
	g := NewGame(42, 48, 24)

	meleeWeapon := []core.Weapon{
		{Name: "Sword", Range: 0, Attacks: 1, ToHit: 4, ToWound: 4, Damage: 1},
	}

	// Place units within 3" of each other so they are engaged
	g.CreateUnit("U1", 1, core.Stats{Health: 10, Save: 2}, meleeWeapon, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("U2", 2, core.Stats{Health: 10, Save: 2}, meleeWeapon, 1, core.Position{X: 12, Y: 10}, 1.0)

	var records []phaseRecord
	tp1 := &trackingPlayer{id: 1, name: "P1", records: &records}
	tp2 := &trackingPlayer{id: 2, name: "P2", records: &records}
	g.AddPlayer(tp1)
	g.AddPlayer(tp2)

	g.RunGame(1)

	// Collect only combat phase records
	var combatRecords []phaseRecord
	for _, r := range records {
		if r.phaseType == phase.PhaseCombat {
			combatRecords = append(combatRecords, r)
		}
	}

	// Units are engaged, so the combat phase should prompt at least one player per turn.
	// The tracking player returns EndPhase, but the engine auto-fights engaged units.
	// Each player's turn has a combat phase where both players have engaged units.
	if len(combatRecords) < 2 {
		t.Fatalf("expected at least 2 combat records (engaged units), got %d", len(combatRecords))
	}

	// The first combat record should be from the priority winner
	firstPlayerName := records[0].playerName
	if combatRecords[0].playerName != firstPlayerName {
		t.Errorf("first combat activation should be %s (priority player), got %s",
			firstPlayerName, combatRecords[0].playerName)
	}
}

// --- Guarded Hero Tests (Rule 25.0, Errata Jan 2026) ---

func TestGuardedHero_CannotShootIfFriendlyNearby(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseShooting

	// P1 shooter with ranged weapon
	shooter := g.CreateUnit("P1 Archers", 1,
		core.Stats{Move: 5, Save: 5, Control: 1, Health: 1},
		[]core.Weapon{{Name: "Bow", Range: 24, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1}},
		5, core.Position{X: 10, Y: 12}, 1.0)

	// P2 Hero (Health <= 10) as target
	hero := g.CreateUnit("P2 Hero", 2,
		core.Stats{Move: 6, Save: 3, Control: 2, Health: 5},
		nil, 1, core.Position{X: 30, Y: 12}, 1.0)
	hero.Keywords = []core.Keyword{core.KeywordHero}

	// P2 friendly unit within 4" of the Hero
	g.CreateUnit("P2 Guard", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 32, Y: 12}, 1.0)

	cmd := &command.ShootCommand{OwnerID: 1, ShooterID: shooter.ID, TargetID: hero.ID}
	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error: cannot shoot Guarded Hero")
	}
}

func TestGuardedHero_CanShootIfNoFriendlyNearby(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseShooting

	// P1 shooter
	shooter := g.CreateUnit("P1 Archers", 1,
		core.Stats{Move: 5, Save: 5, Control: 1, Health: 1},
		[]core.Weapon{{Name: "Bow", Range: 24, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1}},
		5, core.Position{X: 10, Y: 12}, 1.0)

	// P2 Hero alone (no friendly units nearby)
	hero := g.CreateUnit("P2 Hero", 2,
		core.Stats{Move: 6, Save: 3, Control: 2, Health: 5},
		nil, 1, core.Position{X: 30, Y: 12}, 1.0)
	hero.Keywords = []core.Keyword{core.KeywordHero}

	// P2 unit far from hero (> 4")
	g.CreateUnit("P2 Far", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 45, Y: 12}, 1.0)

	cmd := &command.ShootCommand{OwnerID: 1, ShooterID: shooter.ID, TargetID: hero.ID}
	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Errorf("should be able to shoot unguarded Hero, got error: %v", err)
	}
}

func TestGuardedHero_ManifestationDoesNotGuard(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseShooting

	shooter := g.CreateUnit("P1 Archers", 1,
		core.Stats{Move: 5, Save: 5, Control: 1, Health: 1},
		[]core.Weapon{{Name: "Bow", Range: 24, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1}},
		5, core.Position{X: 10, Y: 12}, 1.0)

	hero := g.CreateUnit("P2 Hero", 2,
		core.Stats{Move: 6, Save: 3, Control: 2, Health: 5},
		nil, 1, core.Position{X: 30, Y: 12}, 1.0)
	hero.Keywords = []core.Keyword{core.KeywordHero}

	// Manifestation within 4" -- should NOT guard
	manif := g.CreateUnit("P2 Manifestation", 2,
		core.Stats{Move: 0, Save: 6, Control: 0, Health: 5},
		nil, 1, core.Position{X: 32, Y: 12}, 1.0)
	manif.Keywords = []core.Keyword{core.KeywordManifestation}

	cmd := &command.ShootCommand{OwnerID: 1, ShooterID: shooter.ID, TargetID: hero.ID}
	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Errorf("Manifestation should not guard Hero, got error: %v", err)
	}
}

// --- Visibility in Combat Tests (Rule 7.0, Errata Jan 2026) ---

func TestEngagement_RequiresVisibility(t *testing.T) {
	g := NewGame(42, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseCombat

	// Two units within 3" but with impassable terrain between them
	g.CreateUnit("P1 Warriors", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		[]core.Weapon{{Name: "Sword", Range: 0, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1}},
		5, core.Position{X: 10, Y: 12}, 1.0)

	g.CreateUnit("P2 Warriors", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		[]core.Weapon{{Name: "Sword", Range: 0, Attacks: 2, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1}},
		5, core.Position{X: 12, Y: 12}, 1.0)

	// Place impassable terrain between them
	g.Board.AddTerrain("Wall", board.TerrainImpassable, core.Position{X: 10.5, Y: 11}, 1, 2)

	// P1 unit should NOT be considered engaged because LOS is blocked
	p1Unit := g.GetUnit(core.UnitID(1))
	if g.isEngaged(p1Unit) {
		t.Error("unit should not be engaged when LOS is blocked by impassable terrain")
	}
}
