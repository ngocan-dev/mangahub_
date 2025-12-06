package stats

import (
	"fmt"

	"github.com/spf13/cobra"
)

func runDateRange(cmd *cobra.Command, args []string) error {
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")

	if from != "" || to != "" {
		if from == "" || to == "" {
			return fmt.Errorf("both --from and --to must be provided")
		}
		cmd.Println(fmt.Sprintf("Reading Statistics (%s → %s)", from, to))
		cmd.Println()
		cmd.Println("Total Chapters Read: 5,903")
		cmd.Println("Active Days: 182")
		cmd.Println("Average per Active Day: 32.4")
		cmd.Println("Most-Read Manga:")
		cmd.Println("  • One Piece — 540 chapters")
		cmd.Println("  • Jujutsu Kaisen — 220 chapters")
		cmd.Println("  • Chainsaw Man — 140 chapters")
		cmd.Println()
		cmd.Println("Reading Heatmap:")
		cmd.Println("[Generate a 7x53 weekly reading visual or simplified summary]")
		return nil
	}

	return cmd.Help()
}

func init() {
	StatsCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	StatsCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
}
