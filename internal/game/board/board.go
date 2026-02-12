package board

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// Board represents the battlefield.
type Board struct {
	Width  float64 // Width in inches
	Height float64 // Height in inches
}

// NewBoard creates a new board with the given dimensions.
func NewBoard(width, height float64) *Board {
	return &Board{
		Width:  width,
		Height: height,
	}
}

// IsInBounds checks if a position is within the board boundaries.
func (b *Board) IsInBounds(pos core.Position) bool {
	return pos.X >= 0 && pos.X <= b.Width &&
		pos.Y >= 0 && pos.Y <= b.Height
}

// Clamp restricts a position to be within board boundaries.
func (b *Board) Clamp(pos core.Position) core.Position {
	result := pos
	if result.X < 0 {
		result.X = 0
	}
	if result.X > b.Width {
		result.X = b.Width
	}
	if result.Y < 0 {
		result.Y = 0
	}
	if result.Y > b.Height {
		result.Y = b.Height
	}
	return result
}
