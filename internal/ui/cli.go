package ui

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

const (
	mapWidth  = 60 // characters wide
	mapHeight = 18 // characters tall
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
	p.displayHeader(view)
	p.displayMap(view)
	p.displayUnits(view)
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

		cmd, parseErr := p.parseCommand(line, currentPhase, view)
		if parseErr != nil {
			if parseErr.Error() != "" {
				fmt.Fprintf(p.writer, "  Error: %s\n", parseErr)
			}
			continue
		}
		return cmd
	}
}

// --- Display functions ---

func (p *CLIPlayer) displayHeader(view *game.GameView) {
	fmt.Fprintf(p.writer, "\n")
	fmt.Fprintf(p.writer, "+------------------------------------------------------------+\n")
	fmt.Fprintf(p.writer, "|  BATTLE ROUND %-2d                      %18s  |\n", view.BattleRound, view.CurrentPhase)
	fmt.Fprintf(p.writer, "+------------------------------------------------------------+\n")
}

func (p *CLIPlayer) displayMap(view *game.GameView) {
	// Build a grid representing the board
	grid := make([][]rune, mapHeight)
	for y := range grid {
		grid[y] = make([]rune, mapWidth)
		for x := range grid[y] {
			grid[y][x] = '.'
		}
	}

	// Collect all units sorted by ID for stable rendering
	type unitInfo struct {
		view  game.UnitView
		label rune
		mine  bool
	}
	var allUnits []unitInfo

	for ownerID, units := range view.Units {
		for _, u := range units {
			label := rune('0' + u.ID)
			if u.ID > 9 {
				label = rune('A' + u.ID - 10)
			}
			allUnits = append(allUnits, unitInfo{u, label, ownerID == p.id})
		}
	}
	sort.Slice(allUnits, func(i, j int) bool { return allUnits[i].view.ID < allUnits[j].view.ID })

	// Place units on grid
	for _, ui := range allUnits {
		gx := int(math.Round(ui.view.Position[0] / view.BoardWidth * float64(mapWidth-1)))
		gy := int(math.Round(ui.view.Position[1] / view.BoardHeight * float64(mapHeight-1)))
		if gx < 0 {
			gx = 0
		}
		if gx >= mapWidth {
			gx = mapWidth - 1
		}
		if gy < 0 {
			gy = 0
		}
		if gy >= mapHeight {
			gy = mapHeight - 1
		}
		grid[gy][gx] = ui.label
	}

	// Render
	fmt.Fprintf(p.writer, "\n  Battlefield (%.0f\" x %.0f\"):\n", view.BoardWidth, view.BoardHeight)
	fmt.Fprintf(p.writer, "  +%s+\n", strings.Repeat("-", mapWidth))
	for y := 0; y < mapHeight; y++ {
		fmt.Fprintf(p.writer, "  |%s|\n", string(grid[y]))
	}
	fmt.Fprintf(p.writer, "  +%s+\n", strings.Repeat("-", mapWidth))

	// Legend
	fmt.Fprintf(p.writer, "  Legend:")
	for _, ui := range allUnits {
		tag := "enemy"
		if ui.mine {
			tag = "you"
		}
		fmt.Fprintf(p.writer, " %c=%s(%s)", ui.label, ui.view.Name, tag)
	}
	fmt.Fprintf(p.writer, "\n")
}

func (p *CLIPlayer) displayUnits(view *game.GameView) {
	fmt.Fprintf(p.writer, "\n  YOUR ARMY:\n")
	if units, ok := view.Units[p.id]; ok {
		for _, u := range units {
			p.displayUnitCard(u, true)
		}
	}

	fmt.Fprintf(p.writer, "\n  ENEMY ARMY:\n")
	for ownerID, units := range view.Units {
		if ownerID == p.id {
			continue
		}
		for _, u := range units {
			p.displayUnitCard(u, false)
		}
	}
}

