package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jruiznavarro/wargamestactics/internal/ai"
	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/army"
	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/ui"
)

func main() {
	mode := flag.String("mode", "pvai", "Game mode: pvp, pvai, aivai")
	seed := flag.Int64("seed", 0, "RNG seed (0 = use current time)")
	rounds := flag.Int("rounds", 5, "Maximum battle rounds")
	dataDir := flag.String("data", "data/factions", "Path to faction data directory")
	faction1 := flag.String("p1faction", "", "Player 1 faction (e.g. seraphon)")
	faction2 := flag.String("p2faction", "", "Player 2 faction (e.g. tzeentch)")
	flag.Parse()

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	fmt.Println("=== AOS Battle Simulator ===")
	fmt.Printf("Mode: %s | Seed: %d | Max Rounds: %d\n\n", *mode, *seed, *rounds)

	// Load factions if data directory exists
	registry := army.NewRegistry()
	if _, err := os.Stat(*dataDir); err == nil {
		if err := registry.LoadAllFactions(*dataDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load factions: %v\n", err)
		} else {
			factionIDs := registry.FactionIDs()
			if len(factionIDs) > 0 {
				fmt.Printf("Loaded factions: %v\n", factionIDs)
			}
		}
	}

	// Use battleplan if factions are loaded
	useFactions := *faction1 != "" && *faction2 != ""
	var g *game.Game

	if useFactions {
		// Set up with a random battleplan and data-driven armies
		bp := board.GetBattleplan(board.BattleplanTable1, 1) // Default battleplan
		g = game.NewGameFromBattleplan(*seed, bp)
		fmt.Printf("Battleplan: %s\n", bp.Name)
	} else {
		g = game.NewGame(*seed, 48, 24)
	}

	switch *mode {
	case "pvp":
		p1 := ui.NewCLIPlayer(1, "Player 1")
		p2 := ui.NewCLIPlayer(2, "Player 2")
		g.AddPlayer(p1)
		g.AddPlayer(p2)
	case "pvai":
		p1 := ui.NewCLIPlayer(1, "Player")
		p2 := ai.NewAIPlayer(2, "AI")
		g.AddPlayer(p1)
		g.AddPlayer(p2)
	case "aivai":
		p1 := ai.NewAIPlayer(1, "AI-1")
		p2 := ai.NewAIPlayer(2, "AI-2")
		g.AddPlayer(p1)
		g.AddPlayer(p2)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s (use pvp, pvai, or aivai)\n", *mode)
		os.Exit(1)
	}

	if useFactions {
		f1 := registry.GetFaction(*faction1)
		f2 := registry.GetFaction(*faction2)
		if f1 == nil {
			fmt.Fprintf(os.Stderr, "Unknown faction: %s\n", *faction1)
			os.Exit(1)
		}
		if f2 == nil {
			fmt.Fprintf(os.Stderr, "Unknown faction: %s\n", *faction2)
			os.Exit(1)
		}
		setupFactionArmy(g, f1, 1)
		setupFactionArmy(g, f2, 2)
		fmt.Printf("P1: %s (%d pts) | P2: %s (%d pts)\n\n", f1.Name, armyPoints(f1, 1), f2.Name, armyPoints(f2, 2))
	} else {
		setupExampleTerrain(g)
		setupExampleArmies(g)
	}

	g.RegisterTerrainRules()
	g.RunGame(*rounds)

	fmt.Println()
	fmt.Println("+============================================================+")
	fmt.Println("|                       BATTLE LOG                           |")
	fmt.Println("+============================================================+")
	for _, entry := range g.Log {
		fmt.Println(entry)
	}
	fmt.Println("+============================================================+")

	if g.Winner >= 0 {
		for _, p := range g.Players {
			if p.ID() == g.Winner {
				fmt.Printf("\n  VICTORY: %s (Player %d) wins!\n\n", p.Name(), g.Winner)
				break
			}
		}
	} else {
		fmt.Println("\n  DRAW: No winner after all battle rounds.")
	}
}

