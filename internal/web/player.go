package web

import (
	"encoding/json"

	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// WebPlayer implements the Player interface for browser-based interaction.
// It communicates via channels: sending game state out and receiving commands in.
type WebPlayer struct {
	id   int
	name string

	// stateCh carries the game view + phase to the WebSocket handler
	stateCh chan *PlayerPrompt

	// commandCh receives parsed commands from the WebSocket handler
	commandCh chan interface{}
}

// PlayerPrompt bundles the game view with the current phase for the web client.
type PlayerPrompt struct {
	View  *game.GameView `json:"view"`
	Phase phase.Phase    `json:"phase"`
}

func NewWebPlayer(id int, name string) *WebPlayer {
	return &WebPlayer{
		id:        id,
		name:      name,
		stateCh:   make(chan *PlayerPrompt, 1),
		commandCh: make(chan interface{}, 1),
	}
}

func (p *WebPlayer) ID() int      { return p.id }
func (p *WebPlayer) Name() string { return p.name }

// GetNextCommand blocks until the web client sends a command.
// It first publishes the current game state for the client to render.
func (p *WebPlayer) GetNextCommand(view *game.GameView, currentPhase phase.Phase) interface{} {
	// Send the state to the WebSocket handler
	p.stateCh <- &PlayerPrompt{View: view, Phase: currentPhase}

	// Block waiting for a command from the client
	cmd, ok := <-p.commandCh
	if !ok {
		return nil
	}
	return cmd
}

// WSCommand is the JSON structure received from the web client.
type WSCommand struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ParseWSCommand converts a WSCommand into a game command object.
func ParseWSCommand(playerID int, msg WSCommand) interface{} {
	switch msg.Type {
	case "move":
		var data struct {
			UnitID int     `json:"unitId"`
			X      float64 `json:"x"`
			Y      float64 `json:"y"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.MoveCommand{
			OwnerID:     playerID,
			UnitID:      core.UnitID(data.UnitID),
			Destination: core.Position{X: data.X, Y: data.Y},
		}

	case "run":
		var data struct {
			UnitID int     `json:"unitId"`
			X      float64 `json:"x"`
			Y      float64 `json:"y"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.RunCommand{
			OwnerID:     playerID,
			UnitID:      core.UnitID(data.UnitID),
			Destination: core.Position{X: data.X, Y: data.Y},
		}

	case "retreat":
		var data struct {
			UnitID int     `json:"unitId"`
			X      float64 `json:"x"`
			Y      float64 `json:"y"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.RetreatCommand{
			OwnerID:     playerID,
			UnitID:      core.UnitID(data.UnitID),
			Destination: core.Position{X: data.X, Y: data.Y},
		}

	case "shoot":
		var data struct {
			ShooterID int `json:"shooterId"`
			TargetID  int `json:"targetId"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.ShootCommand{
			OwnerID:   playerID,
			ShooterID: core.UnitID(data.ShooterID),
			TargetID:  core.UnitID(data.TargetID),
		}

	case "fight":
		var data struct {
			AttackerID int `json:"attackerId"`
			TargetID   int `json:"targetId"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.FightCommand{
			OwnerID:    playerID,
			AttackerID: core.UnitID(data.AttackerID),
			TargetID:   core.UnitID(data.TargetID),
		}

	case "charge":
		var data struct {
			ChargerID int `json:"chargerId"`
			TargetID  int `json:"targetId"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.ChargeCommand{
			OwnerID:   playerID,
			ChargerID: core.UnitID(data.ChargerID),
			TargetID:  core.UnitID(data.TargetID),
		}

	case "pilein":
		var data struct {
			UnitID int `json:"unitId"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.PileInCommand{
			OwnerID: playerID,
			UnitID:  core.UnitID(data.UnitID),
		}

	case "cast":
		var data struct {
			CasterID   int `json:"casterId"`
			SpellIndex int `json:"spellIndex"`
			TargetID   int `json:"targetId"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.CastCommand{
			OwnerID:    playerID,
			CasterID:   core.UnitID(data.CasterID),
			SpellIndex: data.SpellIndex,
			TargetID:   core.UnitID(data.TargetID),
		}

	case "chant":
		var data struct {
			ChanterID   int `json:"chanterId"`
			PrayerIndex int `json:"prayerIndex"`
			TargetID    int `json:"targetId"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.ChantCommand{
			OwnerID:     playerID,
			ChanterID:   core.UnitID(data.ChanterID),
			PrayerIndex: data.PrayerIndex,
			TargetID:    core.UnitID(data.TargetID),
		}

	case "rally":
		var data struct {
			UnitID int `json:"unitId"`
		}
		if json.Unmarshal(msg.Data, &data) != nil {
			return nil
		}
		return &command.RallyCommand{
			OwnerID: playerID,
			UnitID:  core.UnitID(data.UnitID),
		}

	case "skip", "end_phase":
		return &command.EndPhaseCommand{OwnerID: playerID}

	default:
		return nil
	}
}
