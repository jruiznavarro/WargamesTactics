package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jruiznavarro/wargamestactics/internal/ai"
	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/ui"
)

func main() {
	mode := flag.String("mode", "pvai", "Game mode: pvp, pvai, aivai")
	seed := flag.Int64("seed", 0, "RNG seed (0 = use current time)")
	rounds := flag.Int("rounds", 5, "Maximum battle rounds")
	flag.Parse()

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	fmt.Println("=== AOS Battle Simulator ===")
	fmt.Printf("Mode: %s | Seed: %d | Max Rounds: %d\n\n", *mode, *seed, *rounds)

	g := game.NewGame(*seed, 48, 24)

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

	setupExampleTerrain(g)
	setupExampleArmies(g)
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
