/* READING_STATISTICS (cache) */

CREATE TABLE reading_statistics (
    id                       INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id                  INTEGER NOT NULL,
    total_chapters_read      INTEGER DEFAULT 0,
    total_manga_read         INTEGER DEFAULT 0,
    total_manga_reading      INTEGER DEFAULT 0,
    total_manga_planned      INTEGER DEFAULT 0,
    favorite_genres          TEXT, -- JSON
    average_rating           REAL DEFAULT 0,
    total_reading_time_hours REAL DEFAULT 0,
    current_streak_days      INTEGER DEFAULT 0,
    longest_streak_days      INTEGER DEFAULT 0,
    monthly_stats            TEXT, -- JSON
    yearly_stats             TEXT, -- JSON
    last_calculated_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user_id)
);

CREATE INDEX idx_reading_statistics_user
    ON reading_statistics(user_id);

CREATE INDEX idx_reading_statistics_last_calculated
    ON reading_statistics(last_calculated_at);

/* READING_GOALS */

CREATE TABLE reading_goals (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    goal_type       TEXT NOT NULL, -- 'chapters', 'manga', 'reading_time'
    target_value    INTEGER NOT NULL,
    current_value   INTEGER DEFAULT 0,
    period_type     TEXT NOT NULL, -- 'daily', 'weekly', 'monthly', 'yearly'
    period_start    DATETIME NOT NULL,
    period_end      DATETIME NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active', -- 'active', 'completed', 'failed'
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CHECK (goal_type IN ('chapters', 'manga', 'reading_time')),
    CHECK (period_type IN ('daily', 'weekly', 'monthly', 'yearly')),
    CHECK (status IN ('active', 'completed', 'failed'))
);

CREATE INDEX idx_reading_goals_user
    ON reading_goals(user_id, status);

CREATE INDEX idx_reading_goals_period
    ON reading_goals(period_start, period_end);
