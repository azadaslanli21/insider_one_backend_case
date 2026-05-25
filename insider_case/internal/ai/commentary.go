package ai

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/insider/football-simulator/internal/domain"
)

type Input struct {
	Standings        []domain.Standings
	Predictions      []domain.Prediction
	NextWeekFixtures []domain.Match
	RemainingMatches int
}

type Output struct {
	League   string `json:"league"`
	NextWeek string `json:"next_week"`
	Provider string `json:"provider"`
	Status   string `json:"status"` // ok | degraded | disabled
}

type CommentaryGenerator interface {
	Generate(ctx context.Context, input Input) Output
}

type FallbackGenerator struct {
	status string
}

func NewDisabledGenerator() CommentaryGenerator {
	return &FallbackGenerator{status: "disabled"}
}

func NewDegradedGenerator() CommentaryGenerator {
	return &FallbackGenerator{status: "degraded"}
}

func (g *FallbackGenerator) Generate(_ context.Context, input Input) Output {
	leagueLine := "No standings available yet."
	if len(input.Standings) > 0 {
		leader := input.Standings[0]
		chaserName := "no immediate chaser"
		pointsGap := 0
		if len(input.Standings) > 1 {
			chaser := input.Standings[1]
			chaserName = chaser.TeamName
			pointsGap = leader.Points - chaser.Points
		}
		leaderProb, chaserProb := 0.0, 0.0
		for _, p := range input.Predictions {
			if p.TeamID == leader.TeamID {
				leaderProb = p.WinProbability
			}
			if len(input.Standings) > 1 && p.TeamID == input.Standings[1].TeamID {
				chaserProb = p.WinProbability
			}
		}

		leagueLine = fmt.Sprintf(
			"%s are leading by %d points over %s (GD %d), with %.1f%% title odds versus %.1f%% for the nearest challenger. With %d matches left, %s can still close the gap if they convert head-to-head weeks.",
			leader.TeamName, pointsGap, chaserName, leader.GoalDiff, leaderProb, chaserProb, input.RemainingMatches, chaserName,
		)
	}

	nextWeekLine := "No upcoming fixtures."
	if len(input.NextWeekFixtures) > 0 {
		insights := make([]string, 0, len(input.NextWeekFixtures))
		for _, m := range input.NextWeekFixtures {
			home, away := m.HomeTeam, m.AwayTeam
			diff := home.SBase - away.SBase
			switch {
			case diff >= 8:
				insights = append(insights, fmt.Sprintf("%s vs %s looks home-favoured; %s are live for 2+ goals", home.ShortName, away.ShortName, home.ShortName))
			case diff <= -8:
				insights = append(insights, fmt.Sprintf("%s vs %s still leans away side; %s should avoid defeat and can edge it", home.ShortName, away.ShortName, away.ShortName))
			default:
				insights = append(insights, fmt.Sprintf("%s vs %s is tight and draw-leaning unless an early goal breaks shape", home.ShortName, away.ShortName))
			}
		}
		sort.Strings(insights)
		nextWeekLine = strings.Join(insights, " ")
	}

	return Output{
		League:   leagueLine,
		NextWeek: nextWeekLine,
		Provider: "fallback",
		Status:   g.status,
	}
}
