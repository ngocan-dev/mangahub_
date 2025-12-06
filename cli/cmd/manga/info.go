package manga

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info <id>",
	Short:   "Show manga details",
	Long:    "Retrieve detailed information about a specific manga by ID.",
	Example: "mangahub manga info one-piece",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		info, err := client.GetMangaInfo(cmd.Context(), id)
		if err != nil {
			if apiErr, ok := err.(*api.Error); ok && apiErr.Status == http.StatusNotFound {
				if config.Runtime().Quiet {
					return nil
				}
				cmd.Printf("✗ Manga not found: '%s'\n", id)
				cmd.Println("Try searching instead:")
				cmd.Println("mangahub manga search \"manga title\"")
				return nil
			}
			return err
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, info)
			return nil
		}

		if config.Runtime().Quiet {
			cmd.Println(info.ID)
			return nil
		}

		renderMangaInfo(cmd, info)
		return nil
	},
}

func init() {
	MangaCmd.AddCommand(infoCmd)
}

func renderMangaInfo(cmd *cobra.Command, info *api.MangaInfoResponse) {
	title := strings.ToUpper(info.Title)
	bannerWidth := maxInt(69, utils.DisplayWidth(title))
	top := "┌" + strings.Repeat("─", bannerWidth) + "┐"
	bottom := "└" + strings.Repeat("─", bannerWidth) + "┘"
	padding := bannerWidth - utils.DisplayWidth(title)
	leftPad := padding / 2
	rightPad := padding - leftPad

	cmd.Println(top)
	cmd.Printf("│%s%s%s│\n", strings.Repeat(" ", leftPad), title, strings.Repeat(" ", rightPad))
	cmd.Println(bottom)
	cmd.Println()

	altTitle := ""
	if len(info.AltTitles) > 0 {
		altTitle = info.AltTitles[0]
	}

	cmd.Println("Basic Information:")
	cmd.Printf("ID: %s\n", info.ID)
	titleLine := info.Title
	if altTitle != "" {
		titleLine = fmt.Sprintf("%s (%s)", info.Title, altTitle)
	}
	cmd.Printf("Title: %s\n", titleLine)
	cmd.Printf("Author: %s\n", valueOrNA(info.Author))
	cmd.Printf("Artist: %s\n", valueOrNA(info.Artist))
	cmd.Printf("Genres: %s\n", joinOrNA(info.Genres))
	cmd.Printf("Status: %s\n", formatStatus(info.Status))
	if info.Year > 0 {
		cmd.Printf("Year: %d\n", info.Year)
	} else {
		cmd.Println("Year: N/A")
	}
	cmd.Println()

	cmd.Println("Progress:")
	chapters := formatWithPlus(info.Chapters, info.Status)
	volumes := formatWithPlus(info.Volumes, info.Status)
	cmd.Printf("Total Chapters: %s\n", chapters)
	cmd.Printf("Total Volumes: %s\n", volumes)
	cmd.Printf("Serialization: %s\n", valueOrNA(info.Serialization))
	cmd.Printf("Publisher: %s\n", valueOrNA(info.Publisher))

	library := info.Library
	if library != nil {
		cmd.Printf("Your Status: %s\n", userStatusLabel(library.Status))
		cmd.Printf("Current Chapter: %s\n", formatNumber(library.CurrentChapter))
		cmd.Printf("Last Updated: %s\n", valueOrNA(library.LastUpdated))
		cmd.Printf("Started Reading: %s\n", valueOrNA(library.StartedReading))
		if library.Rating > 0 {
			cmd.Printf("Personal Rating: %d/10\n", library.Rating)
		} else {
			cmd.Println("Personal Rating: -")
		}
	} else {
		cmd.Println("Your Status: Not in library")
		cmd.Println("Current Chapter: -")
		cmd.Println("Last Updated: -")
		cmd.Println("Started Reading: -")
		cmd.Println("Personal Rating: -")
	}
	cmd.Println()

	cmd.Println("Description:")
	for _, line := range wrapText(info.Description, 70) {
		cmd.Println(line)
	}
	if len(info.Description) == 0 {
		cmd.Println("No description available.")
	}
	cmd.Println()

	cmd.Println("External Links:")
	for _, link := range sortedLinks(info.Links) {
		cmd.Println(link)
	}
	if len(info.Links) == 0 {
		cmd.Println("No links available.")
	}
	cmd.Println()

	cmd.Println("Actions:")
	nextChapter := "<chapter>"
	ratingValue := "<1-10>"
	if library != nil {
		if library.CurrentChapter > 0 {
			nextChapter = fmt.Sprintf("%d", library.CurrentChapter+1)
		}
		if library.Rating > 0 {
			ratingValue = fmt.Sprintf("%d", library.Rating)
		}
	}
	cmd.Printf("Update Progress: mangahub progress update --manga-id %s --chapter %s\n", info.ID, nextChapter)
	cmd.Printf("Rate/Review:   mangahub library update --manga-id %s --rating %s\n", info.ID, ratingValue)
	cmd.Printf("Remove:        mangahub library remove --manga-id %s\n", info.ID)
}

func valueOrNA(value string) string {
	if strings.TrimSpace(value) == "" {
		return "N/A"
	}
	return value
}

func joinOrNA(values []string) string {
	if len(values) == 0 {
		return "N/A"
	}
	return strings.Join(values, ", ")
}

func formatWithPlus(value int, status string) string {
	if value <= 0 {
		return "-"
	}
	formatted := formatNumber(value)
	if strings.ToLower(status) == "ongoing" {
		formatted += "+"
	}
	return formatted
}

func formatNumber(value int) string {
	if value <= 0 {
		return "-"
	}
	s := fmt.Sprintf("%d", value)
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

func wrapText(text string, width int) []string {
	if text == "" {
		return nil
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}
	var lines []string
	var current strings.Builder
	for _, word := range words {
		if current.Len() == 0 {
			current.WriteString(word)
			continue
		}
		if utils.DisplayWidth(current.String())+1+utils.DisplayWidth(word) > width {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
			continue
		}
		current.WriteString(" ")
		current.WriteString(word)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

func sortedLinks(links map[string]string) []string {
	if len(links) == 0 {
		return nil
	}
	keys := make([]string, 0, len(links))
	for k := range links {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var formatted []string
	for _, key := range keys {
		name := linkLabel(key)
		formatted = append(formatted, fmt.Sprintf("%s: %s", name, links[key]))
	}
	return formatted
}

func linkLabel(key string) string {
	switch strings.ToLower(key) {
	case "mal", "myanimelist":
		return "MyAnimeList"
	case "mangadx":
		return "MangaDx"
	default:
		return strings.Title(key)
	}
}

func userStatusLabel(status string) string {
	switch strings.ToLower(status) {
	case "reading", "currently reading":
		return "Currently Reading"
	case "completed":
		return "Completed"
	case "planned", "plan to read":
		return "Plan to Read"
	default:
		return formatStatus(status)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
