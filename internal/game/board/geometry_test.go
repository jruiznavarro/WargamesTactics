package board

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

func TestInRange(t *testing.T) {
	p1 := core.Position{X: 0, Y: 0}
	p2 := core.Position{X: 5, Y: 0}

	if !InRange(p1, p2, 5.0) {
		t.Error("should be in range at exactly 5\"")
	}
	if !InRange(p1, p2, 6.0) {
		t.Error("should be in range at 6\"")
	}
	if InRange(p1, p2, 4.0) {
		t.Error("should NOT be in range at 4\"")
	}
}

func TestBasesOverlap(t *testing.T) {
	// Two 1" bases, 0.5" apart (centers)
	p1 := core.Position{X: 0, Y: 0}
	p2 := core.Position{X: 0.5, Y: 0}
	if !BasesOverlap(p1, 1.0, p2, 1.0) {
		t.Error("bases should overlap (distance 0.5, min 1.0)")
	}

	// Two 1" bases, exactly touching
	p3 := core.Position{X: 1, Y: 0}
	if BasesOverlap(p1, 1.0, p3, 1.0) {
		t.Error("bases should NOT overlap when just touching")
	}

	// Two 1" bases, far apart
	p4 := core.Position{X: 5, Y: 0}
	if BasesOverlap(p1, 1.0, p4, 1.0) {
		t.Error("bases should NOT overlap when far apart")
	}
}

func TestUnitCoherencyValid(t *testing.T) {
	// Single model: always coherent
	if !UnitCoherencyValid([]core.Position{{X: 0, Y: 0}}, 1.0) {
		t.Error("single model should always be coherent")
	}

	// Two models within 1"
	if !UnitCoherencyValid([]core.Position{{X: 0, Y: 0}, {X: 0.5, Y: 0}}, 1.0) {
		t.Error("two close models should be coherent")
	}

	// Two models too far apart
	if UnitCoherencyValid([]core.Position{{X: 0, Y: 0}, {X: 5, Y: 0}}, 1.0) {
		t.Error("two far models should NOT be coherent")
	}

	// Three models in line: each adjacent pair is within 1"
	positions := []core.Position{{X: 0, Y: 0}, {X: 0.8, Y: 0}, {X: 1.6, Y: 0}}
	if !UnitCoherencyValid(positions, 1.0) {
		t.Error("chain of models should be coherent")
	}

	// Three models where one is isolated
	positions = []core.Position{{X: 0, Y: 0}, {X: 0.5, Y: 0}, {X: 10, Y: 0}}
	if UnitCoherencyValid(positions, 1.0) {
		t.Error("isolated model breaks coherency")
	}
}

func TestMoveDistanceValid(t *testing.T) {
	origin := core.Position{X: 0, Y: 0}

	if !MoveDistanceValid(origin, core.Position{X: 5, Y: 0}, 5.0) {
		t.Error("exact max move should be valid")
	}
	if !MoveDistanceValid(origin, core.Position{X: 3, Y: 0}, 5.0) {
		t.Error("move under max should be valid")
	}
	if MoveDistanceValid(origin, core.Position{X: 6, Y: 0}, 5.0) {
		t.Error("move over max should be invalid")
	}
}

func TestBoard_IsInBounds(t *testing.T) {
	b := NewBoard(48, 24)

	if !b.IsInBounds(core.Position{X: 0, Y: 0}) {
		t.Error("origin should be in bounds")
	}
	if !b.IsInBounds(core.Position{X: 24, Y: 12}) {
		t.Error("center should be in bounds")
	}
	if b.IsInBounds(core.Position{X: -1, Y: 0}) {
		t.Error("negative X should be out of bounds")
	}
	if b.IsInBounds(core.Position{X: 49, Y: 12}) {
		t.Error("past width should be out of bounds")
	}
}

func TestBoard_Clamp(t *testing.T) {
	b := NewBoard(48, 24)

	clamped := b.Clamp(core.Position{X: -5, Y: 30})
	if clamped.X != 0 || clamped.Y != 24 {
		t.Errorf("expected (0, 24), got (%f, %f)", clamped.X, clamped.Y)
	}
}
