# Golang Football Simulator

High-performance 4-team football league simulator built in Go with clean package boundaries.

## What It Includes

- Match engine with momentum + Poisson goal simulation
- League table rebuild logic (Points > Goal Difference > Goals Scored)
- Monte Carlo predictor (10,000 concurrent simulations)
- Gin HTTP API
- PostgreSQL persistence via GORM
- AI commentary layer (Gemini + deterministic fallback)

## Project Structure

- `cmd/api`: app entrypoint, DI wiring, router
- `internal/domain`: core entities and errors
- `internal/league_engine`: fixture generation + standings rules
- `internal/simulator`: match simulation logic
- `internal/predictor`: Monte Carlo engine
- `internal/repository`: Postgres/GORM repositories
- `internal/service`: use-case orchestration
- `internal/ai`: commentary abstraction + Gemini integration

## Prerequisites

- Go 1.23+
- PostgreSQL (local or containerized)

## Environment

Copy `.env.example` to `.env` and set values.

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=football
DB_PASSWORD=football
DB_NAME=football_simulator
SERVER_PORT=8080
FOOTBALL_DATA_API_KEY=
GEMINI_API_KEY=
GEMINI_MODEL=gemini-1.5-flash
```

## Run

### Local

```bash
go mod tidy
go run ./cmd/api
```

### Docker

```bash
docker compose up --build
```

## API Endpoints

### `POST /league/init`

Seed teams (current hardcoded strengths) and generate fixtures.

### `POST /league/simulate-next`

Simulate the next unplayed week.

Response includes:
- `results`
- `standings`
- `predictions`
- `week_id`
- `ai_commentary` (`league`, `next_week`, `provider`, `status`)

### `POST /league/simulate-all`

Simulate to season end.

Response includes:
- `weeks` (week-by-week simulation output)
- `final_standings`
- `final_predictions`
- `final_ai_summary`

### `GET /league/table`

Return latest standings snapshot.

### `GET /league/fixtures?mode=all|next`

- `mode=all`: return full fixture list
- `mode=next`: return fixtures for next unplayed week only
- default mode is `all`

### `PATCH /league/match/:id`

Override a match score and rebuild standings.

Request body:

```json
{
  "home_goals": 2,
  "away_goals": 1,
  "reasoning": "manual correction"
}
```

### `POST /league/reset`

Clear teams, matches, standings snapshots, and prediction snapshots.

## Quick Reviewer Flow

```bash
curl -X POST http://localhost:8080/league/init
curl -X GET "http://localhost:8080/league/fixtures?mode=all"
curl -X POST http://localhost:8080/league/simulate-next
curl -X GET "http://localhost:8080/league/fixtures?mode=next"
curl -X GET http://localhost:8080/league/table
curl -X POST http://localhost:8080/league/simulate-all
curl -X PATCH http://localhost:8080/league/match/1 -H "Content-Type: application/json" -d "{\"home_goals\":2,\"away_goals\":1,\"reasoning\":\"manual correction\"}"
curl -X POST http://localhost:8080/league/reset
```

## AI Commentary Behavior

- If `GEMINI_API_KEY` is set and Gemini is reachable: `status=ok`, `provider=gemini`
- If key is missing: `status=disabled`, fallback deterministic commentary
- If Gemini request fails/timeouts: `status=degraded`, fallback deterministic commentary

## Notes

- External real-data strength fetcher is intentionally deferred; `POST /league/init` currently uses hardcoded values.
- AI commentary is response-only and not stored in DB.
- `schema.sql` is included for handover/review.
