/* TAGS */

CREATE TABLE tags (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name    TEXT NOT NULL UNIQUE
);

/* MANGAS */

CREATE TABLE mangas (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    slug            TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    alt_title       TEXT,
    cover_url       TEXT,
    author          TEXT,
    artist          TEXT,
    status          TEXT NOT NULL DEFAULT 'ongoing', -- ongoing, completed, hiatus, cancelled
    synopsis        TEXT,
    language        TEXT DEFAULT 'ja',
    created_by      INTEGER,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rating_average  REAL NOT NULL DEFAULT 0,
    rating_count    INTEGER NOT NULL DEFAULT 0,
    last_chapter    INTEGER,
    last_chapter_at DATETIME,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_mangas_status ON mangas(status);
CREATE INDEX idx_mangas_title ON mangas(title);

/* MANGA_TAGS (many-to-many) */

CREATE TABLE manga_tags (
    manga_id    INTEGER NOT NULL,
    tag_id      INTEGER NOT NULL,
    PRIMARY KEY (manga_id, tag_id),
    FOREIGN KEY (manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

/* CHAPTERS */

CREATE TABLE chapters (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    manga_id        INTEGER NOT NULL,
    number          INTEGER NOT NULL,
    title           TEXT,
    language        TEXT DEFAULT 'ja',
    volume_number   INTEGER,
    pages_count     INTEGER,
    content_text    TEXT,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    uploaded_by     INTEGER,
    FOREIGN KEY (manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE (manga_id, number, language)
);

CREATE INDEX idx_chapters_manga_number ON chapters(manga_id, number);
