package repository

import (
	"fmt"

	"github.com/insider/football-simulator/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a PostgreSQL connection via GORM and runs AutoMigrate so
// all tables exist before the first request arrives.
// Returns domain.ErrDBConnectionFailure on dial error so callers get a
// typed error rather than a raw GORM string.
func Connect(host, port, user, password, dbname string) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, password, dbname,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Info-level logs so every SQL statement appears in server output.
		// Change to logger.Silent in production if too noisy.
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, domain.ErrDBConnectionFailure
	}

	if err := migrate(db); err != nil {
		return nil, domain.NewDBError("auto-migrate", err)
	}

	return db, nil
}

// migrate runs GORM AutoMigrate for every model in the project.
// Order matters when FK constraints are present: teams before matches.
func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&TeamModel{},
		&MatchModel{},
		&PredictionSnapshotModel{},
		&StandingsSnapshotModel{},
	)
}
