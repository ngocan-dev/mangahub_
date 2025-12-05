package grpc

import "github.com/spf13/cobra"

var mangaSearchCmd = &cobra.Command{
	Use:     "search",
	Short:   "Search manga via gRPC",
	Long:    "Search for manga titles using the gRPC API.",
	Example: "mangahub grpc manga search --query 'One Piece'",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement gRPC manga search
		cmd.Println("gRPC manga search is not yet implemented.")
		return nil
	},
}

func init() {
	grpcMangaCmd.AddCommand(mangaSearchCmd)
	mangaSearchCmd.Flags().String("query", "", "Search query")
}
