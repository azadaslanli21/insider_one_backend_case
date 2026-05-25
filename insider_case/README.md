# Golang Football Simulator

4-team football league simulator built with Clean Architecture style boundaries:
- Match simulation engine (Poisson + momentum)
- League rules/standings engine
- Monte Carlo predictor (10,000 concurrent simulations)
- Gin HTTP API
- PostgreSQL persistence with GORM

## Run

### 1) Local

```bash
go mod tidy
go run ./cmd/api
```

### 2) Docker

```bash
docker compose up --build
```

## Environment

Copy `.env.example` to `.env` and adjust values.

## Endpoints

### `POST /league/init`
Seeds teams with hardcoded strengths and generates fixtures.

### `POST /league/simulate-next`
Simulates next unplayed week and returns:
```json
{
  "week_id": 1,
  "results": [],
  "standings": [],
  "predictions": []
}
```

### `POST /league/simulate-all`
Simulates until season end and returns:
```json
{
  "weeks": [],
  "final_standings": [],
  "final_predictions": []
}
```

### `GET /league/table`
Returns latest standings snapshot.

### `GET /league/fixtures?mode=all|next`
Returns full fixture list or only next unplayed week fixtures.

### `POST /league/reset`
Clears teams, matches, standings snapshots, and prediction snapshots.

### `PATCH /league/match/:id`
Overrides a match score and recalculates standings.

Request:
```json
{
  "home_goals": 2,
  "away_goals": 1,
  "reasoning": "manual correction"
}
```

## Notes

- External strength API integration is intentionally deferred for now.
- Current init uses hardcoded strengths via `internal/fetcher`.
- Standings are persisted as weekly snapshots in `standings_snapshots`.
- AI commentary uses Gemini when `GEMINI_API_KEY` is configured.
- If Gemini is unavailable, API returns fallback commentary with `ai_commentary.status=degraded` or `disabled`.
