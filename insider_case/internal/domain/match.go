package domain

import "time"

// Match represents a single fixture between two teams in a given week.
// HomeGoals and AwayGoals are pointers — nil means the match has not been played.
type Match struct {
	ID         uint      `json:"id"`
	WeekID     int       `json:"week_id"` // 1–6 for a 4-team double round-robin
	HomeTeamID uint      `json:"home_team_id"`
	HomeTeam   Team      `json:"home_team"`
	AwayTeamID uint      `json:"away_team_id"`
	AwayTeam   Team      `json:"away_team"`
	HomeGoals  *int      `json:"home_goals"` // nil = not yet played
	AwayGoals  *int      `json:"away_goals"` // nil = not yet played
	Played     bool      `json:"played"`
	Reasoning  string    `json:"reasoning"` // audit log: why this score happened
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MatchResult is a value object returned after a match is simulated.
// It carries the final score and the human-readable reasoning string.
type MatchResult struct {
	Match     Match  `json:"match"`
	HomeGoals int    `json:"home_goals"`
	AwayGoals int    `json:"away_goals"`
	Reasoning string `json:"reasoning"`
}

// Outcome represents the result from one team's perspective.
type Outcome int

const (
	OutcomeWin  Outcome = iota // 3 points
	OutcomeDraw                // 1 point
	OutcomeLoss                // 0 points
)

// Points returns the league points for a given outcome.
func (o Outcome) Points() int {
	switch o {
	case OutcomeWin:
		return PointsWin
	case OutcomeDraw:
		return PointsDraw
	default:
		return PointsLoss
	}
}

// DetermineOutcome returns the outcome for the home team given the final goals.
func DetermineOutcome(homeGoals, awayGoals int) (homeOutcome, awayOutcome Outcome) {
	switch {
	case homeGoals > awayGoals:
		return OutcomeWin, OutcomeLoss
	case homeGoals < awayGoals:
		return OutcomeLoss, OutcomeWin
	default:
		return OutcomeDraw, OutcomeDraw
	}
}
