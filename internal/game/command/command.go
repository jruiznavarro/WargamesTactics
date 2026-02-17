package command

import "fmt"

// CommandType identifies the kind of command.
type CommandType string

const (
	CommandTypeMove     CommandType = "move"
	CommandTypeRun      CommandType = "run"
	CommandTypeRetreat  CommandType = "retreat"
	CommandTypeShoot    CommandType = "shoot"
	CommandTypeFight    CommandType = "fight"
	CommandTypeCharge   CommandType = "charge"
	CommandTypePileIn   CommandType = "pile_in"
	CommandTypeCast                CommandType = "cast"
	CommandTypeChant               CommandType = "chant"
	CommandTypeRally               CommandType = "rally"
	CommandTypeMagicalIntervention CommandType = "magical_intervention"
	CommandTypeEndPhase            CommandType = "end_phase"
)

// Result holds the outcome of an executed command.
type Result struct {
	Description string
	Success     bool
}

func (r Result) String() string {
	if r.Success {
		return fmt.Sprintf("OK: %s", r.Description)
	}
	return fmt.Sprintf("FAILED: %s", r.Description)
}
