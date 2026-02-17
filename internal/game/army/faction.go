package army

// Faction represents a complete army faction with its warscrolls and rules.
type Faction struct {
	ID           string            `json:"id"`           // Faction key (e.g. "seraphon")
	Name         string            `json:"name"`         // Display name (e.g. "Seraphon")
	GrandAlliance string           `json:"grandAlliance"` // "Order", "Chaos", "Death", "Destruction"
	Warscrolls   []Warscroll       `json:"warscrolls"`   // All available unit warscrolls
	BattleTraits []FactionTrait    `json:"battleTraits"`  // Faction battle traits
	SpellLore    []WarscrollSpell  `json:"spellLore"`     // Faction spell lore (available to all wizards)
	PrayerLore   []WarscrollPrayer `json:"prayerLore"`    // Faction prayer lore (available to all priests)
}

// FactionTrait represents a faction-wide special rule.
type FactionTrait struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Phase       string `json:"phase"`  // "passive", "hero", etc.
	Effect      string `json:"effect"` // Machine-readable effect key
	Value       int    `json:"value"`
}

// GetWarscroll returns the warscroll with the given ID, or nil.
func (f *Faction) GetWarscroll(id string) *Warscroll {
	for i := range f.Warscrolls {
		if f.Warscrolls[i].ID == id {
			return &f.Warscrolls[i]
		}
	}
	return nil
}

// GetWarscrollByName returns the first warscroll with the given name, or nil.
func (f *Faction) GetWarscrollByName(name string) *Warscroll {
	for i := range f.Warscrolls {
		if f.Warscrolls[i].Name == name {
			return &f.Warscrolls[i]
		}
	}
	return nil
}

// Heroes returns all Hero warscrolls in the faction.
func (f *Faction) Heroes() []Warscroll {
	var heroes []Warscroll
	for _, ws := range f.Warscrolls {
		if ws.HasKeyword("Hero") {
			heroes = append(heroes, ws)
		}
	}
	return heroes
}

// NonHeroes returns all non-Hero warscrolls in the faction.
func (f *Faction) NonHeroes() []Warscroll {
	var units []Warscroll
	for _, ws := range f.Warscrolls {
		if !ws.HasKeyword("Hero") {
			units = append(units, ws)
		}
	}
	return units
}
