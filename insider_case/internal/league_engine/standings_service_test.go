package league

import (
	"testing"

	"github.com/insider/football-simulator/internal/domain"
)

func intPtr(v int) *int { return &v }

func TestBuildStandingsSortsByPointsThenGDThenGF(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "A", ShortName: "A"},
		{ID: 2, Name: "B", ShortName: "B"},
		{ID: 3, Name: "C", ShortName: "C"},
	}

	matches := []domain.Match{
		{HomeTeamID: 1, AwayTeamID: 2, HomeGoals: intPtr(2), AwayGoals: intPtr(0), Played: true},
		{HomeTeamID: 3, AwayTeamID: 2, HomeGoals: intPtr(2), AwayGoals: intPtr(0), Played: true},
		{HomeTeamID: 1, AwayTeamID: 3, HomeGoals: intPtr(1), AwayGoals: intPtr(1), Played: true},
	}

	table := BuildStandings(teams, matches)
	if len(table) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(table))
	}
	if table[0].TeamID != 1 {
		t.Fatalf("expected team 1 first, got %d", table[0].TeamID)
	}
	if table[1].TeamID != 3 {
		t.Fatalf("expected team 3 second, got %d", table[1].TeamID)
	}
}
