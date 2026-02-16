package board

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
)

// TerrainRules generates all rules implied by the terrain features on a board.
// AoS4 Advanced Rules: Terrain 1.2.
func TerrainRules(b *Board) []rules.Rule {
	var result []rules.Rule

	for _, t := range b.Terrain {
		switch t.Type {
		case TerrainObstacle:
			// Cover + Unstable (Rule 1.4.1)
			result = append(result, coverRule(t)...)
			result = append(result, unstableRule(t)...)
		case TerrainObscuring:
			// Cover + Obscuring + Unstable (Rule 1.4.2)
			result = append(result, coverRule(t)...)
			result = append(result, obscuringRule(t)...)
			result = append(result, unstableRule(t)...)
		case TerrainArea:
			// Cover only (Rule 1.4.3)
			result = append(result, coverRule(t)...)
		case TerrainPlaceOfPower:
			// Cover + Place of Power + Unstable (Rule 1.4.4)
			result = append(result, coverRule(t)...)
			result = append(result, unstableRule(t)...)
		case TerrainImpassable:
			result = append(result, impassableRules(t)...)
		}
	}

	return result
}

// coverRule: AoS4 Rule 1.2 Cover.
// Subtract 1 from HIT ROLLS for attacks that target a unit behind or wholly
// on this terrain feature, unless the target charged this turn or has Fly.
func coverRule(t *TerrainFeature) []rules.Rule {
	return []rules.Rule{
		{
			Name:    t.Name + ":Cover",
			Trigger: rules.BeforeHitRoll,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				if ctx.Defender == nil {
					return false
				}
				if ctx.Defender.HasCharged {
					return false
				}
				if ctx.Defender.HasKeyword(core.KeywordFly) {
					return false
				}
				return t.Contains(ctx.Defender.Position())
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.HitMod -= 1
			},
		},
	}
}

// obscuringRule: AoS4 Rule 1.2 Obscuring.
// A unit cannot be targeted by shooting attacks from enemies not within
// its combat range if it is behind or wholly on this terrain, unless it has Fly.
func obscuringRule(t *TerrainFeature) []rules.Rule {
	return []rules.Rule{
		{
			Name:    t.Name + ":Obscuring",
			Trigger: rules.BeforeShoot,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				if ctx.Defender == nil || ctx.Attacker == nil {
					return false
				}
				if !t.Contains(ctx.Defender.Position()) {
					return false
				}
				if ctx.Defender.HasKeyword(core.KeywordFly) {
					return false
				}
				dist := core.Distance(ctx.Attacker.Position(), ctx.Defender.Position())
				if dist <= 3.0 {
					return false
				}
				return true
			},
			Apply: func(ctx *rules.Context) {
				ctx.Blocked = true
				ctx.BlockMessage = "target is obscured by " + t.Name
			},
		},
	}
}

// unstableRule: AoS4 Rule 1.2 Unstable.
// Models cannot end moves on terrain >1" tall. Simplified: block ending moves inside.
func unstableRule(t *TerrainFeature) []rules.Rule {
	return []rules.Rule{
		{
			Name:    t.Name + ":Unstable",
			Trigger: rules.BeforeMove,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				return t.Contains(ctx.Destination)
			},
			Apply: func(ctx *rules.Context) {
				ctx.Blocked = true
				ctx.BlockMessage = "cannot end move on " + t.Name + " (unstable)"
			},
		},
	}
}

// impassableRules: block movement and charging into terrain entirely.
func impassableRules(t *TerrainFeature) []rules.Rule {
	return []rules.Rule{
		{
			Name:    t.Name + ":BlockMove",
			Trigger: rules.BeforeMove,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				return t.Contains(ctx.Destination)
			},
			Apply: func(ctx *rules.Context) {
				ctx.Blocked = true
				ctx.BlockMessage = "cannot move into " + t.Name + " (impassable)"
			},
		},
		{
			Name:    t.Name + ":BlockCharge",
			Trigger: rules.BeforeCharge,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				if ctx.Defender == nil {
					return false
				}
				return t.Contains(ctx.Defender.Position())
			},
			Apply: func(ctx *rules.Context) {
				ctx.Blocked = true
				ctx.BlockMessage = "cannot charge into " + t.Name + " (impassable)"
			},
		},
	}
}
