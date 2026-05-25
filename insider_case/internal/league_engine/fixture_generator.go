package league

import (
	"math/rand"
	"time"

	"github.com/insider/football-simulator/internal/domain"
)

// GenerateFixtures produces a full double round-robin fixture list for the
// given teams and returns them as domain.Match values ready to be persisted.
//
// Algorithm:
//  1. Build every ordered pair (home, away) — 4 teams → 12 pairs.
//  2. Group pairs into 6 weeks of 2 matches each using the Berger
//     round-robin table so every team plays exactly once per week.
//  3. Shuffle the week order so the season feels random each reset.
func GenerateFixtures(teams []domain.Team) []domain.Match {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Step 1 — build the 6-week Berger schedule for 4 teams.
	// Indices into the teams slice; -1 represents the "bye" pivot.
	// For 4 teams the canonical Berger weeks are:
	//   Week A: (0 v 3), (1 v 2)
	//   Week B: (0 v 2), (3 v 1)
	//   Week C: (0 v 1), (2 v 3)
	// Then repeat with home/away swapped for the return legs (weeks D-F).
	type pair struct{ home, away int }

	firstLeg := [][]pair{
		{{0, 3}, {1, 2}},
		{{0, 2}, {3, 1}},
		{{0, 1}, {2, 3}},
	}

	// Build return legs by swapping home/away in every first-leg week.
	allWeeks := make([][]pair, 0, 6)
	allWeeks = append(allWeeks, firstLeg...)
	for _, week := range firstLeg {
		var returnWeek []pair
		for _, p := range week {
			returnWeek = append(returnWeek, pair{p.away, p.home})
		}
		allWeeks = append(allWeeks, returnWeek)
	}

	// Step 2 — shuffle week order so the calendar differs each season.
	rng.Shuffle(len(allWeeks), func(i, j int) {
		allWeeks[i], allWeeks[j] = allWeeks[j], allWeeks[i]
	})

	// Step 3 — flatten into domain.Match values with 1-based WeekID.
	matches := make([]domain.Match, 0, domain.TotalWeeks*2)
	for weekIdx, week := range allWeeks {
		weekID := weekIdx + 1
		for _, p := range week {
			matches = append(matches, domain.Match{
				WeekID:     weekID,
				HomeTeamID: teams[p.home].ID,
				HomeTeam:   teams[p.home],
				AwayTeamID: teams[p.away].ID,
				AwayTeam:   teams[p.away],
				Played:     false,
			})
		}
	}

	return matches
}
