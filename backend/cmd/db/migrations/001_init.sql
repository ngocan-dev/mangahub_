PRAGMA foreign_keys = ON;

/* ============================================================
   DROP ALL OBJECTS (Reset database)
============================================================ */

DROP VIEW IF EXISTS vw_Novel_With_Tags;
DROP VIEW IF EXISTS vw_Novel_Rating;

DROP TRIGGER IF EXISTS Delete_Child_Comments;

DROP TABLE IF EXISTS Notification_Subscriptions;
DROP TABLE IF EXISTS Sync_Sessions;
DROP TABLE IF EXISTS Chat_Messages;
DROP TABLE IF EXISTS Chat_Rooms;
DROP TABLE IF EXISTS Progress_History;
DROP TABLE IF EXISTS Reading_Progress;
DROP TABLE IF EXISTS User_Library;
DROP TABLE IF EXISTS User_Sessions;
DROP TABLE IF EXISTS Novel_External_Links;

DROP TABLE IF EXISTS Novel_Tags;
DROP TABLE IF EXISTS Tags;
DROP TABLE IF EXISTS Comment_System;
DROP TABLE IF EXISTS Rating_System;
DROP TABLE IF EXISTS Chapters;
DROP TABLE IF EXISTS Novels;
DROP TABLE IF EXISTS Users;
DROP TABLE IF EXISTS Roles;


/* ============================================================
   1. ROLES & USERS & SESSIONS (AUTH)
============================================================ */

CREATE TABLE Roles (
    RoleId      INTEGER PRIMARY KEY AUTOINCREMENT,
    RoleName    TEXT NOT NULL UNIQUE
);

CREATE TABLE Users (
    UserId        INTEGER PRIMARY KEY AUTOINCREMENT,
    Username      TEXT NOT NULL UNIQUE,
    PasswordHash  TEXT NOT NULL,
    Email         TEXT NOT NULL UNIQUE,
    RoleId        INTEGER,
    Created_Date  DATETIME DEFAULT CURRENT_TIMESTAMP,
    Status        TEXT NOT NULL DEFAULT 'active',

    FOREIGN KEY(RoleId) REFERENCES Roles(RoleId)
);

CREATE TABLE User_Sessions (
    Session_Id    INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id       INTEGER NOT NULL,
    Access_Token  TEXT NOT NULL UNIQUE,
    Refresh_Token TEXT,
    User_Agent    TEXT,
    Created_At    DATETIME DEFAULT CURRENT_TIMESTAMP,
    Expires_At    DATETIME,
    Revoked_At    DATETIME,

    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE
);

CREATE INDEX IX_User_Sessions_User ON User_Sessions(User_Id);


/* ============================================================
   2. NOVELS / CHAPTERS / TAGS / EXTERNAL LINKS
============================================================ */

CREATE TABLE Novels (
    Novel_Id       INTEGER PRIMARY KEY AUTOINCREMENT,
    Novel_Name     TEXT NOT NULL,
    Image          TEXT,
    Genre          TEXT,
    Title          TEXT,
    Description    TEXT,
    Author         TEXT,
    Status         TEXT DEFAULT 'Ongoing',
    Rating_Point   REAL DEFAULT 0,
    Date_Updated   DATETIME DEFAULT CURRENT_TIMESTAMP,
    Created_By     INTEGER,
    
    FOREIGN KEY(Created_By) REFERENCES Users(UserId)
);

CREATE TABLE Chapters (
    Chapter_Id     INTEGER PRIMARY KEY AUTOINCREMENT,
    Novel_Id       INTEGER NOT NULL,
    Chapter_Number INTEGER NOT NULL,
    Chapter_Title  TEXT,
    Content        TEXT,
    Date_Updated   DATETIME DEFAULT CURRENT_TIMESTAMP,
    Uploaded_By    INTEGER,

    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE,
    FOREIGN KEY(Uploaded_By) REFERENCES Users(UserId),

    UNIQUE(Novel_Id, Chapter_Number)
);

