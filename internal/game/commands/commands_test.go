package commands

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

func TestNewPlayerState(t *testing.T) {
	ps := NewPlayerState(1, 4)
	if ps.CommandPoints != 4 {
		t.Errorf("expected 4 CP, got %d", ps.CommandPoints)
	}
	if ps.PlayerID != 1 {
		t.Errorf("expected player ID 1, got %d", ps.PlayerID)
	}
}

func TestSpend_DeductsCPAndTracksUsage(t *testing.T) {
	ps := NewPlayerState(1, 4)

	err := ps.Spend(CmdRally, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ps.CommandPoints != 3 {
		t.Errorf("expected 3 CP after spending 1, got %d", ps.CommandPoints)
	}
}

func TestSpend_CannotAfford(t *testing.T) {
	ps := NewPlayerState(1, 1)

	// Counter-charge costs 2
	err := ps.Spend(CmdCounterCharge, 1)
	if err == nil {
		t.Error("expected error for insufficient CP")
	}
}

func TestSpend_CommandOncePerPhase(t *testing.T) {
	ps := NewPlayerState(1, 4)

	err := ps.Spend(CmdRally, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Same command on different unit -- should fail (once per army per phase)
	err = ps.Spend(CmdRally, 2)
	if err == nil {
		t.Error("expected error: same command used twice in one phase")
	}
}

func TestSpend_UnitOncePerPhase(t *testing.T) {
	ps := NewPlayerState(1, 4)

	err := ps.Spend(CmdRally, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Same unit, different command -- should fail (one command per unit per phase)
	err = ps.Spend(CmdAllOutAttack, 1)
	if err == nil {
		t.Error("expected error: unit already used a command this phase")
	}
}

func TestResetPhase_ClearsTracking(t *testing.T) {
	ps := NewPlayerState(1, 4)
	ps.Spend(CmdRally, 1)
	ps.ResetPhase()

	// After reset, Rally should be usable again (on a different unit)
	err := ps.CanUse(CmdRally, 2)
	if err != nil {
		t.Errorf("expected Rally to be available after phase reset: %v", err)
	}

	// CP should NOT be restored
	if ps.CommandPoints != 3 {
		t.Errorf("expected 3 CP (spent persists), got %d", ps.CommandPoints)
	}
}

func TestCommandTracker_InitRound(t *testing.T) {
	ct := NewCommandTracker()
	ct.InitRound([]int{1, 2}, 4, 2) // Player 2 is underdog

	if ct.GetState(1).CommandPoints != 4 {
		t.Errorf("expected player 1 to have 4 CP, got %d", ct.GetState(1).CommandPoints)
	}
	if ct.GetState(2).CommandPoints != 5 {
		t.Errorf("expected player 2 (underdog) to have 5 CP, got %d", ct.GetState(2).CommandPoints)
	}
}

func TestCommandTracker_ResetPhase(t *testing.T) {
	ct := NewCommandTracker()
	ct.InitRound([]int{1, 2}, 4, -1) // No underdog

	ct.GetState(1).Spend(CmdRally, 1)
	ct.ResetPhase()

	// After phase reset, commands are available again
	err := ct.GetState(1).CanUse(CmdRally, 2)
	if err != nil {
		t.Errorf("expected Rally available after phase reset: %v", err)
	}
}

func TestAvailableCommands_FiltersCorrectly(t *testing.T) {
	ct := NewCommandTracker()
	ct.InitRound([]int{1, 2}, 4, -1)

	// Player 1's turn, Hero Phase -- should include Rally
	available := ct.AvailableCommands(1, core.UnitID(1), phase.PhaseHero, true)

	hasRally := false
	for _, id := range available {
		if id == CmdRally {
			hasRally = true
		}
		// Redeploy is enemy movement phase, should NOT be here
		if id == CmdRedeploy {
			t.Error("Redeploy should not be available in hero phase")
		}
	}
	if !hasRally {
		t.Error("Rally should be available in own hero phase")
	}
}

func TestAvailableCommands_EnemyTurn(t *testing.T) {
	ct := NewCommandTracker()
	ct.InitRound([]int{1, 2}, 4, -1)

	// Player 2 during Player 1's movement phase
	available := ct.AvailableCommands(2, core.UnitID(3), phase.PhaseMovement, false)

	hasRedeploy := false
	hasAtTheDouble := false
	for _, id := range available {
		if id == CmdRedeploy {
			hasRedeploy = true
		}
		if id == CmdAtTheDouble {
			hasAtTheDouble = true
		}
	}
	if !hasRedeploy {
		t.Error("Redeploy should be available in enemy movement phase")
	}
	if hasAtTheDouble {
		t.Error("At the Double is your-turn only, should not be available in enemy turn")
	}
}

func TestCounterCharge_Costs2CP(t *testing.T) {
	ps := NewPlayerState(1, 2)

	err := ps.Spend(CmdCounterCharge, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ps.CommandPoints != 0 {
		t.Errorf("expected 0 CP after counter-charge (cost 2), got %d", ps.CommandPoints)
	}
}

func TestAllOutAttack_AnyTurn(t *testing.T) {
	ct := NewCommandTracker()
	ct.InitRound([]int{1, 2}, 4, -1)

	// All-out Attack should be available in combat phase regardless of whose turn it is
	available := ct.AvailableCommands(1, core.UnitID(1), phase.PhaseCombat, true)
	hasAOA := false
	for _, id := range available {
		if id == CmdAllOutAttack {
			hasAOA = true
		}
	}
	// All-out Attack has Phase="" (any phase with attacks), but we match empty phase == any
	if !hasAOA {
		t.Error("All-out Attack should be available in combat phase")
	}
}
