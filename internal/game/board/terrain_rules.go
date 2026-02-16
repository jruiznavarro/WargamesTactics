package board

import (
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
)

// TerrainRules generates all rules implied by the terrain features on a board.
// Call this after all terrain has been placed and pass the result to Engine.AddRule.
func TerrainRules(b *Board) []rules.Rule {
	var result []rules.Rule

	for _, t := range b.Terrain {
		switch t.Type {
		case TerrainWoods:
			result = append(result, woodsRules(b, t)...)
		case TerrainObstacle:
			result = append(result, obstacleRules(b, t)...)
		case TerrainRuins:
			result = append(result, ruinsRules(b, t)...)
		case TerrainImpassable:
			result = append(result, impassableRules(b, t)...)
		}
	}

	return result
}

// woodsRules: -2" movement when moving through woods, +1 save for defender inside.
func woodsRules(b *Board, t *TerrainFeature) []rules.Rule {
	return []rules.Rule{
		{
			Name:    t.Name + ":MovePenalty",
			Trigger: rules.BeforeMove,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				// Applies if destination is inside the woods
				return t.Contains(ctx.Destination)
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.MoveMod -= 2
			},
		},
		{
			Name:    t.Name + ":Cover",
			Trigger: rules.BeforeSaveRoll,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				// Defender gets cover if inside the woods
				return ctx.Defender != nil && t.Contains(ctx.Defender.Position())
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.SaveMod += 1
			},
		},
	}
}

// obstacleRules: +1 save for defender behind obstacle (shooting only).
func obstacleRules(b *Board, t *TerrainFeature) []rules.Rule {
	return []rules.Rule{
		{
			Name:    t.Name + ":Cover",
			Trigger: rules.BeforeSaveRoll,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				// Defender gets cover if inside/behind the obstacle
				return ctx.Defender != nil &&
					ctx.Weapon != nil && ctx.Weapon.IsRanged() &&
					t.Contains(ctx.Defender.Position())
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.SaveMod += 1
			},
		},
	}
}

// ruinsRules: +1 save for defender inside ruins (both melee and ranged).
func ruinsRules(b *Board, t *TerrainFeature) []rules.Rule {
	return []rules.Rule{
		{
			Name:    t.Name + ":Cover",
			Trigger: rules.BeforeSaveRoll,
			Source:  rules.SourceTerrain,
			Condition: func(ctx *rules.Context) bool {
				return ctx.Defender != nil && t.Contains(ctx.Defender.Position())
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.SaveMod += 1
			},
		},
	}
}

// impassableRules: block movement and charging into the terrain.
func impassableRules(b *Board, t *TerrainFeature) []rules.Rule {
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
				// Block charge if the charger would end up in impassable terrain
				if ctx.Defender == nil {
					return false
				}
				// The charge moves toward the target. If the target is in impassable terrain,
				// block the charge.
				return t.Contains(ctx.Defender.Position())
			},
			Apply: func(ctx *rules.Context) {
				ctx.Blocked = true
				ctx.BlockMessage = "cannot charge into " + t.Name + " (impassable)"
			},
		},
	}
}
