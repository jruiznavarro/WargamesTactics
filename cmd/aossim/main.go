package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jruiznavarro/wargamestactics/internal/ai"
	"github.com/jruiznavarro/wargamestactics/internal/game"
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

	// Setup example armies
	setupExampleArmies(g)

	// Run the game
	g.RunGame(*rounds)

	// Print log
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

func setupExampleArmies(g *game.Game) {
	// Player 1: Melee-focused army
	g.CreateUnit(
		"Warriors",
		1,
		core.Stats{Move: 5, Save: 4, Bravery: 7, Wounds: 1},
		[]core.Weapon{
			{Name: "Broadsword", Range: 0, Attacks: 2, ToHit: 3, ToWound: 4, Rend: -1, Damage: 1},
		},
		5,
		core.Position{X: 12, Y: 12},
		1.0,
	)
	g.CreateUnit(
		"Knights",
		1,
		core.Stats{Move: 10, Save: 3, Bravery: 8, Wounds: 3},
		[]core.Weapon{
			{Name: "Lance", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Rend: -2, Damage: 2},
		},
		3,
		core.Position{X: 8, Y: 12},
		1.0,
	)

	// Player 2: Ranged + melee army
	g.CreateUnit(
		"Bowmen",
		2,
		core.Stats{Move: 5, Save: 5, Bravery: 6, Wounds: 1},
		[]core.Weapon{
			{Name: "Longbow", Range: 24, Attacks: 1, ToHit: 4, ToWound: 4, Rend: 0, Damage: 1},
			{Name: "Dagger", Range: 0, Attacks: 1, ToHit: 4, ToWound: 5, Rend: 0, Damage: 1},
		},
		5,
		core.Position{X: 36, Y: 12},
		1.0,
	)
	g.CreateUnit(
		"Brutes",
		2,
		core.Stats{Move: 4, Save: 4, Bravery: 7, Wounds: 3},
		[]core.Weapon{
			{Name: "Choppa", Range: 0, Attacks: 3, ToHit: 3, ToWound: 3, Rend: -1, Damage: 2},
		},
		3,
		core.Position{X: 38, Y: 12},
		1.0,
	)
}
