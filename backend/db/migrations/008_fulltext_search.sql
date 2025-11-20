-- Enable full-text search for advanced manga queries
CREATE VIRTUAL TABLE IF NOT EXISTS NovelSearch USING fts5(
    Title,
    Description,
    content='Novels',
    content_rowid='Novel_Id'
);

-- Seed FTS index from existing data
INSERT INTO NovelSearch(rowid, Title, Description)
SELECT Novel_Id, COALESCE(Title, ''), COALESCE(Description, '')
FROM Novels;

-- Keep FTS index in sync with Novels table
CREATE TRIGGER IF NOT EXISTS novels_ai_fts AFTER INSERT ON Novels BEGIN
    INSERT INTO NovelSearch(rowid, Title, Description)
    VALUES (new.Novel_Id, new.Title, new.Description);
END;

CREATE TRIGGER IF NOT EXISTS novels_au_fts AFTER UPDATE ON Novels BEGIN
    INSERT INTO NovelSearch(NovelSearch, rowid, Title, Description)
    VALUES('delete', old.Novel_Id, old.Title, old.Description);
    INSERT INTO NovelSearch(rowid, Title, Description)
    VALUES (new.Novel_Id, new.Title, new.Description);
END;

CREATE TRIGGER IF NOT EXISTS novels_ad_fts AFTER DELETE ON Novels BEGIN
    INSERT INTO NovelSearch(NovelSearch, rowid, Title, Description)
    VALUES('delete', old.Novel_Id, old.Title, old.Description);
END;
