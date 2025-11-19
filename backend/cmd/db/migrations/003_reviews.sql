-- Migration: Add Reviews table for manga reviews
-- Reviews are separate from comments - they require completed status and have ratings 1-10

CREATE TABLE IF NOT EXISTS Reviews (
    Review_Id    INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id      INTEGER NOT NULL,
    Novel_Id     INTEGER NOT NULL,
    Rating       INTEGER NOT NULL CHECK (Rating BETWEEN 1 AND 10),
    Content      TEXT NOT NULL,
    Created_At   DATETIME DEFAULT CURRENT_TIMESTAMP,
    Updated_At   DATETIME,
    
    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE,
    
    UNIQUE(User_Id, Novel_Id)
);

CREATE INDEX IF NOT EXISTS IX_Reviews_Novel ON Reviews(Novel_Id);
CREATE INDEX IF NOT EXISTS IX_Reviews_User ON Reviews(User_Id);
CREATE INDEX IF NOT EXISTS IX_Reviews_Created ON Reviews(Created_At DESC);