CREATE INDEX IX_Chapters_Novel_Chapter
       ON Chapters(Novel_Id, Chapter_Number);

CREATE TABLE Tags (
    TagId    INTEGER PRIMARY KEY AUTOINCREMENT,
    TagName  TEXT NOT NULL UNIQUE
);

CREATE TABLE Novel_Tags (
    Novel_Id INTEGER NOT NULL,
    TagId    INTEGER NOT NULL,

    PRIMARY KEY (Novel_Id, TagId),
    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE,
    FOREIGN KEY(TagId) REFERENCES Tags(TagId) ON DELETE CASCADE
);

CREATE TABLE Novel_External_Links (
    Link_Id    INTEGER PRIMARY KEY AUTOINCREMENT,
    Novel_Id   INTEGER NOT NULL,
    Site       TEXT NOT NULL,
    Url        TEXT NOT NULL,
    Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE,
    UNIQUE (Novel_Id, Site)
);


/* ============================================================
   3. RATING & COMMENT
============================================================ */

CREATE TABLE Rating_System (
    Rating_Id    INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id      INTEGER NOT NULL,
    Novel_Id     INTEGER NOT NULL,
    Rating_Value INTEGER CHECK (Rating_Value BETWEEN 1 AND 5),
    Rating_Date  DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE,

    UNIQUE(User_Id, Novel_Id)
);

CREATE TABLE Comment_System (
    Comment_Id         INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id            INTEGER NOT NULL,
    Novel_Id           INTEGER,
    Chapter_Id         INTEGER,
    Parent_Comment_Id  INTEGER,
    Content            TEXT NOT NULL,
    Comment_Date       DATETIME DEFAULT CURRENT_TIMESTAMP,
    Is_Deleted         INTEGER DEFAULT 0,

    FOREIGN KEY(User_Id) REFERENCES Users(UserId),
    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id),
    FOREIGN KEY(Chapter_Id) REFERENCES Chapters(Chapter_Id),
    FOREIGN KEY(Parent_Comment_Id) REFERENCES Comment_System(Comment_Id)
);

/* SQLite Trigger version */
CREATE TRIGGER Delete_Child_Comments
AFTER DELETE ON Comment_System
BEGIN
    DELETE FROM Comment_System WHERE Parent_Comment_Id = OLD.Comment_Id;
END;


/* ============================================================
   4. LIBRARY & READING PROGRESS
============================================================ */

CREATE TABLE User_Library (
    Library_Id      INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id         INTEGER NOT NULL,
    Novel_Id        INTEGER NOT NULL,
    Status          TEXT NOT NULL DEFAULT 'plan_to_read',
    Rating          INTEGER,
    Is_Favorite     INTEGER NOT NULL DEFAULT 0,
    Started_At      DATETIME,
    Completed_At    DATETIME,
    Last_Updated_At DATETIME DEFAULT CURRENT_TIMESTAMP,
    Notes           TEXT,

    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE,

    UNIQUE(User_Id, Novel_Id)
);

CREATE INDEX IX_UserLibrary_User_Status
    ON User_Library(User_Id, Status);

CREATE TABLE Reading_Progress (
    Progress_Id        INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id            INTEGER NOT NULL,
    Novel_Id           INTEGER NOT NULL,
    Current_Chapter    INTEGER NOT NULL,
    Current_Chapter_Id INTEGER,
    Current_Volume     INTEGER,
    Last_Read_At       DATETIME DEFAULT CURRENT_TIMESTAMP,
    Source             TEXT,
    Notes              TEXT,

    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE,
    FOREIGN KEY(Current_Chapter_Id) REFERENCES Chapters(Chapter_Id),

    UNIQUE(User_Id, Novel_Id)
);

CREATE INDEX IX_ReadingProgress_User_Novel
    ON Reading_Progress(User_Id, Novel_Id);

