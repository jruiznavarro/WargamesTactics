package command

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// PileInCommand moves a unit up to 3" toward the nearest enemy model.
type PileInCommand struct {
	OwnerID int
	UnitID  core.UnitID
}

func (c *PileInCommand) Type() CommandType { return CommandTypePileIn }
func (c *PileInCommand) PlayerID() int     { return c.OwnerID }
