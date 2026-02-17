package army

import (
	"math"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/rules"
)

// RegisterFactionRules registers all battle trait rules for a faction into the rules engine.
// This should be called once per player during game setup.
func RegisterFactionRules(engine *rules.Engine, faction *Faction, ownerID int) {
	switch faction.ID {
	case "seraphon":
		registerSeraphonRules(engine, ownerID)
	case "tzeentch":
		registerTzeentchRules(engine, ownerID)
	}
}

// RegisterFormationRules registers rules from a battle formation into the engine.
func RegisterFormationRules(engine *rules.Engine, faction *Faction, formationIdx int, ownerID int) {
	if formationIdx < 0 || formationIdx >= len(faction.Formations) {
		return
	}
	formation := &faction.Formations[formationIdx]
	switch faction.ID {
	case "seraphon":
		registerSeraphonFormation(engine, formation, ownerID)
	case "tzeentch":
		registerTzeentchFormation(engine, formation, ownerID)
	}
}

// --- Seraphon Battle Traits ---

func registerSeraphonRules(engine *rules.Engine, ownerID int) {
	// Scaly Skin: SAURUS units have Ward 6+.
	// Applied via BeforeWardSave — if defender is Saurus with no better ward, grant 6+.
	engine.AddRule(rules.Rule{
		Name:    "Scaly Skin",
		Trigger: rules.BeforeWardSave,
		Source:  rules.SourceFaction,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Defender != nil &&
				ctx.Defender.OwnerID == ownerID &&
				ctx.Defender.FactionKeyword == "seraphon" &&
				ctx.Defender.HasTag("Saurus")
		},
		Apply: func(ctx *rules.Context) {
			// Grant Ward 6+ if no better ward exists
			if ctx.Defender.WardSave == 0 || ctx.Defender.WardSave > 6 {
				ctx.WardOverride = 6
			}
		},
	})

	// Predatory Fighters: +1 attack to melee weapons for SAURUS units that charged this turn.
	engine.AddRule(rules.Rule{
		Name:    "Predatory Fighters",
		Trigger: rules.BeforeAttackCount,
		Source:  rules.SourceFaction,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Attacker != nil &&
				ctx.Attacker.OwnerID == ownerID &&
				ctx.Attacker.FactionKeyword == "seraphon" &&
				ctx.Attacker.HasTag("Saurus") &&
				ctx.Attacker.HasCharged &&
				!ctx.IsShooting &&
				ctx.Weapon != nil && ctx.Weapon.IsMelee()
		},
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.AttacksMod += ctx.Attacker.AliveModels()
		},
	})

	// Cold-blooded: Seraphon units wholly within 12" of a Seraphon Hero
	// can ignore negative modifiers. Simplified: grant +1 to counteract -1 modifiers
	// when near a friendly hero.
	engine.AddRule(rules.Rule{
		Name:    "Cold-blooded (Hit)",
		Trigger: rules.BeforeHitRoll,
		Source:  rules.SourceFaction,
		Condition: func(ctx *rules.Context) bool {
			if ctx.Attacker == nil || ctx.Attacker.OwnerID != ownerID ||
				ctx.Attacker.FactionKeyword != "seraphon" {
				return false
			}
			// Check if near a friendly Seraphon Hero
			return isNearFriendlyHero(ctx.Attacker, ctx.AllUnits, "seraphon", 12.0)
		},
		Apply: func(ctx *rules.Context) {
			// Counteract negative hit modifiers
			if ctx.Modifiers.HitMod < 0 {
				ctx.Modifiers.HitMod = 0
			}
		},
	})

	engine.AddRule(rules.Rule{
		Name:    "Cold-blooded (Wound)",
		Trigger: rules.BeforeWoundRoll,
		Source:  rules.SourceFaction,
		Condition: func(ctx *rules.Context) bool {
			if ctx.Attacker == nil || ctx.Attacker.OwnerID != ownerID ||
				ctx.Attacker.FactionKeyword != "seraphon" {
				return false
			}
			return isNearFriendlyHero(ctx.Attacker, ctx.AllUnits, "seraphon", 12.0)
		},
		Apply: func(ctx *rules.Context) {
			if ctx.Modifiers.WoundMod < 0 {
				ctx.Modifiers.WoundMod = 0
			}
		},
	})
}

// --- Seraphon Battle Formations ---

