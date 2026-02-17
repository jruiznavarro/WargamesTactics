package game

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// --- Battle Tactic Card & Tracker Tests ---

func TestAllBattleTacticCards_Returns6(t *testing.T) {
	cards := AllBattleTacticCards()
	if len(cards) != 6 {
		t.Errorf("expected 6 battle tactic cards, got %d", len(cards))
	}
}

func TestGetBattleTacticCard_ValidIDs(t *testing.T) {
	for id := CardSavageSpearhead; id <= CardAttunedToGhyran; id++ {
		card := GetBattleTacticCard(id)
		if card == nil {
			t.Errorf("expected card for ID %d, got nil", id)
		}
	}
}

func TestGetBattleTacticCard_InvalidID(t *testing.T) {
	card := GetBattleTacticCard(99)
	if card != nil {
		t.Errorf("expected nil for invalid card ID 99, got %s", card.Name)
	}
}

func TestBattleTacticCard_GetTactic(t *testing.T) {
	card := GetBattleTacticCard(CardSavageSpearhead)
	if card == nil {
		t.Fatal("expected Savage Spearhead card")
	}

	affray := card.GetTactic(TierAffray)
	if affray.Name != "Aggressive Expansion" {
		t.Errorf("expected Affray 'Aggressive Expansion', got '%s'", affray.Name)
	}
	if affray.VP != BattleTacticVP {
		t.Errorf("expected VP %d, got %d", BattleTacticVP, affray.VP)
	}

	strike := card.GetTactic(TierStrike)
	if strike.Name != "Seize the Centre" {
		t.Errorf("expected Strike 'Seize the Centre', got '%s'", strike.Name)
	}

	dom := card.GetTactic(TierDomination)
	if dom.Name != "Total Conquest" {
		t.Errorf("expected Domination 'Total Conquest', got '%s'", dom.Name)
	}
}

func TestBattleTacticTier_String(t *testing.T) {
	tests := []struct {
		tier     BattleTacticTier
		expected string
	}{
		{TierAffray, "Affray"},
		{TierStrike, "Strike"},
		{TierDomination, "Domination"},
	}
	for _, tt := range tests {
		if tt.tier.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.tier.String())
		}
	}
}

func TestBattleTacticTracker_NewHas6Cards(t *testing.T) {
	tracker := NewBattleTacticTracker()
	if len(tracker.AvailableCards) != 6 {
		t.Errorf("expected 6 available cards, got %d", len(tracker.AvailableCards))
	}
	if tracker.ActiveTactic != nil {
		t.Error("expected no active tactic on new tracker")
	}
}

func TestBattleTacticTracker_SelectTactic(t *testing.T) {
	tracker := NewBattleTacticTracker()
	err := tracker.SelectTactic(CardBrokenRanks, TierAffray)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tracker.ActiveTactic == nil {
		t.Fatal("expected active tactic after selection")
	}
	if tracker.ActiveTactic.Tactic.Name != "Broken Ranks" {
		t.Errorf("expected 'Broken Ranks', got '%s'", tracker.ActiveTactic.Tactic.Name)
	}
}

func TestBattleTacticTracker_CannotSelectTwice(t *testing.T) {
	tracker := NewBattleTacticTracker()
	_ = tracker.SelectTactic(CardBrokenRanks, TierAffray)
	err := tracker.SelectTactic(CardSavageSpearhead, TierAffray)
	if err == nil {
		t.Error("expected error when selecting a second tactic in same round")
	}
}

func TestBattleTacticTracker_CannotReuseCard(t *testing.T) {
	tracker := NewBattleTacticTracker()
	_ = tracker.SelectTactic(CardBrokenRanks, TierAffray)
	tracker.CompleteTactic()
	tracker.ResetRound()

	err := tracker.SelectTactic(CardBrokenRanks, TierStrike)
	if err == nil {
		t.Error("expected error when reusing completed card")
	}
}

func TestBattleTacticTracker_FailedCardAlsoRemoved(t *testing.T) {
	tracker := NewBattleTacticTracker()
	_ = tracker.SelectTactic(CardConquerAndHold, TierDomination)
	tracker.FailTactic()
	tracker.ResetRound()

	err := tracker.SelectTactic(CardConquerAndHold, TierAffray)
	if err == nil {
		t.Error("expected error when reusing failed card")
	}
	if len(tracker.FailedTactics) != 1 {
		t.Errorf("expected 1 failed tactic, got %d", len(tracker.FailedTactics))
	}
}

