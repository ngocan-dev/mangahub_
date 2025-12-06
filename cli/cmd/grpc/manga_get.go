package grpc

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/grpcclient"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

// GRPCCmd groups gRPC commands.
var GRPCCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Interact with MangaHub via gRPC",
	Long:  "Use gRPC clients to retrieve manga data and update progress.",
}

// grpcMangaCmd groups manga-specific gRPC operations.
var grpcMangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Manga operations over gRPC",
	Long:  "Perform manga retrieval and search over the gRPC API.",
}

var mangaGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get manga via gRPC",
	Long:    "Fetch manga details through the gRPC API using an identifier.",
	Example: "mangahub grpc manga get --id one-piece",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("--id is required")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client, err := grpcclient.New(cmd.Context(), cfg)
		if err != nil {
			return fmt.Errorf("✗ gRPC connection error: %w", err)
		}
		defer client.Close()

		resp, err := client.Manga().GetManga(cmd.Context(), &grpcclient.GetMangaRequest{Id: id})
		if err != nil {
			return fmt.Errorf("✗ gRPC error: %v", err)
		}

		output.PrintJSON(cmd, resp)

		if config.Runtime().Quiet {
			cmd.Println(resp.Id)
			return nil
		}

		cmd.Println("gRPC: Fetching manga details...")
		cmd.Println("")
		cmd.Printf("ID: %s\n", resp.Id)
		cmd.Printf("Title: %s (%s)\n", resp.Title, resp.OriginalTitle)
		cmd.Printf("Author: %s\n", resp.Author)
		cmd.Printf("Status: %s\n", resp.Status)
		cmd.Printf("Chapters: %s\n", resp.Chapters)
		if len(resp.Genres) > 0 {
			cmd.Printf("Genres: %s\n", strings.Join(resp.Genres, ", "))
		}
		cmd.Println("")
		cmd.Println("To see full CLI view:")
		cmd.Printf("mangahub manga info %s\n", resp.Id)
		return nil
	},
}

func init() {
	GRPCCmd.AddCommand(grpcMangaCmd)
	grpcMangaCmd.AddCommand(mangaGetCmd)
	mangaGetCmd.Flags().String("id", "", "Manga identifier")
	mangaGetCmd.MarkFlagRequired("id")
}
