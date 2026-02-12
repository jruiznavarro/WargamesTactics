package command

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// MoveCommand moves a unit to a new position.
type MoveCommand struct {
	OwnerID     int
	UnitID      core.UnitID
	Destination core.Position
}

func (c *MoveCommand) Type() CommandType    { return CommandTypeMove }
func (c *MoveCommand) PlayerID() int        { return c.OwnerID }
func (c *MoveCommand) GetUnitID() core.UnitID { return c.UnitID }
func (c *MoveCommand) GetDestination() core.Position { return c.Destination }
