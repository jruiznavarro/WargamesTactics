package board

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// Territory represents a deployment zone on the battlefield.
type Territory struct {
	Name   string        // e.g. "Player 1 Territory", "Player 2 Territory"
	MinPos core.Position // Bottom-left corner of the rectangular zone
	MaxPos core.Position // Top-right corner of the rectangular zone
}

// Contains returns true if the given position is inside the territory.
func (t *Territory) Contains(pos core.Position) bool {
	return pos.X >= t.MinPos.X && pos.X <= t.MaxPos.X &&
		pos.Y >= t.MinPos.Y && pos.Y <= t.MaxPos.Y
}

// ObjectiveConfig describes where to place an objective in a battleplan.
type ObjectiveConfig struct {
	Position      core.Position
	GhyraniteType GhyraniteType
	PairID        int // Objectives of same subtype share a PairID
}

// BattleplanTable identifies which table a battleplan belongs to.
type BattleplanTable int

const (
	BattleplanTable1 BattleplanTable = 1
	BattleplanTable2 BattleplanTable = 2
)

// Battleplan describes the complete setup for a matched play battle.
// GH 2025-26: 12 battleplans across 2 tables (D6 each).
type Battleplan struct {
	Name        string            // Battleplan name
	Table       BattleplanTable   // Which table (1 or 2)
	Roll        int               // D6 roll for this battleplan (1-6)
	BoardWidth  float64           // Board width in inches
	BoardHeight float64           // Board height in inches
	Territories [2]Territory      // Deployment zones ([0] = attacker/P1, [1] = defender/P2)
	Objectives  []ObjectiveConfig // Objective placement configs
	Description string            // Flavor text / brief description
}

// SetupBoard creates a board with the battleplan's dimensions, objectives, and returns it.
func (bp *Battleplan) SetupBoard() *Board {
	b := NewBoard(bp.BoardWidth, bp.BoardHeight)
	for _, oc := range bp.Objectives {
		b.AddGhyraniteObjective(oc.Position, oc.GhyraniteType, oc.PairID)
	}
	return b
}

// AllBattleplans returns all 12 battleplans from the GH 2025-26.
func AllBattleplans() []Battleplan {
	return append(Table1Battleplans(), Table2Battleplans()...)
}

// GetBattleplan returns the battleplan for a given table and roll.
func GetBattleplan(table BattleplanTable, roll int) *Battleplan {
	var plans []Battleplan
	if table == BattleplanTable1 {
		plans = Table1Battleplans()
	} else {
		plans = Table2Battleplans()
	}
	for i := range plans {
		if plans[i].Roll == roll {
			return &plans[i]
		}
	}
	return nil
}

// Table1Battleplans returns the 6 battleplans from Table 1 (GH 2025-26).
func Table1Battleplans() []Battleplan {
	return []Battleplan{
		battleplanPassingSeasons(),
		battleplanPathsOfTheFey(),
		battleplanRoilingRoots(),
		battleplanCyclicShifts(),
		battleplanSurgeOfSlaughter(),
		battleplanLinkedLeyLines(),
	}
}

// Table2Battleplans returns the 6 battleplans from Table 2 (GH 2025-26).
func Table2Battleplans() []Battleplan {
	return []Battleplan{
		battleplanNoxiousNexus(),
		battleplanTheLiferoots(),
		battleplanBountifulEquinox(),
		battleplanLifecycle(),
		battleplanCreepingCorruption(),
		battleplanGraspOfThorns(),
	}
}

// Standard board dimensions for matched play
const (
	StandardBoardWidth  = 60.0 // inches
	StandardBoardHeight = 44.0 // inches
)

// Standard deployment depths
const (
	DeploymentShort = 9.0  // Short-edge deployment depth
	DeploymentLong  = 12.0 // Long-edge deployment depth
)

// --- Table 1 Battleplans ---

func battleplanPassingSeasons() Battleplan {
	return Battleplan{
		Name:        "Passing Seasons",
		Table:       BattleplanTable1,
		Roll:        1,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 10, Y: 10}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 50, Y: 34}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 50, Y: 10}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 10, Y: 34}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 20, Y: 22}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 40, Y: 22}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "The Ghyranite objectives shift and cycle as the seasons change. Balanced deployment with diagonal pair scoring.",
	}
}

func battleplanPathsOfTheFey() Battleplan {
	return Battleplan{
		Name:        "Paths of the Fey",
		Table:       BattleplanTable1,
		Roll:        2,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 15, Y: 15}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 45, Y: 29}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 30, Y: 15}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 29}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 45, Y: 15}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 15, Y: 29}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "Ancient fey paths connect the objectives. A balanced layout with objectives mirrored across the center.",
	}
}

func battleplanRoilingRoots() Battleplan {
	return Battleplan{
		Name:        "Roiling Roots",
		Table:       BattleplanTable1,
		Roll:        3,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: DeploymentLong, Y: StandardBoardHeight}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: StandardBoardWidth - DeploymentLong, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 6, Y: 22}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 54, Y: 22}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 20, Y: 11}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 40, Y: 33}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 16}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 30, Y: 28}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "Short-edge deployment with objectives spread across the battlefield. The roots shift and roil beneath.",
	}
}

