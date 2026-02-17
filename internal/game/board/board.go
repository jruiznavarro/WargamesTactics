package board

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// Board represents the battlefield.
type Board struct {
	Width      float64           // Width in inches
	Height     float64           // Height in inches
	Terrain    []*TerrainFeature // Terrain features on the board
	Objectives []*Objective      // Scoring objectives on the board
}

// NewBoard creates a new board with the given dimensions.
func NewBoard(width, height float64) *Board {
	return &Board{
		Width:  width,
		Height: height,
	}
}

// AddTerrain adds a terrain feature to the board and returns it.
func (b *Board) AddTerrain(name string, terrainType TerrainType, pos core.Position, width, height float64) *TerrainFeature {
	t := &TerrainFeature{
		ID:     len(b.Terrain) + 1,
		Name:   name,
		Type:   terrainType,
		Pos:    pos,
		Width:  width,
		Height: height,
	}
	b.Terrain = append(b.Terrain, t)
	return t
}

// TerrainAt returns all terrain features that contain the given position.
func (b *Board) TerrainAt(pos core.Position) []*TerrainFeature {
	var result []*TerrainFeature
	for _, t := range b.Terrain {
		if t.Contains(pos) {
			result = append(result, t)
		}
	}
	return result
}

// HasTerrainType checks if any terrain of the given type exists at a position.
func (b *Board) HasTerrainType(pos core.Position, terrainType TerrainType) bool {
	for _, t := range b.Terrain {
		if t.Type == terrainType && t.Contains(pos) {
			return true
		}
	}
	return false
}

// IsInBounds checks if a position is within the board boundaries.
func (b *Board) IsInBounds(pos core.Position) bool {
	return pos.X >= 0 && pos.X <= b.Width &&
		pos.Y >= 0 && pos.Y <= b.Height
}

// AddObjective adds a scoring objective to the board and returns it.
func (b *Board) AddObjective(pos core.Position, radius float64) *Objective {
	o := &Objective{
		ID:       len(b.Objectives) + 1,
		Position: pos,
		Radius:   radius,
	}
	b.Objectives = append(b.Objectives, o)
	return o
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
