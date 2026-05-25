package repository

import (
	"context"
	"errors"

	"github.com/insider/football-simulator/internal/domain"
	"gorm.io/gorm"
)

type matchRepo struct {
	db *gorm.DB
}

// NewMatchRepository returns a PostgreSQL-backed MatchRepository.
func NewMatchRepository(db *gorm.DB) MatchRepository {
	return &matchRepo{db: db}
}

func (r *matchRepo) CreateBatch(ctx context.Context, matches []domain.Match) error {
	models := make([]MatchModel, len(matches))
	for i, m := range matches {
		models[i] = toMatchModel(m)
	}
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return domain.NewDBError("match.CreateBatch", err)
	}
	return nil
}

// hydrate loads the two teams for every MatchModel in a single IN query
// to avoid N+1 problems.
func (r *matchRepo) hydrate(ctx context.Context, models []MatchModel) ([]domain.Match, error) {
	if len(models) == 0 {
		return []domain.Match{}, nil
	}

	// Collect unique team IDs
	idSet := make(map[uint]struct{})
	for _, m := range models {
		idSet[m.HomeTeamID] = struct{}{}
		idSet[m.AwayTeamID] = struct{}{}
	}
	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	var teamModels []TeamModel
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&teamModels).Error; err != nil {
		return nil, domain.NewDBError("match.hydrate.teams", err)
	}

	teamMap := make(map[uint]domain.Team, len(teamModels))
	for _, t := range teamModels {
		teamMap[t.ID] = toTeamDomain(t)
	}

	result := make([]domain.Match, len(models))
	for i, m := range models {
		result[i] = toMatchDomain(m, teamMap[m.HomeTeamID], teamMap[m.AwayTeamID])
	}
	return result, nil
}

func (r *matchRepo) GetAll(ctx context.Context) ([]domain.Match, error) {
	var models []MatchModel
	if err := r.db.WithContext(ctx).Order("week_id, id").Find(&models).Error; err != nil {
		return nil, domain.NewDBError("match.GetAll", err)
	}
	return r.hydrate(ctx, models)
}

func (r *matchRepo) GetByWeek(ctx context.Context, weekID int) ([]domain.Match, error) {
	var models []MatchModel
	if err := r.db.WithContext(ctx).Where("week_id = ?", weekID).Find(&models).Error; err != nil {
		return nil, domain.NewDBError("match.GetByWeek", err)
	}
	return r.hydrate(ctx, models)
}

func (r *matchRepo) GetAllPlayed(ctx context.Context) ([]domain.Match, error) {
	var models []MatchModel
	if err := r.db.WithContext(ctx).Where("played = true").Order("week_id, id").Find(&models).Error; err != nil {
		return nil, domain.NewDBError("match.GetAllPlayed", err)
	}
	return r.hydrate(ctx, models)
}

func (r *matchRepo) GetUnplayed(ctx context.Context) ([]domain.Match, error) {
	var models []MatchModel
	if err := r.db.WithContext(ctx).Where("played = false").Order("week_id, id").Find(&models).Error; err != nil {
		return nil, domain.NewDBError("match.GetUnplayed", err)
	}
	return r.hydrate(ctx, models)
}

func (r *matchRepo) GetNextUnplayedWeekID(ctx context.Context) (int, error) {
	var model MatchModel
	err := r.db.WithContext(ctx).
		Where("played = false").
		Order("week_id ASC").
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, domain.ErrNoMatchesRemaining
	}
	if err != nil {
		return 0, domain.NewDBError("match.GetNextUnplayedWeekID", err)
	}
	return model.WeekID, nil
}

func (r *matchRepo) GetByID(ctx context.Context, id uint) (domain.Match, error) {
	var model MatchModel
	err := r.db.WithContext(ctx).First(&model, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Match{}, domain.ErrMatchNotFound
	}
	if err != nil {
		return domain.Match{}, domain.NewDBError("match.GetByID", err)
	}

	matches, err := r.hydrate(ctx, []MatchModel{model})
	if err != nil {
		return domain.Match{}, err
	}
	return matches[0], nil
}

func (r *matchRepo) UpdateResult(ctx context.Context, matchID uint, homeGoals, awayGoals int, reasoning string) error {
	result := r.db.WithContext(ctx).
		Model(&MatchModel{}).
		Where("id = ?", matchID).
		Updates(map[string]any{
			"home_goals": homeGoals,
			"away_goals": awayGoals,
			"played":     true,
			"reasoning":  reasoning,
		})

	if result.Error != nil {
		return domain.NewDBError("match.UpdateResult", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrMatchNotFound
	}
	return nil
}

func (r *matchRepo) CountAll(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&MatchModel{}).Count(&count).Error; err != nil {
		return 0, domain.NewDBError("match.CountAll", err)
	}
	return count, nil
}

func (r *matchRepo) DeleteAll(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("1 = 1").Delete(&MatchModel{}).Error; err != nil {
		return domain.NewDBError("match.DeleteAll", err)
	}
	return nil
}

// GetRecentByTeam returns the last `limit` played matches for the given team
// (regardless of home or away side), ordered by ID descending (most recent first).
// The simulator calls this with limit=3 to compute momentum M.
func (r *matchRepo) GetRecentByTeam(ctx context.Context, teamID uint, limit int) ([]domain.Match, error) {
	var models []MatchModel
	err := r.db.WithContext(ctx).
		Where("(home_team_id = ? OR away_team_id = ?) AND played = true", teamID, teamID).
		Order("id DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, domain.NewDBError("match.GetRecentByTeam", err)
	}
	return r.hydrate(ctx, models)
}
