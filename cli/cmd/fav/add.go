package fav

import (
	"errors"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <novelID>",
	Short:   "Add a novel to favorites",
	Long:    "Bookmark a novel by adding it to your favorites list.",
	Example: "mangahub fav add 42",
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
		resp, err := client.AddToLibrary(cmd.Context(), novelID, "plan_to_read", nil)
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, resp)
			return nil
		}

		if config.Runtime().Quiet {
			cmd.Println(novelID)
			return nil
		}

		cmd.Println("✓ Added to favorites.")
		cmd.Printf("Novel ID: %s\n", novelID)
		cmd.Println("Use 'mangahub fav list' to see your favorites.")
		return nil
	},
}

func init() {
	FavCmd.AddCommand(addCmd)
	output.AddFlag(addCmd)
}
