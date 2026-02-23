package command

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

const CommandTypeAllOutAttack CommandType = "all_out_attack"

// AllOutAttackCommand spends 1 CP to give a unit +1 to hit rolls this phase.
type AllOutAttackCommand struct {
	OwnerID int
	UnitID  core.UnitID
}

func (c *AllOutAttackCommand) Type() CommandType  { return CommandTypeAllOutAttack }
func (c *AllOutAttackCommand) PlayerID() int       { return c.OwnerID }
func (c *AllOutAttackCommand) GetUnitID() core.UnitID { return c.UnitID }
