package core

// Stats represents the base characteristics of a unit (AoS4).
type Stats struct {
	Move    int // Movement in inches
	Save    int // Save roll threshold (e.g. 3+ = 3)
	Control int // Control characteristic for objective scoring
	Health  int // Health (wounds per model)
}
