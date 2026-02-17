package army

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FactionRegistry holds all loaded factions indexed by ID.
type FactionRegistry struct {
	factions map[string]*Faction
}

// NewRegistry creates an empty faction registry.
func NewRegistry() *FactionRegistry {
	return &FactionRegistry{
		factions: make(map[string]*Faction),
	}
}

// LoadFaction loads a faction from a JSON file and registers it.
func (r *FactionRegistry) LoadFaction(path string) (*Faction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading faction file %s: %w", path, err)
	}

	var faction Faction
	if err := json.Unmarshal(data, &faction); err != nil {
		return nil, fmt.Errorf("parsing faction file %s: %w", path, err)
	}

	// Set faction ID on each warscroll
	for i := range faction.Warscrolls {
		if faction.Warscrolls[i].Faction == "" {
			faction.Warscrolls[i].Faction = faction.ID
		}
	}

	r.factions[faction.ID] = &faction
	return &faction, nil
}

// LoadAllFactions loads all JSON files from a directory.
func (r *FactionRegistry) LoadAllFactions(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading faction directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if _, err := r.LoadFaction(path); err != nil {
			return err
		}
	}
	return nil
}

// GetFaction returns the faction with the given ID, or nil.
func (r *FactionRegistry) GetFaction(id string) *Faction {
	return r.factions[id]
}

// AllFactions returns all registered factions.
func (r *FactionRegistry) AllFactions() []*Faction {
	var result []*Faction
	for _, f := range r.factions {
		result = append(result, f)
	}
	return result
}

// FactionIDs returns all registered faction IDs.
func (r *FactionRegistry) FactionIDs() []string {
	var ids []string
	for id := range r.factions {
		ids = append(ids, id)
	}
	return ids
}

// LoadRoster loads an army roster from a JSON file.
func LoadRoster(path string) (*ArmyRoster, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading roster file %s: %w", path, err)
	}

	var roster ArmyRoster
	if err := json.Unmarshal(data, &roster); err != nil {
		return nil, fmt.Errorf("parsing roster file %s: %w", path, err)
	}
	return &roster, nil
}

// ParseFactionJSON parses faction data from raw JSON bytes (useful for embedding).
func ParseFactionJSON(data []byte) (*Faction, error) {
	var faction Faction
	if err := json.Unmarshal(data, &faction); err != nil {
		return nil, fmt.Errorf("parsing faction JSON: %w", err)
	}
	for i := range faction.Warscrolls {
		if faction.Warscrolls[i].Faction == "" {
			faction.Warscrolls[i].Faction = faction.ID
		}
	}
	return &faction, nil
}
