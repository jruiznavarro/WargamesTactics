package command

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// ChantCommand orders a Priest to chant a prayer.
// AoS4 Rule 19.2: Roll D6. On 1 = fail + lose D3 ritual points.
// On 2+: choose to bank ritual points (= roll) or spend them
// (add ritual points to roll; if total >= ChantingValue, prayer is answered).
type ChantCommand struct {
	OwnerID     int
	ChanterID   core.UnitID
	PrayerIndex int         // Index into the unit's Prayers slice
	TargetID    core.UnitID // Target unit (used when spending/answering)
	BankPoints  bool        // true = bank ritual points, false = attempt to answer the prayer
}

func (c *ChantCommand) Type() CommandType        { return CommandTypeChant }
func (c *ChantCommand) PlayerID() int            { return c.OwnerID }
func (c *ChantCommand) GetChanterID() core.UnitID { return c.ChanterID }
func (c *ChantCommand) GetTargetID() core.UnitID   { return c.TargetID }
