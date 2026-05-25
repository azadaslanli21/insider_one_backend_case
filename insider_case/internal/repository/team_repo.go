package repository

import (
	"context"
	"errors"

	"github.com/insider/football-simulator/internal/domain"
	"gorm.io/gorm"
)

type teamRepo struct {
	db *gorm.DB
}

// NewTeamRepository returns a PostgreSQL-backed TeamRepository.
func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepo{db: db}
}

func (r *teamRepo) CreateBatch(ctx context.Context, teams []domain.Team) error {
	models := make([]TeamModel, len(teams))
	for i, t := range teams {
		models[i] = toTeamModel(t)
	}

	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return domain.NewDBError("team.CreateBatch", err)
	}
	return nil
}

func (r *teamRepo) GetAll(ctx context.Context) ([]domain.Team, error) {
	var models []TeamModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, domain.NewDBError("team.GetAll", err)
	}

	teams := make([]domain.Team, len(models))
	for i, m := range models {
		teams[i] = toTeamDomain(m)
	}
	return teams, nil
}

func (r *teamRepo) GetByID(ctx context.Context, id uint) (domain.Team, error) {
	var m TeamModel
	err := r.db.WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Team{}, domain.NewNotFoundError("team", id)
	}
	if err != nil {
		return domain.Team{}, domain.NewDBError("team.GetByID", err)
	}
	return toTeamDomain(m), nil
}

func (r *teamRepo) DeleteAll(ctx context.Context) error {
	// Unscoped hard delete; WHERE 1=1 required by GORM for table-wide deletes.
	if err := r.db.WithContext(ctx).Where("1 = 1").Delete(&TeamModel{}).Error; err != nil {
		return domain.NewDBError("team.DeleteAll", err)
	}
	return nil
}
