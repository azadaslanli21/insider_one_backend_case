CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    short_name VARCHAR(5) NOT NULL,
    s_base DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS matches (
    id BIGSERIAL PRIMARY KEY,
    week_id INT NOT NULL,
    home_team_id BIGINT NOT NULL REFERENCES teams(id),
    away_team_id BIGINT NOT NULL REFERENCES teams(id),
    home_goals INT NULL,
    away_goals INT NULL,
    played BOOLEAN NOT NULL DEFAULT FALSE,
    reasoning TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_matches_week_id ON matches(week_id);
CREATE INDEX IF NOT EXISTS idx_matches_played ON matches(played);
CREATE INDEX IF NOT EXISTS idx_matches_played_week_id ON matches(played, week_id);

CREATE TABLE IF NOT EXISTS prediction_snapshots (
    id BIGSERIAL PRIMARY KEY,
    week_simulated INT NOT NULL,
    team_id BIGINT NOT NULL REFERENCES teams(id),
    team_name TEXT NOT NULL,
    win_probability DOUBLE PRECISION NOT NULL,
    simulations INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prediction_snapshots_week ON prediction_snapshots(week_simulated);

CREATE TABLE IF NOT EXISTS standings_snapshots (
    id BIGSERIAL PRIMARY KEY,
    week_id INT NOT NULL,
    position INT NOT NULL,
    team_id BIGINT NOT NULL REFERENCES teams(id),
    team_name TEXT NOT NULL,
    team_short TEXT NOT NULL,
    played INT NOT NULL,
    won INT NOT NULL,
    drawn INT NOT NULL,
    lost INT NOT NULL,
    goals_for INT NOT NULL,
    goals_against INT NOT NULL,
    goal_diff INT NOT NULL,
    points INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_standings_snapshots_week ON standings_snapshots(week_id);

-- Useful review queries
-- Latest standings table:
-- SELECT * FROM standings_snapshots WHERE week_id = (SELECT MAX(week_id) FROM standings_snapshots) ORDER BY position;
-- Latest prediction snapshot:
-- SELECT * FROM prediction_snapshots WHERE week_simulated = (SELECT MAX(week_simulated) FROM prediction_snapshots) ORDER BY win_probability DESC;
-- Full fixture list:
-- SELECT * FROM matches ORDER BY week_id, id;
-- Next unplayed week fixtures (same behavior as mode=next):
-- SELECT * FROM matches WHERE week_id = (
--   SELECT MIN(week_id) FROM matches WHERE played = false
-- ) ORDER BY id;
-- Note: AI commentary is response-only by design; no DB table is used for AI outputs.
