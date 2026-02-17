package board

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

func TestAllBattleplans_Returns12(t *testing.T) {
	plans := AllBattleplans()
	if len(plans) != 12 {
		t.Errorf("expected 12 battleplans, got %d", len(plans))
	}
}

func TestTable1Battleplans_Returns6(t *testing.T) {
	plans := Table1Battleplans()
	if len(plans) != 6 {
		t.Errorf("expected 6 Table 1 battleplans, got %d", len(plans))
	}
	for _, p := range plans {
		if p.Table != BattleplanTable1 {
			t.Errorf("battleplan '%s' should be Table 1, got Table %d", p.Name, p.Table)
		}
	}
}

func TestTable2Battleplans_Returns6(t *testing.T) {
	plans := Table2Battleplans()
	if len(plans) != 6 {
		t.Errorf("expected 6 Table 2 battleplans, got %d", len(plans))
	}
	for _, p := range plans {
		if p.Table != BattleplanTable2 {
			t.Errorf("battleplan '%s' should be Table 2, got Table %d", p.Name, p.Table)
		}
	}
}

func TestGetBattleplan_ValidRolls(t *testing.T) {
	for roll := 1; roll <= 6; roll++ {
		bp := GetBattleplan(BattleplanTable1, roll)
		if bp == nil {
			t.Errorf("expected battleplan for Table 1 Roll %d, got nil", roll)
		}
		bp2 := GetBattleplan(BattleplanTable2, roll)
		if bp2 == nil {
			t.Errorf("expected battleplan for Table 2 Roll %d, got nil", roll)
		}
	}
}

func TestGetBattleplan_InvalidRoll(t *testing.T) {
	bp := GetBattleplan(BattleplanTable1, 7)
	if bp != nil {
		t.Errorf("expected nil for invalid roll 7, got %s", bp.Name)
	}
}

func TestBattleplan_SetupBoard_CreatesCorrectObjectives(t *testing.T) {
	plans := AllBattleplans()
	for _, p := range plans {
		b := p.SetupBoard()
		if len(b.Objectives) != 6 {
			t.Errorf("battleplan '%s': expected 6 objectives, got %d", p.Name, len(b.Objectives))
		}

		// Verify Ghyranite type distribution: 2 Oakenbrow, 2 Gnarlroot, 1 Winterleaf, 1 Heartwood
		typeCount := make(map[GhyraniteType]int)
		for _, o := range b.Objectives {
			typeCount[o.GhyraniteType]++
		}

		if typeCount[GhyraniteOakenbrow] != 2 {
			t.Errorf("battleplan '%s': expected 2 Oakenbrow, got %d", p.Name, typeCount[GhyraniteOakenbrow])
		}
		if typeCount[GhyraniteGnarlroot] != 2 {
			t.Errorf("battleplan '%s': expected 2 Gnarlroot, got %d", p.Name, typeCount[GhyraniteGnarlroot])
		}
		if typeCount[GhyraniteWinterleaf] != 1 {
			t.Errorf("battleplan '%s': expected 1 Winterleaf, got %d", p.Name, typeCount[GhyraniteWinterleaf])
		}
		if typeCount[GhyraniteHeartwood] != 1 {
			t.Errorf("battleplan '%s': expected 1 Heartwood, got %d", p.Name, typeCount[GhyraniteHeartwood])
		}
	}
}

func TestBattleplan_SetupBoard_Has3Pairs(t *testing.T) {
	plans := AllBattleplans()
	for _, p := range plans {
		b := p.SetupBoard()
		pairIDs := b.PairIDs()
		if len(pairIDs) != 3 {
			t.Errorf("battleplan '%s': expected 3 pairs, got %d", p.Name, len(pairIDs))
		}
		// Each pair should have exactly 2 objectives
		for _, pid := range pairIDs {
			pair := b.ObjectivePair(pid)
			if len(pair) != 2 {
				t.Errorf("battleplan '%s' pair %d: expected 2 objectives, got %d", p.Name, pid, len(pair))
			}
		}
	}
}

func TestBattleplan_ObjectivesInBounds(t *testing.T) {
	plans := AllBattleplans()
	for _, p := range plans {
		b := p.SetupBoard()
		for _, o := range b.Objectives {
			if !b.IsInBounds(o.Position) {
				t.Errorf("battleplan '%s': objective %d at (%.1f, %.1f) is out of bounds (%.0f x %.0f)",
					p.Name, o.ID, o.Position.X, o.Position.Y, b.Width, b.Height)
			}
		}
	}
}

func TestBattleplan_BoardDimensions(t *testing.T) {
	plans := AllBattleplans()
	for _, p := range plans {
		if p.BoardWidth != StandardBoardWidth {
			t.Errorf("battleplan '%s': expected width %.0f, got %.0f", p.Name, StandardBoardWidth, p.BoardWidth)
		}
		if p.BoardHeight != StandardBoardHeight {
			t.Errorf("battleplan '%s': expected height %.0f, got %.0f", p.Name, StandardBoardHeight, p.BoardHeight)
		}
	}
}

func TestBattleplan_TerritoriesInBounds(t *testing.T) {
	plans := AllBattleplans()
	for _, p := range plans {
		for i, terr := range p.Territories {
			if terr.MinPos.X < 0 || terr.MinPos.Y < 0 ||
				terr.MaxPos.X > p.BoardWidth || terr.MaxPos.Y > p.BoardHeight {
				t.Errorf("battleplan '%s' territory %d: out of bounds", p.Name, i)
			}
		}
	}
}

