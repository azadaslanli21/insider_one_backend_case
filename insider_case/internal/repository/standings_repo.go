package repository

import (
	"context"

	"github.com/insider/football-simulator/internal/domain"
	"gorm.io/gorm"
)

type standingsRepo struct {
	db *gorm.DB
}

// NewStandingsRepository returns a PostgreSQL-backed StandingsRepository.
func NewStandingsRepository(db *gorm.DB) StandingsRepository {
	return &standingsRepo{db: db}
}

func (r *standingsRepo) SaveSnapshot(ctx context.Context, weekID int, standings []domain.Standings) error {
	models := make([]StandingsSnapshotModel, len(standings))
	for i, s := range standings {
		models[i] = toStandingsSnapshotModel(weekID, s)
	}
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return domain.NewDBError("standings.SaveSnapshot", err)
	}
	return nil
}

func (r *standingsRepo) GetLatest(ctx context.Context) ([]domain.Standings, error) {
	var maxWeek struct{ Week int }
	if err := r.db.WithContext(ctx).
		Model(&StandingsSnapshotModel{}).
		Select("COALESCE(MAX(week_id), 0) AS week").
		Scan(&maxWeek).Error; err != nil {
		return nil, domain.NewDBError("standings.GetLatest.max", err)
	}
	if maxWeek.Week == 0 {
		return []domain.Standings{}, nil
	}

	var models []StandingsSnapshotModel
	if err := r.db.WithContext(ctx).
		Where("week_id = ?", maxWeek.Week).
		Order("position ASC").
		Find(&models).Error; err != nil {
		return nil, domain.NewDBError("standings.GetLatest", err)
	}

	out := make([]domain.Standings, len(models))
	for i, m := range models {
		out[i] = toStandingsDomain(m)
	}
	return out, nil
}

func (r *standingsRepo) DeleteAll(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("1 = 1").Delete(&StandingsSnapshotModel{}).Error; err != nil {
		return domain.NewDBError("standings.DeleteAll", err)
	}
	return nil
}
