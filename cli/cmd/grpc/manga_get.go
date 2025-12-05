package grpc

import "github.com/spf13/cobra"

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
	Example: "mangahub grpc manga get --id 123",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement gRPC manga get
		cmd.Println("gRPC manga get is not yet implemented.")
		return nil
	},
}

func init() {
	GRPCCmd.AddCommand(grpcMangaCmd)
	grpcMangaCmd.AddCommand(mangaGetCmd)
	mangaGetCmd.Flags().String("id", "", "Manga identifier")
}
