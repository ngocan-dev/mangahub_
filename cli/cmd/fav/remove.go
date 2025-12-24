package fav

import (
	"errors"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <novelID>",
	Short:   "Remove a novel from favorites",
	Long:    "Remove a novel from your favorites list.",
	Example: "mangahub fav remove 42",
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
		if err := client.RemoveFromLibrary(cmd.Context(), novelID); err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"removed": novelID})
			return nil
		}

		if config.Runtime().Quiet {
			cmd.Println(novelID)
			return nil
		}

		cmd.Println("✓ Removed from favorites.")
		cmd.Printf("Novel ID: %s\n", novelID)
		return nil
	},
}

func init() {
	FavCmd.AddCommand(removeCmd)
	output.AddFlag(removeCmd)
}
