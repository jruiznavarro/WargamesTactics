package simulation

import (
	"fmt"
	"time"
)

// GameResult holds the outcome of a single simulated game.
type GameResult struct {
	Seed          int64          // RNG seed used
	Winner        int            // Player ID of winner (-1 = draw)
	VictoryPoints map[int]int    // Final VP per player
	FinalRound    int            // Round at which game ended
	TotalRounds   int            // Max rounds configured
	Duration      time.Duration  // Wall-clock time for this game
	ImmediateWin  bool           // True if won by destroying all enemy units
	UnitsAlive    map[int]int    // Alive units per player at game end
}

// IsDraw returns true if the game ended in a draw.
func (r *GameResult) IsDraw() bool {
	return r.Winner == -1
}

// VPMargin returns the VP difference between winner and loser.
// Returns 0 for draws.
func (r *GameResult) VPMargin() int {
	if r.IsDraw() {
		return 0
	}
	winnerVP := r.VictoryPoints[r.Winner]
	maxLoserVP := 0
	for id, vp := range r.VictoryPoints {
		if id != r.Winner && vp > maxLoserVP {
			maxLoserVP = vp
		}
	}
	return winnerVP - maxLoserVP
}

// MatchupStats holds aggregated results for a series of simulated games
// between two specific armies/factions.
type MatchupStats struct {
	Player1ID   int    // Player 1 ID
	Player2ID   int    // Player 2 ID
	Player1Name string // Player 1 faction/name
	Player2Name string // Player 2 faction/name

	TotalGames  int
	Player1Wins int
	Player2Wins int
	Draws       int

	// VP distribution
	TotalVPPlayer1   int
	TotalVPPlayer2   int
	MaxVPPlayer1     int
	MaxVPPlayer2     int
	MinVPPlayer1     int
	MinVPPlayer2     int

	// Win condition breakdown
	ImmediateWins1 int // P1 wins by destruction
	ImmediateWins2 int // P2 wins by destruction

	// Round distribution
	TotalRoundsPlayed int

	// VP margins
	TotalVPMargin int // Sum of VP margins across all games (for avg)
	MaxVPMargin   int

	// Individual results for detailed analysis
	Results []GameResult
}

// NewMatchupStats creates a new MatchupStats for two players.
func NewMatchupStats(p1ID int, p1Name string, p2ID int, p2Name string) *MatchupStats {
	return &MatchupStats{
		Player1ID:    p1ID,
		Player2ID:    p2ID,
		Player1Name:  p1Name,
		Player2Name:  p2Name,
		MinVPPlayer1: 1<<31 - 1, // MaxInt
		MinVPPlayer2: 1<<31 - 1,
	}
}

// AddResult incorporates a game result into the statistics.
func (s *MatchupStats) AddResult(r GameResult) {
	s.TotalGames++
	s.Results = append(s.Results, r)

	switch r.Winner {
	case s.Player1ID:
		s.Player1Wins++
		if r.ImmediateWin {
			s.ImmediateWins1++
		}
	case s.Player2ID:
		s.Player2Wins++
		if r.ImmediateWin {
			s.ImmediateWins2++
		}
	default:
		s.Draws++
	}

	vp1 := r.VictoryPoints[s.Player1ID]
	vp2 := r.VictoryPoints[s.Player2ID]

	s.TotalVPPlayer1 += vp1
	s.TotalVPPlayer2 += vp2

	if vp1 > s.MaxVPPlayer1 {
		s.MaxVPPlayer1 = vp1
	}
	if vp2 > s.MaxVPPlayer2 {
		s.MaxVPPlayer2 = vp2
	}
	if vp1 < s.MinVPPlayer1 {
		s.MinVPPlayer1 = vp1
	}
	if vp2 < s.MinVPPlayer2 {
		s.MinVPPlayer2 = vp2
	}

	s.TotalRoundsPlayed += r.FinalRound

	margin := r.VPMargin()
	s.TotalVPMargin += margin
	if margin > s.MaxVPMargin {
		s.MaxVPMargin = margin
	}
}

// WinRate returns P1 win rate as a float [0,1].
func (s *MatchupStats) WinRate(playerID int) float64 {
	if s.TotalGames == 0 {
		return 0
	}
	wins := 0
	if playerID == s.Player1ID {
		wins = s.Player1Wins
	} else {
		wins = s.Player2Wins
	}
	return float64(wins) / float64(s.TotalGames)
}

// DrawRate returns draw rate as a float [0,1].
func (s *MatchupStats) DrawRate() float64 {
	if s.TotalGames == 0 {
		return 0
	}
	return float64(s.Draws) / float64(s.TotalGames)
}

// AvgVP returns the average VP scored by a player.
func (s *MatchupStats) AvgVP(playerID int) float64 {
	if s.TotalGames == 0 {
		return 0
	}
	if playerID == s.Player1ID {
		return float64(s.TotalVPPlayer1) / float64(s.TotalGames)
	}
	return float64(s.TotalVPPlayer2) / float64(s.TotalGames)
}

// AvgRounds returns the average number of rounds games lasted.
func (s *MatchupStats) AvgRounds() float64 {
	if s.TotalGames == 0 {
		return 0
	}
	return float64(s.TotalRoundsPlayed) / float64(s.TotalGames)
}

// AvgVPMargin returns the average VP margin in decided games.
func (s *MatchupStats) AvgVPMargin() float64 {
	decided := s.TotalGames - s.Draws
	if decided == 0 {
		return 0
	}
	return float64(s.TotalVPMargin) / float64(decided)
}

// Summary returns a human-readable summary of the matchup.
func (s *MatchupStats) Summary() string {
	if s.TotalGames == 0 {
		return "No games played"
	}

	return fmt.Sprintf(
		`=== Matchup: %s vs %s ===
Games: %d
%s wins: %d (%.1f%%) [%d by destruction]
%s wins: %d (%.1f%%) [%d by destruction]
Draws: %d (%.1f%%)

VP Statistics:
  %s: avg %.1f (min %d, max %d)
  %s: avg %.1f (min %d, max %d)
  Avg margin: %.1f (max %d)

Game Duration: avg %.1f rounds`,
		s.Player1Name, s.Player2Name,
		s.TotalGames,
		s.Player1Name, s.Player1Wins, s.WinRate(s.Player1ID)*100, s.ImmediateWins1,
		s.Player2Name, s.Player2Wins, s.WinRate(s.Player2ID)*100, s.ImmediateWins2,
		s.Draws, s.DrawRate()*100,
		s.Player1Name, s.AvgVP(s.Player1ID), s.MinVPPlayer1, s.MaxVPPlayer1,
		s.Player2Name, s.AvgVP(s.Player2ID), s.MinVPPlayer2, s.MaxVPPlayer2,
		s.AvgVPMargin(), s.MaxVPMargin,
		s.AvgRounds(),
	)
}