func TestBattleTacticTracker_CompleteTactic_ReturnsVP(t *testing.T) {
	tracker := NewBattleTacticTracker()
	_ = tracker.SelectTactic(CardSavageSpearhead, TierDomination)
	vp := tracker.CompleteTactic()
	if vp != BattleTacticVP {
		t.Errorf("expected %d VP, got %d", BattleTacticVP, vp)
	}
	if len(tracker.CompletedTactics) != 1 {
		t.Errorf("expected 1 completed tactic, got %d", len(tracker.CompletedTactics))
	}
}

func TestBattleTacticTracker_CompleteNilTactic_Returns0(t *testing.T) {
	tracker := NewBattleTacticTracker()
	vp := tracker.CompleteTactic()
	if vp != 0 {
		t.Errorf("expected 0 VP for nil tactic, got %d", vp)
	}
}

func TestBattleTacticTracker_ResetRound(t *testing.T) {
	tracker := NewBattleTacticTracker()
	_ = tracker.SelectTactic(CardBrokenRanks, TierAffray)
	tracker.CompleteTactic()
	tracker.ResetRound()
	if tracker.ActiveTactic != nil {
		t.Error("expected nil active tactic after reset")
	}
	// Should have 5 available cards now (1 used)
	if len(tracker.AvailableCards) != 5 {
		t.Errorf("expected 5 available cards after completing 1, got %d", len(tracker.AvailableCards))
	}
}

// --- Battle Tactic Condition Evaluation Tests ---

// setupBattleTacticGame creates a game with a battleplan for testing conditions.
// Uses "Passing Seasons" battleplan:
//   Board: 60x44, P1 territory: (0,0)-(60,12), P2 territory: (0,32)-(60,44)
//   Objectives: Oakenbrow pair 1 at (10,10)/(50,34), Gnarlroot pair 2 at (50,10)/(10,34),
//               Winterleaf/Heartwood pair 3 at (20,22)/(40,22)
func setupBattleTacticGame(seed int64) *Game {
	bp := board.GetBattleplan(board.BattleplanTable1, 1) // Passing Seasons
	g := NewGameFromBattleplan(seed, bp)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseEndOfTurn
	g.InitBattleTactics()
	return g
}

// --- Card 1: Savage Spearhead ---

func TestSavageSpearhead_Affray_ControlMore(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P1 controls 2 objectives (near pair 1 oakenbrow at 10,10 and gnarlroot at 50,10)
	g.CreateUnit("P1 Unit1", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("P1 Unit2", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 50, Y: 10}, 1.0)
	// P2 controls 1
	g.CreateUnit("P2 Unit1", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 3, core.Position{X: 50, Y: 34}, 1.0)

	tactic := GetBattleTacticCard(CardSavageSpearhead).GetTactic(TierAffray)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("P1 controls more objectives, Affray should succeed")
	}
	if g.EvaluateBattleTactic(2, tactic) {
		t.Error("P2 controls fewer objectives, Affray should fail")
	}
}

func TestSavageSpearhead_Strike_CentreObjective(t *testing.T) {
	g := setupBattleTacticGame(1)

	// Winterleaf objective at (20,22) is within 12" of centre (30,22) => distance = 10
	g.CreateUnit("P1 Centre", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 20, Y: 22}, 1.0)
	g.CreateUnit("P1 Extra", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 40, Y: 22}, 1.0)
	// P2 controls none
	g.CreateUnit("P2 Far", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 1, core.Position{X: 5, Y: 42}, 1.0)

	tactic := GetBattleTacticCard(CardSavageSpearhead).GetTactic(TierStrike)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("P1 controls more objectives AND one near centre, Strike should succeed")
	}
}

func TestSavageSpearhead_Domination_AllObjectives(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P1 controls ALL 6 objectives
	positions := []core.Position{
		{X: 10, Y: 10}, {X: 50, Y: 34}, {X: 50, Y: 10},
		{X: 10, Y: 34}, {X: 20, Y: 22}, {X: 40, Y: 22},
	}
	for i, pos := range positions {
		g.CreateUnit("P1 Unit"+string(rune('A'+i)), 1,
			core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
			nil, 5, pos, 1.0)
	}

	tactic := GetBattleTacticCard(CardSavageSpearhead).GetTactic(TierDomination)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("P1 controls all objectives, Domination should succeed")
	}
}

