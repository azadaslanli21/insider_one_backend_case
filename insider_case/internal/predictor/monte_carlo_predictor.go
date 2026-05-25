package predictor

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/insider/football-simulator/internal/domain"
	league "github.com/insider/football-simulator/internal/league_engine"
	"github.com/insider/football-simulator/internal/simulator"
)

// MonteCarloInput holds all data the predictor needs.
// The service layer (Phase 8) pre-fetches everything from the DB so
// RunMonteCarlo performs zero DB calls — it is a pure in-memory computation.
type MonteCarloInput struct {
	Teams            []domain.Team    // all 4 teams in the league
	PlayedMatches    []domain.Match   // completed fixtures (goals already set)
	RemainingMatches []domain.Match   // unplayed fixtures to simulate
	TeamSFinals      map[uint]float64 // S_final per team ID, frozen before Monte Carlo runs
}

// RunMonteCarlo spins up cfg.Simulations goroutines (default 10,000).
// Each goroutine independently simulates all remaining matches and reports
// the simulated league winner through a buffered channel.
//
// Concurrency pattern used:
//   - sync.WaitGroup to coordinate goroutine lifecycle
//   - Buffered channel (size = Simulations) to collect winner IDs
//   - A separate closer goroutine that closes the channel after WaitGroup is done
//   - Main goroutine drains the closed channel to tally results
//
// This demonstrates both WaitGroup and channel usage as required by the spec.
func RunMonteCarlo(input MonteCarloInput, cfg domain.MonteCarloConfig) []domain.Prediction {
	// Buffered so goroutines never block on send.
	results := make(chan uint, cfg.Simulations)

	var wg sync.WaitGroup
	baseNano := time.Now().UnixNano()

	for i := 0; i < cfg.Simulations; i++ {
		wg.Add(1)
		// Use a prime stride so seeds are well-separated even when called
		// in rapid succession; avoids correlated Rand sequences.
		seed := baseNano + int64(i)*1_000_003
		go func(seed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed))
			results <- simulateOnce(rng, input)
		}(seed)
	}

	// Closer: waits for all workers, then signals the drainer below.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Drainer: range over the channel until it is closed and empty.
	winCounts := make(map[uint]int, len(input.Teams))
	for teamID := range results {
		winCounts[teamID]++
	}

	total := float64(cfg.Simulations)
	predictions := make([]domain.Prediction, 0, len(input.Teams))
	for _, t := range input.Teams {
		predictions = append(predictions, domain.Prediction{
			TeamID:         t.ID,
			TeamName:       t.Name,
			TeamShort:      t.ShortName,
			WinProbability: float64(winCounts[t.ID]) / total * 100.0,
		})
	}

	// Sort descending so the API response always leads with the favourite.
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].WinProbability > predictions[j].WinProbability
	})

	return predictions
}

// simulateOnce runs one complete hypothetical season from the current state.
// It is called inside a goroutine and receives its own *rand.Rand so there
// is zero shared mutable state between goroutines.
//
// Steps:
//  1. Copy already-played matches (immutable source of truth).
//  2. For every remaining fixture, sample home and away goals via Poisson.
//  3. Call BuildStandings on the full 12-match slate.
//  4. Return the TeamID of the team that finishes top.
//
// Note: S_final values are frozen in input.TeamSFinals — momentum is not
// re-computed per simulated result because tracking rolling form across
// 10,000 hypothetical futures would add O(n²) complexity with minimal
// accuracy gain at this season length (6 weeks, 4 teams).
func simulateOnce(rng *rand.Rand, input MonteCarloInput) uint {
	// Pre-allocate with played capacity so the append below never reallocates.
	allMatches := make(
		[]domain.Match,
		len(input.PlayedMatches),
		len(input.PlayedMatches)+len(input.RemainingMatches),
	)
	copy(allMatches, input.PlayedMatches)

	for _, m := range input.RemainingMatches {
		homeS := input.TeamSFinals[m.HomeTeamID]
		awayS := input.TeamSFinals[m.AwayTeamID]

		homeXG := simulator.ComputeXG(homeS, awayS, true)
		awayXG := simulator.ComputeXG(awayS, homeS, false)

		hg := simulator.PoissonSample(rng, homeXG)
		ag := simulator.PoissonSample(rng, awayXG)

		// Local copies so we can take their addresses safely.
		hgCopy, agCopy := hg, ag
		allMatches = append(allMatches, domain.Match{
			HomeTeamID: m.HomeTeamID,
			AwayTeamID: m.AwayTeamID,
			HomeGoals:  &hgCopy,
			AwayGoals:  &agCopy,
			Played:     true,
		})
	}

	standings := league.BuildStandings(input.Teams, allMatches)
	if len(standings) == 0 {
		return 0
	}
	return standings[0].TeamID
}
