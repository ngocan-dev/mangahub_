package manga

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	schema := `
    CREATE TABLE Novels (
        Novel_Id INTEGER PRIMARY KEY AUTOINCREMENT,
        Novel_Name TEXT NOT NULL,
        Title TEXT,
        Author TEXT,
        Genre TEXT,
        Status TEXT,
        Description TEXT,
        Image TEXT,
        Rating_Point REAL,
        Date_Updated DATETIME
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
		name        string
		title       string
		author      string
		genre       string
		status      string
		description string
		image       string
		rating      float64
	}{
		{"Hero Saga", "Hero Saga", "AuthorA", "Action", "Ongoing", "Epic hero journey", "hero.jpg", 4.5},
		{"Mystery Tales", "Mystery Tales", "AuthorB", "Mystery", "Completed", "Mysterious cases", "mystery.jpg", 4.8},
		{"Action Hero", "Action Hero", "AuthorC", "Action", "Ongoing", "Another hero story", "action.jpg", 3.2},
	}

	for _, novel := range inserts {
		_, err := db.Exec(`
            INSERT INTO Novels (Novel_Name, Title, Author, Genre, Status, Description, Image, Rating_Point)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `, novel.name, novel.title, novel.author, novel.genre, novel.status, novel.description, novel.image, novel.rating)
		if err != nil {
			t.Fatalf("failed to insert novel %s: %v", novel.name, err)
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
		Status: "Ongoing",
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

	if resp[0].Name != "Hero Saga" {
		t.Fatalf("expected highest rated hero manga first, got %s", resp[0].Name)
	}

	respPage2, _, err := repo.Search(ctx, SearchRequest{
		Query:  "Hero",
		Genres: []string{"Action"},
		Status: "Ongoing",
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

	if respPage2[0].Name != "Action Hero" {
		t.Fatalf("expected second hero manga on page 2, got %s", respPage2[0].Name)
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
