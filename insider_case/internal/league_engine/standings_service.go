package league

import (
	"sort"

	"github.com/insider/football-simulator/internal/domain"
)

// BuildStandings computes the current league table from scratch using every
// played match in the slice. It intentionally ignores unplayed matches.
//
// Calling this after every simulation — rather than incrementally updating
// counters — guarantees correctness when match results are edited via
// PATCH /league/match/:id without any cache-invalidation logic.
//
// Sort order (Premier League rules):
//
//  1. Points (descending)
//  2. Goal Difference (descending)
//  3. Goals Scored / Goals For (descending)
func BuildStandings(teams []domain.Team, matches []domain.Match) []domain.Standings {
	// Seed one row per team using a map for O(1) lookup while processing matches.
	rows := make(map[uint]*domain.Standings, len(teams))
	for _, t := range teams {
		rows[t.ID] = &domain.Standings{
			TeamID:    t.ID,
			TeamName:  t.Name,
			TeamShort: t.ShortName,
		}
	}

	// Accumulate stats from every played match.
	for _, m := range matches {
		if !m.Played || m.HomeGoals == nil || m.AwayGoals == nil {
			continue
		}

		hg := *m.HomeGoals
		ag := *m.AwayGoals

		home := rows[m.HomeTeamID]
		away := rows[m.AwayTeamID]

		if home == nil || away == nil {
			// Defensive: skip if a team is missing from the seed map.
			continue
		}

		home.Played++
		away.Played++

		home.GoalsFor += hg
		home.GoalsAgainst += ag
		away.GoalsFor += ag
		away.GoalsAgainst += hg

		homeOutcome, awayOutcome := domain.DetermineOutcome(hg, ag)

		applyOutcome(home, homeOutcome)
		applyOutcome(away, awayOutcome)
	}

	// Compute goal difference for every row and collect into a slice.
	result := make([]domain.Standings, 0, len(rows))
	for _, row := range rows {
		row.GoalDiff = row.GoalsFor - row.GoalsAgainst
		result = append(result, *row)
	}

	// Sort: Points → GD → GF, all descending.
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if a.GoalDiff != b.GoalDiff {
			return a.GoalDiff > b.GoalDiff
		}
		return a.GoalsFor > b.GoalsFor
	})

	// Assign positions after sort so they are always 1-based and contiguous.
	for i := range result {
		result[i].Position = i + 1
	}

	return result
}

// applyOutcome increments win/draw/loss counters and adds the correct points.
func applyOutcome(row *domain.Standings, outcome domain.Outcome) {
	switch outcome {
	case domain.OutcomeWin:
		row.Won++
	case domain.OutcomeDraw:
		row.Drawn++
	case domain.OutcomeLoss:
		row.Lost++
	}
	row.Points += outcome.Points()
}