func (p *CLIPlayer) displayUnitCard(u game.UnitView, showDetails bool) {
	// Status flags
	var flags []string
	if u.HasMoved {
		flags = append(flags, "MOVED")
	}
	if u.HasShot {
		flags = append(flags, "SHOT")
	}
	if u.HasFought {
		flags = append(flags, "FOUGHT")
	}
	if u.HasCharged {
		flags = append(flags, "CHARGED")
	}
	statusStr := ""
	if len(flags) > 0 {
		statusStr = " [" + strings.Join(flags, ",") + "]"
	}

	// Health bar
	healthBar := renderBar(u.CurrentWounds, u.MaxWounds, 15)

	fmt.Fprintf(p.writer, "  +----------------------------------------------+\n")
	fmt.Fprintf(p.writer, "  | [%d] %-20s  Models: %d/%-2d     |\n", u.ID, u.Name, u.AliveModels, u.TotalModels)
	fmt.Fprintf(p.writer, "  | HP: %s %d/%-3d           |\n", healthBar, u.CurrentWounds, u.MaxWounds)
	fmt.Fprintf(p.writer, "  | Pos: (%.1f, %.1f)  Move: %d\"  Save: %d+%s\n",
		u.Position[0], u.Position[1], u.MoveSpeed, u.Save, statusStr)

	if showDetails && len(u.Weapons) > 0 {
		fmt.Fprintf(p.writer, "  | Weapons:\n")
		for _, w := range u.Weapons {
			rangeStr := "Melee"
			if w.Range > 0 {
				rangeStr = fmt.Sprintf("%d\"", w.Range)
			}
			rendStr := ""
			if w.Rend != 0 {
				rendStr = fmt.Sprintf(" Rend:%d", w.Rend)
			}
			fmt.Fprintf(p.writer, "  |   %-14s %5s  A:%d  %d+/%d+  D:%d%s\n",
				w.Name, rangeStr, w.Attacks, w.ToHit, w.ToWound, w.Damage, rendStr)
		}
	}
	fmt.Fprintf(p.writer, "  +----------------------------------------------+\n")
}

func (p *CLIPlayer) displayPrompt(currentPhase phase.Phase) {
	fmt.Fprintf(p.writer, "\n  Commands:")
	for _, ct := range currentPhase.AllowedCommands {
		switch ct {
		case command.CommandTypeMove:
			fmt.Fprintf(p.writer, " move <id> <x> <y>")
		case command.CommandTypeShoot:
			fmt.Fprintf(p.writer, " shoot <id> <target>")
		case command.CommandTypeFight:
			fmt.Fprintf(p.writer, " fight <id> <target>")
		case command.CommandTypeCharge:
			fmt.Fprintf(p.writer, " charge <id> <target>")
		case command.CommandTypeEndPhase:
			fmt.Fprintf(p.writer, " skip")
		}
	}
	fmt.Fprintf(p.writer, " | map | help\n")
}

// --- Parsing ---

func (p *CLIPlayer) parseCommand(line string, currentPhase phase.Phase, view *game.GameView) (interface{}, error) {
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

	case "map":
		p.displayMap(view)
		return nil, fmt.Errorf("")

	case "help":
		fmt.Fprintf(p.writer, "\n  Available commands:\n")
		fmt.Fprintf(p.writer, "    move <unit_id> <x> <y>      Move unit to position\n")
		fmt.Fprintf(p.writer, "    shoot <unit_id> <target_id>  Shoot at enemy unit\n")
		fmt.Fprintf(p.writer, "    fight <unit_id> <target_id>  Melee attack enemy unit\n")
		fmt.Fprintf(p.writer, "    charge <unit_id> <target_id> Declare a charge\n")
		fmt.Fprintf(p.writer, "    skip                         End current phase\n")
		fmt.Fprintf(p.writer, "    map                          Show battlefield map\n")
		fmt.Fprintf(p.writer, "    help                         Show this help\n")
		return nil, fmt.Errorf("")

	default:
		return nil, fmt.Errorf("unknown command: %s (type 'help' for available commands)", parts[0])
	}
}

// --- Rendering helpers ---

func renderBar(current, max, width int) string {
	if max <= 0 {
		return "[" + strings.Repeat(" ", width) + "]"
	}
	filled := current * width / max
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	empty := width - filled
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", empty) + "]"
}

// FormatCombatLog renders a combat result as a visual block for the game log.
func FormatCombatLog(attackerName, defenderName string, weaponName string, attacks, hits, wounds, savesFailed, damage, slain int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("    %s attacks %s with %s\n", attackerName, defenderName, weaponName))

	// Dice pipeline visualization
	sb.WriteString(fmt.Sprintf("    Attacks: %-3d ", attacks))
	sb.WriteString(renderPipeline(attacks, hits, "HIT"))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("    Hits:    %-3d ", hits))
	sb.WriteString(renderPipeline(hits, wounds, "WND"))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("    Wounds:  %-3d ", wounds))
	sb.WriteString(renderPipeline(wounds, savesFailed, "SAV"))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("    Unsaved: %-3d x Damage -> %d total", savesFailed, damage))
	if slain > 0 {
		sb.WriteString(fmt.Sprintf("  [%d slain]", slain))
	}
	sb.WriteString("\n")

	return sb.String()
}

func renderPipeline(total, successes int, label string) string {
	if total == 0 {
		return fmt.Sprintf("--(%s)--> 0", label)
	}
	return fmt.Sprintf("--(%s)--> %d", label, successes)
}