func TestSavageSpearhead_Domination_NotAllObjectives(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P1 controls 5 of 6
	positions := []core.Position{
		{X: 10, Y: 10}, {X: 50, Y: 34}, {X: 50, Y: 10},
		{X: 10, Y: 34}, {X: 20, Y: 22},
	}
	for i, pos := range positions {
		g.CreateUnit("P1 Unit"+string(rune('A'+i)), 1,
			core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
			nil, 5, pos, 1.0)
	}
	// P2 controls 1
	g.CreateUnit("P2 Unit", 2, core.Stats{Move: 5, Save: 4, Control: 3, Health: 1},
		nil, 5, core.Position{X: 40, Y: 22}, 1.0)

	tactic := GetBattleTacticCard(CardSavageSpearhead).GetTactic(TierDomination)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("P1 doesn't control ALL objectives, Domination should fail")
	}
}

// --- Card 2: Broken Ranks ---

func TestBrokenRanks_Affray_1Kill(t *testing.T) {
	g := setupBattleTacticGame(1)
	g.UnitsDestroyedThisTurnMap[1] = 1

	tactic := GetBattleTacticCard(CardBrokenRanks).GetTactic(TierAffray)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("1 enemy unit destroyed, Affray should succeed")
	}
}

func TestBrokenRanks_Strike_2Kills(t *testing.T) {
	g := setupBattleTacticGame(1)
	g.UnitsDestroyedThisTurnMap[1] = 2

	tactic := GetBattleTacticCard(CardBrokenRanks).GetTactic(TierStrike)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("2 enemy units destroyed, Strike should succeed")
	}

	g.UnitsDestroyedThisTurnMap[1] = 1
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("only 1 kill, Strike should fail")
	}
}

func TestBrokenRanks_Domination_3Kills(t *testing.T) {
	g := setupBattleTacticGame(1)
	g.UnitsDestroyedThisTurnMap[1] = 3

	tactic := GetBattleTacticCard(CardBrokenRanks).GetTactic(TierDomination)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 enemy units destroyed, Domination should succeed")
	}

	g.UnitsDestroyedThisTurnMap[1] = 2
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("only 2 kills, Domination should fail")
	}
}

// --- Card 3: Conquer and Hold ---

func TestConquerAndHold_Affray_2InEnemyTerritory(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P2 territory is (0,32)-(60,44). Place P1 units there.
	g.CreateUnit("P1 Invader1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 35}, 1.0)
	g.CreateUnit("P1 Invader2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 40}, 1.0)
	g.CreateUnit("P2 Home", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 3, core.Position{X: 30, Y: 34}, 1.0)

	tactic := GetBattleTacticCard(CardConquerAndHold).GetTactic(TierAffray)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("2 P1 units in P2 territory, Affray should succeed")
	}
}

func TestConquerAndHold_Strike_3InEnemyTerritory(t *testing.T) {
	g := setupBattleTacticGame(1)

	// Place 3 P1 units in P2 territory (0,32)-(60,44)
	g.CreateUnit("P1 Inv1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 35}, 1.0)
	g.CreateUnit("P1 Inv2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 40}, 1.0)
	g.CreateUnit("P1 Inv3", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 50, Y: 38}, 1.0)

	tactic := GetBattleTacticCard(CardConquerAndHold).GetTactic(TierStrike)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 P1 units in P2 territory, Strike should succeed")
	}
}

func TestConquerAndHold_Domination_3AndControlObjective(t *testing.T) {
	g := setupBattleTacticGame(1)

	// 3 P1 units in P2 territory, one on the Oakenbrow obj at (50,34)
	g.CreateUnit("P1 Inv1", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 50, Y: 34}, 1.0) // on objective
	g.CreateUnit("P1 Inv2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 40}, 1.0)
	g.CreateUnit("P1 Inv3", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 35}, 1.0)

	tactic := GetBattleTacticCard(CardConquerAndHold).GetTactic(TierDomination)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 in enemy territory + control objective there, Domination should succeed")
	}
}

