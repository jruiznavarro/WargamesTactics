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

// AddGhyraniteObjective adds a Ghyranite objective (GH 2025-26) with subtype and pair ID.
// Standard control zone is 3" for Ghyranite objectives.
func (b *Board) AddGhyraniteObjective(pos core.Position, ghyraniteType GhyraniteType, pairID int) *Objective {
	o := &Objective{
		ID:            len(b.Objectives) + 1,
		Position:      pos,
		Radius:        3.0, // GH 2025-26: 40mm base, 3" control zone
		GhyraniteType: ghyraniteType,
		PairID:        pairID,
	}
	b.Objectives = append(b.Objectives, o)
	return o
}

// GhyraniteObjectives returns all Ghyranite objectives on the board.
func (b *Board) GhyraniteObjectives() []*Objective {
	var result []*Objective
	for _, o := range b.Objectives {
		if o.IsGhyranite() {
			result = append(result, o)
		}
	}
	return result
}

// ObjectivePair returns both objectives belonging to a given pair ID, or nil if not found.
func (b *Board) ObjectivePair(pairID int) []*Objective {
	var result []*Objective
	for _, o := range b.Objectives {
		if o.PairID == pairID {
			result = append(result, o)
		}
	}
	return result
}

// PairIDs returns all unique pair IDs present on the board.
func (b *Board) PairIDs() []int {
	seen := make(map[int]bool)
	var ids []int
	for _, o := range b.Objectives {
		if o.PairID > 0 && !seen[o.PairID] {
			seen[o.PairID] = true
			ids = append(ids, o.PairID)
		}
	}
	return ids
}

// IsVisible checks if two positions have line of sight to each other.
// AoS4 Rule 6.0 / 7.0 (Errata Jan 2026): visibility requires no impassable terrain
// directly between the two positions. Simplified LOS check.
func (b *Board) IsVisible(from, to core.Position) bool {
	for _, t := range b.Terrain {
		if t.Type != TerrainImpassable {
			continue
		}
		if lineIntersectsRect(from, to, t.Pos, t.Width, t.Height) {
			return false
		}
	}
	return true
}

// lineIntersectsRect checks if a line segment from p1 to p2 passes through a rectangle.
func lineIntersectsRect(p1, p2 core.Position, rectPos core.Position, rectW, rectH float64) bool {
	// Check if either point is inside the rectangle
	if pointInRect(p1, rectPos, rectW, rectH) || pointInRect(p2, rectPos, rectW, rectH) {
		return true
	}

	// Check line segment against each edge of the rectangle
	r1 := rectPos
	r2 := core.Position{X: rectPos.X + rectW, Y: rectPos.Y}
	r3 := core.Position{X: rectPos.X + rectW, Y: rectPos.Y + rectH}
	r4 := core.Position{X: rectPos.X, Y: rectPos.Y + rectH}

	return segmentsIntersect(p1, p2, r1, r2) ||
		segmentsIntersect(p1, p2, r2, r3) ||
		segmentsIntersect(p1, p2, r3, r4) ||
		segmentsIntersect(p1, p2, r4, r1)
}

func pointInRect(p core.Position, rectPos core.Position, w, h float64) bool {
	return p.X >= rectPos.X && p.X <= rectPos.X+w &&
		p.Y >= rectPos.Y && p.Y <= rectPos.Y+h
}

// segmentsIntersect checks if line segment (a1,a2) intersects (b1,b2).
func segmentsIntersect(a1, a2, b1, b2 core.Position) bool {
	d1 := cross(b1, b2, a1)
	d2 := cross(b1, b2, a2)
	d3 := cross(a1, a2, b1)
	d4 := cross(a1, a2, b2)

	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}
	return false
}

func cross(a, b, c core.Position) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
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
