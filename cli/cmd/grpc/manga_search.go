package grpc

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/grpcclient"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var mangaSearchCmd = &cobra.Command{
	Use:     "search",
	Short:   "Search manga via gRPC",
	Long:    "Search for manga titles using the gRPC API.",
	Example: "mangahub grpc manga search --query 'One Piece'",
	RunE: func(cmd *cobra.Command, args []string) error {
		query, _ := cmd.Flags().GetString("query")
		if strings.TrimSpace(query) == "" {
			return fmt.Errorf("--query is required")
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

		resp, err := client.Manga().SearchManga(cmd.Context(), &grpcclient.SearchMangaRequest{Query: query})
		if err != nil {
			return fmt.Errorf("✗ gRPC error: %v", err)
		}

		output.PrintJSON(cmd, resp)
		if config.Runtime().Quiet {
			for _, r := range resp.Results {
				cmd.Println(r.Id)
			}
			return nil
		}

		cmd.Printf("gRPC: Searching for \"%s\"...\n", query)
		cmd.Println("")

		if len(resp.Results) == 0 {
			cmd.Println("No results found.")
			return nil
		}

		cmd.Printf("Found %d results:\n", len(resp.Results))
		for _, r := range resp.Results {
			cmd.Printf("- %-20s | %-30s | %-10s | %s chapters\n", r.Id, r.Title, r.Status, r.Chapters)
		}
		return nil
	},
}

func init() {
	grpcMangaCmd.AddCommand(mangaSearchCmd)
	mangaSearchCmd.Flags().String("query", "", "Search query")
	mangaSearchCmd.MarkFlagRequired("query")
}
