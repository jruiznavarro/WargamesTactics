package rules

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

func TestEngineAddAndEvaluate(t *testing.T) {
	e := NewEngine()

	// Add a rule that gives +1 to save when defender has Save >= 4
	e.AddRule(Rule{
		Name:    "Light Cover",
		Trigger: BeforeSaveRoll,
		Source:  SourceTerrain,
		Condition: func(ctx *Context) bool {
			return ctx.Defender != nil && ctx.Defender.Stats.Save >= 4
		},
		Apply: func(ctx *Context) {
			ctx.Modifiers.SaveMod += 1
		},
	})

	if e.RuleCount() != 1 {
		t.Fatalf("expected 1 rule, got %d", e.RuleCount())
	}

	// Test with a defender that qualifies
	defender := &core.Unit{Stats: core.Stats{Save: 4}}
	ctx := &Context{Defender: defender}
	e.Evaluate(BeforeSaveRoll, ctx)

	if ctx.Modifiers.SaveMod != 1 {
		t.Errorf("expected SaveMod 1, got %d", ctx.Modifiers.SaveMod)
	}

	// Test with a defender that doesn't qualify
	defender2 := &core.Unit{Stats: core.Stats{Save: 3}}
	ctx2 := &Context{Defender: defender2}
	e.Evaluate(BeforeSaveRoll, ctx2)

	if ctx2.Modifiers.SaveMod != 0 {
		t.Errorf("expected SaveMod 0 for save 3+, got %d", ctx2.Modifiers.SaveMod)
	}
}

func TestEngineMultipleRulesStack(t *testing.T) {
	e := NewEngine()

	// Two rules that modify hit rolls
	e.AddRule(Rule{
		Name:    "Charging Bonus",
		Trigger: BeforeHitRoll,
		Source:  SourceUnitAbility,
		Apply: func(ctx *Context) {
			ctx.Modifiers.HitMod += 1
		},
	})
	e.AddRule(Rule{
		Name:    "Mystic Shield",
		Trigger: BeforeHitRoll,
		Source:  SourceGlobal,
		Apply: func(ctx *Context) {
			ctx.Modifiers.HitMod -= 1
		},
	})

	ctx := &Context{}
	e.Evaluate(BeforeHitRoll, ctx)

	// +1 and -1 should cancel out
	if ctx.Modifiers.HitMod != 0 {
		t.Errorf("expected HitMod 0, got %d", ctx.Modifiers.HitMod)
	}
}

func TestEngineRemoveBySource(t *testing.T) {
	e := NewEngine()

	e.AddRule(Rule{
		Name:    "Cover",
		Trigger: BeforeSaveRoll,
		Source:  SourceTerrain,
		Apply:   func(ctx *Context) { ctx.Modifiers.SaveMod += 1 },
	})
	e.AddRule(Rule{
		Name:    "Tough Skin",
		Trigger: BeforeSaveRoll,
		Source:  SourceUnitAbility,
		Apply:   func(ctx *Context) { ctx.Modifiers.SaveMod += 1 },
	})

	if e.RuleCount() != 2 {
		t.Fatalf("expected 2 rules, got %d", e.RuleCount())
	}

	e.RemoveRulesBySource(SourceTerrain, "Cover")

	if e.RuleCount() != 1 {
		t.Errorf("expected 1 rule after removal, got %d", e.RuleCount())
	}

	ctx := &Context{}
	e.Evaluate(BeforeSaveRoll, ctx)
	if ctx.Modifiers.SaveMod != 1 {
		t.Errorf("expected SaveMod 1 (only Tough Skin), got %d", ctx.Modifiers.SaveMod)
	}
}

func TestEngineBlockAction(t *testing.T) {
	e := NewEngine()

	e.AddRule(Rule{
		Name:    "Impassable",
		Trigger: BeforeMove,
		Source:  SourceTerrain,
		Apply: func(ctx *Context) {
			ctx.Blocked = true
			ctx.BlockMessage = "terrain is impassable"
		},
	})

	ctx := &Context{}
	e.Evaluate(BeforeMove, ctx)

	if !ctx.Blocked {
		t.Error("expected move to be blocked")
	}
	if ctx.BlockMessage != "terrain is impassable" {
		t.Errorf("unexpected block message: %s", ctx.BlockMessage)
	}
}

func TestEngineNoRulesForTrigger(t *testing.T) {
	e := NewEngine()

	ctx := &Context{}
	result := e.Evaluate(BeforeHitRoll, ctx)

	// Should return context unchanged
	if result.Modifiers.HitMod != 0 {
		t.Error("expected no modifications with no rules")
	}
}
