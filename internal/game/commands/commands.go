package commands

import (
	"fmt"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// CommandID identifies a specific command ability.
type CommandID string

const (
	// Hero Phase
	CmdRally               CommandID = "rally"
	CmdMagicalIntervention CommandID = "magical_intervention"

	// Movement Phase (enemy)
	CmdRedeploy   CommandID = "redeploy"
	CmdAtTheDouble CommandID = "at_the_double"

	// Shooting Phase (enemy)
	CmdCoveringFire CommandID = "covering_fire"

	// Charge Phase
	CmdCounterCharge      CommandID = "counter_charge"
	CmdForwardToVictory   CommandID = "forward_to_victory"

	// Attack (Shooting & Combat)
	CmdAllOutAttack CommandID = "all_out_attack"

	// Defence
	CmdAllOutDefence CommandID = "all_out_defence"

	// End of Turn
	CmdPowerThrough CommandID = "power_through"
)

// CommandDef defines the properties of a command ability.
type CommandDef struct {
	ID          CommandID
	Name        string
	Cost        int            // Command point cost
	Phase       phase.PhaseType // Phase where it can be used
	IsReaction  bool           // True if it's a reaction to an opponent action
	YourTurn    bool           // True = your turn only, false = enemy turn only
	AnyTurn     bool           // True = usable in any turn
	Description string
}

// Registry holds all command definitions.
var Registry = map[CommandID]CommandDef{
	CmdRally: {
		ID: CmdRally, Name: "Rally", Cost: 1,
		Phase: phase.PhaseHero, YourTurn: true,
		Description: "Roll 6D6 for a non-combat unit. Each 4+ = 1 rally point. Spend to Heal(1) or return slain model.",
	},
	CmdMagicalIntervention: {
		ID: CmdMagicalIntervention, Name: "Magical Intervention", Cost: 1,
		Phase: phase.PhaseHero,
		Description: "Wizard/Priest uses spell/prayer in enemy hero phase (-1 to casting/chanting).",
	},
	CmdRedeploy: {
		ID: CmdRedeploy, Name: "Redeploy", Cost: 1,
		Phase: phase.PhaseMovement,
		Description: "Non-combat unit moves up to D6\" in enemy movement phase.",
	},
	CmdAtTheDouble: {
		ID: CmdAtTheDouble, Name: "At the Double", Cost: 1,
		Phase: phase.PhaseMovement, IsReaction: true, YourTurn: true,
		Description: "Reaction to Run: add 6\" instead of D6 to move.",
	},
	CmdCoveringFire: {
		ID: CmdCoveringFire, Name: "Covering Fire", Cost: 1,
		Phase: phase.PhaseShooting,
		Description: "Non-combat unit shoots closest enemy with -1 hit in enemy shooting phase.",
	},
	CmdCounterCharge: {
		ID: CmdCounterCharge, Name: "Counter-charge", Cost: 2,
		Phase: phase.PhaseCharging,
		Description: "Non-combat unit charges in enemy charge phase.",
	},
	CmdForwardToVictory: {
		ID: CmdForwardToVictory, Name: "Forward to Victory", Cost: 1,
		Phase: phase.PhaseCharging, IsReaction: true, YourTurn: true,
		Description: "Reaction to Charge: re-roll the charge roll.",
	},
	CmdAllOutAttack: {
		ID: CmdAllOutAttack, Name: "All-out Attack", Cost: 1,
		Phase: "", IsReaction: true, AnyTurn: true,
		Description: "Reaction to Attack: +1 to hit rolls for attacking unit.",
	},
	CmdAllOutDefence: {
		ID: CmdAllOutDefence, Name: "All-out Defence", Cost: 1,
		Phase: "", IsReaction: true, AnyTurn: true,
		Description: "Reaction to opponent Attack: +1 to save rolls for defending unit.",
	},
	CmdPowerThrough: {
		ID: CmdPowerThrough, Name: "Power Through", Cost: 1,
		Phase: phase.PhaseEndOfTurn, AnyTurn: true,
		Description: "Unit that charged deals D3 mortal to lower-Health enemy in combat, then moves up to Move.",
	},
}

// PlayerState tracks command point economy and usage for one player per round.
type PlayerState struct {
	PlayerID       int
	CommandPoints  int
	UsedThisPhase  map[CommandID]bool            // Each command used at most once per army per phase
	UnitUsedPhase  map[core.UnitID]bool          // Each unit used at most one command per phase
}

// NewPlayerState creates fresh command state for a player.
func NewPlayerState(playerID int, cp int) *PlayerState {
	return &PlayerState{
		PlayerID:      playerID,
		CommandPoints: cp,
		UsedThisPhase: make(map[CommandID]bool),
		UnitUsedPhase: make(map[core.UnitID]bool),
	}
}

// ResetPhase clears per-phase tracking (called at start of each phase).
func (ps *PlayerState) ResetPhase() {
	ps.UsedThisPhase = make(map[CommandID]bool)
	ps.UnitUsedPhase = make(map[core.UnitID]bool)
}

// CanUse checks if a command can be used by this player for a given unit.
func (ps *PlayerState) CanUse(cmdID CommandID, unitID core.UnitID) error {
	def, ok := Registry[cmdID]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmdID)
	}
	if ps.CommandPoints < def.Cost {
		return fmt.Errorf("not enough command points (%d available, %d required)", ps.CommandPoints, def.Cost)
	}
	if ps.UsedThisPhase[cmdID] {
		return fmt.Errorf("%s already used this phase", def.Name)
	}
	if ps.UnitUsedPhase[unitID] {
		return fmt.Errorf("unit already used a command this phase")
	}
	return nil
}

