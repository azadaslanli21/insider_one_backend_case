package main

import (
	"os"

	"github.com/insider/football-simulator/internal/ai"
	"github.com/insider/football-simulator/internal/fetcher"
	"github.com/insider/football-simulator/internal/repository"
	"github.com/insider/football-simulator/internal/service"
)

func buildLeagueService() (*service.LeagueService, error) {
	db, err := repository.Connect(
		envOrDefault("DB_HOST", "localhost"),
		envOrDefault("DB_PORT", "5432"),
		envOrDefault("DB_USER", "football"),
		envOrDefault("DB_PASSWORD", "football"),
		envOrDefault("DB_NAME", "football_simulator"),
	)
	if err != nil {
		return nil, err
	}

	strengthProvider := fetcher.NewStrengthProvider(os.Getenv("FOOTBALL_DATA_API_KEY"))
	commentary := ai.NewGeminiGenerator(os.Getenv("GEMINI_API_KEY"), os.Getenv("GEMINI_MODEL"))
	return service.NewLeagueService(
		db,
		repository.NewTeamRepository(db),
		repository.NewMatchRepository(db),
		repository.NewPredictionRepository(db),
		repository.NewStandingsRepository(db),
		strengthProvider,
		commentary,
	), nil
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
