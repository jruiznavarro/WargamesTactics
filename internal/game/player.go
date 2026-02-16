package game

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// Player is the interface that both human and AI players implement.
type Player interface {
	// GetNextCommand asks the player for their next command during a phase.
	// Returning nil signals the player wants to end the phase.
	GetNextCommand(view *GameView, currentPhase phase.Phase) interface{}
	// ID returns the player's unique identifier.
	ID() int
	// Name returns the player's display name.
	Name() string
}

// TerrainView is a read-only view of a terrain feature.
type TerrainView struct {
	Name   string
	Type   string     // "Obstacle", "Woods", "Ruins", "Impassable", "Open"
	Symbol rune
	Pos    [2]float64 // Top-left X, Y
	Width  float64
	Height float64
}

// GameView provides a read-only snapshot of the game state for a specific player.
type GameView struct {
	Units        map[int][]UnitView // Units by owner player ID
	Terrain      []TerrainView
	BoardWidth   float64
	BoardHeight  float64
	CurrentPhase phase.PhaseType
	BattleRound  int
	ActivePlayer int
}

// UnitView is a read-only view of a unit for display/AI purposes.
type UnitView struct {
	ID            int
	Name          string
	OwnerID       int
	Position      [2]float64 // X, Y
	AliveModels   int
	TotalModels   int
	CurrentWounds int
	MaxWounds     int
	MoveSpeed     int
	Save          int
	Weapons       []WeaponView
	StrikeOrder   core.StrikeOrder
	HasMoved      bool
	HasShot       bool
	HasFought     bool
	HasCharged    bool
	HasPiledIn    bool
	IsEngaged     bool
}

// WeaponView is a read-only view of a weapon.
type WeaponView struct {
	Name    string
	Range   int
	Attacks int
	ToHit   int
	ToWound int
	Rend    int
	Damage  int
}

// AllowedCommands returns the command types valid for the current phase.
func (v *GameView) AllowedCommands() []command.CommandType {
	p := phase.Phase{Type: v.CurrentPhase}
	// Rebuild allowed commands from phase type
	for _, sp := range phase.StandardTurnSequence() {
		if sp.Type == v.CurrentPhase {
			p = sp
			break
		}
	}
	return p.AllowedCommands
}
