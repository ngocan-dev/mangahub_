package library

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var batchAddCmd = &cobra.Command{
	Use:     "batch-add",
	Short:   "Add multiple manga entries",
	Long:    "Batch add manga entries to your library from a list of IDs or file.",
	Example: "mangahub library batch-add --file manga_ids.txt",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		status, _ := cmd.Flags().GetString("status")

		if strings.TrimSpace(file) == "" {
			return errors.New("--file is required")
		}

		status = strings.ToLower(strings.TrimSpace(status))
		if err := validateStatus(status); err != nil {
			return err
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

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

		ids := make([]string, 0)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			ids = append(ids, line)
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		if len(ids) == 0 {
			cmd.Println("No manga IDs found in file.")
			return nil
		}

		quiet := config.Runtime().Quiet
		if !quiet {
			cmd.Println("Batch adding manga to library...")
			cmd.Printf("Source file: %s\n", file)
			cmd.Printf("Status: %s\n\n", status)
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)

		total := len(ids)
		success := 0
		failed := 0

		for _, id := range ids {
			_, err := client.AddToLibrary(cmd.Context(), id, status, nil)
			if err != nil {
				_ = handleAddError(cmd, err, id)
				if quiet {
					cmd.Printf("%s\n", id)
				}
				failed++
				continue
			}

			if quiet {
				cmd.Println(id)
				success++
				continue
			}

			cmd.Printf("✓ Added: %s\n", id)
			success++
		}

		if !quiet {
			cmd.Println("\nSummary:")
			cmd.Printf("Total: %d\n", total)
			cmd.Printf("Success: %d\n", success)
			cmd.Printf("Failed: %d\n", failed)
			cmd.Println("")
			cmd.Printf("Use 'mangahub library list --status %s' to review.\n", status)
		}

		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(batchAddCmd)
	batchAddCmd.Flags().String("file", "", "File containing manga IDs")
	batchAddCmd.Flags().String("status", "reading", "Reading status (reading|completed|plan-to-read|on-hold|dropped)")
	batchAddCmd.MarkFlagRequired("file")
}
