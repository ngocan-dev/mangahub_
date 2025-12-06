/* FRIENDS (social graph) */

CREATE TABLE friends (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    requester_id    INTEGER NOT NULL,
    addressee_id    INTEGER NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending', -- pending, accepted, blocked
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (requester_id, addressee_id),
    FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (addressee_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_friends_user ON friends(requester_id, addressee_id);
