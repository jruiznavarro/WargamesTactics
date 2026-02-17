package command

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// CastCommand orders a Wizard to cast a spell at a target.
type CastCommand struct {
	OwnerID    int
	CasterID   core.UnitID
	SpellIndex int         // Index into the unit's Spells slice
	TargetID   core.UnitID // Target unit (friendly or enemy depending on spell)
}

func (c *CastCommand) Type() CommandType      { return CommandTypeCast }
func (c *CastCommand) PlayerID() int          { return c.OwnerID }
func (c *CastCommand) GetCasterID() core.UnitID { return c.CasterID }
func (c *CastCommand) GetTargetID() core.UnitID { return c.TargetID }
