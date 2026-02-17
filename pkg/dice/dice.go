package dice

import "math/rand"

// Roller provides deterministic dice rolling using a seeded RNG.
type Roller struct {
	rng *rand.Rand
}

// NewRoller creates a new Roller with the given seed.
func NewRoller(seed int64) *Roller {
	return &Roller{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// RollD6 returns a random number between 1 and 6.
func (r *Roller) RollD6() int {
	return r.rng.Intn(6) + 1
}

// Roll2D6 returns the sum of two D6 rolls.
func (r *Roller) Roll2D6() int {
	return r.RollD6() + r.RollD6()
}

// RollD3 returns a random number between 1 and 3.
func (r *Roller) RollD3() int {
	return r.rng.Intn(3) + 1
}

// RollMultipleD6 rolls n D6s and returns all results.
func (r *Roller) RollMultipleD6(n int) []int {
	results := make([]int, n)
	for i := range results {
		results[i] = r.RollD6()
	}
	return results
}

// RollWithThreshold rolls a D6 and returns true if the result is >= threshold.
// A natural 1 always fails. A modified result is clamped: natural 1 always fails.
func (r *Roller) RollWithThreshold(threshold int) (int, bool) {
	roll := r.RollD6()
	if roll == 1 {
		return roll, false
	}
	return roll, roll >= threshold
}

// RollWithModifier rolls a D6, applies a modifier, and checks against threshold.
// A natural 1 always fails regardless of modifier.
func (r *Roller) RollWithModifier(threshold, modifier int) (natural int, modified int, success bool) {
	natural = r.RollD6()
	if natural == 1 {
		return natural, natural + modifier, false
	}
	modified = natural + modifier
	return natural, modified, modified >= threshold
}

// RerollType indicates the type of re-roll allowed.
type RerollType int

const (
	RerollNone      RerollType = iota // No re-roll
	RerollFailed                      // Re-roll failed dice
	RerollAll                         // Re-roll all dice
	RerollOnes                        // Re-roll results of 1
)

// RollD6WithReroll rolls a D6, optionally re-rolling once based on the RerollType.
// AoS4 Rule 2.2 (Errata Jan 2026): A die cannot be re-rolled more than once.
// Re-rolls happen before modifiers are applied.
func (r *Roller) RollD6WithReroll(reroll RerollType, threshold int) int {
	roll := r.RollD6()
	switch reroll {
	case RerollFailed:
		if roll < threshold {
			roll = r.RollD6()
		}
	case RerollAll:
		roll = r.RollD6() // Always take second roll
	case RerollOnes:
		if roll == 1 {
			roll = r.RollD6()
		}
	}
	return roll
}

// Roll2D6WithReroll rolls 2D6 with optional re-roll of the full roll.
// Re-roll happens before modifiers (Rule 2.2).
func (r *Roller) Roll2D6WithReroll(reroll bool, threshold int) int {
	total := r.Roll2D6()
	if reroll && total < threshold {
		total = r.Roll2D6()
	}
	return total
}
