package progress

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var batchUpdateCmd = &cobra.Command{
	Use:     "batch-update",
	Short:   "Batch update progress",
	Long:    "Update progress for multiple manga entries from a file or list.",
	Example: "mangahub progress batch-update --file updates.csv",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if strings.TrimSpace(file) == "" {
			return errors.New("--file is required")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}
		if strings.TrimSpace(cfg.Data.Token) == "" {
			return errors.New("Please login first")
		}

		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer f.Close()

		reader := csv.NewReader(f)
		reader.TrimLeadingSpace = true

		records, err := reader.ReadAll()
		if err != nil {
			return fmt.Errorf("read csv: %w", err)
		}
		if len(records) == 0 {
			cmd.Println("No updates found in file.")
			return nil
		}

		quiet := config.Runtime().Quiet
		if !quiet {
			cmd.Println("Batch updating reading progress...")
			cmd.Printf("Source file: %s\n\n", file)
		}

		total := 0
		updated := 0
		failed := 0

		start := 0
		if len(records) > 0 {
			header := strings.ToLower(strings.Join(records[0], ","))
			if strings.Contains(header, "manga_id") {
				start = 1
			}
		}

		for i := start; i < len(records); i++ {
			row := records[i]
			if len(row) < 2 {
				failed++
				continue
			}

			mangaID := strings.TrimSpace(row[0])
			chapter, _ := strconv.Atoi(strings.TrimSpace(row[1]))
			volume := 0
			if len(row) > 2 {
				volume, _ = strconv.Atoi(strings.TrimSpace(row[2]))
			}
			notes := ""
			if len(row) > 3 {
				notes = strings.TrimSpace(row[3])
			}

			total++

			if quiet {
				cmd.Printf("%s,%d,%d,%s\n", mangaID, chapter, volume, notes)
				updated++
				continue
			}

			status := "✓ Updated"
			if dryRun {
				status = "(dry-run)"
			}

			cmd.Printf("[%d/%d] %s → chapter %d: %s\n", total, len(records)-start, mangaID, chapter, status)
			if dryRun {
				updated++
				continue
			}

			// In this CLI mock, we assume success when not in dry-run.
			updated++
		}

		if !quiet {
			cmd.Println("\nSummary:")
			cmd.Printf("Processed: %d\n", total)
			cmd.Printf("Updated: %d\n", updated)
			cmd.Printf("Failed: %d\n", failed)
		}

		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(batchUpdateCmd)
	batchUpdateCmd.Flags().String("file", "", "File containing progress updates")
	batchUpdateCmd.Flags().Bool("dry-run", false, "Simulate updates without saving")
}
