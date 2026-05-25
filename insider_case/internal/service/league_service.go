package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/insider/football-simulator/internal/ai"
	"github.com/insider/football-simulator/internal/domain"
	"github.com/insider/football-simulator/internal/fetcher"
	league "github.com/insider/football-simulator/internal/league_engine"
	"github.com/insider/football-simulator/internal/predictor"
	"github.com/insider/football-simulator/internal/repository"
	"github.com/insider/football-simulator/internal/simulator"
	"github.com/olekukonko/tablewriter"
	"gorm.io/gorm"
)

type LeagueService struct {
	db             *gorm.DB
	teamRepo       repository.TeamRepository
	matchRepo      repository.MatchRepository
	predictionRepo repository.PredictionRepository
	standingsRepo  repository.StandingsRepository
	strengths      fetcher.StrengthProvider
	commentary     ai.CommentaryGenerator
}

func NewLeagueService(
	db *gorm.DB,
	teamRepo repository.TeamRepository,
	matchRepo repository.MatchRepository,
	predictionRepo repository.PredictionRepository,
	standingsRepo repository.StandingsRepository,
	strengths fetcher.StrengthProvider,
	commentary ai.CommentaryGenerator,
) *LeagueService {
	if commentary == nil {
		commentary = ai.NewDisabledGenerator()
	}
	return &LeagueService{
		db:             db,
		teamRepo:       teamRepo,
		matchRepo:      matchRepo,
		predictionRepo: predictionRepo,
		standingsRepo:  standingsRepo,
		strengths:      strengths,
		commentary:     commentary,
	}
}

type AICommentary = ai.Output

type SimulateWeekResponse struct {
	WeekID       int                 `json:"week_id"`
	Results      []domain.Match      `json:"results"`
	Standings    []domain.Standings  `json:"standings"`
	Predictions  []domain.Prediction `json:"predictions"`
	AICommentary AICommentary        `json:"ai_commentary"`
}

type SimulateAllResponse struct {
	Weeks            []SimulateWeekResponse `json:"weeks"`
	FinalStandings   []domain.Standings     `json:"final_standings"`
	FinalPredictions []domain.Prediction    `json:"final_predictions"`
	FinalAISummary   AICommentary           `json:"final_ai_summary"`
}

type FixturesResponse struct {
	Mode     string         `json:"mode"`
	WeekID   *int           `json:"week_id,omitempty"`
	Fixtures []domain.Match `json:"fixtures"`
}

