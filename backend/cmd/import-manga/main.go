package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"
	"unicode"

	dbpkg "github.com/ngocan-dev/mangahub/backend/db"
	"github.com/ngocan-dev/mangahub/backend/domain/manga"
	"github.com/ngocan-dev/mangahub/backend/internal/config"
	chapterrepository "github.com/ngocan-dev/mangahub/backend/internal/repository/chapter"
	chapterservice "github.com/ngocan-dev/mangahub/backend/internal/service/chapter"
)

type MangaSeed struct {
	Title       string
	AltTitle    string
	Author      string
	Artist      string
	Genres      []string
	Status      string
	Description string
	Rating      float64
	Views       int64
	CoverURL    string
	Slug        string
}

type ChapterSeed struct {
	Number      int
	Title       string
	ContentText string
	Language    string
}

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := dbpkg.Open(cfg.DB.Driver, cfg.DB.DSN, nil)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	mangaService := manga.NewService(db)
	chapterRepo := chapterrepository.NewRepository(db)
	chapterSvc := chapterservice.NewService(chapterRepo)
	mangaService.SetChapterService(chapterSvc)

	ctx := context.Background()
	seeds := generateMangaSeeds(45)

	for _, seed := range seeds {
		existing, err := mangaService.GetByTitle(ctx, seed.Title)
		if err != nil {
			log.Printf("skip %s due to lookup error: %v", seed.Title, err)
			continue
		}
		if existing != nil {
			log.Printf("skip existing manga: %s", seed.Title)
			continue
		}

		chapterSeeds := generateChapters(seed)
		req := manga.CreateMangaRequest{
			Title:       seed.Title,
			AltTitle:    seed.AltTitle,
			Slug:        seed.Slug,
			CoverURL:    seed.CoverURL,
			Author:      seed.Author,
			Artist:      seed.Artist,
			Status:      seed.Status,
			Synopsis:    seed.Description,
			Genres:      seed.Genres,
			Rating:      seed.Rating,
			Views:       seed.Views,
			Language:    "ja",
			LastChapter: len(chapterSeeds),
		}

		mangaID, err := mangaService.CreateManga(ctx, req)
		if err != nil {
			log.Printf("failed to create manga %s: %v", seed.Title, err)
			continue
		}
		log.Printf("created manga [%d]: %s", mangaID, seed.Title)

		for _, ch := range chapterSeeds {
			if _, err := chapterSvc.CreateChapter(ctx, mangaID, ch.Number, ch.Title, ch.ContentText, ch.Language); err != nil {
				log.Printf("failed to create chapter %d for %s: %v", ch.Number, seed.Title, err)
				continue
			}
			log.Printf("  chapter %d added: %s", ch.Number, ch.Title)
		}
	}
}

func generateMangaSeeds(count int) []MangaSeed {
	titles := make(map[string]struct{})
	seeds := make([]MangaSeed, 0, count)

	for len(seeds) < count {
		title := randomTitle()
		if _, exists := titles[title]; exists {
			continue
		}
		titles[title] = struct{}{}

		slug := slugify(title)
		alt := randomAltTitle(title)
		genres := randomGenres()
		status := randomStatus()
		rating := randomRating()
		views := randomViews()
		author := randomAuthor()
		artist := randomArtist()
		desc := randomDescription(genres, status)
		coverURL := fmt.Sprintf("https://cdn.mangahub.fake/covers/%s.jpg", slug)

		seeds = append(seeds, MangaSeed{
			Title:       title,
			AltTitle:    alt,
			Author:      author,
			Artist:      artist,
			Genres:      genres,
			Status:      status,
			Description: desc,
			Rating:      rating,
			Views:       views,
			CoverURL:    coverURL,
			Slug:        slug,
		})
	}

	return seeds
}

func generateChapters(seed MangaSeed) []ChapterSeed {
	total := rand.Intn(16) + 5 // 5–20 chapters
	chapters := make([]ChapterSeed, 0, total)

	for i := 1; i <= total; i++ {
		chapters = append(chapters, ChapterSeed{
			Number:      i,
			Title:       fmt.Sprintf("Chapter %d: %s", i, randomChapterTitle()),
			ContentText: fmt.Sprintf("Chapter %d content for %s.\n\nThis is placeholder demo text generated during import.", i, seed.Title),
			Language:    "ja",
		})
	}

	return chapters
}

