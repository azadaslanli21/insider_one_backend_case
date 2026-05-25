package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/insider/football-simulator/internal/domain"
)

func TestDisabledGenerator(t *testing.T) {
	g := NewDisabledGenerator()
	out := g.Generate(context.Background(), Input{})
	if out.Status != "disabled" {
		t.Fatalf("expected disabled status, got %s", out.Status)
	}
	if out.Provider != "fallback" {
		t.Fatalf("expected fallback provider, got %s", out.Provider)
	}
}

func TestFallbackIncludesLeaderAndFixtures(t *testing.T) {
	g := NewDegradedGenerator()
	out := g.Generate(context.Background(), Input{
		Standings: []domain.Standings{
			{TeamName: "Arsenal", TeamShort: "ARS", Points: 12, GoalDiff: 7},
		},
		NextWeekFixtures: []domain.Match{
			{
				HomeTeam: domain.Team{ShortName: "ARS"},
				AwayTeam: domain.Team{ShortName: "LIV"},
			},
		},
		RemainingMatches: 4,
	})
	if out.Status != "degraded" {
		t.Fatalf("expected degraded status, got %s", out.Status)
	}
	if !strings.Contains(out.League, "Arsenal") {
		t.Fatalf("league commentary missing leader: %s", out.League)
	}
	if !strings.Contains(out.NextWeek, "ARS vs LIV") {
		t.Fatalf("next week commentary missing fixture: %s", out.NextWeek)
	}
}
