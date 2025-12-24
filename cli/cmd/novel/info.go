package novel

import (
	"errors"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info <novelID>",
	Short:   "Show novel details",
	Long:    "Retrieve detailed information about a specific novel by ID.",
	Example: "mangahub novel info 42",
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
			output.PrintJSON(cmd, info)
			return nil
		}

		if config.Runtime().Quiet {
			cmd.Println(info.ID)
			return nil
		}

		cmd.Println("Novel Details")
		cmd.Println("--------------")
		cmd.Printf("ID: %d\n", info.ID)
		cmd.Printf("Title: %s\n", fallbackString(info.Title, info.Name))
		cmd.Printf("Author: %s\n", info.Author)
		cmd.Printf("Status: %s\n", formatNovelStatus(info.Status))
		cmd.Printf("Genre: %s\n", info.Genre)
		cmd.Printf("Rating: %.1f\n", info.RatingPoint)
		cmd.Printf("Chapters: %d\n", info.ChapterCount)
		if strings.TrimSpace(info.Description) != "" {
			cmd.Println("\nDescription:")
			cmd.Println(info.Description)
		}
		cmd.Println("\nNext steps:")
		cmd.Printf("- View chapters: mangahub chapter list %d\n", info.ID)
		return nil
	},
}

func init() {
	NovelCmd.AddCommand(infoCmd)
	output.AddFlag(infoCmd)
}

func fallbackString(primary, secondary string) string {
	primary = strings.TrimSpace(primary)
	secondary = strings.TrimSpace(secondary)
	if primary != "" {
		return primary
	}
	return secondary
}
