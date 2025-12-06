package grpc

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/grpcclient"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var grpcProgressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Progress operations over gRPC",
	Long:  "Update reading progress through the gRPC API.",
}

var progressUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update progress via gRPC",
	Long:    "Send reading progress updates over gRPC.",
	Example: "mangahub grpc progress update --manga-id one-piece --chapter 1095",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")

		if strings.TrimSpace(mangaID) == "" {
			return fmt.Errorf("--manga-id is required")
		}
		if chapter <= 0 {
			return fmt.Errorf("--chapter must be greater than 0")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}
		if strings.TrimSpace(cfg.Data.Token) == "" {
			return fmt.Errorf("✗ You must be logged in to update progress via gRPC. Please login first.")
		}

		client, err := grpcclient.New(cmd.Context(), cfg)
		if err != nil {
			return fmt.Errorf("✗ gRPC connection error: %w", err)
		}
		defer client.Close()

		resp, err := client.Progress().UpdateProgress(cmd.Context(), &grpcclient.UpdateProgressRequest{
			MangaID: mangaID,
			Chapter: chapter,
			Token:   cfg.Data.Token,
		})
		if err != nil {
			return fmt.Errorf("✗ gRPC: Progress update failed: %v", err)
		}

		output.PrintJSON(cmd, resp)
		if config.Runtime().Quiet {
			cmd.Printf("%s %d\n", mangaID, chapter)
			return nil
		}

		cmd.Println("gRPC: Updating reading progress...")
		cmd.Println("")
		cmd.Println("✓ Progress updated via gRPC!")
		cmd.Printf("Manga: %s\n", resp.MangaTitle)
		cmd.Printf("Current chapter: %,d\n", resp.Chapter)
		return nil
	},
}

func init() {
	GRPCCmd.AddCommand(grpcProgressCmd)
	grpcProgressCmd.AddCommand(progressUpdateCmd)
	progressUpdateCmd.Flags().String("manga-id", "", "Manga identifier")
	progressUpdateCmd.Flags().Int("chapter", 0, "Current chapter")
	progressUpdateCmd.MarkFlagRequired("manga-id")
	progressUpdateCmd.MarkFlagRequired("chapter")
}
