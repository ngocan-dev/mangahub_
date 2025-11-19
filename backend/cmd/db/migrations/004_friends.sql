-- Migration: Add Friends table for user friendships
-- Friendships are bidirectional (mutual)

CREATE TABLE IF NOT EXISTS Friends (
    Friendship_Id    INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id          INTEGER NOT NULL,
    Friend_Id        INTEGER NOT NULL,
    Status           TEXT NOT NULL DEFAULT 'pending', -- pending, accepted, blocked
    Created_At       DATETIME DEFAULT CURRENT_TIMESTAMP,
    Accepted_At      DATETIME,
    
    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    FOREIGN KEY(Friend_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    
    UNIQUE(User_Id, Friend_Id),
    CHECK(User_Id != Friend_Id),
    CHECK(Status IN ('pending', 'accepted', 'blocked'))
);

CREATE INDEX IF NOT EXISTS IX_Friends_User ON Friends(User_Id, Status);
CREATE INDEX IF NOT EXISTS IX_Friends_Friend ON Friends(Friend_Id, Status);

