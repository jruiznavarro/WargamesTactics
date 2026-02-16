package board_test

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
)

func setupEngine(b *board.Board) *rules.Engine {
	e := rules.NewEngine()
	for _, r := range board.TerrainRules(b) {
		e.AddRule(r)
	}
	return e
}

func TestWoods_MovePenalty(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Forest", board.TerrainWoods, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

	// Moving INTO the woods should apply -2 movement
	ctx := &rules.Context{
		Attacker:    &core.Unit{ID: 1},
		Origin:      core.Position{X: 5, Y: 12},
		Destination: core.Position{X: 12, Y: 12},
	}
	e.Evaluate(rules.BeforeMove, ctx)

	if ctx.Modifiers.MoveMod != -2 {
		t.Errorf("expected MoveMod -2 in woods, got %d", ctx.Modifiers.MoveMod)
	}

	// Moving to position OUTSIDE the woods should NOT apply penalty
	ctx2 := &rules.Context{
		Attacker:    &core.Unit{ID: 1},
		Origin:      core.Position{X: 5, Y: 12},
		Destination: core.Position{X: 8, Y: 12},
	}
	e.Evaluate(rules.BeforeMove, ctx2)

	if ctx2.Modifiers.MoveMod != 0 {
		t.Errorf("expected MoveMod 0 outside woods, got %d", ctx2.Modifiers.MoveMod)
	}
}

func TestWoods_CoverBonus(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Forest", board.TerrainWoods, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

	defender := &core.Unit{
		ID: 2,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 12, Y: 12}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	ctx := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Weapon:   &core.Weapon{Name: "Bow", Range: 18},
	}
	e.Evaluate(rules.BeforeSaveRoll, ctx)

	if ctx.Modifiers.SaveMod != 1 {
		t.Errorf("expected SaveMod +1 in woods, got %d", ctx.Modifiers.SaveMod)
	}
}

func TestObstacle_CoverOnlyRanged(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Wall", board.TerrainObstacle, core.Position{X: 10, Y: 10}, 4, 2)
	e := setupEngine(b)

	defender := &core.Unit{
		ID: 2,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 12, Y: 11}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	// Ranged weapon should get cover
	ctx := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Weapon:   &core.Weapon{Name: "Bow", Range: 18},
	}
	e.Evaluate(rules.BeforeSaveRoll, ctx)

	if ctx.Modifiers.SaveMod != 1 {
		t.Errorf("expected SaveMod +1 behind obstacle (ranged), got %d", ctx.Modifiers.SaveMod)
	}

	// Melee weapon should NOT get cover from obstacle
	ctx2 := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Weapon:   &core.Weapon{Name: "Sword", Range: 0},
	}
	e.Evaluate(rules.BeforeSaveRoll, ctx2)

	if ctx2.Modifiers.SaveMod != 0 {
		t.Errorf("expected SaveMod 0 behind obstacle (melee), got %d", ctx2.Modifiers.SaveMod)
	}
}

func TestRuins_CoverBothMeleeAndRanged(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Ruins", board.TerrainRuins, core.Position{X: 10, Y: 10}, 5, 5)
	e := setupEngine(b)

	defender := &core.Unit{
		ID: 2,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 12, Y: 12}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	// Ranged
	ctx := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Weapon:   &core.Weapon{Name: "Bow", Range: 18},
	}
	e.Evaluate(rules.BeforeSaveRoll, ctx)
	if ctx.Modifiers.SaveMod != 1 {
		t.Errorf("expected SaveMod +1 in ruins (ranged), got %d", ctx.Modifiers.SaveMod)
	}

	// Melee
	ctx2 := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Weapon:   &core.Weapon{Name: "Sword", Range: 0},
	}
	e.Evaluate(rules.BeforeSaveRoll, ctx2)
	if ctx2.Modifiers.SaveMod != 1 {
		t.Errorf("expected SaveMod +1 in ruins (melee), got %d", ctx2.Modifiers.SaveMod)
	}
}

func TestImpassable_BlocksMovement(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Lava", board.TerrainImpassable, core.Position{X: 20, Y: 10}, 4, 4)
	e := setupEngine(b)

	ctx := &rules.Context{
		Attacker:    &core.Unit{ID: 1},
		Origin:      core.Position{X: 15, Y: 12},
		Destination: core.Position{X: 22, Y: 12},
	}
	e.Evaluate(rules.BeforeMove, ctx)

	if !ctx.Blocked {
		t.Error("expected movement into impassable terrain to be blocked")
	}
	if ctx.BlockMessage == "" {
		t.Error("expected a block message")
	}
}

func TestImpassable_BlocksCharge(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Lava", board.TerrainImpassable, core.Position{X: 20, Y: 10}, 4, 4)
	e := setupEngine(b)

	defender := &core.Unit{
		ID: 2,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 22, Y: 12}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	ctx := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Origin:   core.Position{X: 15, Y: 12},
	}
	e.Evaluate(rules.BeforeCharge, ctx)

	if !ctx.Blocked {
		t.Error("expected charge into impassable terrain to be blocked")
	}
}

func TestOpenTerrain_NoEffect(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Field", board.TerrainOpen, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

	ctx := &rules.Context{
		Attacker:    &core.Unit{ID: 1},
		Origin:      core.Position{X: 5, Y: 12},
		Destination: core.Position{X: 12, Y: 12},
	}
	e.Evaluate(rules.BeforeMove, ctx)

	if ctx.Modifiers.MoveMod != 0 {
		t.Errorf("expected no move penalty for open terrain, got %d", ctx.Modifiers.MoveMod)
	}
	if ctx.Blocked {
		t.Error("open terrain should not block movement")
	}
}

func TestMultipleTerrain_StackingRules(t *testing.T) {
	b := board.NewBoard(48, 24)
	// Two overlapping terrain pieces at same position
	b.AddTerrain("Woods1", board.TerrainWoods, core.Position{X: 10, Y: 10}, 6, 6)
	b.AddTerrain("Ruins1", board.TerrainRuins, core.Position{X: 12, Y: 12}, 4, 4)
	e := setupEngine(b)

	// Defender is in both woods and ruins
	defender := &core.Unit{
		ID: 2,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 13, Y: 13}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	ctx := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Weapon:   &core.Weapon{Name: "Bow", Range: 18},
	}
	e.Evaluate(rules.BeforeSaveRoll, ctx)

	// Both woods (+1) and ruins (+1) should stack
	if ctx.Modifiers.SaveMod != 2 {
		t.Errorf("expected SaveMod +2 from stacked terrain, got %d", ctx.Modifiers.SaveMod)
	}
}

func TestTerrainRules_Count(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Woods", board.TerrainWoods, core.Position{X: 0, Y: 0}, 4, 4)
	b.AddTerrain("Wall", board.TerrainObstacle, core.Position{X: 10, Y: 10}, 2, 2)
	b.AddTerrain("Ruins", board.TerrainRuins, core.Position{X: 20, Y: 20}, 3, 3)
	b.AddTerrain("Lava", board.TerrainImpassable, core.Position{X: 30, Y: 0}, 2, 2)
	b.AddTerrain("Field", board.TerrainOpen, core.Position{X: 40, Y: 0}, 4, 4)

	rulesList := board.TerrainRules(b)

	// Woods: 2 (move penalty + cover)
	// Obstacle: 1 (cover ranged only)
	// Ruins: 1 (cover)
	// Impassable: 2 (block move + block charge)
	// Open: 0
	// Total: 6
	expected := 6
	if len(rulesList) != expected {
		t.Errorf("expected %d terrain rules, got %d", expected, len(rulesList))
	}
}
