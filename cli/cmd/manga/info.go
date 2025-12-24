package manga

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	Example: "mangahub manga info 42",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

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

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"result": info})
			return nil
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
	output.AddFlag(infoCmd)
}

func renderMangaInfo(cmd *cobra.Command, info *api.MangaInfoResponse) {
	titleText := info.Title
	if strings.TrimSpace(titleText) == "" {
		titleText = info.Name
	}
	title := strings.ToUpper(titleText)
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

	cmd.Println("Basic Information:")
	cmd.Printf("ID: %d\n", info.ID)
	cmd.Printf("Title: %s\n", valueOrNA(titleText))
	cmd.Printf("Author: %s\n", valueOrNA(info.Author))
	cmd.Printf("Genres: %s\n", joinOrNA(splitGenres(info.Genre)))
	cmd.Printf("Status: %s\n", formatStatus(info.Status))
	cmd.Println()

	cmd.Println("Availability:")
	chapters := formatWithPlus(info.ChapterCount, info.Status)
	cmd.Printf("Total Chapters: %s\n", chapters)

	library := info.Library
	progress := info.Progress
	cmd.Println()
	cmd.Println("Your Progress:")
	if library != nil {
		cmd.Printf("Your Status: %s\n", userStatusLabel(library.Status))
		if progress != nil {
			cmd.Printf("Current Chapter: %s\n", formatNumber(progress.CurrentChapter))
			cmd.Printf("Last Read: %s\n", progress.LastReadAt.UTC().Format("2006-01-02 15:04:05 MST"))
		} else {
			cmd.Println("Current Chapter: -")
			cmd.Println("Last Read: -")
		}
		cmd.Printf("Started Reading: %s\n", formatOptionalTime(library.StartedAt))
		cmd.Printf("Completed At: %s\n", formatOptionalTime(library.CompletedAt))
	} else {
		cmd.Println("Your Status: Not in library")
		cmd.Println("Current Chapter: -")
		cmd.Println("Last Read: -")
		cmd.Println("Started Reading: -")
		cmd.Println("Completed At: -")
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

	cmd.Println("Actions:")
	if library == nil {
		cmd.Printf("Add to Library: mangahub library add --manga-id %d --status reading\n", info.ID)
		cmd.Printf("Update Progress: mangahub progress update --manga-id %d --chapter <chapter>\n", info.ID)
	} else {
		nextChapter := "<chapter>"
		if progress != nil && progress.CurrentChapter > 0 {
			nextChapter = fmt.Sprintf("%d", progress.CurrentChapter+1)
		}
		cmd.Printf("Update Progress: mangahub progress update --manga-id %d --chapter %s\n", info.ID, nextChapter)
		cmd.Printf("Change Status:  mangahub library update --manga-id %d --status completed\n", info.ID)
		cmd.Printf("Remove:         mangahub library remove --manga-id %d\n", info.ID)
	}
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

func splitGenres(genre string) []string {
	genre = strings.TrimSpace(genre)
	if genre == "" {
		return nil
	}
	parts := strings.Split(genre, ",")
	var cleaned []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.UTC().Format("2006-01-02 15:04:05 MST")
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
