package manga

import (
	"errors"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

// MangaCmd groups manga-related commands.
var MangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Interact with manga metadata",
	Long:  "Search and retrieve manga information from MangaHub services.",
}

// searchCmd performs manga searches against the backend service.
var searchCmd = &cobra.Command{
	Use:     "search <query>",
	Short:   "Search for manga titles",
	Long:    "Search MangaHub for manga titles using keywords.",
	Example: "mangahub manga search \"one piece\"",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}
		if cfg.Data.Token == "" {
			return errors.New("Please login first")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		results, err := client.SearchManga(cmd.Context(), query)
		if err != nil {
			return err
		}

		if len(results) == 0 && !config.Runtime().Quiet {
			cmd.Println("No manga found")
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, results)
			return nil
		}

		if config.Runtime().Quiet {
			return nil
		}

		for _, m := range results {
			cmd.Printf("- %s: %s (%s)\n", m.ID, m.Title, m.Status)
		}
		return nil
	},
}

func init() {
	MangaCmd.AddCommand(searchCmd)
}
