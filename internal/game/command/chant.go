package command

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// ChantCommand orders a Priest to chant a prayer at a target.
type ChantCommand struct {
	OwnerID     int
	ChanterID   core.UnitID
	PrayerIndex int         // Index into the unit's Prayers slice
	TargetID    core.UnitID // Target unit (friendly or enemy depending on prayer)
}

func (c *ChantCommand) Type() CommandType        { return CommandTypeChant }
func (c *ChantCommand) PlayerID() int            { return c.OwnerID }
func (c *ChantCommand) GetChanterID() core.UnitID { return c.ChanterID }
func (c *ChantCommand) GetTargetID() core.UnitID   { return c.TargetID }