func (s *LeagueService) InitLeague(ctx context.Context) error {
	count, err := s.matchRepo.CountAll(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	strengthMap, err := s.strengths.FetchStrengths()
	if err != nil {
		return err
	}

	teams := make([]domain.Team, 0, len(domain.KnownTeams))
	for _, t := range domain.KnownTeams {
		teams = append(teams, domain.Team{Name: t.Name, ShortName: t.ShortName})
	}
	teams = fetcher.ApplyToTeams(teams, strengthMap)
	if err := s.teamRepo.CreateBatch(ctx, teams); err != nil {
		return err
	}

	persistedTeams, err := s.teamRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	sort.Slice(persistedTeams, func(i, j int) bool { return persistedTeams[i].ID < persistedTeams[j].ID })

	fixtures := league.GenerateFixtures(persistedTeams)
	return s.matchRepo.CreateBatch(ctx, fixtures)
}

func (s *LeagueService) SimulateNextWeek(ctx context.Context) (SimulateWeekResponse, error) {
	var out SimulateWeekResponse
	matchCount, err := s.matchRepo.CountAll(ctx)
	if err != nil {
		return out, err
	}
	if matchCount == 0 {
		return out, domain.ErrLeagueNotInitialized
	}

	weekID, err := s.matchRepo.GetNextUnplayedWeekID(ctx)
	if err != nil {
		return out, err
	}

	weekMatches, err := s.matchRepo.GetByWeek(ctx, weekID)
	if err != nil {
		return out, err
	}

	for _, m := range weekMatches {
		if m.Played {
			continue
		}
		recentHome, err := s.matchRepo.GetRecentByTeam(ctx, m.HomeTeamID, 3)
		if err != nil {
			return out, err
		}
		recentAway, err := s.matchRepo.GetRecentByTeam(ctx, m.AwayTeamID, 3)
		if err != nil {
			return out, err
		}

		simOut := simulator.SimulateMatch(simulator.SimulateMatchInput{
			HomeTeam:   m.HomeTeam,
			AwayTeam:   m.AwayTeam,
			RecentHome: recentHome,
			RecentAway: recentAway,
		})
		if err := s.matchRepo.UpdateResult(ctx, m.ID, simOut.HomeGoals, simOut.AwayGoals, simOut.Reasoning); err != nil {
			return out, err
		}
	}

	played, err := s.matchRepo.GetAllPlayed(ctx)
	if err != nil {
		return out, err
	}
	teams, err := s.teamRepo.GetAll(ctx)
	if err != nil {
		return out, err
	}
	standings := league.BuildStandings(teams, played)
	if err := s.standingsRepo.SaveSnapshot(ctx, weekID, standings); err != nil {
		return out, err
	}

	remaining, err := s.matchRepo.GetUnplayed(ctx)
	if err != nil {
		return out, err
	}
	teamSFinals := make(map[uint]float64, len(teams))
	for _, t := range teams {
		recent, err := s.matchRepo.GetRecentByTeam(ctx, t.ID, 3)
		if err != nil {
			return out, err
		}
		m := simulator.ComputeMomentum(t.ID, recent)
		teamSFinals[t.ID] = simulator.ComputeSFinal(t.SBase, m)
	}

	preds := predictor.RunMonteCarlo(predictor.MonteCarloInput{
		Teams:            teams,
		PlayedMatches:    played,
		RemainingMatches: remaining,
		TeamSFinals:      teamSFinals,
	}, domain.DefaultMonteCarloConfig())
	if err := s.predictionRepo.SaveSnapshot(ctx, weekID, preds); err != nil {
		return out, err
	}

	currentWeekResults, err := s.matchRepo.GetByWeek(ctx, weekID)
	if err != nil {
		return out, err
	}
	nextWeekMatches, _ := s.getNextWeekFixtures(ctx)
	aiOut := s.commentary.Generate(ctx, ai.Input{
		Standings:        standings,
		Predictions:      preds,
		NextWeekFixtures: nextWeekMatches,
		RemainingMatches: len(remaining),
	})

	s.logStandingsTable(weekID, standings)
	out = SimulateWeekResponse{
		WeekID:       weekID,
		Results:      currentWeekResults,
		Standings:    standings,
		Predictions:  preds,
		AICommentary: aiOut,
	}
	return out, nil
}

func (s *LeagueService) SimulateAll(ctx context.Context) (SimulateAllResponse, error) {
	var out SimulateAllResponse
	matchCount, err := s.matchRepo.CountAll(ctx)
	if err != nil {
		return out, err
	}
	if matchCount == 0 {
		return out, domain.ErrLeagueNotInitialized
	}

	for {
		week, err := s.SimulateNextWeek(ctx)
		if err != nil {
			if errors.Is(err, domain.ErrNoMatchesRemaining) {
				if len(out.Weeks) == 0 {
					return out, domain.ErrSeasonAlreadyComplete
				}
				break
			}
			return out, err
		}
		out.Weeks = append(out.Weeks, week)
	}

	table, err := s.GetTable(ctx)
	if err != nil {
		return out, err
	}
	out.FinalStandings = table

	latestPreds, err := s.predictionRepo.GetLatest(ctx)
	if err != nil {
		return out, err
	}
	out.FinalPredictions = make([]domain.Prediction, 0, len(latestPreds))
	for _, p := range latestPreds {
		out.FinalPredictions = append(out.FinalPredictions, domain.Prediction{
			TeamID:         p.TeamID,
			TeamName:       p.TeamName,
			TeamShort:      "",
			WinProbability: p.WinProbability,
		})
	}
	nextWeekMatches, _ := s.getNextWeekFixtures(ctx)
	unplayed, _ := s.matchRepo.GetUnplayed(ctx)
	out.FinalAISummary = s.commentary.Generate(ctx, ai.Input{
		Standings:        out.FinalStandings,
		Predictions:      out.FinalPredictions,
		NextWeekFixtures: nextWeekMatches,
		RemainingMatches: len(unplayed),
	})
	return out, nil
}

func (s *LeagueService) GetFixtures(ctx context.Context, mode string) (FixturesResponse, error) {
	var out FixturesResponse
	matchCount, err := s.matchRepo.CountAll(ctx)
	if err != nil {
		return out, err
	}
	if matchCount == 0 {
		return out, domain.ErrLeagueNotInitialized
	}

	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "" {
		mode = "all"
	}
	switch mode {
	case "all":
		matches, err := s.matchRepo.GetAll(ctx)
		if err != nil {
			return out, err
		}
		return FixturesResponse{
			Mode:     "all",
			Fixtures: matches,
		}, nil
	case "next":
		matches, weekID := s.getNextWeekFixtures(ctx)
		return FixturesResponse{
			Mode:     "next",
			WeekID:   weekID,
			Fixtures: matches,
		}, nil
	default:
		return out, domain.ErrInvalidMode
	}
}

func (s *LeagueService) GetTable(ctx context.Context) ([]domain.Standings, error) {
	table, err := s.standingsRepo.GetLatest(ctx)
	if err != nil {
		return nil, err
	}
	if len(table) > 0 {
		return table, nil
	}

	teams, err := s.teamRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	played, err := s.matchRepo.GetAllPlayed(ctx)
	if err != nil {
		return nil, err
	}
	return league.BuildStandings(teams, played), nil
}

func (s *LeagueService) ResetLeague(ctx context.Context) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&repository.MatchModel{}).Error; err != nil {
			return domain.NewDBError("reset.delete.matches", err)
		}
		if err := tx.Where("1 = 1").Delete(&repository.StandingsSnapshotModel{}).Error; err != nil {
			return domain.NewDBError("reset.delete.standings", err)
		}
		if err := tx.Where("1 = 1").Delete(&repository.PredictionSnapshotModel{}).Error; err != nil {
			return domain.NewDBError("reset.delete.predictions", err)
		}
		if err := tx.Where("1 = 1").Delete(&repository.TeamModel{}).Error; err != nil {
			return domain.NewDBError("reset.delete.teams", err)
		}
		return nil
	})
}

