-- Migration: Add Reading Statistics cache table
-- Stores pre-calculated reading statistics for performance

CREATE TABLE IF NOT EXISTS Reading_Statistics (
    Stat_Id              INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id              INTEGER NOT NULL,
    Total_Chapters_Read  INTEGER DEFAULT 0,
    Total_Manga_Read     INTEGER DEFAULT 0,
    Total_Manga_Reading  INTEGER DEFAULT 0,
    Total_Manga_Planned  INTEGER DEFAULT 0,
    Favorite_Genres      TEXT, -- JSON array of genres with counts
    Average_Rating       REAL DEFAULT 0,
    Total_Reading_Time_Hours REAL DEFAULT 0, -- Estimated reading time
    Current_Streak_Days   INTEGER DEFAULT 0,
    Longest_Streak_Days   INTEGER DEFAULT 0,
    Monthly_Stats        TEXT, -- JSON object with monthly statistics
    Yearly_Stats         TEXT, -- JSON object with yearly statistics
    Last_Calculated_At    DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    
    UNIQUE(User_Id)
);

CREATE INDEX IF NOT EXISTS IX_ReadingStatistics_User ON Reading_Statistics(User_Id);
CREATE INDEX IF NOT EXISTS IX_ReadingStatistics_LastCalculated ON Reading_Statistics(Last_Calculated_At);

