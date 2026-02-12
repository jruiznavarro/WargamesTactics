package command

// EndPhaseCommand signals that a player wants to end the current phase.
type EndPhaseCommand struct {
	OwnerID int
}

func (c *EndPhaseCommand) Type() CommandType { return CommandTypeEndPhase }
func (c *EndPhaseCommand) PlayerID() int     { return c.OwnerID }
