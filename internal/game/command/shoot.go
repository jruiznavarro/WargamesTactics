package command

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// ShootCommand fires ranged weapons at a target unit.
type ShootCommand struct {
	OwnerID   int
	ShooterID core.UnitID
	TargetID  core.UnitID
}

func (c *ShootCommand) Type() CommandType        { return CommandTypeShoot }
func (c *ShootCommand) PlayerID() int            { return c.OwnerID }
func (c *ShootCommand) GetShooterID() core.UnitID { return c.ShooterID }
func (c *ShootCommand) GetTargetID() core.UnitID   { return c.TargetID }
