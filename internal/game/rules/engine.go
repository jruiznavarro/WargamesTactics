package rules

// Engine stores all active rules and evaluates them at hook points.
type Engine struct {
	rules map[Trigger][]Rule
}

// NewEngine creates an empty rule engine.
func NewEngine() *Engine {
	return &Engine{
		rules: make(map[Trigger][]Rule),
	}
}

// AddRule registers a rule in the engine.
func (e *Engine) AddRule(r Rule) {
	e.rules[r.Trigger] = append(e.rules[r.Trigger], r)
}

// RemoveRulesBySource removes all rules from a given source.
// Useful when terrain is removed or a buff expires.
func (e *Engine) RemoveRulesBySource(source Source, name string) {
	for trigger, ruleList := range e.rules {
		filtered := ruleList[:0]
		for _, r := range ruleList {
			if !(r.Source == source && r.Name == name) {
				filtered = append(filtered, r)
			}
		}
		e.rules[trigger] = filtered
	}
}

// Evaluate runs all rules for the given trigger against the context.
// Rules whose condition matches will apply their effect to ctx.
// Returns the (possibly modified) context.
func (e *Engine) Evaluate(trigger Trigger, ctx *Context) *Context {
	ruleList, ok := e.rules[trigger]
	if !ok {
		return ctx
	}

	for _, r := range ruleList {
		if r.Condition != nil && !r.Condition(ctx) {
			continue
		}
		r.Apply(ctx)
	}

	return ctx
}

// HasRulesFor returns true if there are any rules registered for a trigger.
func (e *Engine) HasRulesFor(trigger Trigger) bool {
	ruleList, ok := e.rules[trigger]
	return ok && len(ruleList) > 0
}

// RuleCount returns the total number of registered rules.
func (e *Engine) RuleCount() int {
	total := 0
	for _, ruleList := range e.rules {
		total += len(ruleList)
	}
	return total
}
