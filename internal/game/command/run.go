package command

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// RunCommand declares a run: normal move + D6" extra, but cannot shoot or charge.
// AoS4 Rule 14.1.
type RunCommand struct {
	OwnerID     int
	UnitID      core.UnitID
	Destination core.Position
}

func (c *RunCommand) Type() CommandType              { return CommandTypeRun }
func (c *RunCommand) PlayerID() int                  { return c.OwnerID }
func (c *RunCommand) GetUnitID() core.UnitID         { return c.UnitID }
func (c *RunCommand) GetDestination() core.Position  { return c.Destination }
