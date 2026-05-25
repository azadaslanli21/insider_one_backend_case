package simulator

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/insider/football-simulator/internal/domain"
)

func ip(v int) *int { return &v }

func TestComputeMomentumWindows(t *testing.T) {
	teamID := uint(1)
	if m := ComputeMomentum(teamID, nil); m != 0 {
		t.Fatalf("expected 0, got %f", m)
	}

	one := []domain.Match{{HomeTeamID: teamID, AwayTeamID: 2, HomeGoals: ip(1), AwayGoals: ip(0), Played: true}}
	if m := ComputeMomentum(teamID, one); m != 1 {
		t.Fatalf("expected 1, got %f", m)
	}

	two := []domain.Match{
		{HomeTeamID: teamID, AwayTeamID: 2, HomeGoals: ip(1), AwayGoals: ip(0), Played: true},
		{HomeTeamID: 2, AwayTeamID: teamID, HomeGoals: ip(1), AwayGoals: ip(1), Played: true},
	}
	if m := ComputeMomentum(teamID, two); m < 0.66 || m > 0.67 {
		t.Fatalf("expected ~0.666, got %f", m)
	}
}

func TestSimulateMatchOutput(t *testing.T) {
	out := SimulateMatch(SimulateMatchInput{
		HomeTeam: domain.Team{ID: 1, Name: "A", ShortName: "A", SBase: 80},
		AwayTeam: domain.Team{ID: 2, Name: "B", ShortName: "B", SBase: 75},
	})
	if out.HomeGoals < 0 || out.AwayGoals < 0 {
		t.Fatal("goals cannot be negative")
	}
	if strings.TrimSpace(out.Reasoning) == "" {
		t.Fatal("reasoning must not be empty")
	}
}

func TestPoissonSampleNonNegative(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 100; i++ {
		if v := PoissonSample(r, 1.4); v < 0 {
			t.Fatalf("poisson sample negative: %d", v)
		}
	}
}
