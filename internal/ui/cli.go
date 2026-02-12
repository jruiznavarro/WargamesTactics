package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// CLIPlayer implements the Player interface for human input via CLI.
type CLIPlayer struct {
	id     int
	name   string
	reader *bufio.Reader
	writer io.Writer
}

// NewCLIPlayer creates a new CLI player.
func NewCLIPlayer(id int, name string) *CLIPlayer {
	return &CLIPlayer{
		id:     id,
		name:   name,
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// NewCLIPlayerWithIO creates a CLI player with custom I/O (useful for testing).
func NewCLIPlayerWithIO(id int, name string, reader io.Reader, writer io.Writer) *CLIPlayer {
	return &CLIPlayer{
		id:     id,
		name:   name,
		reader: bufio.NewReader(reader),
		writer: writer,
	}
}

func (p *CLIPlayer) ID() int      { return p.id }
func (p *CLIPlayer) Name() string { return p.name }

func (p *CLIPlayer) GetNextCommand(view *game.GameView, currentPhase phase.Phase) interface{} {
	p.displayState(view)
	p.displayPrompt(currentPhase)

	for {
		fmt.Fprint(p.writer, "> ")
		line, err := p.reader.ReadString('\n')
		if err != nil {
			return nil
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		cmd, parseErr := p.parseCommand(line, currentPhase)
		if parseErr != nil {
			fmt.Fprintf(p.writer, "Error: %s\n", parseErr)
			continue
		}
		return cmd
	}
}

func (p *CLIPlayer) displayState(view *game.GameView) {
	fmt.Fprintf(p.writer, "\n=== Battle Round %d - %s ===\n", view.BattleRound, view.CurrentPhase)
	fmt.Fprintf(p.writer, "\nYour units:\n")
	if units, ok := view.Units[p.id]; ok {
		for _, u := range units {
			status := ""
			if u.HasMoved {
				status += " [moved]"
			}
			if u.HasShot {
				status += " [shot]"
			}
			if u.HasFought {
				status += " [fought]"
			}
			fmt.Fprintf(p.writer, "  [%d] %s (%d/%d models) at (%.1f, %.1f) - Move: %d\" - Wounds: %d/%d%s\n",
				u.ID, u.Name, u.AliveModels, u.TotalModels, u.Position[0], u.Position[1],
				u.MoveSpeed, u.CurrentWounds, u.MaxWounds, status)
		}
	}

	fmt.Fprintf(p.writer, "\nEnemy units:\n")
	for ownerID, units := range view.Units {
		if ownerID == p.id {
			continue
		}
		for _, u := range units {
			fmt.Fprintf(p.writer, "  [%d] %s (%d/%d models) at (%.1f, %.1f) - Wounds: %d/%d\n",
				u.ID, u.Name, u.AliveModels, u.TotalModels, u.Position[0], u.Position[1],
				u.CurrentWounds, u.MaxWounds)
		}
	}
}

func (p *CLIPlayer) displayPrompt(currentPhase phase.Phase) {
	fmt.Fprintf(p.writer, "\nCommands:")
	for _, ct := range currentPhase.AllowedCommands {
		switch ct {
		case command.CommandTypeMove:
			fmt.Fprintf(p.writer, " move <unit_id> <x> <y> |")
		case command.CommandTypeShoot:
			fmt.Fprintf(p.writer, " shoot <unit_id> <target_id> |")
		case command.CommandTypeFight:
			fmt.Fprintf(p.writer, " fight <unit_id> <target_id> |")
		case command.CommandTypeCharge:
			fmt.Fprintf(p.writer, " charge <unit_id> <target_id> |")
		case command.CommandTypeEndPhase:
			fmt.Fprintf(p.writer, " skip |")
		}
	}
	fmt.Fprintf(p.writer, " help\n")
}

func (p *CLIPlayer) parseCommand(line string, currentPhase phase.Phase) (interface{}, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	switch parts[0] {
	case "move":
		if !currentPhase.IsCommandAllowed(command.CommandTypeMove) {
			return nil, fmt.Errorf("move not allowed in %s", currentPhase.Type)
		}
		if len(parts) != 4 {
			return nil, fmt.Errorf("usage: move <unit_id> <x> <y>")
		}
		unitID, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid unit ID: %s", parts[1])
		}
		x, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid X coordinate: %s", parts[2])
		}
		y, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Y coordinate: %s", parts[3])
		}
		return &command.MoveCommand{
			OwnerID:     p.id,
			UnitID:      core.UnitID(unitID),
			Destination: core.Position{X: x, Y: y},
		}, nil

	case "shoot":
		if !currentPhase.IsCommandAllowed(command.CommandTypeShoot) {
			return nil, fmt.Errorf("shoot not allowed in %s", currentPhase.Type)
		}
		if len(parts) != 3 {
			return nil, fmt.Errorf("usage: shoot <unit_id> <target_id>")
		}
		shooterID, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid shooter ID: %s", parts[1])
		}
		targetID, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid target ID: %s", parts[2])
		}
		return &command.ShootCommand{
			OwnerID:   p.id,
			ShooterID: core.UnitID(shooterID),
			TargetID:  core.UnitID(targetID),
		}, nil

	case "fight":
		if !currentPhase.IsCommandAllowed(command.CommandTypeFight) {
			return nil, fmt.Errorf("fight not allowed in %s", currentPhase.Type)
		}
		if len(parts) != 3 {
			return nil, fmt.Errorf("usage: fight <unit_id> <target_id>")
		}
		attackerID, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid attacker ID: %s", parts[1])
		}
		targetID, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid target ID: %s", parts[2])
		}
		return &command.FightCommand{
			OwnerID:    p.id,
			AttackerID: core.UnitID(attackerID),
			TargetID:   core.UnitID(targetID),
		}, nil

	case "charge":
		if !currentPhase.IsCommandAllowed(command.CommandTypeCharge) {
			return nil, fmt.Errorf("charge not allowed in %s", currentPhase.Type)
		}
		if len(parts) != 3 {
			return nil, fmt.Errorf("usage: charge <unit_id> <target_id>")
		}
		chargerID, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid charger ID: %s", parts[1])
		}
		targetID, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid target ID: %s", parts[2])
		}
		return &command.ChargeCommand{
			OwnerID:   p.id,
			ChargerID: core.UnitID(chargerID),
			TargetID:  core.UnitID(targetID),
		}, nil

	case "skip", "end", "done":
		return &command.EndPhaseCommand{OwnerID: p.id}, nil

	case "help":
		fmt.Fprintf(p.writer, "\nAvailable commands:\n")
		fmt.Fprintf(p.writer, "  move <unit_id> <x> <y>     - Move a unit to position (x, y)\n")
		fmt.Fprintf(p.writer, "  shoot <unit_id> <target_id> - Shoot at target with unit\n")
		fmt.Fprintf(p.writer, "  fight <unit_id> <target_id> - Melee attack target with unit\n")
		fmt.Fprintf(p.writer, "  charge <unit_id> <target_id> - Declare a charge\n")
		fmt.Fprintf(p.writer, "  skip                        - End current phase\n")
		fmt.Fprintf(p.writer, "  help                        - Show this help\n")
		return nil, fmt.Errorf("") // Not a real error, just re-prompt

	default:
		return nil, fmt.Errorf("unknown command: %s (type 'help' for available commands)", parts[0])
	}
}
