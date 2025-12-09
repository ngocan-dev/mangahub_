/* COMMENTS */

CREATE TABLE comments (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id             INTEGER NOT NULL,
    manga_id            INTEGER,
    chapter_id          INTEGER,
    parent_comment_id   INTEGER,
    content             TEXT NOT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted          INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_comment_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE INDEX idx_comments_manga   ON comments(manga_id);
CREATE INDEX idx_comments_chapter ON comments(chapter_id);
CREATE INDEX idx_comments_user    ON comments(user_id);

/* RATINGS (1–5 + optional review text) */

CREATE TABLE ratings (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    manga_id        INTEGER NOT NULL,
    score           INTEGER NOT NULL CHECK (score BETWEEN 1 AND 5),
    review          TEXT,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
    UNIQUE (user_id, manga_id)
);

CREATE INDEX idx_ratings_manga ON ratings(manga_id);

/* FAVORITES (simple follow) */

CREATE TABLE favorites (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    manga_id        INTEGER NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
    UNIQUE (user_id, manga_id)
);

CREATE INDEX idx_favorites_user ON favorites(user_id);

/* LIBRARIES (mal-style shelf) */

CREATE TABLE libraries (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    manga_id        INTEGER NOT NULL,
    status          TEXT NOT NULL DEFAULT 'plan_to_read', -- plan_to_read, reading, completed, dropped, on_hold
    is_favorite     INTEGER NOT NULL DEFAULT 0,
    score           INTEGER,
    notes           TEXT,
    started_at      DATETIME,
    completed_at    DATETIME,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES mangas(id) ON DELETE CASCADE,
    UNIQUE (user_id, manga_id)
);

CREATE INDEX idx_libraries_user_status ON libraries(user_id, status);
