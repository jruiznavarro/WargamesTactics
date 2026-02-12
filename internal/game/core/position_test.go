package core

import (
	"math"
	"testing"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name string
		p1   Position
		p2   Position
		want float64
	}{
		{"same point", Position{X: 0, Y: 0}, Position{X: 0, Y: 0}, 0},
		{"horizontal", Position{X: 0, Y: 0}, Position{X: 3, Y: 0}, 3},
		{"vertical", Position{X: 0, Y: 0}, Position{X: 0, Y: 4}, 4},
		{"diagonal 3-4-5", Position{X: 0, Y: 0}, Position{X: 3, Y: 4}, 5},
		{"negative coords", Position{X: -1, Y: -1}, Position{X: 2, Y: 3}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Distance(tt.p1, tt.p2)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Distance(%v, %v) = %f, want %f", tt.p1, tt.p2, got, tt.want)
			}
		})
	}
}

func TestPosition_Towards(t *testing.T) {
	origin := Position{X: 0, Y: 0}
	target := Position{X: 10, Y: 0}

	// Move 5 inches toward target
	result := origin.Towards(target, 5)
	if math.Abs(result.X-5.0) > 1e-9 || math.Abs(result.Y) > 1e-9 {
		t.Errorf("expected (5, 0), got (%f, %f)", result.X, result.Y)
	}

	// Move more than distance returns target
	result = origin.Towards(target, 15)
	if math.Abs(result.X-10.0) > 1e-9 || math.Abs(result.Y) > 1e-9 {
		t.Errorf("expected (10, 0), got (%f, %f)", result.X, result.Y)
	}
}

func TestPosition_Add(t *testing.T) {
	p := Position{X: 3, Y: 4}
	result := p.Add(1.5, -2.0)
	if math.Abs(result.X-4.5) > 1e-9 || math.Abs(result.Y-2.0) > 1e-9 {
		t.Errorf("expected (4.5, 2.0), got (%f, %f)", result.X, result.Y)
	}
}
