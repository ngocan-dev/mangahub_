`
CREATE TABLE IF NOT EXISTS users (
  user_id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  email TEXT,
  role TEXT CHECK(role IN ('admin','user')) DEFAULT 'user',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS manga (
  manga_id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  author TEXT,
  genre TEXT,
  description TEXT,
  cover_url TEXT,
  status TEXT CHECK(status IN ('ongoing','completed','hiatus')) DEFAULT 'ongoing',
  created_by INTEGER,
  updated_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (created_by) REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS chapter (
  chapter_id INTEGER PRIMARY KEY AUTOINCREMENT,
  manga_id INTEGER,
  title TEXT,
  number INTEGER,
  content_url TEXT,
  release_date DATETIME,
  FOREIGN KEY (manga_id) REFERENCES manga(manga_id)
);

CREATE TABLE IF NOT EXISTS comment (
  comment_id INTEGER PRIMARY KEY AUTOINCREMENT,
  manga_id INTEGER,
  user_id INTEGER,
  content TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (manga_id) REFERENCES manga(manga_id),
  FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS rating (
  rating_id INTEGER PRIMARY KEY AUTOINCREMENT,
  manga_id INTEGER,
  user_id INTEGER,
  score INTEGER CHECK(score BETWEEN 1 AND 5),
  review TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (manga_id) REFERENCES manga(manga_id),
  FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS reading_history (
  history_id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER,
  chapter_id INTEGER,
  last_read_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id),
  FOREIGN KEY (chapter_id) REFERENCES chapter(chapter_id)
);

CREATE TABLE IF NOT EXISTS notifications (
  notif_id INTEGER PRIMARY KEY AUTOINCREMENT,
  sender_id INTEGER,
  message TEXT,
  delivery_type TEXT CHECK(delivery_type IN ('udp','web')),
  sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (sender_id) REFERENCES users(user_id)
);`