package command

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// FightCommand initiates melee combat between two units.
type FightCommand struct {
	OwnerID    int
	AttackerID core.UnitID
	TargetID   core.UnitID
}

func (c *FightCommand) Type() CommandType       { return CommandTypeFight }
func (c *FightCommand) PlayerID() int           { return c.OwnerID }
func (c *FightCommand) GetAttackerID() core.UnitID { return c.AttackerID }
func (c *FightCommand) GetTargetID() core.UnitID   { return c.TargetID }
