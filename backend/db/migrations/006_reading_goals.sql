-- Migration: Add Reading Goals table for user reading goals
-- Users can set goals for chapters, manga, or reading time

CREATE TABLE IF NOT EXISTS Reading_Goals (
    Goal_Id          INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id          INTEGER NOT NULL,
    Goal_Type        TEXT NOT NULL, -- 'chapters', 'manga', 'reading_time'
    Target_Value     INTEGER NOT NULL, -- Target number (chapters, manga, or hours)
    Current_Value    INTEGER DEFAULT 0, -- Current progress
    Period_Type      TEXT NOT NULL, -- 'daily', 'weekly', 'monthly', 'yearly'
    Period_Start     DATETIME NOT NULL, -- Start of the goal period
    Period_End       DATETIME NOT NULL, -- End of the goal period
    Status           TEXT NOT NULL DEFAULT 'active', -- 'active', 'completed', 'failed'
    Created_At       DATETIME DEFAULT CURRENT_TIMESTAMP,
    Updated_At       DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    
    CHECK(Goal_Type IN ('chapters', 'manga', 'reading_time')),
    CHECK(Period_Type IN ('daily', 'weekly', 'monthly', 'yearly')),
    CHECK(Status IN ('active', 'completed', 'failed'))
);

CREATE INDEX IF NOT EXISTS IX_ReadingGoals_User ON Reading_Goals(User_Id, Status);
CREATE INDEX IF NOT EXISTS IX_ReadingGoals_Period ON Reading_Goals(Period_Start, Period_End);

