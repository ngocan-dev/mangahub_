/* FULL-TEXT SEARCH FOR MANGAS */

CREATE VIRTUAL TABLE IF NOT EXISTS manga_search USING fts5(
    title,
    synopsis,
    content='mangas',
    content_rowid='id'
);

/* Seed FTS from existing mangas */

INSERT INTO manga_search(rowid, title, synopsis)
SELECT id, COALESCE(title, ''), COALESCE(synopsis, '')
FROM mangas;

/* Triggers keep FTS in sync */

CREATE TRIGGER IF NOT EXISTS mangas_ai_fts
AFTER INSERT ON mangas
BEGIN
    INSERT INTO manga_search(rowid, title, synopsis)
    VALUES (new.id, new.title, new.synopsis);
END;

CREATE TRIGGER IF NOT EXISTS mangas_au_fts
AFTER UPDATE ON mangas
BEGIN
    INSERT INTO manga_search(manga_search, rowid, title, synopsis)
    VALUES ('delete', old.id, old.title, old.synopsis);
    INSERT INTO manga_search(rowid, title, synopsis)
    VALUES (new.id, new.title, new.synopsis);
END;

CREATE TRIGGER IF NOT EXISTS mangas_ad_fts
AFTER DELETE ON mangas
BEGIN
    INSERT INTO manga_search(manga_search, rowid, title, synopsis)
    VALUES ('delete', old.id, old.title, old.synopsis);
END;
