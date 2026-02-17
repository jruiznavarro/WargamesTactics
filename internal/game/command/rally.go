package command

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// RallyCommand orders a unit to rally during the hero phase.
// AoS4 Rule 20.0: Costs 1 CP. Pick a friendly unit not in combat.
// Roll 6D6: each 4+ = 1 rally point.
// Spend rally points: Heal(1) per point, or return a slain model (costs Health characteristic).
type RallyCommand struct {
	OwnerID int
	UnitID  core.UnitID
}

func (c *RallyCommand) Type() CommandType     { return CommandTypeRally }
func (c *RallyCommand) PlayerID() int         { return c.OwnerID }
func (c *RallyCommand) GetUnitID() core.UnitID { return c.UnitID }
