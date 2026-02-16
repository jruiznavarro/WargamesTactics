package board

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// TerrainType identifies the kind of terrain feature.
type TerrainType int

const (
	TerrainObstacle   TerrainType = iota // Low walls, barricades -- cover for shooting
	TerrainWoods                         // Dense forests -- cover + slows movement
	TerrainRuins                         // Ruined buildings -- cover, defensible
	TerrainImpassable                    // Lava, deep water, etc. -- blocks movement
	TerrainOpen                          // Open ground -- no effect (decorative)
)

func (t TerrainType) String() string {
	switch t {
	case TerrainObstacle:
		return "Obstacle"
	case TerrainWoods:
		return "Woods"
	case TerrainRuins:
		return "Ruins"
	case TerrainImpassable:
		return "Impassable"
	case TerrainOpen:
		return "Open"
	default:
		return "Unknown"
	}
}

// TerrainFeature represents a terrain piece on the battlefield.
// It is modeled as an axis-aligned rectangle for simplicity.
type TerrainFeature struct {
	ID     int
	Name   string
	Type   TerrainType
	Pos    core.Position // Top-left corner
	Width  float64       // Width in inches (along X)
	Height float64       // Height in inches (along Y)
}

// Contains returns true if the given position is inside this terrain feature.
func (t *TerrainFeature) Contains(pos core.Position) bool {
	return pos.X >= t.Pos.X && pos.X <= t.Pos.X+t.Width &&
		pos.Y >= t.Pos.Y && pos.Y <= t.Pos.Y+t.Height
}

// Center returns the center point of the terrain feature.
func (t *TerrainFeature) Center() core.Position {
	return core.Position{
		X: t.Pos.X + t.Width/2,
		Y: t.Pos.Y + t.Height/2,
	}
}

// Symbol returns a character for minimap display.
func (t *TerrainFeature) Symbol() rune {
	switch t.Type {
	case TerrainObstacle:
		return '#'
	case TerrainWoods:
		return 'W'
	case TerrainRuins:
		return 'R'
	case TerrainImpassable:
		return 'X'
	case TerrainOpen:
		return '~'
	default:
		return '?'
	}
}
