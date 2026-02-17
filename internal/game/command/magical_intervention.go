package command

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// MagicalInterventionCommand allows a Wizard or Priest to use a spell/prayer
// during the enemy's hero phase at -1 to their roll.
// AoS4: Costs 1 CP. Declare a friendly Wizard or Priest.
// They may use one Spell or Prayer ability as if it were your hero phase,
// but with -1 to the casting/chanting roll.
type MagicalInterventionCommand struct {
	OwnerID    int
	CasterID   core.UnitID
	SpellIndex int         // -1 if using a prayer instead
	PrayerIndex int        // -1 if using a spell instead
	TargetID   core.UnitID
	BankPoints bool        // Only relevant if using a prayer
}

func (c *MagicalInterventionCommand) Type() CommandType     { return CommandTypeMagicalIntervention }
func (c *MagicalInterventionCommand) PlayerID() int         { return c.OwnerID }
func (c *MagicalInterventionCommand) GetCasterID() core.UnitID { return c.CasterID }
