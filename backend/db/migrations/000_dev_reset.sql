PRAGMA foreign_keys = OFF;

/* Drop views */
DROP VIEW IF EXISTS vw_manga_with_tags;
DROP VIEW IF EXISTS vw_manga_rating;

/* Drop tables (ngược dependency) */
DROP TABLE IF EXISTS notification_subscriptions;
DROP TABLE IF EXISTS sync_sessions;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_rooms;
DROP TABLE IF EXISTS reading_history;
DROP TABLE IF EXISTS reading_progress;
DROP TABLE IF EXISTS libraries;
DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS ratings;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS chapters;
DROP TABLE IF EXISTS manga_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS friends;
DROP TABLE IF EXISTS mangas;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

PRAGMA foreign_keys = ON;
