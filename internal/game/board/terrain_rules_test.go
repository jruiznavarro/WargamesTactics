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

// --- Obstacle: Cover + Unstable ---

func TestObstacle_CoverReducesHitRoll(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Wall", board.TerrainObstacle, core.Position{X: 10, Y: 10}, 4, 4)
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
	e.Evaluate(rules.BeforeHitRoll, ctx)

	if ctx.Modifiers.HitMod != -1 {
		t.Errorf("expected HitMod -1 from cover, got %d", ctx.Modifiers.HitMod)
	}
}

func TestObstacle_CoverIgnoredIfCharged(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Wall", board.TerrainObstacle, core.Position{X: 10, Y: 10}, 4, 4)
	e := setupEngine(b)

	defender := &core.Unit{
		ID:         2,
		HasCharged: true,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 12, Y: 12}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	ctx := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Weapon:   &core.Weapon{Name: "Bow", Range: 18},
	}
	e.Evaluate(rules.BeforeHitRoll, ctx)

	if ctx.Modifiers.HitMod != 0 {
		t.Errorf("expected HitMod 0 (charged unit ignores cover), got %d", ctx.Modifiers.HitMod)
	}
}

func TestObstacle_CoverIgnoredIfFly(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Wall", board.TerrainObstacle, core.Position{X: 10, Y: 10}, 4, 4)
	e := setupEngine(b)

	defender := &core.Unit{
		ID:       2,
		Keywords: []core.Keyword{core.KeywordFly},
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 12, Y: 12}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	ctx := &rules.Context{
		Attacker: &core.Unit{ID: 1},
		Defender: defender,
		Weapon:   &core.Weapon{Name: "Bow", Range: 18},
	}
	e.Evaluate(rules.BeforeHitRoll, ctx)

	if ctx.Modifiers.HitMod != 0 {
		t.Errorf("expected HitMod 0 (Fly ignores cover), got %d", ctx.Modifiers.HitMod)
	}
}

func TestObstacle_UnstableBlocksMove(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Wall", board.TerrainObstacle, core.Position{X: 10, Y: 10}, 4, 4)
	e := setupEngine(b)

	ctx := &rules.Context{
		Attacker:    &core.Unit{ID: 1},
		Origin:      core.Position{X: 5, Y: 12},
		Destination: core.Position{X: 12, Y: 12},
	}
	e.Evaluate(rules.BeforeMove, ctx)

	if !ctx.Blocked {
		t.Error("expected move into obstacle terrain to be blocked (unstable)")
	}
}

// --- Obscuring: Cover + Obscuring + Unstable ---

func TestObscuring_CoverReducesHitRoll(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Forest", board.TerrainObscuring, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

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
	e.Evaluate(rules.BeforeHitRoll, ctx)

	if ctx.Modifiers.HitMod != -1 {
		t.Errorf("expected HitMod -1 from cover in obscuring terrain, got %d", ctx.Modifiers.HitMod)
	}
}

func TestObscuring_BlocksShootingFromDistance(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Forest", board.TerrainObscuring, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 0, Y: 13}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}
	defender := &core.Unit{
		ID: 2,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 13, Y: 13}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	ctx := &rules.Context{
		Attacker: attacker,
		Defender: defender,
	}
	e.Evaluate(rules.BeforeShoot, ctx)

	if !ctx.Blocked {
		t.Error("expected shooting to be blocked by obscuring terrain (attacker >3\" away)")
	}
}

func TestObscuring_AllowsShootingFromCombatRange(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Forest", board.TerrainObscuring, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

	// Attacker is within 3" of defender
	attacker := &core.Unit{
		ID: 1,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 11, Y: 13}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}
	defender := &core.Unit{
		ID: 2,
		Models: []core.Model{
			{ID: 0, Position: core.Position{X: 13, Y: 13}, CurrentWounds: 1, MaxWounds: 1, IsAlive: true},
		},
	}

	ctx := &rules.Context{
		Attacker: attacker,
		Defender: defender,
	}
	e.Evaluate(rules.BeforeShoot, ctx)

	if ctx.Blocked {
		t.Error("shooting should not be blocked when attacker is within 3\" of defender")
	}
}

func TestObscuring_UnstableBlocksMove(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Forest", board.TerrainObscuring, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

	ctx := &rules.Context{
		Attacker:    &core.Unit{ID: 1},
		Origin:      core.Position{X: 5, Y: 12},
		Destination: core.Position{X: 13, Y: 13},
	}
	e.Evaluate(rules.BeforeMove, ctx)

	if !ctx.Blocked {
		t.Error("expected move into obscuring terrain to be blocked (unstable)")
	}
}

// --- Area: Cover only ---

func TestArea_CoverReducesHitRoll(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Hill", board.TerrainArea, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

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
	e.Evaluate(rules.BeforeHitRoll, ctx)

	if ctx.Modifiers.HitMod != -1 {
		t.Errorf("expected HitMod -1 from area terrain cover, got %d", ctx.Modifiers.HitMod)
	}
}

func TestArea_DoesNotBlockMove(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Hill", board.TerrainArea, core.Position{X: 10, Y: 10}, 6, 6)
	e := setupEngine(b)

	ctx := &rules.Context{
		Attacker:    &core.Unit{ID: 1},
		Origin:      core.Position{X: 5, Y: 12},
		Destination: core.Position{X: 13, Y: 13},
	}
	e.Evaluate(rules.BeforeMove, ctx)

	if ctx.Blocked {
		t.Error("area terrain should not block movement")
	}
}

// --- Impassable ---

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

// --- Open ---

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

// --- Stacking ---

func TestMultipleTerrain_CoverStacks(t *testing.T) {
	b := board.NewBoard(48, 24)
	// Two overlapping cover-granting terrain pieces
	b.AddTerrain("Forest", board.TerrainObscuring, core.Position{X: 10, Y: 10}, 6, 6)
	b.AddTerrain("Hill", board.TerrainArea, core.Position{X: 12, Y: 12}, 4, 4)
	e := setupEngine(b)

	// Defender is in both terrain pieces
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
	e.Evaluate(rules.BeforeHitRoll, ctx)

	// Both cover rules stack: -1 + -1 = -2
	if ctx.Modifiers.HitMod != -2 {
		t.Errorf("expected HitMod -2 from stacked cover, got %d", ctx.Modifiers.HitMod)
	}
}

// --- Rule Count ---

func TestTerrainRules_Count(t *testing.T) {
	b := board.NewBoard(48, 24)
	b.AddTerrain("Forest", board.TerrainObscuring, core.Position{X: 0, Y: 0}, 4, 4)    // 3 rules: Cover + Obscuring + Unstable
	b.AddTerrain("Wall", board.TerrainObstacle, core.Position{X: 10, Y: 10}, 2, 2)      // 2 rules: Cover + Unstable
	b.AddTerrain("Hill", board.TerrainArea, core.Position{X: 20, Y: 20}, 3, 3)           // 1 rule: Cover
	b.AddTerrain("Lava", board.TerrainImpassable, core.Position{X: 30, Y: 0}, 2, 2)     // 2 rules: BlockMove + BlockCharge
	b.AddTerrain("Field", board.TerrainOpen, core.Position{X: 40, Y: 0}, 4, 4)           // 0 rules

	rulesList := board.TerrainRules(b)

	// Obscuring: 3 + Obstacle: 2 + Area: 1 + Impassable: 2 + Open: 0 = 8
	expected := 8
	if len(rulesList) != expected {
		t.Errorf("expected %d terrain rules, got %d", expected, len(rulesList))
	}
}
