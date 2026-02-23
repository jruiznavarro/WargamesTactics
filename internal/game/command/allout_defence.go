package command

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

const CommandTypeAllOutDefence CommandType = "all_out_defence"

// AllOutDefenceCommand spends 1 CP to give a unit +1 to save rolls this phase.
type AllOutDefenceCommand struct {
	OwnerID int
	UnitID  core.UnitID
}

func (c *AllOutDefenceCommand) Type() CommandType  { return CommandTypeAllOutDefence }
func (c *AllOutDefenceCommand) PlayerID() int       { return c.OwnerID }
func (c *AllOutDefenceCommand) GetUnitID() core.UnitID { return c.UnitID }
