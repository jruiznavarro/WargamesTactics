package game

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
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

	stats := core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 2}
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
	stats := core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 1}
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
	stats := core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 1}
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
	stats := core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 1}
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
		{Name: "Sword", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Rend: -1, Damage: 1},
	}

	g.CreateUnit("Attackers", 1, core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 1}, meleeWeapon, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Defenders", 2, core.Stats{Move: 4, Save: 4, Bravery: 6, Wounds: 3}, nil, 1, core.Position{X: 11, Y: 10}, 1.0)

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

	g.CreateUnit("Attackers", 1, core.Stats{Wounds: 1}, meleeWeapon, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Defenders", 2, core.Stats{Wounds: 3, Save: 4}, nil, 1, core.Position{X: 20, Y: 10}, 1.0) // 10" away

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

	g.CreateUnit("Archers", 1, core.Stats{Move: 5, Save: 5, Bravery: 6, Wounds: 1}, rangedWeapon, 3, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Target", 2, core.Stats{Move: 4, Save: 4, Bravery: 6, Wounds: 2}, nil, 2, core.Position{X: 20, Y: 10}, 1.0)

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
	g.CreateUnit("Chargers", 1, core.Stats{Move: 5, Wounds: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Target", 2, core.Stats{Move: 4, Wounds: 2, Save: 4}, nil, 1, core.Position{X: 18, Y: 10}, 1.0)

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

	g.CreateUnit("Warriors", 1, core.Stats{Wounds: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Enemies", 2, core.Stats{Wounds: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

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
	g.CreateUnit("Warriors", 1, core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 1}, weapons, 2, core.Position{X: 10, Y: 10}, 1.0)
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
	g.CreateUnit("Unit1", 1, core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 2}, meleeWeapon, 3, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("Unit2", 2, core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 2}, meleeWeapon, 3, core.Position{X: 30, Y: 10}, 1.0)

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

	g.CreateUnit("U1", 1, core.Stats{Wounds: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("U2", 2, core.Stats{Wounds: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	g.RunGame(1)

	// With alternating combat, the first player's turn produces:
	// 5 solo phases (Hero, Movement, Shooting, Charge, Battleshock) for first player
	// + 2 combat records (first player then second player, both end immediately)
	// Then the second player's turn produces the same pattern.
	// Total: at least 14 records.

	if len(records) < 14 {
		t.Fatalf("expected at least 14 phase records, got %d", len(records))
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

func TestCombatPhase_AlternatingActivation(t *testing.T) {
	var records []phaseRecord

	g := NewGame(42, 48, 24)

	tp1 := &trackingPlayer{id: 1, name: "P1", records: &records}
	tp2 := &trackingPlayer{id: 2, name: "P2", records: &records}
	g.AddPlayer(tp1)
	g.AddPlayer(tp2)

	g.CreateUnit("U1", 1, core.Stats{Wounds: 1}, nil, 1, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("U2", 2, core.Stats{Wounds: 1}, nil, 1, core.Position{X: 30, Y: 10}, 1.0)

	g.RunGame(1)

	// Collect only combat phase records
	var combatRecords []phaseRecord
	for _, r := range records {
		if r.phaseType == phase.PhaseCombat {
			combatRecords = append(combatRecords, r)
		}
	}

	// Each combat phase should have both players participating
	// (tracking players end immediately, so we get one record per player per combat phase)
	// There are 2 combat phases per round (one in each player's turn)
	if len(combatRecords) < 4 {
		t.Fatalf("expected at least 4 combat records (2 players x 2 turns), got %d", len(combatRecords))
	}

	// In each combat phase, both players should appear
	// First combat phase (first player's turn): first player picks first
	firstPlayerName := records[0].playerName // whoever won priority
	if combatRecords[0].playerName != firstPlayerName {
		t.Errorf("first combat activation should be %s (active player), got %s",
			firstPlayerName, combatRecords[0].playerName)
	}
	if combatRecords[1].playerName == combatRecords[0].playerName {
		t.Errorf("second combat activation should be the other player, got %s again",
			combatRecords[1].playerName)
	}
}
