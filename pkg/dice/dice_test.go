package dice

import (
	"testing"
)

func TestRollD6_Deterministic(t *testing.T) {
	r1 := NewRoller(42)
	r2 := NewRoller(42)

	for i := 0; i < 100; i++ {
		a := r1.RollD6()
		b := r2.RollD6()
		if a != b {
			t.Fatalf("roll %d: got %d and %d with same seed", i, a, b)
		}
	}
}

func TestRollD6_Range(t *testing.T) {
	r := NewRoller(12345)
	for i := 0; i < 1000; i++ {
		roll := r.RollD6()
		if roll < 1 || roll > 6 {
			t.Fatalf("D6 roll out of range: %d", roll)
		}
	}
}

func TestRoll2D6_Range(t *testing.T) {
	r := NewRoller(99999)
	for i := 0; i < 1000; i++ {
		roll := r.Roll2D6()
		if roll < 2 || roll > 12 {
			t.Fatalf("2D6 roll out of range: %d", roll)
		}
	}
}

func TestRollD3_Range(t *testing.T) {
	r := NewRoller(7777)
	for i := 0; i < 1000; i++ {
		roll := r.RollD3()
		if roll < 1 || roll > 3 {
			t.Fatalf("D3 roll out of range: %d", roll)
		}
	}
}

func TestRollMultipleD6(t *testing.T) {
	r := NewRoller(55555)
	results := r.RollMultipleD6(5)
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
	for i, roll := range results {
		if roll < 1 || roll > 6 {
			t.Fatalf("roll %d out of range: %d", i, roll)
		}
	}
}

func TestRollWithThreshold(t *testing.T) {
	// With seed 42, let's verify specific behavior
	r := NewRoller(42)

	// Roll many times and verify:
	// - natural 1 always fails
	// - results >= threshold succeed
	successes := 0
	total := 1000
	threshold := 4

	for i := 0; i < total; i++ {
		roll, success := r.RollWithThreshold(threshold)
		if roll == 1 && success {
			t.Fatal("natural 1 should always fail")
		}
		if roll >= threshold && !success {
			t.Fatalf("roll %d should succeed with threshold %d", roll, threshold)
		}
		if success {
			successes++
		}
	}

	// Statistically, about 50% should succeed (4,5,6 pass, 1,2,3 fail)
	ratio := float64(successes) / float64(total)
	if ratio < 0.35 || ratio > 0.65 {
		t.Fatalf("unexpected success ratio: %.2f (expected ~0.50)", ratio)
	}
}

func TestRollWithModifier(t *testing.T) {
	r := NewRoller(42)

	// Test that natural 1 always fails even with modifier
	for i := 0; i < 1000; i++ {
		natural, _, success := r.RollWithModifier(2, +5)
		if natural == 1 && success {
			t.Fatal("natural 1 should always fail even with positive modifier")
		}
	}
}

func TestDifferentSeeds_DifferentResults(t *testing.T) {
	r1 := NewRoller(1)
	r2 := NewRoller(2)

	same := true
	for i := 0; i < 20; i++ {
		if r1.RollD6() != r2.RollD6() {
			same = false
			break
		}
	}
	if same {
		t.Fatal("different seeds should produce different sequences")
	}
}