func registerSeraphonFormation(engine *rules.Engine, formation *BattleFormation, ownerID int) {
	switch formation.Name {
	case "Sunclaw Temple-host":
		// Saurus units: +1 Rend on charge turn for melee weapons.
		engine.AddRule(rules.Rule{
			Name:    "Sunclaw Temple-host: Savage Charge",
			Trigger: rules.BeforeSaveRoll,
			Source:  rules.SourceFormation,
			Condition: func(ctx *rules.Context) bool {
				return ctx.Attacker != nil &&
					ctx.Attacker.OwnerID == ownerID &&
					ctx.Attacker.FactionKeyword == "seraphon" &&
					ctx.Attacker.HasTag("Saurus") &&
					ctx.Attacker.HasCharged &&
					!ctx.IsShooting
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.RendMod++
			},
		})

	case "Starborne Host":
		// Seraphon units within 12" of a friendly Wizard have Ward 6+.
		engine.AddRule(rules.Rule{
			Name:    "Starborne Host: Celestial Ward",
			Trigger: rules.BeforeWardSave,
			Source:  rules.SourceFormation,
			Condition: func(ctx *rules.Context) bool {
				if ctx.Defender == nil || ctx.Defender.OwnerID != ownerID ||
					ctx.Defender.FactionKeyword != "seraphon" {
					return false
				}
				return isNearFriendlyWizard(ctx.Defender, ctx.AllUnits, ownerID, 12.0)
			},
			Apply: func(ctx *rules.Context) {
				if ctx.Defender.WardSave == 0 || ctx.Defender.WardSave > 6 {
					ctx.WardOverride = 6
				}
			},
		})

	case "Shadowstrike Starhost":
		// Skink units get +1 to charge rolls (simplified ambush).
		engine.AddRule(rules.Rule{
			Name:    "Shadowstrike Starhost: Swift Strike",
			Trigger: rules.BeforeCharge,
			Source:  rules.SourceFormation,
			Condition: func(ctx *rules.Context) bool {
				return ctx.Attacker != nil &&
					ctx.Attacker.OwnerID == ownerID &&
					ctx.Attacker.FactionKeyword == "seraphon" &&
					ctx.Attacker.HasTag("Skink")
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.ChargeMod++
			},
		})
	}
}

// --- Tzeentch Battle Traits ---

func registerTzeentchRules(engine *rules.Engine, ownerID int) {
	// Locus of Change: -1 to wound rolls for attacks targeting DAEMON units
	// that are wholly within 9" of a friendly Tzeentch Hero.
	engine.AddRule(rules.Rule{
		Name:    "Locus of Change",
		Trigger: rules.BeforeWoundRoll,
		Source:  rules.SourceFaction,
		Condition: func(ctx *rules.Context) bool {
			if ctx.Defender == nil || ctx.Defender.OwnerID != ownerID ||
				ctx.Defender.FactionKeyword != "tzeentch" ||
				!ctx.Defender.HasTag("Daemon") {
				return false
			}
			return isNearFriendlyHero(ctx.Defender, ctx.AllUnits, "tzeentch", 9.0)
		},
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.WoundMod--
		},
	})

	// Masters of Destiny: Implemented at Game level via DestinyDice pool.
	// The rules engine hook is in BeforeHitRoll/BeforeWoundRoll/BeforeSaveRoll
	// but actual substitution is handled by the game loop. We register a marker rule.
	// (Destiny Dice mechanics are managed in the Game struct, not the rules engine.)
}

// --- Tzeentch Battle Formations ---

func registerTzeentchFormation(engine *rules.Engine, formation *BattleFormation, ownerID int) {
	switch formation.Name {
	case "Arcanite Cabal":
		// Kairic Acolytes and Tzaangor units near a hero get +1 to casting.
		// Simplified: all Arcanite-tagged units near a hero get +1 hit on ranged attacks.
		engine.AddRule(rules.Rule{
			Name:    "Arcanite Cabal: Arcane Synergy",
			Trigger: rules.BeforeHitRoll,
			Source:  rules.SourceFormation,
			Condition: func(ctx *rules.Context) bool {
				if ctx.Attacker == nil || ctx.Attacker.OwnerID != ownerID ||
					ctx.Attacker.FactionKeyword != "tzeentch" {
					return false
				}
				if !ctx.Attacker.HasTag("Arcanite") {
					return false
				}
				return ctx.IsShooting && isNearFriendlyHero(ctx.Attacker, ctx.AllUnits, "tzeentch", 12.0)
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.HitMod++
			},
		})

	case "Wyrdflame Host":
		// Flamers and Exalted Flamers get +1 Rend on ranged weapons.
		engine.AddRule(rules.Rule{
			Name:    "Wyrdflame Host: Searing Flames",
			Trigger: rules.BeforeSaveRoll,
			Source:  rules.SourceFormation,
			Condition: func(ctx *rules.Context) bool {
				return ctx.Attacker != nil &&
					ctx.Attacker.OwnerID == ownerID &&
					ctx.Attacker.FactionKeyword == "tzeentch" &&
					ctx.Attacker.HasTag("Flamer") &&
					ctx.IsShooting
			},
			Apply: func(ctx *rules.Context) {
				ctx.Modifiers.RendMod++
			},
		})

	case "Omniscient Oracles":
		// Provides Destiny Dice manipulation — handled at Game level.
		// No rules engine hook needed here.
	}
}

