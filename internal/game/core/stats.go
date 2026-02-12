package core

// Stats represents the base characteristics of a unit.
type Stats struct {
	Move    int // Movement in inches
	Save    int // Save roll threshold (e.g. 3+ = 3)
	Bravery int // Bravery/morale value
	Wounds  int // Maximum wounds per model
}
