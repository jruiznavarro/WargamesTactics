package army

import "github.com/jruiznavarro/wargamestactics/internal/game/core"

// Warscroll represents a unit's datasheet definition.
// This is the data-driven equivalent of hardcoded unit creation.
type Warscroll struct {
	ID           string            `json:"id"`            // Unique warscroll identifier (e.g. "seraphon_saurus_warriors")
	Name         string            `json:"name"`          // Display name (e.g. "Saurus Warriors")
	Faction      string            `json:"faction"`       // Faction key (e.g. "seraphon")
	Points       int               `json:"points"`        // Matched play points cost
	UnitSize     int               `json:"unitSize"`      // Base number of models
	MaxSize      int               `json:"maxSize"`       // Reinforced size (0 = cannot reinforce)
	BaseSizeMM   int               `json:"baseSizeMM"`    // Base diameter in millimeters
	Keywords     []string          `json:"keywords"`      // Unit keywords (Hero, Infantry, etc.)
	Tags         []string          `json:"tags"`          // Faction sub-keywords (Saurus, Skink, Daemon, etc.)
	Stats        WarscrollStats    `json:"stats"`         // Core stats
	Weapons      []WarscrollWeapon `json:"weapons"`       // Weapon profiles
	WardSave     int               `json:"wardSave"`      // Ward save (0 = none)
	PowerLevel   int               `json:"powerLevel"`    // Wizard(X) or Priest(X) level (0 = not a caster)
	Spells       []WarscrollSpell  `json:"spells"`        // Known spells (Wizard only)
	Prayers      []WarscrollPrayer `json:"prayers"`       // Known prayers (Priest only)
	Abilities    []WarscrollAbility `json:"abilities"`    // Special abilities
	Unique       bool              `json:"unique"`        // Named/unique character (limit 1)
}

// WarscrollStats maps to core.Stats.
type WarscrollStats struct {
	Move    int `json:"move"`    // Movement in inches
	Save    int `json:"save"`    // Save roll (e.g. 3 = 3+)
	Control int `json:"control"` // Control characteristic
	Health  int `json:"health"`  // Wounds per model
}

// WarscrollWeapon defines a weapon profile.
type WarscrollWeapon struct {
	Name      string   `json:"name"`
	Range     int      `json:"range"`     // 0 = melee
	Attacks   int      `json:"attacks"`
	ToHit     int      `json:"hit"`       // e.g. 3 = 3+
	ToWound   int      `json:"wound"`     // e.g. 4 = 4+
	Rend      int      `json:"rend"`
	Damage    int      `json:"damage"`
	Abilities []string `json:"abilities"` // String ability tags
}

// WarscrollSpell defines a spell known by the unit.
type WarscrollSpell struct {
	Name         string `json:"name"`
	CastingValue int    `json:"castingValue"`
	Range        int    `json:"range"`
	Effect       string `json:"effect"`      // "damage", "heal", "buff"
	EffectValue  int    `json:"effectValue"` // Magnitude
	TargetAlly   bool   `json:"targetAlly"`
	Unlimited    bool   `json:"unlimited"`
}

// WarscrollPrayer defines a prayer known by the unit.
type WarscrollPrayer struct {
	Name          string `json:"name"`
	ChantingValue int    `json:"chantingValue"`
	Range         int    `json:"range"`
	Effect        string `json:"effect"`      // "damage", "heal", "buff"
	EffectValue   int    `json:"effectValue"`
	TargetAlly    bool   `json:"targetAlly"`
	Unlimited     bool   `json:"unlimited"`
}

// WarscrollAbility represents a special rule on the warscroll.
type WarscrollAbility struct {
	Name        string `json:"name"`
	Description string `json:"description"` // Human-readable text
	Phase       string `json:"phase"`       // When it triggers: "passive", "hero", "movement", "shooting", "charge", "combat", "end"
	Effect      string `json:"effect"`      // Machine-readable effect key (e.g. "ward", "strikeFirst", "fly")
	Value       int    `json:"value"`       // Numeric value for the effect (e.g. ward save threshold)
}

// BaseSizeInches converts millimeter base size to inches.
func (w *Warscroll) BaseSizeInches() float64 {
	return float64(w.BaseSizeMM) / 25.4
}

// HasKeyword returns true if the warscroll has the given keyword string.
func (w *Warscroll) HasKeyword(kw string) bool {
	for _, k := range w.Keywords {
		if k == kw {
			return true
		}
	}
	return false
}