func TestConquerAndHold_Domination_3ButNoObjective(t *testing.T) {
	g := setupBattleTacticGame(1)

	// 3 P1 units in P2 territory but NOT on any objective
	g.CreateUnit("P1 Inv1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 5, Y: 35}, 1.0)
	g.CreateUnit("P1 Inv2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 40}, 1.0)
	g.CreateUnit("P1 Inv3", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 55, Y: 42}, 1.0)
	// P2 controls the obj at (50,34)
	g.CreateUnit("P2 Guard", 2, core.Stats{Move: 5, Save: 4, Control: 5, Health: 1},
		nil, 10, core.Position{X: 50, Y: 34}, 1.0)

	tactic := GetBattleTacticCard(CardConquerAndHold).GetTactic(TierDomination)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 in enemy territory but no objective controlled there, Domination should fail")
	}
}

// --- Card 4: Ferocious Advance ---

func TestFerocousAdvance_Affray_3RanOrCharged(t *testing.T) {
	g := setupBattleTacticGame(1)

	u1 := g.CreateUnit("P1 A", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	u2 := g.CreateUnit("P1 B", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 20, Y: 10}, 1.0)
	u3 := g.CreateUnit("P1 C", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 10}, 1.0)

	u1.HasRun = true
	u2.HasCharged = true
	u3.HasRun = true

	tactic := GetBattleTacticCard(CardFerocousAdvance).GetTactic(TierAffray)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 units ran/charged, Affray should succeed")
	}
}

func TestFerocousAdvance_Affray_Only2(t *testing.T) {
	g := setupBattleTacticGame(1)

	u1 := g.CreateUnit("P1 A", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	u2 := g.CreateUnit("P1 B", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 20, Y: 10}, 1.0)
	g.CreateUnit("P1 C", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 10}, 1.0)

	u1.HasRun = true
	u2.HasCharged = true
	// u3 has not run or charged

	tactic := GetBattleTacticCard(CardFerocousAdvance).GetTactic(TierAffray)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("only 2 units ran/charged, Affray should fail")
	}
}

func TestFerocousAdvance_Domination_3ChargedAnd1Kill(t *testing.T) {
	g := setupBattleTacticGame(1)

	u1 := g.CreateUnit("P1 A", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	u2 := g.CreateUnit("P1 B", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 20, Y: 10}, 1.0)
	u3 := g.CreateUnit("P1 C", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 10}, 1.0)

	u1.HasCharged = true
	u2.HasCharged = true
	u3.HasCharged = true
	g.UnitsDestroyedThisTurnMap[1] = 1

	tactic := GetBattleTacticCard(CardFerocousAdvance).GetTactic(TierDomination)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 charged + 1 kill, Domination should succeed")
	}
}

func TestFerocousAdvance_Domination_3ChargedButNoKill(t *testing.T) {
	g := setupBattleTacticGame(1)

	u1 := g.CreateUnit("P1 A", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	u2 := g.CreateUnit("P1 B", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 20, Y: 10}, 1.0)
	u3 := g.CreateUnit("P1 C", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 10}, 1.0)

	u1.HasCharged = true
	u2.HasCharged = true
	u3.HasCharged = true
	// No kills

	tactic := GetBattleTacticCard(CardFerocousAdvance).GetTactic(TierDomination)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 charged but 0 kills, Domination should fail")
	}
}

// --- Card 5: Scouting Force ---

func TestScoutingForce_Affray_3NonHeroOutside(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P1 territory is (0,0)-(60,12). Place non-Hero units outside.
	g.CreateUnit("P1 Scout1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 20}, 1.0) // outside P1 territory
	g.CreateUnit("P1 Scout2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 22}, 1.0) // outside
	g.CreateUnit("P1 Scout3", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 50, Y: 30}, 1.0) // outside

	tactic := GetBattleTacticCard(CardScoutingForce).GetTactic(TierAffray)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 non-Hero units outside own territory, Affray should succeed")
	}
}

