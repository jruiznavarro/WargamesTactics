package board

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// Objective represents a scoring objective on the battlefield.
// Units contest an objective if they are within its radius.
type Objective struct {
	ID       int
	Position core.Position
	Radius   float64 // Distance in inches from center that counts as "on" the objective
}

// IsContested returns true if the given position is within the objective's radius.
func (o *Objective) IsContested(pos core.Position) bool {
	return core.Distance(o.Position, pos) <= o.Radius
}
