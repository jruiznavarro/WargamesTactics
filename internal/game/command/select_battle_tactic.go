package command

const CommandTypeSelectBattleTactic CommandType = "select_battle_tactic"

// SelectBattleTacticCommand selects a battle tactic card and tier for the round.
type SelectBattleTacticCommand struct {
	OwnerID int
	CardID  int // BattleTacticCardID (1-6)
	Tier    int // BattleTacticTier (0=Affray, 1=Strike, 2=Domination)
}

func (c *SelectBattleTacticCommand) Type() CommandType { return CommandTypeSelectBattleTactic }
func (c *SelectBattleTacticCommand) PlayerID() int     { return c.OwnerID }
func (c *SelectBattleTacticCommand) GetCardID() int    { return c.CardID }
func (c *SelectBattleTacticCommand) GetTier() int      { return c.Tier }