func TestScoutingForce_Affray_HeroesDontCount(t *testing.T) {
	g := setupBattleTacticGame(1)

	// 2 non-heroes outside + 1 Hero outside = only 2 count
	g.CreateUnit("P1 Scout1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 20}, 1.0)
	g.CreateUnit("P1 Scout2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 22}, 1.0)
	hero := g.CreateUnit("P1 Hero", 1, core.Stats{Move: 5, Save: 3, Control: 2, Health: 5},
		nil, 1, core.Position{X: 50, Y: 30}, 1.0)
	hero.Keywords = append(hero.Keywords, core.KeywordHero)

	tactic := GetBattleTacticCard(CardScoutingForce).GetTactic(TierAffray)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("only 2 non-Hero units outside, Affray should fail")
	}
}

func TestScoutingForce_Strike_2NonHeroInEnemy(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P2 territory (0,32)-(60,44)
	g.CreateUnit("P1 Scout1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 35}, 1.0) // in P2 territory
	g.CreateUnit("P1 Scout2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 40}, 1.0) // in P2 territory

	tactic := GetBattleTacticCard(CardScoutingForce).GetTactic(TierStrike)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("2 non-Hero units in enemy territory, Strike should succeed")
	}
}

func TestScoutingForce_Domination_3InEnemyAndNoEnemyHome(t *testing.T) {
	g := setupBattleTacticGame(1)

	// 3 P1 non-Hero in P2 territory (0,32)-(60,44)
	g.CreateUnit("P1 Scout1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 35}, 1.0)
	g.CreateUnit("P1 Scout2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 40}, 1.0)
	g.CreateUnit("P1 Scout3", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 50, Y: 38}, 1.0)
	// P2 units NOT in P1 territory (0,0)-(60,12)
	g.CreateUnit("P2 Far", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 3, core.Position{X: 30, Y: 22}, 1.0) // In no-man's-land

	tactic := GetBattleTacticCard(CardScoutingForce).GetTactic(TierDomination)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("3 non-Hero in enemy territory, no enemy in own territory, Domination should succeed")
	}
}

func TestScoutingForce_Domination_FailsIfEnemyInMyTerritory(t *testing.T) {
	g := setupBattleTacticGame(1)

	// 3 P1 non-Hero in P2 territory
	g.CreateUnit("P1 Scout1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 35}, 1.0)
	g.CreateUnit("P1 Scout2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 30, Y: 40}, 1.0)
	g.CreateUnit("P1 Scout3", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 50, Y: 38}, 1.0)
	// P2 unit inside P1 territory (0,0)-(60,12)
	g.CreateUnit("P2 Invader", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 3, core.Position{X: 30, Y: 5}, 1.0)

	tactic := GetBattleTacticCard(CardScoutingForce).GetTactic(TierDomination)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("enemy unit in own territory, Domination should fail")
	}
}

// --- Card 6: Attuned to Ghyran ---

func TestAttunedToGhyran_Affray_2NearCentre(t *testing.T) {
	g := setupBattleTacticGame(1)

	// Centre is (30, 22). Place 2 units within 12" and not engaged.
	g.CreateUnit("P1 Unit1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 25, Y: 22}, 1.0) // 5" from centre
	g.CreateUnit("P1 Unit2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 35, Y: 22}, 1.0) // 5" from centre

	tactic := GetBattleTacticCard(CardAttunedToGhyran).GetTactic(TierAffray)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("2 units near centre and not in combat, Affray should succeed")
	}
}

func TestAttunedToGhyran_Affray_EngagedDontCount(t *testing.T) {
	g := setupBattleTacticGame(1)

	// Place 2 P1 units near centre but one is engaged (enemy within 3")
	g.CreateUnit("P1 Unit1", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 25, Y: 22}, 1.0)
	g.CreateUnit("P1 Unit2", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 35, Y: 22}, 1.0)
	// Enemy right next to P1 Unit2 => engaged
	g.CreateUnit("P2 Melee", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 36, Y: 22}, 1.0) // within 3" of Unit2

	tactic := GetBattleTacticCard(CardAttunedToGhyran).GetTactic(TierAffray)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("only 1 non-engaged unit near centre, Affray should fail")
	}
}

