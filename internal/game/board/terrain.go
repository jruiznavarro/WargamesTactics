package board

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// TerrainType identifies terrain types per AoS4 Rule 1.4.
type TerrainType int

const (
	TerrainObstacle     TerrainType = iota // Ruins, debris, barricades: Cover + Unstable
	TerrainObscuring                       // Wyldwood, fortress wall: Cover + Obscuring + Unstable
	TerrainArea                            // Hills, Stormvault: Cover only
	TerrainPlaceOfPower                    // Realmgate, Aqualith: Cover + Place of Power + Unstable
	TerrainImpassable                      // Custom: blocks all movement
	TerrainOpen                            // Open ground: no effect (decorative)
)

func (t TerrainType) String() string {
	switch t {
	case TerrainObstacle:
		return "Obstacle"
	case TerrainObscuring:
		return "Obscuring"
	case TerrainArea:
		return "Area"
	case TerrainPlaceOfPower:
		return "Place of Power"
	case TerrainImpassable:
		return "Impassable"
	case TerrainOpen:
		return "Open"
	default:
		return "Unknown"
	}
}

// TerrainFeature represents a terrain piece on the battlefield.
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
	case TerrainObscuring:
		return 'W'
	case TerrainArea:
		return 'H'
	case TerrainPlaceOfPower:
		return 'P'
	case TerrainImpassable:
		return 'X'
	case TerrainOpen:
		return '~'
	default:
		return '?'
	}
}
