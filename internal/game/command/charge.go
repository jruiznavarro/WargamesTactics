package command

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// ChargeCommand declares a charge from one unit to another.
type ChargeCommand struct {
	OwnerID   int
	ChargerID core.UnitID
	TargetID  core.UnitID
}

func (c *ChargeCommand) Type() CommandType        { return CommandTypeCharge }
func (c *ChargeCommand) PlayerID() int            { return c.OwnerID }
func (c *ChargeCommand) GetChargerID() core.UnitID { return c.ChargerID }
func (c *ChargeCommand) GetTargetID() core.UnitID   { return c.TargetID }