func TestAttunedToGhyran_Strike_Control2Pairs(t *testing.T) {
	g := setupBattleTacticGame(1)

	// Control Oakenbrow pair 1: (10,10) and (50,34)
	g.CreateUnit("P1 Oak1", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("P1 Oak2", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 50, Y: 34}, 1.0)
	// Control Gnarlroot pair 2: (50,10) and (10,34)
	g.CreateUnit("P1 Gnar1", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 50, Y: 10}, 1.0)
	g.CreateUnit("P1 Gnar2", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 10, Y: 34}, 1.0)

	tactic := GetBattleTacticCard(CardAttunedToGhyran).GetTactic(TierStrike)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("P1 controls 2 pairs, Strike should succeed")
	}
}

func TestAttunedToGhyran_Domination_NoEnemyInTerritoryAndMajorityPairs(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P1 controls 2 of 3 pairs (majority)
	g.CreateUnit("P1 Oak1", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("P1 Oak2", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 50, Y: 34}, 1.0)
	g.CreateUnit("P1 Gnar1", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 50, Y: 10}, 1.0)
	g.CreateUnit("P1 Gnar2", 1, core.Stats{Move: 5, Save: 4, Control: 2, Health: 1},
		nil, 5, core.Position{X: 10, Y: 34}, 1.0)
	// P2 NOT in P1 territory (stays in no-man's-land)
	g.CreateUnit("P2 Far", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 3, core.Position{X: 30, Y: 22}, 1.0)

	tactic := GetBattleTacticCard(CardAttunedToGhyran).GetTactic(TierDomination)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("majority pairs + no enemy in territory, Domination should succeed")
	}
}

// --- Game-Level Battle Tactic Integration ---

func TestGame_InitBattleTactics(t *testing.T) {
	g := setupBattleTacticGame(1)
	if len(g.BattleTactics) != 2 {
		t.Errorf("expected 2 trackers, got %d", len(g.BattleTactics))
	}
	for _, p := range g.Players {
		if g.BattleTactics[p.ID()] == nil {
			t.Errorf("expected tracker for player %d", p.ID())
		}
	}
}

func TestGame_SelectBattleTactic(t *testing.T) {
	g := setupBattleTacticGame(1)
	err := g.SelectBattleTactic(1, CardBrokenRanks, TierStrike)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.BattleTactics[1].ActiveTactic == nil {
		t.Fatal("expected active tactic after selection")
	}
	if g.BattleTactics[1].ActiveTactic.Tactic.Name != "Shatter the Lines" {
		t.Errorf("expected 'Shatter the Lines', got '%s'", g.BattleTactics[1].ActiveTactic.Tactic.Name)
	}
}

func TestGame_SelectBattleTactic_InvalidPlayer(t *testing.T) {
	g := setupBattleTacticGame(1)
	err := g.SelectBattleTactic(99, CardBrokenRanks, TierStrike)
	if err == nil {
		t.Error("expected error for invalid player")
	}
}

func TestGame_EvaluateAndScoreBattleTactic_Success(t *testing.T) {
	g := setupBattleTacticGame(1)
	g.UnitsDestroyedThisTurnMap[1] = 2

	_ = g.SelectBattleTactic(1, CardBrokenRanks, TierStrike)
	vp := g.EvaluateAndScoreBattleTactic(1)
	if vp != BattleTacticVP {
		t.Errorf("expected %d VP, got %d", BattleTacticVP, vp)
	}
	if g.VictoryPoints[1] != BattleTacticVP {
		t.Errorf("expected total VP %d, got %d", BattleTacticVP, g.VictoryPoints[1])
	}
}

func TestGame_EvaluateAndScoreBattleTactic_Failure(t *testing.T) {
	g := setupBattleTacticGame(1)
	g.UnitsDestroyedThisTurnMap[1] = 0

	_ = g.SelectBattleTactic(1, CardBrokenRanks, TierAffray)
	vp := g.EvaluateAndScoreBattleTactic(1)
	if vp != 0 {
		t.Errorf("expected 0 VP on failure, got %d", vp)
	}
	if len(g.BattleTactics[1].FailedTactics) != 1 {
		t.Errorf("expected 1 failed tactic, got %d", len(g.BattleTactics[1].FailedTactics))
	}
}

func TestGame_EvaluateAndScore_NoActiveTactic(t *testing.T) {
	g := setupBattleTacticGame(1)
	vp := g.EvaluateAndScoreBattleTactic(1)
	if vp != 0 {
		t.Errorf("expected 0 VP with no active tactic, got %d", vp)
	}
}

