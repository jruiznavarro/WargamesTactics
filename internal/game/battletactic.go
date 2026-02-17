package game

import (
	"fmt"

	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

// BattleTacticTier represents the difficulty tier of a battle tactic.
type BattleTacticTier int

const (
	TierAffray      BattleTacticTier = iota // Easiest tier
	TierStrike                               // Medium tier
	TierDomination                           // Hardest tier
)

func (t BattleTacticTier) String() string {
	switch t {
	case TierAffray:
		return "Affray"
	case TierStrike:
		return "Strike"
	case TierDomination:
		return "Domination"
	default:
		return "Unknown"
	}
}

// BattleTacticCardID identifies a battle tactic card (1-6).
type BattleTacticCardID int

const (
	CardSavageSpearhead  BattleTacticCardID = 1
	CardBrokenRanks      BattleTacticCardID = 2
	CardConquerAndHold   BattleTacticCardID = 3
	CardFerocousAdvance  BattleTacticCardID = 4
	CardScoutingForce    BattleTacticCardID = 5
	CardAttunedToGhyran  BattleTacticCardID = 6
)

// BattleTacticVP is the VP value for completing a battle tactic.
const BattleTacticVP = 4

// BattleTactic represents a single tactic option (one tier of one card).
type BattleTactic struct {
	CardID      BattleTacticCardID
	CardName    string
	Tier        BattleTacticTier
	Name        string // e.g. "Aggressive Expansion"
	Description string // Human-readable condition
	VP          int    // Victory points awarded
}

// BattleTacticCard groups the 3 tiers of a single card.
type BattleTacticCard struct {
	ID       BattleTacticCardID
	Name     string
	Affray   BattleTactic
	Strike   BattleTactic
	Domination BattleTactic
}

// GetTactic returns the tactic for the given tier.
func (c *BattleTacticCard) GetTactic(tier BattleTacticTier) BattleTactic {
	switch tier {
	case TierAffray:
		return c.Affray
	case TierStrike:
		return c.Strike
	case TierDomination:
		return c.Domination
	default:
		return c.Affray
	}
}

// ActiveBattleTactic tracks a player's selected tactic for the current battle round.
type ActiveBattleTactic struct {
	Tactic    BattleTactic
	Completed bool
}

// BattleTacticTracker manages battle tactic selection and scoring per player.
type BattleTacticTracker struct {
	// Available cards (initially all 6, reduced as used)
	AvailableCards map[BattleTacticCardID]bool
	// Currently active tactic (nil if not selected yet this round)
	ActiveTactic *ActiveBattleTactic
	// History of completed tactics
	CompletedTactics []BattleTactic
	// History of failed tactics
	FailedTactics []BattleTactic
}

// NewBattleTacticTracker creates a tracker with all 6 cards available.
func NewBattleTacticTracker() *BattleTacticTracker {
	available := make(map[BattleTacticCardID]bool)
	for id := CardSavageSpearhead; id <= CardAttunedToGhyran; id++ {
		available[id] = true
	}
	return &BattleTacticTracker{
		AvailableCards: available,
	}
}

// SelectTactic selects a battle tactic for the current round.
func (bt *BattleTacticTracker) SelectTactic(cardID BattleTacticCardID, tier BattleTacticTier) error {
	if bt.ActiveTactic != nil {
		return fmt.Errorf("already selected tactic '%s' for this round", bt.ActiveTactic.Tactic.Name)
	}
	if !bt.AvailableCards[cardID] {
		return fmt.Errorf("card %d is not available (already used)", cardID)
	}

	card := GetBattleTacticCard(cardID)
	if card == nil {
		return fmt.Errorf("unknown card ID %d", cardID)
	}

	tactic := card.GetTactic(tier)
	bt.ActiveTactic = &ActiveBattleTactic{Tactic: tactic}
	return nil
}

// CompleteTactic marks the current tactic as completed and returns VP scored.
func (bt *BattleTacticTracker) CompleteTactic() int {
	if bt.ActiveTactic == nil || bt.ActiveTactic.Completed {
		return 0
	}
	bt.ActiveTactic.Completed = true
	bt.CompletedTactics = append(bt.CompletedTactics, bt.ActiveTactic.Tactic)
	// Remove card from available
	delete(bt.AvailableCards, bt.ActiveTactic.Tactic.CardID)
	return bt.ActiveTactic.Tactic.VP
}

// FailTactic marks the current tactic as failed (not completed).
func (bt *BattleTacticTracker) FailTactic() {
	if bt.ActiveTactic == nil {
		return
	}
	bt.FailedTactics = append(bt.FailedTactics, bt.ActiveTactic.Tactic)
	// Remove card from available (can't reuse even on failure)
	delete(bt.AvailableCards, bt.ActiveTactic.Tactic.CardID)
}

// ResetRound clears the active tactic for a new round.
func (bt *BattleTacticTracker) ResetRound() {
	bt.ActiveTactic = nil
}

// --- Battle Tactic Card Definitions ---

// AllBattleTacticCards returns all 6 battle tactic cards.
func AllBattleTacticCards() []BattleTacticCard {
	return []BattleTacticCard{
		cardSavageSpearhead(),
		cardBrokenRanks(),
		cardConquerAndHold(),
		cardFerocousAdvance(),
		cardScoutingForce(),
		cardAttunedToGhyran(),
	}
}

// GetBattleTacticCard returns the card with the given ID, or nil.
func GetBattleTacticCard(id BattleTacticCardID) *BattleTacticCard {
	cards := AllBattleTacticCards()
	for i := range cards {
		if cards[i].ID == id {
			return &cards[i]
		}
	}
	return nil
}

func cardSavageSpearhead() BattleTacticCard {
	return BattleTacticCard{
		ID:   CardSavageSpearhead,
		Name: "Savage Spearhead",
		Affray: BattleTactic{
			CardID: CardSavageSpearhead, CardName: "Savage Spearhead",
			Tier: TierAffray, Name: "Aggressive Expansion", VP: BattleTacticVP,
			Description: "You control more objectives than your opponent.",
		},
		Strike: BattleTactic{
			CardID: CardSavageSpearhead, CardName: "Savage Spearhead",
			Tier: TierStrike, Name: "Seize the Centre", VP: BattleTacticVP,
			Description: "You control more objectives than your opponent and at least 1 objective within 12\" of the centre of the battlefield.",
		},
		Domination: BattleTactic{
			CardID: CardSavageSpearhead, CardName: "Savage Spearhead",
			Tier: TierDomination, Name: "Total Conquest", VP: BattleTacticVP,
			Description: "You control all objectives on the battlefield.",
		},
	}
}

func cardBrokenRanks() BattleTacticCard {
	return BattleTacticCard{
		ID:   CardBrokenRanks,
		Name: "Broken Ranks",
		Affray: BattleTactic{
			CardID: CardBrokenRanks, CardName: "Broken Ranks",
			Tier: TierAffray, Name: "Broken Ranks", VP: BattleTacticVP,
			Description: "At least 1 enemy unit was destroyed this turn.",
		},
		Strike: BattleTactic{
			CardID: CardBrokenRanks, CardName: "Broken Ranks",
			Tier: TierStrike, Name: "Shatter the Lines", VP: BattleTacticVP,
			Description: "At least 2 enemy units were destroyed this turn.",
		},
		Domination: BattleTactic{
			CardID: CardBrokenRanks, CardName: "Broken Ranks",
			Tier: TierDomination, Name: "Annihilation", VP: BattleTacticVP,
			Description: "At least 3 enemy units were destroyed this turn.",
		},
	}
}

func cardConquerAndHold() BattleTacticCard {
	return BattleTacticCard{
		ID:   CardConquerAndHold,
		Name: "Conquer and Hold",
		Affray: BattleTactic{
			CardID: CardConquerAndHold, CardName: "Conquer and Hold",
			Tier: TierAffray, Name: "Strategic Positions", VP: BattleTacticVP,
			Description: "At least 2 friendly units are in enemy territory.",
		},
		Strike: BattleTactic{
			CardID: CardConquerAndHold, CardName: "Conquer and Hold",
			Tier: TierStrike, Name: "Deep Strike", VP: BattleTacticVP,
			Description: "At least 3 friendly units are in enemy territory.",
		},
		Domination: BattleTactic{
			CardID: CardConquerAndHold, CardName: "Conquer and Hold",
			Tier: TierDomination, Name: "Stranglehold", VP: BattleTacticVP,
			Description: "At least 3 friendly units are in enemy territory and you control at least 1 objective in enemy territory.",
		},
	}
}

func cardFerocousAdvance() BattleTacticCard {
	return BattleTacticCard{
		ID:   CardFerocousAdvance,
		Name: "Ferocious Advance",
		Affray: BattleTactic{
			CardID: CardFerocousAdvance, CardName: "Ferocious Advance",
			Tier: TierAffray, Name: "Defiant Surge", VP: BattleTacticVP,
			Description: "At least 3 friendly units ran or charged this turn.",
		},
		Strike: BattleTactic{
			CardID: CardFerocousAdvance, CardName: "Ferocious Advance",
			Tier: TierStrike, Name: "Daring Resurgence", VP: BattleTacticVP,
			Description: "You are the underdog and at least 2 friendly units used a fight ability this turn.",
		},
		Domination: BattleTactic{
			CardID: CardFerocousAdvance, CardName: "Ferocious Advance",
			Tier: TierDomination, Name: "Overwhelming Assault", VP: BattleTacticVP,
			Description: "At least 3 friendly units charged this turn and at least 1 enemy unit was destroyed.",
		},
	}
}

func cardScoutingForce() BattleTacticCard {
	return BattleTacticCard{
		ID:   CardScoutingForce,
		Name: "Scouting Force",
		Affray: BattleTactic{
			CardID: CardScoutingForce, CardName: "Scouting Force",
			Tier: TierAffray, Name: "Raiding Party", VP: BattleTacticVP,
			Description: "At least 3 friendly non-Hero units are outside your territory.",
		},
		Strike: BattleTactic{
			CardID: CardScoutingForce, CardName: "Scouting Force",
			Tier: TierStrike, Name: "Bold Explorers", VP: BattleTacticVP,
			Description: "At least 2 friendly non-Hero units are in enemy territory.",
		},
		Domination: BattleTactic{
			CardID: CardScoutingForce, CardName: "Scouting Force",
			Tier: TierDomination, Name: "Behind Enemy Lines", VP: BattleTacticVP,
			Description: "At least 3 friendly non-Hero units are in enemy territory and no enemy units in your territory.",
		},
	}
}

func cardAttunedToGhyran() BattleTacticCard {
	return BattleTacticCard{
		ID:   CardAttunedToGhyran,
		Name: "Attuned to Ghyran",
		Affray: BattleTactic{
			CardID: CardAttunedToGhyran, CardName: "Attuned to Ghyran",
			Tier: TierAffray, Name: "Sacred Centrality", VP: BattleTacticVP,
			Description: "At least 2 friendly units are within 12\" of the centre of the board and not in combat.",
		},
		Strike: BattleTactic{
			CardID: CardAttunedToGhyran, CardName: "Attuned to Ghyran",
			Tier: TierStrike, Name: "Ritual Dominance", VP: BattleTacticVP,
			Description: "You control at least 2 complete objective pairs.",
		},
		Domination: BattleTactic{
			CardID: CardAttunedToGhyran, CardName: "Attuned to Ghyran",
			Tier: TierDomination, Name: "Purification Rites", VP: BattleTacticVP,
			Description: "No enemy units are in your territory and you control the majority of objective pairs.",
		},
	}
}

// --- Battle Tactic Condition Evaluation ---

// EvaluateBattleTactic checks if a battle tactic's condition is met.
func (g *Game) EvaluateBattleTactic(playerID int, tactic BattleTactic) bool {
	switch tactic.CardID {
	case CardSavageSpearhead:
		return g.evalSavageSpearhead(playerID, tactic.Tier)
	case CardBrokenRanks:
		return g.evalBrokenRanks(playerID, tactic.Tier)
	case CardConquerAndHold:
		return g.evalConquerAndHold(playerID, tactic.Tier)
	case CardFerocousAdvance:
		return g.evalFerocousAdvance(playerID, tactic.Tier)
	case CardScoutingForce:
		return g.evalScoutingForce(playerID, tactic.Tier)
	case CardAttunedToGhyran:
		return g.evalAttunedToGhyran(playerID, tactic.Tier)
	default:
		return false
	}
}

func (g *Game) evalSavageSpearhead(playerID int, tier BattleTacticTier) bool {
	// Ensure objective control is calculated
	if g.Battleplan != nil {
		g.CalculateGhyraniteObjectiveControl()
	} else {
		g.CalculateObjectiveControl()
	}

	myObj := g.ObjectivesControlledBy(playerID)
	opponentObj := 0
	for _, p := range g.Players {
		if p.ID() != playerID {
			opponentObj = g.ObjectivesControlledBy(p.ID())
			break
		}
	}

	switch tier {
	case TierAffray:
		// Control more objectives than opponent
		return myObj > opponentObj
	case TierStrike:
		// Control more objectives AND at least 1 within 12" of centre
		if myObj <= opponentObj {
			return false
		}
		centre := core.Position{X: g.Board.Width / 2, Y: g.Board.Height / 2}
		for _, obj := range g.Board.Objectives {
			ctrl, ok := g.ObjectiveControl[obj.ID]
			if ok && ctrl == playerID && core.Distance(obj.Position, centre) <= 12.0 {
				return true
			}
		}
		return false
	case TierDomination:
		// Control ALL objectives on the battlefield
		totalObj := len(g.Board.Objectives)
		return totalObj > 0 && myObj == totalObj
	}
	return false
}

func (g *Game) evalBrokenRanks(playerID int, tier BattleTacticTier) bool {
	destroyed := g.UnitsDestroyedThisTurn(playerID)
	switch tier {
	case TierAffray:
		return destroyed >= 1
	case TierStrike:
		return destroyed >= 2
	case TierDomination:
		return destroyed >= 3
	}
	return false
}

func (g *Game) evalConquerAndHold(playerID int, tier BattleTacticTier) bool {
	if g.Battleplan == nil {
		return false // Need territories
	}

	// Determine enemy territory index
	enemyTerritoryIdx := 1
	if g.Players[1].ID() == playerID {
		enemyTerritoryIdx = 0
	}
	enemyTerritory := g.Battleplan.Territories[enemyTerritoryIdx]

	unitsInEnemyTerritory := 0
	for _, u := range g.Units {
		if u.OwnerID == playerID && !u.IsDestroyed() {
			if enemyTerritory.Contains(u.Position()) {
				unitsInEnemyTerritory++
			}
		}
	}

	switch tier {
	case TierAffray:
		return unitsInEnemyTerritory >= 2
	case TierStrike:
		return unitsInEnemyTerritory >= 3
	case TierDomination:
		if unitsInEnemyTerritory < 3 {
			return false
		}
		// Also need to control at least 1 objective in enemy territory
		if g.Battleplan != nil {
			g.CalculateGhyraniteObjectiveControl()
		} else {
			g.CalculateObjectiveControl()
		}
		for _, obj := range g.Board.Objectives {
			ctrl, ok := g.ObjectiveControl[obj.ID]
			if ok && ctrl == playerID && enemyTerritory.Contains(obj.Position) {
				return true
			}
		}
		return false
	}
	return false
}

func (g *Game) evalFerocousAdvance(playerID int, tier BattleTacticTier) bool {
	switch tier {
	case TierAffray:
		// At least 3 friendly units ran or charged
		count := 0
		for _, u := range g.Units {
			if u.OwnerID == playerID && !u.IsDestroyed() && (u.HasRun || u.HasCharged) {
				count++
			}
		}
		return count >= 3
	case TierStrike:
		// Underdog AND at least 2 units fought
		isUnderdog := g.determineUnderdog() == playerID
		if !isUnderdog {
			return false
		}
		fought := 0
		for _, u := range g.Units {
			if u.OwnerID == playerID && !u.IsDestroyed() && u.HasFought {
				fought++
			}
		}
		return fought >= 2
	case TierDomination:
		// At least 3 charged AND at least 1 enemy destroyed
		charged := 0
		for _, u := range g.Units {
			if u.OwnerID == playerID && !u.IsDestroyed() && u.HasCharged {
				charged++
			}
		}
		return charged >= 3 && g.UnitsDestroyedThisTurn(playerID) >= 1
	}
	return false
}

func (g *Game) evalScoutingForce(playerID int, tier BattleTacticTier) bool {
	if g.Battleplan == nil {
		return false
	}

	// Determine territories
	myTerritoryIdx := 0
	enemyTerritoryIdx := 1
	if g.Players[1].ID() == playerID {
		myTerritoryIdx = 1
		enemyTerritoryIdx = 0
	}
	myTerritory := g.Battleplan.Territories[myTerritoryIdx]
	enemyTerritory := g.Battleplan.Territories[enemyTerritoryIdx]

	nonHeroOutside := 0
	nonHeroInEnemy := 0
	for _, u := range g.Units {
		if u.OwnerID == playerID && !u.IsDestroyed() && !u.HasKeyword(core.KeywordHero) {
			if !myTerritory.Contains(u.Position()) {
				nonHeroOutside++
			}
			if enemyTerritory.Contains(u.Position()) {
				nonHeroInEnemy++
			}
		}
	}

	switch tier {
	case TierAffray:
		return nonHeroOutside >= 3
	case TierStrike:
		return nonHeroInEnemy >= 2
	case TierDomination:
		if nonHeroInEnemy < 3 {
			return false
		}
		// No enemy units in my territory
		for _, u := range g.Units {
			if u.OwnerID != playerID && !u.IsDestroyed() {
				if myTerritory.Contains(u.Position()) {
					return false
				}
			}
		}
		return true
	}
	return false
}

func (g *Game) evalAttunedToGhyran(playerID int, tier BattleTacticTier) bool {
	centre := core.Position{X: g.Board.Width / 2, Y: g.Board.Height / 2}

	switch tier {
	case TierAffray:
		// At least 2 friendly units within 12" of centre, not in combat
		count := 0
		for _, u := range g.Units {
			if u.OwnerID == playerID && !u.IsDestroyed() && !g.isEngaged(u) {
				if core.Distance(u.Position(), centre) <= 12.0 {
					count++
				}
			}
		}
		return count >= 2
	case TierStrike:
		// Control at least 2 complete pairs
		if g.Battleplan == nil {
			return false
		}
		g.CalculateGhyraniteObjectiveControl()
		return g.PairsControlledBy(playerID) >= 2
	case TierDomination:
		// No enemy in your territory AND control majority of pairs
		if g.Battleplan == nil {
			return false
		}
		myTerritoryIdx := 0
		if g.Players[1].ID() == playerID {
			myTerritoryIdx = 1
		}
		myTerritory := g.Battleplan.Territories[myTerritoryIdx]
		for _, u := range g.Units {
			if u.OwnerID != playerID && !u.IsDestroyed() {
				if myTerritory.Contains(u.Position()) {
					return false
				}
			}
		}
		g.CalculateGhyraniteObjectiveControl()
		totalPairs := len(g.Board.PairIDs())
		myPairs := g.PairsControlledBy(playerID)
		return totalPairs > 0 && myPairs*2 > totalPairs
	}
	return false
}
