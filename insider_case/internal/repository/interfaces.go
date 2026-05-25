package repository

import (
	"context"

	"github.com/insider/football-simulator/internal/domain"
)

// TeamRepository handles all persistence operations for teams.
type TeamRepository interface {
	// CreateBatch inserts multiple teams in a single transaction.
	CreateBatch(ctx context.Context, teams []domain.Team) error

	// GetAll returns every team in the league.
	GetAll(ctx context.Context) ([]domain.Team, error)

	// GetByID returns a single team or domain.ErrMatchNotFound if absent.
	GetByID(ctx context.Context, id uint) (domain.Team, error)

	// DeleteAll removes all team rows — used by POST /league/reset.
	DeleteAll(ctx context.Context) error
}

// MatchRepository handles all persistence operations for fixtures and results.
type MatchRepository interface {
	// CreateBatch inserts a full fixture list in a single transaction.
	CreateBatch(ctx context.Context, matches []domain.Match) error

	// GetAll returns every match (played and unplayed) with teams populated.
	GetAll(ctx context.Context) ([]domain.Match, error)

	// GetByWeek returns all matches for a given week with teams populated.
	GetByWeek(ctx context.Context, weekID int) ([]domain.Match, error)

	// GetAllPlayed returns only matches where Played = true.
	GetAllPlayed(ctx context.Context) ([]domain.Match, error)

	// GetUnplayed returns only matches where Played = false.
	GetUnplayed(ctx context.Context) ([]domain.Match, error)

	// GetNextUnplayedWeekID returns the lowest weekID that still has
	// unplayed matches. Returns domain.ErrNoMatchesRemaining if the season
	// is complete.
	GetNextUnplayedWeekID(ctx context.Context) (int, error)

	// GetByID returns a single match with teams populated, or
	// domain.ErrMatchNotFound if absent.
	GetByID(ctx context.Context, id uint) (domain.Match, error)

	// UpdateResult marks a match as played and records its score and reasoning.
	// Used by both the simulator (new result) and PATCH /league/match/:id (edit).
	UpdateResult(ctx context.Context, matchID uint, homeGoals, awayGoals int, reasoning string) error

	// CountAll returns the total number of fixture rows — used to detect
	// whether the league has been initialized.
	CountAll(ctx context.Context) (int64, error)

	// GetRecentByTeam returns the last `limit` played matches for a team
	// (home or away), ordered newest first.
	// Used by the simulator to compute momentum M in Phase 6.
	GetRecentByTeam(ctx context.Context, teamID uint, limit int) ([]domain.Match, error)

	// DeleteAll removes all match rows — used by POST /league/reset.
	DeleteAll(ctx context.Context) error
}

// PredictionRepository handles persistence of Monte Carlo snapshots.
type PredictionRepository interface {
	// SaveSnapshot persists the four team predictions produced after
	// simulating a given week.
	SaveSnapshot(ctx context.Context, weekSimulated int, predictions []domain.Prediction) error

	// GetLatest returns the most recently saved set of prediction snapshots
	// (i.e. those with the highest WeekSimulated value).
	GetLatest(ctx context.Context) ([]domain.PredictionSnapshot, error)

	// DeleteAll removes all snapshot rows — used by POST /league/reset.
	DeleteAll(ctx context.Context) error
}

// StandingsRepository handles persistence of weekly table snapshots.
type StandingsRepository interface {
	// SaveSnapshot persists one standings row per team for a simulated week.
	SaveSnapshot(ctx context.Context, weekID int, standings []domain.Standings) error

	// GetLatest returns standings rows for the latest simulated week.
	GetLatest(ctx context.Context) ([]domain.Standings, error)

	// DeleteAll removes all standings snapshots.
	DeleteAll(ctx context.Context) error
}