// ToCoreKeywords converts string keywords to core.Keyword values.
func (w *Warscroll) ToCoreKeywords() []core.Keyword {
	keywordMap := map[string]core.Keyword{
		"Infantry":      core.KeywordInfantry,
		"Cavalry":       core.KeywordCavalry,
		"Hero":          core.KeywordHero,
		"Monster":       core.KeywordMonster,
		"War Machine":   core.KeywordWarMachine,
		"Wizard":        core.KeywordWizard,
		"Priest":        core.KeywordPriest,
		"Fly":           core.KeywordFly,
		"Manifestation": core.KeywordManifestation,
	}

	var result []core.Keyword
	for _, k := range w.Keywords {
		if ck, ok := keywordMap[k]; ok {
			result = append(result, ck)
		}
	}
	return result
}

// ToCoreStats converts warscroll stats to core.Stats.
func (w *Warscroll) ToCoreStats() core.Stats {
	return core.Stats{
		Move:    w.Stats.Move,
		Save:    w.Stats.Save,
		Control: w.Stats.Control,
		Health:  w.Stats.Health,
	}
}

// ToCoreWeapons converts warscroll weapons to core.Weapon slices.
func (w *Warscroll) ToCoreWeapons() []core.Weapon {
	weapons := make([]core.Weapon, len(w.Weapons))
	for i, ww := range w.Weapons {
		weapons[i] = core.Weapon{
			Name:      ww.Name,
			Range:     ww.Range,
			Attacks:   ww.Attacks,
			ToHit:     ww.ToHit,
			ToWound:   ww.ToWound,
			Rend:      ww.Rend,
			Damage:    ww.Damage,
			Abilities: parseWeaponAbilities(ww.Abilities),
		}
	}
	return weapons
}

// ToCoreSpells converts warscroll spells to core.Spell slices.
func (w *Warscroll) ToCoreSpells() []core.Spell {
	spells := make([]core.Spell, len(w.Spells))
	for i, s := range w.Spells {
		spells[i] = core.Spell{
			Name:           s.Name,
			CastingValue:   s.CastingValue,
			Range:          s.Range,
			Effect:         parseSpellEffect(s.Effect),
			EffectValue:    s.EffectValue,
			TargetFriendly: s.TargetAlly,
			Unlimited:      s.Unlimited,
		}
	}
	return spells
}

// ToCorePrayers converts warscroll prayers to core.Prayer slices.
func (w *Warscroll) ToCorePrayers() []core.Prayer {
	prayers := make([]core.Prayer, len(w.Prayers))
	for i, p := range w.Prayers {
		prayers[i] = core.Prayer{
			Name:           p.Name,
			ChantingValue:  p.ChantingValue,
			Range:          p.Range,
			Effect:         parseSpellEffect(p.Effect),
			EffectValue:    p.EffectValue,
			TargetFriendly: p.TargetAlly,
			Unlimited:      p.Unlimited,
		}
	}
	return prayers
}

// parseWeaponAbilities converts string ability tags to the bitmask.
func parseWeaponAbilities(abilities []string) core.WeaponAbility {
	var result core.WeaponAbility
	abilityMap := map[string]core.WeaponAbility{
		"Anti-Infantry":    core.AbilityAntiInfantry,
		"Anti-Cavalry":     core.AbilityAntiCavalry,
		"Anti-Hero":        core.AbilityAntiHero,
		"Anti-Monster":     core.AbilityAntiMonster,
		"Anti-charge":      core.AbilityAntiCharge,
		"Charge":           core.AbilityCharge,
		"Crit(2 Hits)":     core.AbilityCrit2Hits,
		"Crit(Auto-wound)": core.AbilityCritAutoWound,
		"Crit(Mortal)":     core.AbilityCritMortal,
		"Companion":        core.AbilityCompanion,
		"Shoot in Combat":  core.AbilityShootInCombat,
	}
	for _, a := range abilities {
		if wa, ok := abilityMap[a]; ok {
			result |= wa
		}
	}
	return result
}

func parseSpellEffect(s string) core.SpellEffect {
	switch s {
	case "damage":
		return core.SpellEffectDamage
	case "heal":
		return core.SpellEffectHeal
	case "buff":
		return core.SpellEffectBuff
	default:
		return core.SpellEffectDamage
	}
}
