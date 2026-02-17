package game

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// Player is the interface that both human and AI players implement.
type Player interface {
	GetNextCommand(view *GameView, currentPhase phase.Phase) interface{}
	ID() int
	Name() string
}

// TerrainView is a read-only view of a terrain feature.
type TerrainView struct {
	Name   string
	Type   string
	Symbol rune
	Pos    [2]float64
	Width  float64
	Height float64
}

// ObjectiveView is a read-only view of an objective.
type ObjectiveView struct {
	ID           int
	Position     [2]float64
	Radius       float64
	ControlledBy int // Player ID, -1 if uncontrolled
}

// GameView provides a read-only snapshot of the game state.
type GameView struct {
	Units           map[int][]UnitView
	Terrain         []TerrainView
	Objectives      []ObjectiveView
	BoardWidth      float64
	BoardHeight     float64
	CurrentPhase    phase.PhaseType
	BattleRound     int
	MaxBattleRounds int
	ActivePlayer    int
	CommandPoints   map[int]int // CP remaining per player ID
	VictoryPoints   map[int]int // VP per player ID
}

// UnitView is a read-only view of a unit.
type UnitView struct {
	ID            int
	Name          string
	OwnerID       int
	Position      [2]float64
	AliveModels   int
	TotalModels   int
	CurrentWounds int
	MaxWounds     int
	MoveSpeed     int
	Save          int
	WardSave      int
	Weapons       []WeaponView
	StrikeOrder   core.StrikeOrder
	HasMoved      bool
	HasRun        bool
	HasRetreated  bool
	HasShot       bool
	HasFought     bool
	HasCharged    bool
	HasPiledIn    bool
	IsEngaged     bool
	Spells        []SpellView
	Prayers       []PrayerView
	CanCast       bool
	CanChant      bool
}

// WeaponView is a read-only view of a weapon.
type WeaponView struct {
	Name      string
	Range     int
	Attacks   int
	ToHit     int
	ToWound   int
	Rend      int
	Damage    int
	Abilities core.WeaponAbility
}

// SpellView is a read-only view of a spell.
type SpellView struct {
	Name         string
	CastingValue int
	Range        int
}

// PrayerView is a read-only view of a prayer.
type PrayerView struct {
	Name          string
	ChantingValue int
	Range         int
}

// AllowedCommands returns the command types valid for the current phase.
func (v *GameView) AllowedCommands() []command.CommandType {
	p := phase.Phase{Type: v.CurrentPhase}
	for _, sp := range phase.StandardTurnSequence() {
		if sp.Type == v.CurrentPhase {
			p = sp
			break
		}
	}
	return p.AllowedCommands
}
