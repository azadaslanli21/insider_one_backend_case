package repository

import (
	"context"

	"github.com/insider/football-simulator/internal/domain"
	"gorm.io/gorm"
)

type predictionRepo struct {
	db *gorm.DB
}

// NewPredictionRepository returns a PostgreSQL-backed PredictionRepository.
func NewPredictionRepository(db *gorm.DB) PredictionRepository {
	return &predictionRepo{db: db}
}

// SaveSnapshot persists one prediction row per team for the given week.
// Called by simulate-next after the Monte Carlo engine finishes.
// Always inserts new rows — it never updates, so history is preserved across weeks.
func (r *predictionRepo) SaveSnapshot(ctx context.Context, weekSimulated int, predictions []domain.Prediction) error {
	models := make([]PredictionSnapshotModel, len(predictions))
	cfg := domain.DefaultMonteCarloConfig()

	for i, p := range predictions {
		models[i] = PredictionSnapshotModel{
			WeekSimulated:  weekSimulated,
			TeamID:         p.TeamID,
			TeamName:       p.TeamName,
			WinProbability: p.WinProbability,
			Simulations:    cfg.Simulations,
		}
	}

	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return domain.NewDBError("prediction.SaveSnapshot", err)
	}
	return nil
}

// GetLatest returns the four prediction rows from the most recently simulated week.
// Returns an empty slice (not an error) when no predictions exist yet.
func (r *predictionRepo) GetLatest(ctx context.Context) ([]domain.PredictionSnapshot, error) {
	// One query to find the highest week_simulated stored.
	var maxWeek struct{ Week int }
	if err := r.db.WithContext(ctx).
		Model(&PredictionSnapshotModel{}).
		Select("COALESCE(MAX(week_simulated), 0) AS week").
		Scan(&maxWeek).Error; err != nil {
		return nil, domain.NewDBError("prediction.GetLatest.max", err)
	}

	if maxWeek.Week == 0 {
		return []domain.PredictionSnapshot{}, nil
	}

	var models []PredictionSnapshotModel
	if err := r.db.WithContext(ctx).
		Where("week_simulated = ?", maxWeek.Week).
		Order("win_probability DESC").
		Find(&models).Error; err != nil {
		return nil, domain.NewDBError("prediction.GetLatest", err)
	}

	result := make([]domain.PredictionSnapshot, len(models))
	for i, m := range models {
		result[i] = toPredictionDomain(m)
	}
	return result, nil
}

// DeleteAll removes all snapshot rows. Called by POST /league/reset.
func (r *predictionRepo) DeleteAll(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("1 = 1").Delete(&PredictionSnapshotModel{}).Error; err != nil {
		return domain.NewDBError("prediction.DeleteAll", err)
	}
	return nil
}
