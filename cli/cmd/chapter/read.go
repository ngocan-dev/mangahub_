package chapter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:     "read <chapterID>",
	Short:   "Read a chapter",
	Long:    "Retrieve the full content of a chapter by its identifier.",
	Example: "mangahub chapter read 1001",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		chapterID, err := strconv.ParseInt(strings.TrimSpace(args[0]), 10, 64)
		if err != nil || chapterID <= 0 {
			return errors.New("chapterID must be a valid number")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		chapter, err := client.GetChapter(cmd.Context(), chapterID)
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, chapter)
			return nil
		}

		if config.Runtime().Quiet {
			cmd.Println(chapter.Content)
			return nil
		}

		cmd.Printf("Chapter %d\n", chapter.Number)
		if strings.TrimSpace(chapter.Title) != "" {
			cmd.Printf("Title: %s\n", chapter.Title)
		}
		cmd.Printf("Chapter ID: %d\n", chapter.ID)
		cmd.Println("\nContent:")
		cmd.Println(chapter.Content)
		return nil
	},
}

func init() {
	ChapterCmd.AddCommand(readCmd)
	output.AddFlag(readCmd)
}