CREATE TABLE Progress_History (
    History_Id   INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id      INTEGER NOT NULL,
    Novel_Id     INTEGER NOT NULL,
    From_Chapter INTEGER,
    To_Chapter   INTEGER,
    From_Volume  INTEGER,
    To_Volume    INTEGER,
    Notes        TEXT,
    Created_At   DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE
);

CREATE INDEX IX_ProgressHistory_User
    ON Progress_History(User_Id);

CREATE INDEX IX_ProgressHistory_Novel
    ON Progress_History(Novel_Id);


/* ============================================================
   5. CHAT (WebSocket)
============================================================ */

CREATE TABLE Chat_Rooms (
    Room_Id    INTEGER PRIMARY KEY AUTOINCREMENT,
    Room_Code  TEXT NOT NULL UNIQUE,
    Room_Name  TEXT NOT NULL,
    Novel_Id   INTEGER,
    Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE SET NULL
);

CREATE TABLE Chat_Messages (
    Message_Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Room_Id    INTEGER NOT NULL,
    User_Id    INTEGER NOT NULL,
    Content    TEXT NOT NULL,
    Created_At DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(Room_Id) REFERENCES Chat_Rooms(Room_Id) ON DELETE CASCADE,
    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE
);

CREATE INDEX IX_ChatMessages_Room ON Chat_Messages(Room_Id);
CREATE INDEX IX_ChatMessages_User ON Chat_Messages(User_Id);


/* ============================================================
   6. TCP SYNC SESSIONS
============================================================ */

CREATE TABLE Sync_Sessions (
    Sync_Id      INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id      INTEGER NOT NULL,
    Device_Name  TEXT,
    Device_Type  TEXT,
    Status       TEXT NOT NULL DEFAULT 'active',
    Started_At   DATETIME DEFAULT CURRENT_TIMESTAMP,
    Last_Seen_At DATETIME,
    Last_Ip      TEXT,

    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,

    CHECK (Status IN ('active','closed'))
);

CREATE INDEX IX_SyncSessions_User_Status
    ON Sync_Sessions(User_Id, Status);


/* ============================================================
   7. UDP NOTIFICATIONS
============================================================ */

CREATE TABLE Notification_Subscriptions (
    Subscription_Id INTEGER PRIMARY KEY AUTOINCREMENT,
    User_Id         INTEGER NOT NULL,
    Novel_Id        INTEGER,
    Is_Active       INTEGER NOT NULL DEFAULT 1,
    Created_At      DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(User_Id) REFERENCES Users(UserId) ON DELETE CASCADE,
    FOREIGN KEY(Novel_Id) REFERENCES Novels(Novel_Id) ON DELETE CASCADE
);

CREATE INDEX IX_NotifySubs_User_Novel
    ON Notification_Subscriptions(User_Id, Novel_Id);


/* ============================================================
   8. VIEWS
============================================================ */

CREATE VIEW vw_Novel_Rating AS
SELECT 
    n.Novel_Id,
    n.Novel_Name,
    AVG(r.Rating_Value) AS Average_Rating
FROM Novels n
LEFT JOIN Rating_System r ON n.Novel_Id = r.Novel_Id
GROUP BY n.Novel_Id, n.Novel_Name;

CREATE VIEW vw_Novel_With_Tags AS
SELECT
    n.Novel_Id,
    n.Novel_Name,
    GROUP_CONCAT(t.TagName, ', ') AS Tags
FROM Novels n
LEFT JOIN Novel_Tags nt ON n.Novel_Id = nt.Novel_Id
LEFT JOIN Tags t ON nt.TagId = t.TagId
GROUP BY n.Novel_Id, n.Novel_Name;


/* ============================================================
   9. SEED ROLES
============================================================ */

INSERT OR IGNORE INTO Roles(RoleName) VALUES ('Admin');
INSERT OR IGNORE INTO Roles(RoleName) VALUES ('Editor');
INSERT OR IGNORE INTO Roles(RoleName) VALUES ('User');