// --- Destruction Tracking ---

func TestGame_SnapshotAliveUnits(t *testing.T) {
	g := setupBattleTacticGame(1)
	g.CreateUnit("P1 Unit", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("P2 Unit1", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 40, Y: 10}, 1.0)
	g.CreateUnit("P2 Unit2", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 50, Y: 10}, 1.0)

	snapshot := g.SnapshotAliveUnits(1)
	if len(snapshot) != 2 {
		t.Errorf("expected 2 enemy units in snapshot, got %d", len(snapshot))
	}
}

func TestGame_CountNewDestructions(t *testing.T) {
	g := setupBattleTacticGame(1)
	g.CreateUnit("P1 Unit", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	enemy1 := g.CreateUnit("P2 Unit1", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 40, Y: 10}, 1.0)
	g.CreateUnit("P2 Unit2", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 50, Y: 10}, 1.0)

	snapshot := g.SnapshotAliveUnits(1)

	// Destroy one enemy unit
	for i := range enemy1.Models {
		enemy1.Models[i].IsAlive = false
		enemy1.Models[i].CurrentWounds = 0
	}

	count := g.CountNewDestructions(1, snapshot)
	if count != 1 {
		t.Errorf("expected 1 newly destroyed unit, got %d", count)
	}
}

// --- Seize the Initiative ---

func TestRollOffPriorityWithSeize_Round1_NoSeize(t *testing.T) {
	g := setupBattleTacticGame(42)
	g.BattleRound = 1
	g.PreviousSecondPlayer = -1

	// Round 1: no seize attempt
	first, _ := g.rollOffPriorityWithSeize()
	// Just verify it returns some valid result (depends on dice seed)
	if first != 0 && first != 1 {
		t.Errorf("expected player index 0 or 1, got %d", first)
	}
}

func TestRollOffPriorityWithSeize_Round2_AttemptSeize(t *testing.T) {
	// Use a seed where priority roll gives specific results
	g := setupBattleTacticGame(1)
	g.BattleRound = 2
	g.PreviousSecondPlayer = 1 // P2 (index 1) went second last round
	g.Commands.InitRound([]int{1, 2}, 4, -1)

	// Run it - we can't guarantee the dice result but verify no panic
	first, second := g.rollOffPriorityWithSeize()
	if first == second {
		t.Error("first and second should be different players")
	}
}

// --- GameView Battle Tactics ---

func TestGameView_BattleTactics(t *testing.T) {
	g := setupBattleTacticGame(1)
	_ = g.SelectBattleTactic(1, CardSavageSpearhead, TierAffray)

	view := g.View(1)
	if view.BattleTactics == nil {
		t.Fatal("expected BattleTactics in view")
	}
	if len(view.BattleTactics) != 2 {
		t.Errorf("expected 2 player tactics views, got %d", len(view.BattleTactics))
	}

	p1View := view.BattleTactics[1]
	if p1View == nil {
		t.Fatal("expected P1 tactics view")
	}
	if p1View.ActiveTactic == nil {
		t.Fatal("expected active tactic in P1 view")
	}
	if p1View.ActiveTactic.TacticName != "Aggressive Expansion" {
		t.Errorf("expected 'Aggressive Expansion', got '%s'", p1View.ActiveTactic.TacticName)
	}
	if p1View.ActiveTactic.Tier != "Affray" {
		t.Errorf("expected tier 'Affray', got '%s'", p1View.ActiveTactic.Tier)
	}
	// Card remains available until completed/failed, so still 6
	if len(p1View.AvailableCards) != 6 {
		t.Errorf("expected 6 available cards for P1 (not removed until resolved), got %d", len(p1View.AvailableCards))
	}
}

func TestGameView_BattleTacticsHistory(t *testing.T) {
	g := setupBattleTacticGame(1)

	// Complete one, fail one
	_ = g.SelectBattleTactic(1, CardBrokenRanks, TierAffray)
	g.UnitsDestroyedThisTurnMap[1] = 1
	g.EvaluateAndScoreBattleTactic(1)
	g.BattleTactics[1].ResetRound()

	_ = g.SelectBattleTactic(1, CardSavageSpearhead, TierDomination)
	g.UnitsDestroyedThisTurnMap[1] = 0
	g.EvaluateAndScoreBattleTactic(1) // Should fail (no objectives)
	g.BattleTactics[1].ResetRound()

	view := g.View(1)
	p1View := view.BattleTactics[1]
	if p1View.History.CompletedCount != 1 {
		t.Errorf("expected 1 completed, got %d", p1View.History.CompletedCount)
	}
	if p1View.History.FailedCount != 1 {
		t.Errorf("expected 1 failed, got %d", p1View.History.FailedCount)
	}
}

