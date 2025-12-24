package chapter

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list <novelID>",
	Short:   "List chapters for a novel",
	Long:    "Retrieve a list of chapters for the specified novel.",
	Example: "mangahub chapter list 42",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		novelID := strings.TrimSpace(args[0])
		if novelID == "" {
			return errors.New("novelID cannot be empty")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		info, err := client.GetMangaInfo(cmd.Context(), novelID)
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"chapters": info.Chapters})
			return nil
		}

		if len(info.Chapters) == 0 {
			cmd.Println("No chapters available for this novel yet.")
			return nil
		}

		table := utils.Table{Headers: []string{"ID", "Chapter", "Title", "Updated"}}
		for _, chapter := range info.Chapters {
			updated := ""
			if chapter.UpdatedAt != nil {
				updated = chapter.UpdatedAt.UTC().Format(time.RFC3339)
			}
			table.AddRow(
				fmt.Sprintf("%d", chapter.ID),
				fmt.Sprintf("%d", chapter.Number),
				chapter.Title,
				updated,
			)
		}

		title := fmt.Sprintf("Chapters for %s", formatChapterTitle(info.Title, info.Name))
		cmd.Println(table.RenderWithTitle(title))
		cmd.Println("Use 'mangahub chapter read <chapterID>' to read a chapter")
		return nil
	},
}

func init() {
	ChapterCmd.AddCommand(listCmd)
	output.AddFlag(listCmd)
}

func formatChapterTitle(title, name string) string {
	title = strings.TrimSpace(title)
	name = strings.TrimSpace(name)
	if title != "" {
		return title
	}
	return name
}
