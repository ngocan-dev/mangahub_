package manga

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	schema := `
    CREATE TABLE mangas (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        alt_title TEXT,
        cover_url TEXT,
        author TEXT,
        artist TEXT,
        status TEXT NOT NULL DEFAULT 'ongoing',
        synopsis TEXT,
        language TEXT DEFAULT 'ja',
        rating_average REAL NOT NULL DEFAULT 0,
        rating_count INTEGER NOT NULL DEFAULT 0,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE tags (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );

    CREATE TABLE manga_tags (
        manga_id INTEGER NOT NULL,
        tag_id INTEGER NOT NULL,
        PRIMARY KEY (manga_id, tag_id)
    );
    `

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func seedNovels(t *testing.T, db *sql.DB) {
	t.Helper()

	inserts := []struct {
		slug        string
		title       string
		author      string
		genres      []string
		status      string
		description string
		image       string
		rating      float64
		views       int
	}{
		{"hero-saga", "Hero Saga", "AuthorA", []string{"Action", "Fantasy"}, "ongoing", "Epic hero journey", "hero.jpg", 4.5, 1000},
		{"mystery-tales", "Mystery Tales", "AuthorB", []string{"Mystery", "Thriller"}, "completed", "Mysterious cases", "mystery.jpg", 4.8, 5000},
		{"action-hero", "Action Hero", "AuthorC", []string{"Action"}, "ongoing", "Another hero story", "action.jpg", 3.2, 200},
	}

	for _, novel := range inserts {
		res, err := db.Exec(`
            INSERT INTO mangas (slug, title, author, status, synopsis, cover_url, rating_average, rating_count)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `, novel.slug, novel.title, novel.author, novel.status, novel.description, novel.image, novel.rating, novel.views)
		if err != nil {
			t.Fatalf("failed to insert novel %s: %v", novel.title, err)
		}

		id, err := res.LastInsertId()
		if err != nil {
			t.Fatalf("failed to get inserted id: %v", err)
		}

		for _, genre := range novel.genres {
			if _, err := db.Exec(`INSERT INTO tags (name) VALUES (?) ON CONFLICT(name) DO NOTHING`, genre); err != nil {
				t.Fatalf("failed to insert tag: %v", err)
			}
			var tagID int64
			if err := db.QueryRow(`SELECT id FROM tags WHERE name = ?`, genre).Scan(&tagID); err != nil {
				t.Fatalf("failed to fetch tag id: %v", err)
			}
			if _, err := db.Exec(`INSERT INTO manga_tags (manga_id, tag_id) VALUES (?, ?)`, id, tagID); err != nil {
				t.Fatalf("failed to link tag: %v", err)
			}
		}
	}
}

func TestRepositorySearch_FiltersAndPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedNovels(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	resp, total, err := repo.Search(ctx, SearchRequest{
		Query:  "Hero",
		Genres: []string{"Action"},
		Status: "ongoing",
		Page:   1,
		Limit:  1,
		SortBy: "rating",
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 result on first page, got %d", len(resp))
	}

	if resp[0].Title != "Hero Saga" {
		t.Fatalf("expected highest rated hero manga first, got %s", resp[0].Title)
	}

	respPage2, _, err := repo.Search(ctx, SearchRequest{
		Query:  "Hero",
		Genres: []string{"Action"},
		Status: "ongoing",
		Page:   2,
		Limit:  1,
		SortBy: "rating",
	})
	if err != nil {
		t.Fatalf("search page 2 failed: %v", err)
	}

	if len(respPage2) != 1 {
		t.Fatalf("expected 1 result on second page, got %d", len(respPage2))
	}

	if respPage2[0].Title != "Action Hero" {
		t.Fatalf("expected second hero manga on page 2, got %s", respPage2[0].Title)
	}
}

func TestRepositorySearch_NoResults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedNovels(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	resp, total, err := repo.Search(ctx, SearchRequest{
		Query: "Nonexistent",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if total != 0 {
		t.Fatalf("expected total 0, got %d", total)
	}

	if len(resp) != 0 {
		t.Fatalf("expected no results, got %d", len(resp))
	}
}
