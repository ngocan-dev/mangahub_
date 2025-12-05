/* READING_PROGRESS */

CREATE TABLE reading_progress (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id             INTEGER NOT NULL,
    manga_id            INTEGER NOT NULL,
    current_chapter_id  INTEGER,
    current_page        INTEGER,
    progress_percent    REAL,
    last_read_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source              TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
    FOREIGN KEY (current_chapter_id) REFERENCES chapters(id) ON DELETE SET NULL,
    UNIQUE (user_id, manga_id)
);

CREATE INDEX idx_reading_progress_user ON reading_progress(user_id);

/* READING_HISTORY */

CREATE TABLE reading_history (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    manga_id        INTEGER NOT NULL,
    chapter_id      INTEGER,
    event_type      TEXT NOT NULL, -- opened, finished_chapter, finished_manga
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source          TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
);

CREATE INDEX idx_reading_history_user ON reading_history(user_id);
