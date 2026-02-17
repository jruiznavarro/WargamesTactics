package phase

import "github.com/jruiznavarro/wargamestactics/internal/game/command"

// PhaseType identifies each phase in a battle round.
type PhaseType string

const (
	PhaseHero        PhaseType = "Hero Phase"
	PhaseMovement    PhaseType = "Movement Phase"
	PhaseCharging    PhaseType = "Charge Phase"
	PhaseShooting    PhaseType = "Shooting Phase"
	PhaseCombat    PhaseType = "Combat Phase"
	PhaseEndOfTurn PhaseType = "End of Turn"
)

// Phase defines the interface for a game phase.
type Phase struct {
	Type            PhaseType
	AllowedCommands []command.CommandType
	Alternating     bool // If true, both players alternate activations (e.g. Combat)
}

// NewHeroPhase creates the hero phase.
func NewHeroPhase() Phase {
	return Phase{
		Type: PhaseHero,
		AllowedCommands: []command.CommandType{
			command.CommandTypeCast,
			command.CommandTypeChant,
			command.CommandTypeRally,
			command.CommandTypeMagicalIntervention,
			command.CommandTypeEndPhase,
		},
	}
}

// NewMovementPhase creates the movement phase.
func NewMovementPhase() Phase {
	return Phase{
		Type: PhaseMovement,
		AllowedCommands: []command.CommandType{
			command.CommandTypeMove,
			command.CommandTypeRun,
			command.CommandTypeRetreat,
			command.CommandTypeEndPhase,
		},
	}
}

// NewShootingPhase creates the shooting phase.
func NewShootingPhase() Phase {
	return Phase{
		Type: PhaseShooting,
		AllowedCommands: []command.CommandType{
			command.CommandTypeShoot,
			command.CommandTypeEndPhase,
		},
	}
}

// NewChargePhase creates the charge phase.
func NewChargePhase() Phase {
	return Phase{
		Type: PhaseCharging,
		AllowedCommands: []command.CommandType{
			command.CommandTypeCharge,
			command.CommandTypeEndPhase,
		},
	}
}

// NewCombatPhase creates the combat phase.
// Combat is alternating: the priority player picks a unit to fight first,
// then the other player picks one, and so on.
func NewCombatPhase() Phase {
	return Phase{
		Type: PhaseCombat,
		AllowedCommands: []command.CommandType{
			command.CommandTypePileIn,
			command.CommandTypeFight,
			command.CommandTypeEndPhase,
		},
		Alternating: true,
	}
}

// NewEndOfTurnPhase creates the end of turn phase.
// Used for scoring objectives, triggering end-of-turn abilities, etc.
func NewEndOfTurnPhase() Phase {
	return Phase{
		Type: PhaseEndOfTurn,
		AllowedCommands: []command.CommandType{
			command.CommandTypeEndPhase,
		},
	}
}

// StandardTurnSequence returns the standard 6-phase turn sequence.
func StandardTurnSequence() []Phase {
	return []Phase{
		NewHeroPhase(),
		NewMovementPhase(),
		NewShootingPhase(),
		NewChargePhase(),
		NewCombatPhase(),
		NewEndOfTurnPhase(),
	}
}

// IsCommandAllowed checks if a command type is allowed in this phase.
func (p Phase) IsCommandAllowed(ct command.CommandType) bool {
	for _, allowed := range p.AllowedCommands {
		if allowed == ct {
			return true
		}
	}
	return false
}
