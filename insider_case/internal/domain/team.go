package domain

import "time"

// Team represents a football club participating in the league.
// SBase is the base strength derived from real-world API data at league init
// and is frozen for the entire season — it never changes between weeks.
type Team struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`       // e.g. "Arsenal"
	ShortName string    `json:"short_name"` // e.g. "ARS"
	SBase     float64   `json:"s_base"`     // Base strength [0–100], frozen at init
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// KnownTeams holds the four teams this league is built around.
// ShortNames and display names are fixed; SBase is populated by the fetcher.
var KnownTeams = []struct {
	Name      string
	ShortName string
}{
	{Name: "Arsenal", ShortName: "ARS"},
	{Name: "Liverpool", ShortName: "LIV"},
	{Name: "Chelsea", ShortName: "CHE"},
	{Name: "Manchester United", ShortName: "MUN"},
}
