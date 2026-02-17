package army

// EnhancementType classifies what kind of enhancement this is.
type EnhancementType string

const (
	EnhancementHeroicTrait EnhancementType = "heroicTrait"
	EnhancementArtefact    EnhancementType = "artefact"
)

// Enhancement represents a matched play enhancement (artefact, heroic trait).
// Enhancements modify a hero unit's capabilities.
type Enhancement struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        EnhancementType `json:"type"`
	Effect      string          `json:"effect"` // Machine-readable effect key
	Value       int             `json:"value"`  // Numeric value for the effect
}

// BattleFormation represents a selectable army-wide formation.
// Each faction offers 3 formations; one is chosen during list building.
type BattleFormation struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Effects     []BattleFormationEffect `json:"effects"`
}

// BattleFormationEffect defines a single effect from a battle formation.
type BattleFormationEffect struct {
	Description string `json:"description"`
	TargetTag   string `json:"targetTag"`  // Which units this applies to (e.g. "Saurus", "Flamer")
	Effect      string `json:"effect"`     // Machine-readable effect key
	Value       int    `json:"value"`      // Numeric value
	Condition   string `json:"condition"`  // When it applies: "charged", "always", "nearWizard"
}

// ScourgeOfGhyran represents a GHB universal enhancement available to all factions.
type ScourgeOfGhyran struct {
	HeroicTraits []Enhancement `json:"heroicTraits"`
	Artefacts    []Enhancement `json:"artefacts"`
	SpellLore    []WarscrollSpell  `json:"spellLore"`
	PrayerLore   []WarscrollPrayer `json:"prayerLore"`
}

// DefaultScourgeOfGhyran returns the GHB 2025-26 Scourge of Ghyran enhancements.
func DefaultScourgeOfGhyran() *ScourgeOfGhyran {
	return &ScourgeOfGhyran{
		HeroicTraits: []Enhancement{
			{
				Name:        "Battle-hardened",
				Description: "This unit has Ward 6+.",
				Type:        EnhancementHeroicTrait,
				Effect:      "ward",
				Value:       6,
			},
			{
				Name:        "Skilled Leader",
				Description: "At the start of your hero phase, if this unit is on the battlefield, you receive 1 extra command point.",
				Type:        EnhancementHeroicTrait,
				Effect:      "extraCP",
				Value:       1,
			},
			{
				Name:        "Strategic Genius",
				Description: "After deployment, pick 1 friendly unit and set it up again anywhere wholly within your territory.",
				Type:        EnhancementHeroicTrait,
				Effect:      "redeploy",
				Value:       1,
			},
		},
		Artefacts: []Enhancement{
			{
				Name:        "Amulet of Destiny",
				Description: "This unit has Ward 5+.",
				Type:        EnhancementArtefact,
				Effect:      "ward",
				Value:       5,
			},
			{
				Name:        "Seed of Rebirth",
				Description: "At the start of each hero phase, heal D3 wounds allocated to this unit.",
				Type:        EnhancementArtefact,
				Effect:      "healStart",
				Value:       2, // Average of D3
			},
			{
				Name:        "Arcane Tome",
				Description: "This unit gains the WIZARD keyword and can cast 1 spell (Arcane Bolt - CV5, D3 damage at 18\").",
				Type:        EnhancementArtefact,
				Effect:      "arcaneWizard",
				Value:       1,
			},
		},
		SpellLore: []WarscrollSpell{
			{
				Name:         "Wildform",
				CastingValue: 7,
				Range:        12,
				Effect:       "buff",
				EffectValue:  1,
				TargetAlly:   true,
			},
			{
				Name:         "Lifesurge",
				CastingValue: 6,
				Range:        18,
				Effect:       "heal",
				EffectValue:  3,
				TargetAlly:   true,
			},
		},
		PrayerLore: []WarscrollPrayer{
			{
				Name:          "Heal",
				ChantingValue: 4,
				Range:         12,
				Effect:        "heal",
				EffectValue:   3,
				TargetAlly:    true,
			},
			{
				Name:          "Curse",
				ChantingValue: 5,
				Range:         12,
				Effect:        "damage",
				EffectValue:   3,
				TargetAlly:    false,
			},
		},
	}
}
