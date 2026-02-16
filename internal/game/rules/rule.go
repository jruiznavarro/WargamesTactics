package rules

// Source identifies where a rule comes from.
type Source int

const (
	SourceTerrain     Source = iota // Rule from terrain feature
	SourceUnitAbility              // Rule from a unit's innate ability
	SourceWeapon                   // Rule from a weapon profile
	SourceGlobal                   // Global game rule
)

// Rule defines a single game rule that hooks into the engine.
type Rule struct {
	// Name is a human-readable identifier for logging/debugging.
	Name string

	// Trigger is the hook point where this rule is evaluated.
	Trigger Trigger

	// Source identifies where this rule comes from.
	Source Source

	// Condition returns true if this rule should apply given the current context.
	// If nil, the rule always applies.
	Condition func(ctx *Context) bool

	// Apply modifies the context (typically ctx.Modifiers or ctx.Blocked).
	Apply func(ctx *Context)
}
