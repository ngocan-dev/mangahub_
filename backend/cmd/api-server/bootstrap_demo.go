package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ngocan-dev/mangahub/backend/domain/manga"
	chapterservice "github.com/ngocan-dev/mangahub/backend/internal/service/chapter"
)

type demoMangaSeed struct {
	Title    string
	AltTitle string
	Author   string
	Artist   string
	Status   string
	Synopsis string
	Genres   []string
	Rating   float64
	ViewsMin int64
	ViewsMax int64
	Chapters int
	Language string
}

func bootstrapDemoManga(ctx context.Context, mangaSvc *manga.Service, chapterSvc *chapterservice.Service) error {
	seeds := []demoMangaSeed{
		{
			Title:    "Kogarashi no Tsuki",
			AltTitle: "Moon of Wandering Winds",
			Author:   "Aoki Haruka",
			Artist:   "Sato Yuna",
			Status:   "ongoing",
			Synopsis: "A ronin haunted by a forgotten duel escorts a shrine maiden through provinces filled with spirits. Their pact unravels secrets buried beneath a crescent moon. Old oaths resurface as rival clans close in.",
			Genres:   []string{"Action", "Supernatural", "Historical"},
			Rating:   4.3,
			ViewsMin: 42000,
			ViewsMax: 180000,
			Chapters: 6,
			Language: "ja",
		},
		{
			Title:    "Azure Court Chef",
			AltTitle: "Palace of Sapphire Flames",
			Author:   "Liang Wei",
			Artist:   "Chen Lan",
			Status:   "ongoing",
			Synopsis: "A palace cook with a perfect palate can taste incoming disasters. When the emperor falls ill, her dishes become coded warnings. Assassins and courtiers alike vie for her recipes and her loyalty.",
			Genres:   []string{"Drama", "Slice of Life", "Romance"},
			Rating:   3.9,
			ViewsMin: 38000,
			ViewsMax: 150000,
			Chapters: 5,
			Language: "zh",
		},
		{
			Title:    "Seonbi's Borrowed Blade",
			AltTitle: "Scholar of the Silent Yard",
			Author:   "Han Jisoo",
			Artist:   "Baek Riku",
			Status:   "completed",
			Synopsis: "A Joseon scholar inherits a sword bound to a vengeful spirit. Each duel forces him to trade memories for power. To reclaim his past, he must solve the spirit's unfinished murder.",
			Genres:   []string{"Mystery", "Action", "Historical"},
			Rating:   4.6,
			ViewsMin: 56000,
			ViewsMax: 210000,
			Chapters: 7,
			Language: "ko",
		},
		{
			Title:    "Lanterns of Liyue Harbor",
			AltTitle: "Harbor of Thousand Lights",
			Author:   "Zhao Ming",
			Artist:   "Lin Eri",
			Status:   "ongoing",
			Synopsis: "Two smugglers ferry relics across stormy seas while chasing rumors of a lantern that grants safe passage. Each voyage exposes the lies between them. A naval officer stalks their wake with his own debts to pay.",
			Genres:   []string{"Adventure", "Drama", "Romance"},
			Rating:   4.1,
			ViewsMin: 47000,
			ViewsMax: 160000,
			Chapters: 8,
			Language: "zh",
		},
		{
			Title:    "Snowbound Apothecary",
			AltTitle: "Herbs Under Frost",
			Author:   "Yamamoto Rei",
			Artist:   "Ono Hikari",
			Status:   "completed",
			Synopsis: "An exiled healer opens a shop in a mountain pass where travelers vanish in blizzards. Her remedies reveal the fears her patients hide. The storm's origin may rest in her own forbidden cure.",
			Genres:   []string{"Drama", "Fantasy", "Slice of Life"},
			Rating:   4.5,
			ViewsMin: 52000,
			ViewsMax: 175000,
			Chapters: 4,
			Language: "ja",
		},
		{
			Title:    "Crystal Lotus Syndicate",
			AltTitle: "Gang of Glass Petals",
			Author:   "Nguyen Kaito",
			Artist:   "Tang Mei",
			Status:   "ongoing",
			Synopsis: "A thief with mirrored eyes joins a syndicate that steals memories instead of gold. Each heist leaves cities forgetting their heroes. When her own reflection vanishes, she questions who is erasing whom.",
			Genres:   []string{"Thriller", "Sci-Fi", "Action"},
			Rating:   4.2,
			ViewsMin: 61000,
			ViewsMax: 195000,
			Chapters: 6,
			Language: "vi",
		},
	}

	if len(seeds) == 0 {
		return nil
	}

	for _, seed := range seeds {
		existing, err := mangaSvc.GetByTitle(ctx, seed.Title)
		if err != nil {
			return err
		}
		if existing != nil {
			return nil
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, seed := range seeds {
		slug := slugify(seed.Title)
		views := seed.ViewsMin
		if seed.ViewsMax > seed.ViewsMin {
			views = seed.ViewsMin + rng.Int63n(seed.ViewsMax-seed.ViewsMin+1)
		}

		req := manga.CreateMangaRequest{
			Title:       seed.Title,
			AltTitle:    seed.AltTitle,
			Slug:        slug,
			CoverURL:    "https://cdn.mangahub.demo/covers/" + slug + ".jpg",
			Author:      seed.Author,
			Artist:      seed.Artist,
			Status:      seed.Status,
			Synopsis:    seed.Synopsis,
			Genres:      seed.Genres,
			Rating:      seed.Rating,
			Views:       views,
			Language:    seed.Language,
			LastChapter: seed.Chapters,
		}

		mangaID, err := mangaSvc.CreateManga(ctx, req)
		if err != nil {
			log.Printf("demo bootstrap: failed to create manga %s: %v", seed.Title, err)
			continue
		}
		log.Printf("demo bootstrap: created manga [%d] %s", mangaID, seed.Title)

		for ch := 1; ch <= seed.Chapters; ch++ {
			chapterTitle := buildChapterTitle(ch)
			content := "This is the full content of chapter " + strconv.Itoa(ch) + " for " + seed.Title + ".\n\nIt is auto-generated for demo purposes."
			if _, err := chapterSvc.CreateChapter(ctx, mangaID, ch, chapterTitle, content, seed.Language); err != nil {
				log.Printf("demo bootstrap: failed to create chapter %d for %s: %v", ch, seed.Title, err)
				continue
			}
			log.Printf("demo bootstrap: created chapter %d for %s", ch, seed.Title)
		}
	}

	return nil
}

func buildChapterTitle(number int) string {
	fragments := []string{"Silent Harbor", "Glass Oath", "Lantern Chase", "Frosted Path", "Borrowed Blade", "Moonlit Feast", "Crimson Debt", "Whispered Recipe", "Storm Ledger", "Hidden Shrine"}
	idx := number % len(fragments)
	if idx < 0 {
		idx = 0
	}
	return "Chapter " + strconv.Itoa(number) + ": " + fragments[idx]
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