// --- Warscroll Ability Rules ---

// RegisterWarscrollAbilityRules registers rules for warscroll-specific abilities
// that affect the combat pipeline.
func RegisterWarscrollAbilityRules(engine *rules.Engine, unit *core.Unit, ws *Warscroll) {
	for _, ab := range ws.Abilities {
		switch ab.Effect {
		case "bonusChargeAttacks":
			// +N attacks on the charge (like Predatory Fighters but unit-specific)
			registerBonusChargeAttacks(engine, unit, ab.Value)
		case "rerollCharges":
			registerRerollCharges(engine, unit)
		case "mortalOnCharge":
			registerMortalOnCharge(engine, unit, ab.Value)
		case "fly":
			// Already handled via keyword; no rule needed
		case "shootInCombat":
			// Already handled via weapon ability; no rule needed
		case "healOnKill":
			registerHealOnKill(engine, unit, ab.Value)
		case "minusOneToBeHit":
			registerMinusOneToBeHit(engine, unit)
		}
	}
}

func registerBonusChargeAttacks(engine *rules.Engine, unit *core.Unit, bonus int) {
	unitID := unit.ID
	ownerID := unit.OwnerID
	engine.AddRule(rules.Rule{
		Name:    unit.Name + ": Bonus Charge Attacks",
		Trigger: rules.BeforeAttackCount,
		Source:  rules.SourceUnitAbility,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Attacker != nil &&
				ctx.Attacker.ID == unitID &&
				ctx.Attacker.OwnerID == ownerID &&
				ctx.Attacker.HasCharged &&
				!ctx.IsShooting
		},
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.AttacksMod += bonus * ctx.Attacker.AliveModels()
		},
	})
}

func registerRerollCharges(engine *rules.Engine, unit *core.Unit) {
	unitID := unit.ID
	engine.AddRule(rules.Rule{
		Name:    unit.Name + ": Reroll Charges",
		Trigger: rules.BeforeCharge,
		Source:  rules.SourceUnitAbility,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Attacker != nil && ctx.Attacker.ID == unitID
		},
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.ChargeMod += 2 // Simplified: +2 to charge instead of full reroll
		},
	})
}

func registerMortalOnCharge(engine *rules.Engine, unit *core.Unit, value int) {
	unitID := unit.ID
	engine.AddRule(rules.Rule{
		Name:    unit.Name + ": Mortal on Charge",
		Trigger: rules.BeforeAttackCount,
		Source:  rules.SourceUnitAbility,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Attacker != nil &&
				ctx.Attacker.ID == unitID &&
				ctx.Attacker.HasCharged &&
				!ctx.IsShooting
		},
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.MortalWounds += value
		},
	})
}

func registerHealOnKill(engine *rules.Engine, unit *core.Unit, value int) {
	unitID := unit.ID
	engine.AddRule(rules.Rule{
		Name:    unit.Name + ": Heal on Kill",
		Trigger: rules.AfterCombatResolve,
		Source:  rules.SourceUnitAbility,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Attacker != nil && ctx.Attacker.ID == unitID &&
				ctx.Modifiers.DamageMod > 0 // DamageMod stores total damage dealt
		},
		Apply: func(ctx *rules.Context) {
			// Heal first alive model
			for i := range ctx.Attacker.Models {
				if ctx.Attacker.Models[i].IsAlive && ctx.Attacker.Models[i].CurrentWounds < ctx.Attacker.Models[i].MaxWounds {
					ctx.Attacker.Models[i].CurrentWounds += value
					if ctx.Attacker.Models[i].CurrentWounds > ctx.Attacker.Models[i].MaxWounds {
						ctx.Attacker.Models[i].CurrentWounds = ctx.Attacker.Models[i].MaxWounds
					}
					break
				}
			}
		},
	})
}

