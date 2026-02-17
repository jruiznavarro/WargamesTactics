package board

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// GhyraniteType represents the Ghyranite subtype of an objective (GH 2025-26 Season Rules).
type GhyraniteType int

const (
	GhyraniteNone       GhyraniteType = iota // Standard objective (non-Ghyranite)
	GhyraniteOakenbrow                       // Oakenbrow: defensive subtype
	GhyraniteGnarlroot                       // Gnarlroot: tenacious subtype
	GhyraniteWinterleaf                      // Winterleaf: aggressive subtype
	GhyraniteHeartwood                       // Heartwood: vital subtype
)

func (g GhyraniteType) String() string {
	switch g {
	case GhyraniteOakenbrow:
		return "Oakenbrow"
	case GhyraniteGnarlroot:
		return "Gnarlroot"
	case GhyraniteWinterleaf:
		return "Winterleaf"
	case GhyraniteHeartwood:
		return "Heartwood"
	default:
		return "None"
	}
}

// Objective represents a scoring objective on the battlefield.
// Units contest an objective if they are within its radius.
type Objective struct {
	ID            int
	Position      core.Position
	Radius        float64       // Distance in inches from center that counts as "on" the objective
	GhyraniteType GhyraniteType // Ghyranite subtype (GH 2025-26); GhyraniteNone for standard
	PairID        int           // Objectives of the same subtype share a PairID (0 = not paired)
}

// IsContested returns true if the given position is within the objective's radius.
func (o *Objective) IsContested(pos core.Position) bool {
	return core.Distance(o.Position, pos) <= o.Radius
}

// IsContestedByModel returns true if any model in the unit is within the objective's radius.
// GH 2025-26 Season Rules: Ghyranite objectives are contested per-model.
func (o *Objective) IsContestedByModel(models []core.Model) bool {
	for i := range models {
		if !models[i].IsAlive {
			continue
		}
		if core.Distance(o.Position, models[i].Position) <= o.Radius {
			return true
		}
	}
	return false
}

// IsGhyranite returns true if this objective has a Ghyranite subtype.
func (o *Objective) IsGhyranite() bool {
	return o.GhyraniteType != GhyraniteNone
}

// IsPaired returns true if this objective is part of a pair.
func (o *Objective) IsPaired() bool {
	return o.PairID > 0
}