// --- No Battleplan: tactics need battleplan ---

func TestConquerAndHold_NoBattleplan_ReturnsFalse(t *testing.T) {
	g := NewGame(1, 60, 44)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	tactic := GetBattleTacticCard(CardConquerAndHold).GetTactic(TierAffray)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("Conquer and Hold should fail without battleplan (no territories)")
	}
}

func TestScoutingForce_NoBattleplan_ReturnsFalse(t *testing.T) {
	g := NewGame(1, 60, 44)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	tactic := GetBattleTacticCard(CardScoutingForce).GetTactic(TierAffray)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("Scouting Force should fail without battleplan (no territories)")
	}
}

// --- Underdog (for Ferocious Advance Strike tier) ---

func TestDetermineUnderdog_FewerWounds(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P1 has fewer total wounds remaining
	g.CreateUnit("P1 Small", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 3, core.Position{X: 10, Y: 10}, 1.0) // 3 wounds total
	g.CreateUnit("P2 Big", 2, core.Stats{Move: 5, Save: 3, Control: 2, Health: 3},
		nil, 5, core.Position{X: 40, Y: 34}, 1.0) // 15 wounds total

	underdog := g.determineUnderdog()
	if underdog != 1 {
		t.Errorf("expected P1 (id=1) as underdog, got %d", underdog)
	}
}

func TestDetermineUnderdog_Tied(t *testing.T) {
	g := setupBattleTacticGame(1)

	g.CreateUnit("P1 Unit", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 10, Y: 10}, 1.0)
	g.CreateUnit("P2 Unit", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 40, Y: 34}, 1.0)

	underdog := g.determineUnderdog()
	if underdog != -1 {
		t.Errorf("expected no underdog (-1), got %d", underdog)
	}
}

func TestFerocousAdvance_Strike_UnderdogAndFought(t *testing.T) {
	g := setupBattleTacticGame(1)

	// Make P1 the underdog (fewer wounds)
	u1 := g.CreateUnit("P1 A", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 2, core.Position{X: 10, Y: 10}, 1.0) // 2 wounds
	u2 := g.CreateUnit("P1 B", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 2, core.Position{X: 20, Y: 10}, 1.0) // 2 wounds
	g.CreateUnit("P2 Big", 2, core.Stats{Move: 5, Save: 3, Control: 2, Health: 3},
		nil, 10, core.Position{X: 40, Y: 34}, 1.0) // 30 wounds

	u1.HasFought = true
	u2.HasFought = true

	tactic := GetBattleTacticCard(CardFerocousAdvance).GetTactic(TierStrike)
	if !g.EvaluateBattleTactic(1, tactic) {
		t.Error("P1 is underdog and 2 units fought, Strike should succeed")
	}
}

func TestFerocousAdvance_Strike_NotUnderdogFails(t *testing.T) {
	g := setupBattleTacticGame(1)

	// P1 has MORE wounds (not underdog)
	u1 := g.CreateUnit("P1 A", 1, core.Stats{Move: 5, Save: 3, Control: 2, Health: 3},
		nil, 10, core.Position{X: 10, Y: 10}, 1.0) // 30 wounds
	u2 := g.CreateUnit("P1 B", 1, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 5, core.Position{X: 20, Y: 10}, 1.0) // 5 wounds
	g.CreateUnit("P2 Small", 2, core.Stats{Move: 5, Save: 4, Control: 1, Health: 1},
		nil, 2, core.Position{X: 40, Y: 34}, 1.0) // 2 wounds

	u1.HasFought = true
	u2.HasFought = true

	tactic := GetBattleTacticCard(CardFerocousAdvance).GetTactic(TierStrike)
	if g.EvaluateBattleTactic(1, tactic) {
		t.Error("P1 is NOT underdog, Strike should fail regardless of fighting")
	}
}