func registerMinusOneToBeHit(engine *rules.Engine, unit *core.Unit) {
	unitID := unit.ID
	ownerID := unit.OwnerID
	engine.AddRule(rules.Rule{
		Name:    unit.Name + ": Hard to Hit",
		Trigger: rules.BeforeHitRoll,
		Source:  rules.SourceUnitAbility,
		Condition: func(ctx *rules.Context) bool {
			return ctx.Defender != nil &&
				ctx.Defender.ID == unitID &&
				ctx.Defender.OwnerID == ownerID
		},
		Apply: func(ctx *rules.Context) {
			ctx.Modifiers.HitMod--
		},
	})
}

// --- Destiny Dice System (Tzeentch) ---

// DestinyDicePool manages the Tzeentch Masters of Destiny mechanic.
type DestinyDicePool struct {
	Dice    []int // Current destiny dice values (1-6)
	OwnerID int   // Player who owns this pool
}

// NewDestinyDicePool creates a pool by rolling 9 dice.
func NewDestinyDicePool(ownerID int, rolls []int) *DestinyDicePool {
	pool := &DestinyDicePool{
		OwnerID: ownerID,
		Dice:    make([]int, len(rolls)),
	}
	copy(pool.Dice, rolls)
	return pool
}

// Count returns the number of remaining destiny dice.
func (p *DestinyDicePool) Count() int {
	return len(p.Dice)
}

// HasValue returns true if a die with the given value exists in the pool.
func (p *DestinyDicePool) HasValue(value int) bool {
	for _, d := range p.Dice {
		if d == value {
			return true
		}
	}
	return false
}

// UseValue removes and returns a die with the given value from the pool.
// Returns false if no such die exists.
func (p *DestinyDicePool) UseValue(value int) bool {
	for i, d := range p.Dice {
		if d == value {
			p.Dice = append(p.Dice[:i], p.Dice[i+1:]...)
			return true
		}
	}
	return false
}

// UseBest removes and returns the highest value die from the pool.
// Returns 0 if pool is empty.
func (p *DestinyDicePool) UseBest() int {
	if len(p.Dice) == 0 {
		return 0
	}
	bestIdx := 0
	for i, d := range p.Dice {
		if d > p.Dice[bestIdx] {
			bestIdx = i
		}
	}
	val := p.Dice[bestIdx]
	p.Dice = append(p.Dice[:bestIdx], p.Dice[bestIdx+1:]...)
	return val
}

// UseWorst removes and returns the lowest value die from the pool.
// Returns 0 if pool is empty.
func (p *DestinyDicePool) UseWorst() int {
	if len(p.Dice) == 0 {
		return 0
	}
	worstIdx := 0
	for i, d := range p.Dice {
		if d < p.Dice[worstIdx] {
			worstIdx = i
		}
	}
	val := p.Dice[worstIdx]
	p.Dice = append(p.Dice[:worstIdx], p.Dice[worstIdx+1:]...)
	return val
}

// AddDie adds a die with the given value to the pool.
func (p *DestinyDicePool) AddDie(value int) {
	if value >= 1 && value <= 6 {
		p.Dice = append(p.Dice, value)
	}
}

// --- Helper Functions ---

// isNearFriendlyHero checks if a unit is within the given range of a friendly hero
// with the specified faction keyword.
func isNearFriendlyHero(unit *core.Unit, allUnits []*core.Unit, factionKW string, rangeInches float64) bool {
	for _, other := range allUnits {
		if other.ID == unit.ID || other.OwnerID != unit.OwnerID {
			continue
		}
		if other.FactionKeyword != factionKW || !other.HasKeyword(core.KeywordHero) {
			continue
		}
		if other.IsDestroyed() {
			continue
		}
		if unitDistance(unit, other) <= rangeInches {
			return true
		}
	}
	return false
}

// isNearFriendlyWizard checks if a unit is within the given range of a friendly wizard.
func isNearFriendlyWizard(unit *core.Unit, allUnits []*core.Unit, ownerID int, rangeInches float64) bool {
	for _, other := range allUnits {
		if other.ID == unit.ID || other.OwnerID != ownerID {
			continue
		}
		if !other.HasKeyword(core.KeywordWizard) || other.IsDestroyed() {
			continue
		}
		if unitDistance(unit, other) <= rangeInches {
			return true
		}
	}
	return false
}

// unitDistance returns the distance between two units based on leader positions.
func unitDistance(a, b *core.Unit) float64 {
	posA := a.Position()
	posB := b.Position()
	dx := posA.X - posB.X
	dy := posA.Y - posB.Y
	return math.Sqrt(dx*dx + dy*dy)
}
