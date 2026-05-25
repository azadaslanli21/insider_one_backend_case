package repository

import (
	"time"

	"github.com/insider/football-simulator/internal/domain"
)

// TeamModel is the GORM-annotated DB struct for the teams table.
// Kept separate from domain.Team so the domain package never imports GORM.
type TeamModel struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	Name      string  `gorm:"uniqueIndex;not null"`
	ShortName string  `gorm:"size:5;not null"`
	SBase     float64 `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TeamModel) TableName() string { return "teams" }

func toTeamDomain(m TeamModel) domain.Team {
	return domain.Team{
		ID:        m.ID,
		Name:      m.Name,
		ShortName: m.ShortName,
		SBase:     m.SBase,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toTeamModel(t domain.Team) TeamModel {
	return TeamModel{
		ID:        t.ID,
		Name:      t.Name,
		ShortName: t.ShortName,
		SBase:     t.SBase,
	}
}

// ─────────────────────────────────────────────────────────────────────────────

// MatchModel is the GORM-annotated DB struct for the matches table.
// HomeGoals and AwayGoals are pointers: NULL in DB = match not yet played.
type MatchModel struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	WeekID     int    `gorm:"not null;index"`
	HomeTeamID uint   `gorm:"not null"`
	AwayTeamID uint   `gorm:"not null"`
	HomeGoals  *int   `gorm:"default:null"`
	AwayGoals  *int   `gorm:"default:null"`
	Played     bool   `gorm:"not null;default:false;index"`
	Reasoning  string `gorm:"type:text"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (MatchModel) TableName() string { return "matches" }

// toMatchDomain converts a MatchModel + pre-loaded team models into domain.Match.
// homeTeam and awayTeam are passed in from join/preload logic in the repo.
func toMatchDomain(m MatchModel, homeTeam, awayTeam domain.Team) domain.Match {
	return domain.Match{
		ID:         m.ID,
		WeekID:     m.WeekID,
		HomeTeamID: m.HomeTeamID,
		HomeTeam:   homeTeam,
		AwayTeamID: m.AwayTeamID,
		AwayTeam:   awayTeam,
		HomeGoals:  m.HomeGoals,
		AwayGoals:  m.AwayGoals,
		Played:     m.Played,
		Reasoning:  m.Reasoning,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func toMatchModel(m domain.Match) MatchModel {
	return MatchModel{
		ID:         m.ID,
		WeekID:     m.WeekID,
		HomeTeamID: m.HomeTeamID,
		AwayTeamID: m.AwayTeamID,
		HomeGoals:  m.HomeGoals,
		AwayGoals:  m.AwayGoals,
		Played:     m.Played,
		Reasoning:  m.Reasoning,
	}
}

// ─────────────────────────────────────────────────────────────────────────────

// PredictionSnapshotModel is the GORM-annotated DB struct for prediction_snapshots.
// One row per team per week simulated (4 rows per simulate-next call).
type PredictionSnapshotModel struct {
	ID             uint    `gorm:"primaryKey;autoIncrement"`
	WeekSimulated  int     `gorm:"not null;index"`
	TeamID         uint    `gorm:"not null"`
	TeamName       string  `gorm:"not null"`
	WinProbability float64 `gorm:"not null"`
	Simulations    int     `gorm:"not null"`
	CreatedAt      time.Time
}

func (PredictionSnapshotModel) TableName() string { return "prediction_snapshots" }

func toPredictionDomain(m PredictionSnapshotModel) domain.PredictionSnapshot {
	return domain.PredictionSnapshot{
		ID:             m.ID,
		WeekSimulated:  m.WeekSimulated,
		TeamID:         m.TeamID,
		TeamName:       m.TeamName,
		WinProbability: m.WinProbability,
		Simulations:    m.Simulations,
		CreatedAt:      m.CreatedAt,
	}
}

// StandingsSnapshotModel stores one standings row per team at each week boundary.
type StandingsSnapshotModel struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	WeekID       int    `gorm:"not null;index"`
	Position     int    `gorm:"not null"`
	TeamID       uint   `gorm:"not null"`
	TeamName     string `gorm:"not null"`
	TeamShort    string `gorm:"not null"`
	Played       int    `gorm:"not null"`
	Won          int    `gorm:"not null"`
	Drawn        int    `gorm:"not null"`
	Lost         int    `gorm:"not null"`
	GoalsFor     int    `gorm:"not null"`
	GoalsAgainst int    `gorm:"not null"`
	GoalDiff     int    `gorm:"not null"`
	Points       int    `gorm:"not null"`
	CreatedAt    time.Time
}

func (StandingsSnapshotModel) TableName() string { return "standings_snapshots" }

func toStandingsSnapshotModel(weekID int, s domain.Standings) StandingsSnapshotModel {
	return StandingsSnapshotModel{
		WeekID:       weekID,
		Position:     s.Position,
		TeamID:       s.TeamID,
		TeamName:     s.TeamName,
		TeamShort:    s.TeamShort,
		Played:       s.Played,
		Won:          s.Won,
		Drawn:        s.Drawn,
		Lost:         s.Lost,
		GoalsFor:     s.GoalsFor,
		GoalsAgainst: s.GoalsAgainst,
		GoalDiff:     s.GoalDiff,
		Points:       s.Points,
	}
}

func toStandingsDomain(m StandingsSnapshotModel) domain.Standings {
	return domain.Standings{
		Position:     m.Position,
		TeamID:       m.TeamID,
		TeamName:     m.TeamName,
		TeamShort:    m.TeamShort,
		Played:       m.Played,
		Won:          m.Won,
		Drawn:        m.Drawn,
		Lost:         m.Lost,
		GoalsFor:     m.GoalsFor,
		GoalsAgainst: m.GoalsAgainst,
		GoalDiff:     m.GoalDiff,
		Points:       m.Points,
	}
}