func TestBattleplan_UniqueNames(t *testing.T) {
	plans := AllBattleplans()
	names := make(map[string]bool)
	for _, p := range plans {
		if names[p.Name] {
			t.Errorf("duplicate battleplan name: %s", p.Name)
		}
		names[p.Name] = true
	}
}

func TestBattleplan_UniqueRolls(t *testing.T) {
	// Verify each table has rolls 1-6
	for _, table := range []BattleplanTable{BattleplanTable1, BattleplanTable2} {
		var plans []Battleplan
		if table == BattleplanTable1 {
			plans = Table1Battleplans()
		} else {
			plans = Table2Battleplans()
		}
		rolls := make(map[int]bool)
		for _, p := range plans {
			if rolls[p.Roll] {
				t.Errorf("Table %d: duplicate roll %d", table, p.Roll)
			}
			rolls[p.Roll] = true
		}
		for r := 1; r <= 6; r++ {
			if !rolls[r] {
				t.Errorf("Table %d: missing roll %d", table, r)
			}
		}
	}
}

func TestTerritory_Contains(t *testing.T) {
	terr := Territory{
		Name:   "Test",
		MinPos: core.Position{X: 0, Y: 0},
		MaxPos: core.Position{X: 60, Y: 12},
	}
	if !terr.Contains(core.Position{X: 30, Y: 6}) {
		t.Error("expected (30,6) to be inside territory")
	}
	if terr.Contains(core.Position{X: 30, Y: 20}) {
		t.Error("expected (30,20) to be outside territory")
	}
}

func TestGhyraniteObjective_ControlZone(t *testing.T) {
	b := NewBoard(60, 44)
	obj := b.AddGhyraniteObjective(core.Position{X: 30, Y: 22}, GhyraniteOakenbrow, 1)

	if obj.Radius != 3.0 {
		t.Errorf("expected Ghyranite control zone of 3\", got %.1f", obj.Radius)
	}
	if !obj.IsGhyranite() {
		t.Error("expected objective to be Ghyranite")
	}
	if !obj.IsPaired() {
		t.Error("expected objective to be paired")
	}
}

func TestGhyraniteObjective_IsContestedByModel(t *testing.T) {
	b := NewBoard(60, 44)
	obj := b.AddGhyraniteObjective(core.Position{X: 30, Y: 22}, GhyraniteOakenbrow, 1)

	// Model within 3" should contest
	models := []core.Model{
		{ID: 0, Position: core.Position{X: 32, Y: 22}, IsAlive: true},
	}
	if !obj.IsContestedByModel(models) {
		t.Error("expected model within 3\" to contest")
	}

	// Model at exactly 3" should contest
	models2 := []core.Model{
		{ID: 0, Position: core.Position{X: 33, Y: 22}, IsAlive: true},
	}
	if !obj.IsContestedByModel(models2) {
		t.Error("expected model at exactly 3\" to contest")
	}

	// Model beyond 3" should not contest
	models3 := []core.Model{
		{ID: 0, Position: core.Position{X: 34, Y: 22}, IsAlive: true},
	}
	if obj.IsContestedByModel(models3) {
		t.Error("expected model beyond 3\" NOT to contest")
	}

	// Dead model should not contest
	models4 := []core.Model{
		{ID: 0, Position: core.Position{X: 30, Y: 22}, IsAlive: false},
	}
	if obj.IsContestedByModel(models4) {
		t.Error("expected dead model NOT to contest")
	}
}

func TestGhyraniteType_String(t *testing.T) {
	tests := []struct {
		gt       GhyraniteType
		expected string
	}{
		{GhyraniteNone, "None"},
		{GhyraniteOakenbrow, "Oakenbrow"},
		{GhyraniteGnarlroot, "Gnarlroot"},
		{GhyraniteWinterleaf, "Winterleaf"},
		{GhyraniteHeartwood, "Heartwood"},
	}
	for _, tt := range tests {
		if tt.gt.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.gt.String())
		}
	}
}

func TestBoard_GhyraniteObjectives(t *testing.T) {
	b := NewBoard(60, 44)
	b.AddObjective(core.Position{X: 10, Y: 10}, 6.0) // Standard
	b.AddGhyraniteObjective(core.Position{X: 30, Y: 22}, GhyraniteOakenbrow, 1)
	b.AddGhyraniteObjective(core.Position{X: 50, Y: 22}, GhyraniteGnarlroot, 2)

	ghyranite := b.GhyraniteObjectives()
	if len(ghyranite) != 2 {
		t.Errorf("expected 2 Ghyranite objectives, got %d", len(ghyranite))
	}
}

func TestBoard_ObjectivePair(t *testing.T) {
	b := NewBoard(60, 44)
	b.AddGhyraniteObjective(core.Position{X: 10, Y: 10}, GhyraniteOakenbrow, 1)
	b.AddGhyraniteObjective(core.Position{X: 50, Y: 34}, GhyraniteOakenbrow, 1)
	b.AddGhyraniteObjective(core.Position{X: 30, Y: 22}, GhyraniteGnarlroot, 2)

	pair1 := b.ObjectivePair(1)
	if len(pair1) != 2 {
		t.Errorf("expected pair 1 to have 2 objectives, got %d", len(pair1))
	}

	pair2 := b.ObjectivePair(2)
	if len(pair2) != 1 {
		t.Errorf("expected pair 2 to have 1 objective (incomplete), got %d", len(pair2))
	}

	pair99 := b.ObjectivePair(99)
	if len(pair99) != 0 {
		t.Errorf("expected pair 99 to have 0 objectives, got %d", len(pair99))
	}
}
