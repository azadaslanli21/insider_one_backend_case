package predictor

import (
	"testing"

	"github.com/insider/football-simulator/internal/domain"
)

func TestRunMonteCarloProbabilities(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "A", ShortName: "A"},
		{ID: 2, Name: "B", ShortName: "B"},
		{ID: 3, Name: "C", ShortName: "C"},
		{ID: 4, Name: "D", ShortName: "D"},
	}
	remaining := []domain.Match{
		{HomeTeamID: 1, AwayTeamID: 2},
		{HomeTeamID: 3, AwayTeamID: 4},
	}
	out := RunMonteCarlo(MonteCarloInput{
		Teams:            teams,
		PlayedMatches:    nil,
		RemainingMatches: remaining,
		TeamSFinals: map[uint]float64{
			1: 70, 2: 70, 3: 70, 4: 70,
		},
	}, domain.MonteCarloConfig{Simulations: 1000})

	if len(out) != 4 {
		t.Fatalf("expected 4 predictions, got %d", len(out))
	}
	var sum float64
	for _, p := range out {
		sum += p.WinProbability
	}
	if sum < 99 || sum > 101 {
		t.Fatalf("expected sum around 100, got %.2f", sum)
	}
}
