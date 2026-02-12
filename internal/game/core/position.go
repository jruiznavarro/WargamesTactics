package core

import "math"

// Position represents a point on the battlefield in inches.
type Position struct {
	X float64
	Y float64
}

// Distance returns the Euclidean distance between two positions.
func Distance(p1, p2 Position) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Add returns a new position offset by the given deltas.
func (p Position) Add(dx, dy float64) Position {
	return Position{X: p.X + dx, Y: p.Y + dy}
}

// Towards returns a position moved from p towards target by the given distance.
// If the distance is greater than or equal to the distance to target, returns target.
func (p Position) Towards(target Position, dist float64) Position {
	d := Distance(p, target)
	if d <= dist || d < 1e-9 {
		return target
	}
	ratio := dist / d
	return Position{
		X: p.X + (target.X-p.X)*ratio,
		Y: p.Y + (target.Y-p.Y)*ratio,
	}
}
