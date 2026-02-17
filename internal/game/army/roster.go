package army

import (
	"fmt"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// GH 2025-26 matched play army composition rules.
const (
	DefaultPointsLimit = 2000 // Standard matched play
	MaxHeroes          = 6    // Maximum Hero units
	MaxReinforcements  = 4    // Maximum reinforced units
	GeneralRequired    = true // Must designate a general
)

// RosterEntry represents a single unit selection in an army roster.
type RosterEntry struct {
	WarscrollID string `json:"warscrollId"` // Reference to warscroll
	Reinforced  bool   `json:"reinforced"`  // If true, uses maxSize instead of unitSize
	IsGeneral   bool   `json:"isGeneral"`   // Designated as the army general
}

// ArmyRoster represents a complete army list for one player.
type ArmyRoster struct {
	FactionID   string        `json:"factionId"`   // Faction this army belongs to
	Entries     []RosterEntry `json:"entries"`      // Unit selections
	PointsLimit int           `json:"pointsLimit"`  // Max points (default 2000)
}

// Validate checks the roster against matched play composition rules.
// Returns nil if valid, or a list of validation errors.
func (r *ArmyRoster) Validate(faction *Faction) []error {
	var errs []error

	if r.PointsLimit <= 0 {
		r.PointsLimit = DefaultPointsLimit
	}

	totalPoints := 0
	heroCount := 0
	reinforcedCount := 0
	generalCount := 0
	uniqueUsed := make(map[string]bool)

	for i, entry := range r.Entries {
		ws := faction.GetWarscroll(entry.WarscrollID)
		if ws == nil {
			errs = append(errs, fmt.Errorf("entry %d: unknown warscroll '%s'", i, entry.WarscrollID))
			continue
		}

		// Points
		points := ws.Points
		if entry.Reinforced {
			if ws.MaxSize == 0 {
				errs = append(errs, fmt.Errorf("entry %d: '%s' cannot be reinforced", i, ws.Name))
			}
			// Reinforced units cost double
			points *= 2
			reinforcedCount++
		}
		totalPoints += points

		// Heroes
		if ws.HasKeyword("Hero") {
			heroCount++
		}

		// General
		if entry.IsGeneral {
			generalCount++
			if !ws.HasKeyword("Hero") {
				errs = append(errs, fmt.Errorf("entry %d: general '%s' must be a Hero", i, ws.Name))
			}
		}

		// Unique
		if ws.Unique {
			if uniqueUsed[ws.ID] {
				errs = append(errs, fmt.Errorf("entry %d: unique unit '%s' can only be taken once", i, ws.Name))
			}
			uniqueUsed[ws.ID] = true
		}
	}

	// Total points
	if totalPoints > r.PointsLimit {
		errs = append(errs, fmt.Errorf("army costs %d points, exceeds limit of %d", totalPoints, r.PointsLimit))
	}

	// Hero limit
	if heroCount > MaxHeroes {
		errs = append(errs, fmt.Errorf("too many heroes: %d (max %d)", heroCount, MaxHeroes))
	}

	// Reinforcement limit
	if reinforcedCount > MaxReinforcements {
		errs = append(errs, fmt.Errorf("too many reinforced units: %d (max %d)", reinforcedCount, MaxReinforcements))
	}

	// General requirement
	if len(r.Entries) > 0 && generalCount == 0 {
		errs = append(errs, fmt.Errorf("army must designate a general"))
	}
	if generalCount > 1 {
		errs = append(errs, fmt.Errorf("only 1 general allowed, got %d", generalCount))
	}

	return errs
}

// TotalPoints calculates the total points cost of the roster.
func (r *ArmyRoster) TotalPoints(faction *Faction) int {
	total := 0
	for _, entry := range r.Entries {
		ws := faction.GetWarscroll(entry.WarscrollID)
		if ws == nil {
			continue
		}
		pts := ws.Points
		if entry.Reinforced {
			pts *= 2
		}
		total += pts
	}
	return total
}

// BuildUnits creates core.Unit instances from the roster entries.
// Each entry produces one unit at the given starting positions.
func (r *ArmyRoster) BuildUnits(faction *Faction, ownerID int, positions []core.Position) []*UnitSpec {
	var specs []*UnitSpec
	for i, entry := range r.Entries {
		ws := faction.GetWarscroll(entry.WarscrollID)
		if ws == nil {
			continue
		}

		numModels := ws.UnitSize
		if entry.Reinforced && ws.MaxSize > 0 {
			numModels = ws.MaxSize
		}

		pos := core.Position{X: 0, Y: 0}
		if i < len(positions) {
			pos = positions[i]
		}

		spec := &UnitSpec{
			Warscroll:  ws,
			NumModels:  numModels,
			Position:   pos,
			OwnerID:    ownerID,
			IsGeneral:  entry.IsGeneral,
			Reinforced: entry.Reinforced,
		}
		specs = append(specs, spec)
	}
	return specs
}

// UnitSpec is a fully resolved unit ready to be created in the game.
type UnitSpec struct {
	Warscroll  *Warscroll
	NumModels  int
	Position   core.Position
	OwnerID    int
	IsGeneral  bool
	Reinforced bool
}

// CreateUnit creates a core.Unit from this spec using the game's CreateUnit interface.
// The wardSave and special abilities from the warscroll are applied after creation.
func (s *UnitSpec) ToUnitParams() (name string, ownerID int, stats core.Stats, weapons []core.Weapon, numModels int, position core.Position, baseSize float64) {
	ws := s.Warscroll
	return ws.Name, s.OwnerID, ws.ToCoreStats(), ws.ToCoreWeapons(), s.NumModels, s.Position, ws.BaseSizeInches()
}

// ApplyToUnit applies warscroll-level attributes (keywords, ward, spells, etc.) to a created unit.
func (s *UnitSpec) ApplyToUnit(u *core.Unit) {
	ws := s.Warscroll
	u.Keywords = ws.ToCoreKeywords()
	u.WardSave = ws.WardSave
	u.PowerLevel = ws.PowerLevel
	u.Spells = ws.ToCoreSpells()
	u.Prayers = ws.ToCorePrayers()

	// Apply ability effects
	for _, ab := range ws.Abilities {
		switch ab.Effect {
		case "ward":
			if ab.Value > 0 && (u.WardSave == 0 || ab.Value < u.WardSave) {
				u.WardSave = ab.Value
			}
		case "strikeFirst":
			u.StrikeOrder = core.StrikeFirst
		case "strikeLast":
			u.StrikeOrder = core.StrikeLast
		}
	}
}
