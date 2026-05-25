package simulator

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/insider/football-simulator/internal/domain"
)

// Goal-rate constants derived from Premier League averages.
const (
	baseGoalsPerMatch = 2.7  // PL average ~2.7 goals per match across a season
	homeAdvantage     = 1.10 // home teams score ~10% more on average
	awayDisadvantage  = 0.90
	minXG             = 0.20 // floor so Poisson lambda is never 0
)

// ─── Input / Output ──────────────────────────────────────────────────────────

// SimulateMatchInput carries everything the engine needs to run one fixture.
// RecentHome and RecentAway are fetched by the service via
// MatchRepository.GetRecentByTeam(teamID, 3) before calling SimulateMatch.
type SimulateMatchInput struct {
	HomeTeam   domain.Team
	AwayTeam   domain.Team
	RecentHome []domain.Match // last ≤3 played matches for the home side
	RecentAway []domain.Match // last ≤3 played matches for the away side
}

// SimulateMatchOutput is the result of one match simulation.
type SimulateMatchOutput struct {
	HomeGoals int
	AwayGoals int
	Reasoning string // human-readable audit log for this result
}

// ─── Main entry point ────────────────────────────────────────────────────────

// SimulateMatch runs a single fixture through the full engine:
//  1. Compute momentum M for each side from recent results.
//  2. Derive S_final via the weighted formula.
//  3. Convert S_final values to per-team expected goals (xG).
//  4. Sample actual goals from a Poisson distribution.
//  5. Build a human-readable Reasoning string.
//
// A fresh RNG is seeded from the wall clock so consecutive calls in the
// same second still produce different results.
func SimulateMatch(input SimulateMatchInput) SimulateMatchOutput {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	homeM := ComputeMomentum(input.HomeTeam.ID, input.RecentHome)
	awayM := ComputeMomentum(input.AwayTeam.ID, input.RecentAway)

	homeSFinal := ComputeSFinal(input.HomeTeam.SBase, homeM)
	awaySFinal := ComputeSFinal(input.AwayTeam.SBase, awayM)

	homeXG := ComputeXG(homeSFinal, awaySFinal, true)
	awayXG := ComputeXG(awaySFinal, homeSFinal, false)

	hg := PoissonSample(rng, homeXG)
	ag := PoissonSample(rng, awayXG)

	reasoning := buildReasoning(
		input.HomeTeam, input.AwayTeam,
		homeSFinal, awaySFinal,
		homeM, awayM,
		homeXG, awayXG,
		hg, ag,
	)

	return SimulateMatchOutput{HomeGoals: hg, AwayGoals: ag, Reasoning: reasoning}
}

// ─── Exported helpers (reused by the Monte Carlo predictor) ──────────────────

// ComputeMomentum returns M ∈ [0, 1] representing recent form.
//
// Window rules (avoids dividing by 9 when fewer than 3 matches exist):
//
//	Week 1  (N=0): M = 0.0              — no prior data
//	Week 2  (N=1): points / 3           — one match window
//	Week 3  (N=2): points / 6           — two match window
//	Week 4+ (N=3): points / 9           — full three match window
//
// The caller is responsible for fetching the right window via
// MatchRepository.GetRecentByTeam(teamID, 3).
func ComputeMomentum(teamID uint, recentMatches []domain.Match) float64 {
	n := len(recentMatches)
	if n == 0 {
		return 0.0
	}

	totalPoints := 0
	counted := 0
	for _, m := range recentMatches {
		if !m.Played || m.HomeGoals == nil || m.AwayGoals == nil {
			continue
		}
		homeOutcome, awayOutcome := domain.DetermineOutcome(*m.HomeGoals, *m.AwayGoals)
		if m.HomeTeamID == teamID {
			totalPoints += homeOutcome.Points()
		} else {
			totalPoints += awayOutcome.Points()
		}
		counted++
	}

	if counted == 0 {
		return 0.0
	}
	// Normalise against the actual number of counted matches × 3 (max points)
	return float64(totalPoints) / float64(counted*domain.PointsWin)
}

// ComputeSFinal applies the weighted strength formula:
//
//	S_final = (S_base × 0.8) + (M × 20)
//
// With S_base in [0,100] and M in [0,1]:
//   - No momentum:   S_final = S_base × 0.8
//   - Full momentum: S_final = S_base × 0.8 + 20
func ComputeSFinal(sBase, momentum float64) float64 {
	return (sBase * 0.8) + (momentum * 20.0)
}

// ComputeXG converts a team's S_final into expected goals for one match.
// Exported so the Monte Carlo engine uses the identical formula.
//
// Formula:
//
//	share   = sFinal / (sFinal + oppSFinal)
//	xG      = baseGoalsPerMatch × share × homeAdv (or awayDisadv)
//	xG      = max(xG, minXG)
func ComputeXG(sFinal, oppSFinal float64, isHome bool) float64 {
	total := sFinal + oppSFinal
	if total == 0 {
		return minXG
	}
	xg := baseGoalsPerMatch * (sFinal / total)
	if isHome {
		xg *= homeAdvantage
	} else {
		xg *= awayDisadvantage
	}
	return math.Max(xg, minXG)
}

// PoissonSample draws one random integer from a Poisson(lambda) distribution.
// Uses Knuth's exact algorithm; efficient and unbiased for lambda < 20,
// which covers all realistic football xG values (0.2 – 3.5).
// Exported so the Monte Carlo predictor reuses the same sampling method.
func PoissonSample(rng *rand.Rand, lambda float64) int {
	if lambda <= 0 {
		return 0
	}
	L := math.Exp(-lambda)
	k := 0
	p := 1.0
	for {
		k++
		p *= rng.Float64()
		if p <= L {
			break
		}
	}
	return k - 1
}

// ─── Internal ────────────────────────────────────────────────────────────────

// buildReasoning constructs the audit string stored on every match row.
//
// Example output:
//
//	Arsenal beat Man United 3-1
//	[S_final: 74.4 vs 54.4 | xG: 1.87 vs 1.19 | Momentum — ARS: +6.7  MUN: +0.0]
func buildReasoning(
	home, away domain.Team,
	homeSFinal, awaySFinal,
	homeM, awayM,
	homeXG, awayXG float64,
	hg, ag int,
) string {
	homeOutcome, _ := domain.DetermineOutcome(hg, ag)

	var verb string
	switch homeOutcome {
	case domain.OutcomeWin:
		verb = fmt.Sprintf("%s beat %s", home.Name, away.Name)
	case domain.OutcomeLoss:
		verb = fmt.Sprintf("%s beat %s", away.Name, home.Name)
	default:
		verb = fmt.Sprintf("%s drew with %s", home.Name, away.Name)
	}

	return fmt.Sprintf(
		"%s %d-%d [S_final: %.1f vs %.1f | xG: %.2f vs %.2f | Momentum — %s: +%.1f  %s: +%.1f]",
		verb, hg, ag,
		homeSFinal, awaySFinal,
		homeXG, awayXG,
		home.ShortName, homeM*20.0,
		away.ShortName, awayM*20.0,
	)
}
