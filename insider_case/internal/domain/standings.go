package domain

// League scoring constants — Premier League rules.
const (
	PointsWin  = 3
	PointsDraw = 1
	PointsLoss = 0
)

// TotalWeeks is the number of weeks in a 4-team full double round-robin season.
// 4 teams × 3 opponents × 2 legs = 12 matches across 6 weeks (2 matches/week).
const TotalWeeks = 6

// TeamsInLeague is fixed for this simulation.
const TeamsInLeague = 4

// Standings holds the computed league table row for a single team.
// It is always rebuilt from scratch by the StandingsService — never stored
// incrementally — to prevent drift from match edits.
type Standings struct {
	Position     int    `json:"position"`
	TeamID       uint   `json:"team_id"`
	TeamName     string `json:"team_name"`
	TeamShort    string `json:"team_short"`
	Played       int    `json:"played"`
	Won          int    `json:"won"`
	Drawn        int    `json:"drawn"`
	Lost         int    `json:"lost"`
	GoalsFor     int    `json:"goals_for"`
	GoalsAgainst int    `json:"goals_against"`
	GoalDiff     int    `json:"goal_diff"`
	Points       int    `json:"points"`
}

// StandingsSnapshot is a persisted copy of the standings at a given week boundary.
// Used to reconstruct history and serve the /league/table endpoint.
type StandingsSnapshot struct {
	WeekID    int         `json:"week_id"`
	Standings []Standings `json:"standings"`
}