// setupFactionArmy creates a sample army for a player from a faction.
// Uses a selection of units up to ~1000 points for quick demonstration.
func setupFactionArmy(g *game.Game, faction *army.Faction, ownerID int) {
	// Deploy along P1 bottom or P2 top depending on owner
	baseY := 6.0
	if ownerID == 2 {
		baseY = 38.0
	}

	pointsSpent := 0
	pointsLimit := 1000
	xPos := 10.0

	// Pick 1 Hero first
	for _, ws := range faction.Warscrolls {
		if !ws.HasKeyword("Hero") || pointsSpent+ws.Points > pointsLimit {
			continue
		}
		pos := core.Position{X: xPos, Y: baseY}
		u := g.CreateUnitFromSpec(ws.Name, ownerID, ws.ToCoreStats(), ws.ToCoreWeapons(),
			ws.UnitSize, pos, ws.BaseSizeInches(),
			ws.ToCoreKeywords(), ws.WardSave, ws.PowerLevel,
			ws.ToCoreSpells(), ws.ToCorePrayers())
		applyAbilities(u, &ws)
		pointsSpent += ws.Points
		xPos += 8.0
		break
	}

	// Fill with non-Hero units
	for _, ws := range faction.Warscrolls {
		if ws.HasKeyword("Hero") || pointsSpent+ws.Points > pointsLimit {
			continue
		}
		if xPos > 55.0 {
			break
		}
		pos := core.Position{X: xPos, Y: baseY}
		u := g.CreateUnitFromSpec(ws.Name, ownerID, ws.ToCoreStats(), ws.ToCoreWeapons(),
			ws.UnitSize, pos, ws.BaseSizeInches(),
			ws.ToCoreKeywords(), ws.WardSave, ws.PowerLevel,
			ws.ToCoreSpells(), ws.ToCorePrayers())
		applyAbilities(u, &ws)
		pointsSpent += ws.Points
		xPos += 8.0
	}
}

func applyAbilities(u *core.Unit, ws *army.Warscroll) {
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

func armyPoints(faction *army.Faction, _ int) int {
	// Estimate for display (simplified - same as setupFactionArmy logic)
	pointsSpent := 0
	pointsLimit := 1000
	heroTaken := false
	count := 0

	for _, ws := range faction.Warscrolls {
		if ws.HasKeyword("Hero") && !heroTaken && pointsSpent+ws.Points <= pointsLimit {
			pointsSpent += ws.Points
			heroTaken = true
			count++
			continue
		}
		if !ws.HasKeyword("Hero") && pointsSpent+ws.Points <= pointsLimit && count < 7 {
			pointsSpent += ws.Points
			count++
		}
	}
	return pointsSpent
}

func setupExampleTerrain(g *game.Game) {
	g.Board.AddTerrain("Dark Woods", board.TerrainObscuring, core.Position{X: 20, Y: 8}, 8, 8)
	g.Board.AddTerrain("Old Ruins", board.TerrainObstacle, core.Position{X: 34, Y: 2}, 5, 4)
	g.Board.AddTerrain("Rocky Hill", board.TerrainArea, core.Position{X: 14, Y: 18}, 6, 2)
	g.Board.AddTerrain("Lava Pit", board.TerrainImpassable, core.Position{X: 22, Y: 0}, 4, 3)
}

func setupExampleArmies(g *game.Game) {
	warriors := g.CreateUnit(
		"Warriors", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		[]core.Weapon{
			{Name: "Broadsword", Range: 0, Attacks: 2, ToHit: 3, ToWound: 4, Rend: 1, Damage: 1, Abilities: core.AbilityAntiInfantry},
		},
		5, core.Position{X: 12, Y: 12}, 1.0,
	)
	warriors.Keywords = []core.Keyword{core.KeywordInfantry}

	knights := g.CreateUnit(
		"Knights", 1,
		core.Stats{Move: 10, Save: 3, Control: 2, Health: 3},
		[]core.Weapon{
			{Name: "Lance", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Rend: 2, Damage: 2, Abilities: core.AbilityCharge},
		},
		3, core.Position{X: 8, Y: 12}, 1.0,
	)
	knights.Keywords = []core.Keyword{core.KeywordCavalry}

	bowmen := g.CreateUnit(
		"Bowmen", 2,
		core.Stats{Move: 5, Save: 5, Control: 1, Health: 1},
		[]core.Weapon{
			{Name: "Longbow", Range: 24, Attacks: 1, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1, Abilities: core.AbilityCrit2Hits},
			{Name: "Dagger", Range: 0, Attacks: 1, ToHit: 4, ToWound: 5, Rend: 0, Damage: 1},
		},
		5, core.Position{X: 36, Y: 12}, 1.0,
	)
	bowmen.Keywords = []core.Keyword{core.KeywordInfantry}

	brutes := g.CreateUnit(
		"Brutes", 2,
		core.Stats{Move: 4, Save: 4, Control: 1, Health: 3},
		[]core.Weapon{
			{Name: "Choppa", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Rend: 1, Damage: 2, Abilities: core.AbilityCritMortal},
		},
		3, core.Position{X: 38, Y: 12}, 1.0,
	)
	brutes.Keywords = []core.Keyword{core.KeywordInfantry}
}
