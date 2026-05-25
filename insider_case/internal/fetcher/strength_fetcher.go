package fetcher

import "github.com/insider/football-simulator/internal/domain"

// StrengthProvider is the single interface both the hardcoded fallback
// and the future API-backed implementation must satisfy.
// The DI layer depends only on this interface, so Phase 2b can slot in
// a real provider without touching any other package.
type StrengthProvider interface {
	// FetchStrengths returns a map of team ShortName → S_base value.
	// Called exactly once at POST /league/init; values are frozen for the season.
	FetchStrengths() (map[string]float64, error)
}

// hardcodedProvider is the Phase 2a implementation.
// S_base values are derived from a composite of recent Premier League
// performance across the 2022–2025 seasons:
//
//	Base 50 + weighted scoring across four dimensions:
//	  • Avg league position (last 3 seasons)   → up to +20 pts
//	  • Avg points-per-game                    → up to +15 pts
//	  • Avg seasonal goal difference           → up to +10 pts
//	  • Squad quality / UCL participation      → up to  +5 pts
//
//	Liverpool  88 — title winner 23/24, UCL pedigree, best GD in PL
//	Arsenal    84 — 2nd place 23/24, highest press intensity, deep squad
//	Chelsea    74 — inconsistent results but elite squad value, top-6 finishes
//	Man United 63 — 8th place 23/24 (worst modern-era finish), poor GD (-18)
type hardcodedProvider struct{}

var hardcodedStrengths = map[string]float64{
	"LIV": 88.0,
	"ARS": 84.0,
	"CHE": 74.0,
	"MUN": 63.0,
}

func (h *hardcodedProvider) FetchStrengths() (map[string]float64, error) {
	// Return a copy so the caller cannot mutate the package-level map
	result := make(map[string]float64, len(hardcodedStrengths))
	for k, v := range hardcodedStrengths {
		result[k] = v
	}
	return result, nil
}

// NewStrengthProvider is the factory used by the DI layer.
// apiKey is reserved for Phase 2b — pass an empty string for now.
// When Phase 2b is ready, the body becomes:
//
//	if apiKey != "" { return newFootballDataProvider(apiKey) }
func NewStrengthProvider(apiKey string) StrengthProvider {
	return &hardcodedProvider{}
}

// ApplyToTeams merges a strengths map (ShortName → S_base) into a team slice.
// Used by the init service after calling FetchStrengths().
func ApplyToTeams(teams []domain.Team, strengths map[string]float64) []domain.Team {
	for i, t := range teams {
		if s, ok := strengths[t.ShortName]; ok {
			teams[i].SBase = s
		}
	}
	return teams
}