func battleplanCyclicShifts() Battleplan {
	return Battleplan{
		Name:        "Cyclic Shifts",
		Table:       BattleplanTable1,
		Roll:        4,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 15, Y: 8}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 45, Y: 36}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 15, Y: 36}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 45, Y: 8}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 22}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 30, Y: 22}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "The Ghyranite energy cycles through the objectives. Center objective is contested territory.",
	}
}

func battleplanSurgeOfSlaughter() Battleplan {
	return Battleplan{
		Name:        "Surge of Slaughter",
		Table:       BattleplanTable1,
		Roll:        5,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 10, Y: 22}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 50, Y: 22}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 20, Y: 8}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 40, Y: 36}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 14}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 30, Y: 30}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "Aggressive battleplan rewarding push into enemy territory.",
	}
}

func battleplanLinkedLeyLines() Battleplan {
	return Battleplan{
		Name:        "Linked Ley Lines",
		Table:       BattleplanTable1,
		Roll:        6,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 15, Y: 22}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 45, Y: 22}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 30, Y: 8}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 36}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 10, Y: 14}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 50, Y: 30}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "Ley lines of Ghyranite energy link paired objectives across the battlefield.",
	}
}

// --- Table 2 Battleplans ---

func battleplanNoxiousNexus() Battleplan {
	return Battleplan{
		Name:        "Noxious Nexus",
		Table:       BattleplanTable2,
		Roll:        1,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 20, Y: 14}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 40, Y: 30}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 40, Y: 14}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 20, Y: 30}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 22}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 10, Y: 22}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "A noxious nexus of Ghyranite energy corrupts the center of the battlefield.",
	}
}

func battleplanTheLiferoots() Battleplan {
	return Battleplan{
		Name:        "The Liferoots",
		Table:       BattleplanTable2,
		Roll:        2,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: DeploymentLong, Y: StandardBoardHeight}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: StandardBoardWidth - DeploymentLong, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 6, Y: 14}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 54, Y: 30}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 6, Y: 30}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 54, Y: 14}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 24, Y: 22}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 36, Y: 22}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "Short-edge deployment. The liferoots spread through the battlefield connecting distant objectives.",
	}
}

func battleplanBountifulEquinox() Battleplan {
	return Battleplan{
		Name:        "Bountiful Equinox",
		Table:       BattleplanTable2,
		Roll:        3,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 10, Y: 14}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 50, Y: 30}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 30, Y: 14}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 30}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 50, Y: 14}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 10, Y: 30}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "The equinox brings balance to the Ghyranite energy. Symmetrical layout rewards control of both flanks.",
	}
}

func battleplanLifecycle() Battleplan {
	return Battleplan{
		Name:        "Lifecycle",
		Table:       BattleplanTable2,
		Roll:        4,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 15, Y: 11}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 45, Y: 33}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 45, Y: 11}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 15, Y: 33}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 18}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 30, Y: 26}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "The cycle of life and death flows through the battlefield as objectives wax and wane.",
	}
}

func battleplanCreepingCorruption() Battleplan {
	return Battleplan{
		Name:        "Creeping Corruption",
		Table:       BattleplanTable2,
		Roll:        5,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 20, Y: 8}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 40, Y: 36}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 10, Y: 22}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 50, Y: 22}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 40, Y: 8}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 20, Y: 36}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "Corruption spreads from the edges inward. Objectives in each player's territory create pressure to advance.",
	}
}

func battleplanGraspOfThorns() Battleplan {
	return Battleplan{
		Name:        "Grasp of Thorns",
		Table:       BattleplanTable2,
		Roll:        6,
		BoardWidth:  StandardBoardWidth,
		BoardHeight: StandardBoardHeight,
		Territories: [2]Territory{
			{Name: "Player 1 Territory", MinPos: core.Position{X: 0, Y: 0}, MaxPos: core.Position{X: StandardBoardWidth, Y: DeploymentLong}},
			{Name: "Player 2 Territory", MinPos: core.Position{X: 0, Y: StandardBoardHeight - DeploymentLong}, MaxPos: core.Position{X: StandardBoardWidth, Y: StandardBoardHeight}},
		},
		Objectives: []ObjectiveConfig{
			{Position: core.Position{X: 15, Y: 22}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 45, Y: 22}, GhyraniteType: GhyraniteOakenbrow, PairID: 1},
			{Position: core.Position{X: 10, Y: 8}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 50, Y: 36}, GhyraniteType: GhyraniteGnarlroot, PairID: 2},
			{Position: core.Position{X: 30, Y: 11}, GhyraniteType: GhyraniteWinterleaf, PairID: 3},
			{Position: core.Position{X: 30, Y: 33}, GhyraniteType: GhyraniteHeartwood, PairID: 3},
		},
		Description: "Thorny vines grasp at the objectives. Center flanks are key to controlling the battlefield.",
	}
}
