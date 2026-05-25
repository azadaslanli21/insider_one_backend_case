package domain

import "fmt"

// AppError is the base custom error type for all domain-level errors.
// It carries a machine-readable Code for programmatic handling and
// a human-readable Message for API responses.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Sentinel errors — compare with errors.Is() or direct pointer equality.
var (
	// ErrNoMatchesRemaining is returned by simulate-next when all 12 fixtures
	// in the season have already been played.
	ErrNoMatchesRemaining = &AppError{
		Code:    "NO_MATCHES_REMAINING",
		Message: "no unplayed matches remaining — the season is complete",
	}

	// ErrSeasonAlreadyComplete is returned by simulate-all when called on
	// a finished season.
	ErrSeasonAlreadyComplete = &AppError{
		Code:    "SEASON_ALREADY_COMPLETE",
		Message: "the season is already complete; call POST /league/reset to start over",
	}

	// ErrLeagueNotInitialized is returned when any simulation endpoint is called
	// before POST /league/init has been executed.
	ErrLeagueNotInitialized = &AppError{
		Code:    "LEAGUE_NOT_INITIALIZED",
		Message: "league has not been initialized; call POST /league/init first",
	}

	// ErrMatchNotFound is returned by the edit endpoint when the given match ID
	// does not exist in the database.
	ErrMatchNotFound = &AppError{
		Code:    "MATCH_NOT_FOUND",
		Message: "match not found",
	}

	// ErrMatchAlreadyPlayed is returned when trying to simulate a match that
	// has already been played (use PATCH /league/match/:id to edit instead).
	ErrMatchAlreadyPlayed = &AppError{
		Code:    "MATCH_ALREADY_PLAYED",
		Message: "match has already been played; use PATCH /league/match/:id to edit the result",
	}

	// ErrDBConnectionFailure is returned when the application cannot reach
	// the PostgreSQL database on startup or during a query.
	ErrDBConnectionFailure = &AppError{
		Code:    "DB_CONNECTION_FAILURE",
		Message: "failed to connect to the database",
	}

	// ErrInvalidGoalCount is returned by the edit endpoint when a caller
	// submits a negative goal count.
	ErrInvalidGoalCount = &AppError{
		Code:    "INVALID_GOAL_COUNT",
		Message: "goal counts must be zero or positive integers",
	}

	// ErrInvalidMode is returned when a query mode is not one of the supported values.
	ErrInvalidMode = &AppError{
		Code:    "INVALID_MODE",
		Message: "invalid mode; supported values are 'all' or 'next'",
	}
)

// NewNotFoundError returns a contextual not-found error for any entity.
func NewNotFoundError(entity string, id any) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s with id %v not found", entity, id),
	}
}

// NewDBError wraps a raw database error with context.
func NewDBError(op string, cause error) *AppError {
	return &AppError{
		Code:    "DB_ERROR",
		Message: fmt.Sprintf("database error during %s: %v", op, cause),
	}
}
