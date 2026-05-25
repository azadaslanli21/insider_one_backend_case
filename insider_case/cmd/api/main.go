package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found; using environment variables")
	}

	leagueService, err := buildLeagueService()
	if err != nil {
		log.Fatalf("failed to initialize dependencies: %v", err)
	}

	port := envOrDefault("SERVER_PORT", "8080")
	router := setupRouter(leagueService)
	log.Printf("football simulator listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