// Spend deducts CP and records usage.
func (ps *PlayerState) Spend(cmdID CommandID, unitID core.UnitID) error {
	if err := ps.CanUse(cmdID, unitID); err != nil {
		return err
	}
	def := Registry[cmdID]
	ps.CommandPoints -= def.Cost
	ps.UsedThisPhase[cmdID] = true
	ps.UnitUsedPhase[unitID] = true
	return nil
}

// CommandTracker manages command state for all players across a battle round.
type CommandTracker struct {
	States map[int]*PlayerState // Keyed by player ID
}

// NewCommandTracker creates a new tracker for the round.
func NewCommandTracker() *CommandTracker {
	return &CommandTracker{
		States: make(map[int]*PlayerState),
	}
}

// InitRound sets up command points for a new battle round.
// Each player gets baseCP (normally 4). The underdog gets +1.
func (ct *CommandTracker) InitRound(playerIDs []int, baseCP int, underdogID int) {
	for _, pid := range playerIDs {
		cp := baseCP
		if pid == underdogID {
			cp++
		}
		ct.States[pid] = NewPlayerState(pid, cp)
	}
}

// ResetPhase clears per-phase tracking for all players.
func (ct *CommandTracker) ResetPhase() {
	for _, ps := range ct.States {
		ps.ResetPhase()
	}
}

// GetState returns the command state for a player.
func (ct *CommandTracker) GetState(playerID int) *PlayerState {
	return ct.States[playerID]
}

// AvailableCommands returns command IDs a player can afford for a given unit in the current phase.
func (ct *CommandTracker) AvailableCommands(playerID int, unitID core.UnitID, currentPhase phase.PhaseType, isMyTurn bool) []CommandID {
	ps := ct.States[playerID]
	if ps == nil {
		return nil
	}

	var available []CommandID
	for id, def := range Registry {
		// Phase check (empty phase = any attack/defence phase)
		if def.Phase != "" && def.Phase != currentPhase {
			continue
		}

		// Turn ownership check
		if !def.AnyTurn {
			if def.YourTurn && !isMyTurn {
				continue
			}
			if !def.YourTurn && isMyTurn {
				continue
			}
		}

		if ps.CanUse(id, unitID) == nil {
			available = append(available, id)
		}
	}
	return available
}