func (s *LeagueService) OverrideMatchResult(ctx context.Context, matchID uint, homeGoals, awayGoals int, reasoning string) ([]domain.Standings, error) {
	if homeGoals < 0 || awayGoals < 0 {
		return nil, domain.ErrInvalidGoalCount
	}
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	overrideReason := strings.TrimSpace(reasoning)
	if overrideReason == "" {
		overrideReason = "manual override"
	}
	overrideReason = fmt.Sprintf("%s | override %d-%d", overrideReason, homeGoals, awayGoals)

	if err := s.matchRepo.UpdateResult(ctx, match.ID, homeGoals, awayGoals, overrideReason); err != nil {
		return nil, err
	}

	teams, err := s.teamRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	played, err := s.matchRepo.GetAllPlayed(ctx)
	if err != nil {
		return nil, err
	}
	standings := league.BuildStandings(teams, played)

	weekID := match.WeekID
	if err := s.standingsRepo.SaveSnapshot(ctx, weekID, standings); err != nil {
		return nil, err
	}
	s.logStandingsTable(weekID, standings)
	return standings, nil
}

func (s *LeagueService) logStandingsTable(weekID int, standings []domain.Standings) {
	var b strings.Builder
	t := tablewriter.NewWriter(&b)
	t.SetHeader([]string{"Pos", "Team", "P", "W", "D", "L", "GF", "GA", "GD", "Pts"})
	for _, row := range standings {
		t.Append([]string{
			fmt.Sprintf("%d", row.Position),
			row.TeamShort,
			fmt.Sprintf("%d", row.Played),
			fmt.Sprintf("%d", row.Won),
			fmt.Sprintf("%d", row.Drawn),
			fmt.Sprintf("%d", row.Lost),
			fmt.Sprintf("%d", row.GoalsFor),
			fmt.Sprintf("%d", row.GoalsAgainst),
			fmt.Sprintf("%d", row.GoalDiff),
			fmt.Sprintf("%d", row.Points),
		})
	}
	t.Render()
	log.Printf("Standings after week %d:\n%s", weekID, b.String())
}

func (s *LeagueService) getNextWeekFixtures(ctx context.Context) ([]domain.Match, *int) {
	weekID, err := s.matchRepo.GetNextUnplayedWeekID(ctx)
	if err != nil {
		return []domain.Match{}, nil
	}
	matches, err := s.matchRepo.GetByWeek(ctx, weekID)
	if err != nil {
		return []domain.Match{}, nil
	}
	return matches, &weekID
}
