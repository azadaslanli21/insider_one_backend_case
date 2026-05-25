package domain

import "time"

// Prediction is the output of the Monte Carlo engine for a single team.
// WinProbability is expressed as a percentage (0.0–100.0).
type Prediction struct {
	TeamID         uint    `json:"team_id"`
	TeamName       string  `json:"team_name"`
	TeamShort      string  `json:"team_short"`
	WinProbability float64 `json:"win_probability"` // e.g. 42.37 means 42.37%
}

// PredictionSnapshot is the persisted record of Monte Carlo results
// taken after a specific week is simulated.
type PredictionSnapshot struct {
	ID             uint      `json:"id"`
	WeekSimulated  int       `json:"week_simulated"` // week after which this was computed
	TeamID         uint      `json:"team_id"`
	TeamName       string    `json:"team_name"`
	WinProbability float64   `json:"win_probability"`
	Simulations    int       `json:"simulations"` // always 10_000, stored for auditability
	CreatedAt      time.Time `json:"created_at"`
}

// MonteCarloConfig holds the parameters for a Monte Carlo run.
// Kept as a struct so it is easy to tune without touching engine code.
type MonteCarloConfig struct {
	Simulations int // default: 10_000
}

// DefaultMonteCarloConfig returns the standard production config.
func DefaultMonteCarloConfig() MonteCarloConfig {
	return MonteCarloConfig{
		Simulations: 10_000,
	}
}
