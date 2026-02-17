package command

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// RetreatCommand declares a retreat: move away from combat, suffers D3 mortal damage.
// AoS4 Rule 14.1.
type RetreatCommand struct {
	OwnerID     int
	UnitID      core.UnitID
	Destination core.Position
}

func (c *RetreatCommand) Type() CommandType              { return CommandTypeRetreat }
func (c *RetreatCommand) PlayerID() int                  { return c.OwnerID }
func (c *RetreatCommand) GetUnitID() core.UnitID         { return c.UnitID }
func (c *RetreatCommand) GetDestination() core.Position  { return c.Destination }
