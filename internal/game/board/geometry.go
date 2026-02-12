package board

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

const FloatTolerance = 1e-9

// InRange checks if two positions are within the given range (inclusive).
func InRange(p1, p2 core.Position, rangeInches float64) bool {
	return core.Distance(p1, p2) <= rangeInches+FloatTolerance
}

// BasesOverlap checks if two circular bases overlap.
// Each base is defined by a center position and a diameter.
func BasesOverlap(p1 core.Position, diameter1 float64, p2 core.Position, diameter2 float64) bool {
	dist := core.Distance(p1, p2)
	minDist := (diameter1 + diameter2) / 2.0
	return dist < minDist-FloatTolerance
}

// WithinCoherency checks if a model is within coherency distance of at least
// one other model in the unit. Coherency in AoS is typically 1 inch.
func WithinCoherency(model core.Position, others []core.Position, coherencyRange float64) bool {
	for _, other := range others {
		if InRange(model, other, coherencyRange) {
			return true
		}
	}
	return false
}

// UnitCoherencyValid checks that every model in the unit is within coherency
// range of at least one other model. A single model is always coherent.
func UnitCoherencyValid(positions []core.Position, coherencyRange float64) bool {
	if len(positions) <= 1 {
		return true
	}
	for i, pos := range positions {
		others := make([]core.Position, 0, len(positions)-1)
		for j, other := range positions {
			if i != j {
				others = append(others, other)
			}
		}
		if !WithinCoherency(pos, others, coherencyRange) {
			return false
		}
	}
	return true
}

// MoveDistanceValid checks if moving from origin to destination is within
// the allowed movement distance.
func MoveDistanceValid(origin, destination core.Position, maxMove float64) bool {
	return core.Distance(origin, destination) <= maxMove+FloatTolerance
}