func randomTitle() string {
	prefixes := []string{"Crimson", "Azure", "Silent", "Eternal", "Shattered", "Hidden", "Rising", "Lonely", "Iron", "Crystal", "Moonlit", "Frost", "Scarlet", "Emerald", "Storm", "Celestial"}
	nouns := []string{"Blade", "Lantern", "Chronicle", "Whisper", "Path", "Saga", "Promise", "Realm", "Soul", "Bloom", "Vow", "Journey", "Tide", "Oracle", "Harbor", "Legend"}
	suffixes := []string{"of Kyoto", "in Edo", "of the Nine Isles", "from Hokkaido", "of the Azure Sky", "from Silla", "of the Jade Court", "of Liyue", "from Joseon", "of the Bamboo Sea", "of the Silver Dawn", "from Moon Valley"}

	return fmt.Sprintf("%s %s %s", prefixes[rand.Intn(len(prefixes))], nouns[rand.Intn(len(nouns))], suffixes[rand.Intn(len(suffixes))])
}

func randomAltTitle(title string) string {
	fragments := []string{"Monogatari", "Gaiden", "Senki", "Densetsu", "No Yoru", "no Kaze", "no Hoshi", "no Umi"}
	return fmt.Sprintf("%s %s", title, fragments[rand.Intn(len(fragments))])
}

func randomGenres() []string {
	genres := []string{"Action", "Adventure", "Fantasy", "Romance", "Seinen", "Josei", "Drama", "Thriller", "Sci-Fi", "Historical", "Mystery", "Supernatural"}
	selected := make([]string, 0, 4)
	count := rand.Intn(3) + 2 // 2–4 genres

	rand.Shuffle(len(genres), func(i, j int) { genres[i], genres[j] = genres[j], genres[i] })
	for i := 0; i < count && i < len(genres); i++ {
		selected = append(selected, genres[i])
	}
	return selected
}

func randomStatus() string {
	if rand.Intn(100) < 65 {
		return "ongoing"
	}
	return "completed"
}

func randomRating() float64 {
	val := 3.0 + rand.Float64()*2.0
	return math.Round(val*10) / 10
}

func randomViews() int64 {
	return int64(rand.Intn(1_000_000-1_000+1) + 1_000)
}

func randomAuthor() string {
	family := []string{"Aoki", "Kimura", "Choi", "Wei", "Suzuki", "Han", "Nguyen", "Lin", "Yamamoto", "Park", "Zhao", "Fujimoto", "Ito"}
	given := []string{"Haruka", "Ren", "Minseo", "Jia", "Tatsuya", "Eun", "Ling", "Yuto", "Seoyeon", "Hikari", "Ming", "Kaito"}
	return fmt.Sprintf("%s %s", family[rand.Intn(len(family))], given[rand.Intn(len(given))])
}

func randomArtist() string {
	family := []string{"Sato", "Lee", "Chen", "Kang", "Nakano", "Pham", "Tang", "Mori", "Baek", "Ono", "Gao", "Abe"}
	given := []string{"Yuna", "Haru", "Sojin", "Wei", "Kana", "Jiho", "Lan", "Riku", "Mei", "Daisuke", "Eri", "Hyeon"}
	return fmt.Sprintf("%s %s", family[rand.Intn(len(family))], given[rand.Intn(len(given))])
}

func randomDescription(genres []string, status string) string {
	openers := []string{
		"A forgotten prophecy awakens in a distant province.",
		"A wandering swordsman stumbles upon an ancient secret.",
		"In a city of neon lanterns, a quiet rebellion brews.",
		"A royal heir hides among commoners, chasing freedom.",
		"A scholar deciphers a map that should not exist.",
	}
	middles := []string{
		"Allies with clashing ideals must travel together.",
		"A mysterious guild offers uneasy sanctuary.",
		"The spirit world whispers warnings through dreams.",
		"An old mentor returns with unsettling news.",
		"Rival clans watch every move from the shadows.",
	}
	closers := []string{
		"Each choice pulls them deeper into court intrigue.",
		"The road to the capital is paved with betrayal.",
		"Ancient relics reshape the meaning of loyalty.",
		"The border between myth and reality begins to blur.",
		"A final duel threatens to change the realm forever.",
	}

	genreText := strings.Join(genres, ", ")
	return fmt.Sprintf("%s %s %s Genres: %s. Status: %s.", openers[rand.Intn(len(openers))], middles[rand.Intn(len(middles))], closers[rand.Intn(len(closers))], genreText, status)
}

func randomChapterTitle() string {
	fragments := []string{"Moonlit Market", "Silent Citadel", "Broken Oath", "Hidden Shrine", "Frozen River", "Ember Festival", "Thunder Path", "Emerald Gate", "Glass Library", "Azure Requiem", "Shadow Banquet", "Jade Courtyard"}
	return fragments[rand.Intn(len(fragments))]
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if r == ' ' || r == '_' || r == '-' {
			if b.Len() > 0 && !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
