package grpc

import "github.com/spf13/cobra"

var grpcProgressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Progress operations over gRPC",
	Long:  "Update reading progress through the gRPC API.",
}

var progressUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update progress via gRPC",
	Long:    "Send reading progress updates over gRPC.",
	Example: "mangahub grpc progress update --id 123 --chapter 5",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement gRPC progress update
		cmd.Println("gRPC progress update is not yet implemented.")
		return nil
	},
}

func init() {
	GRPCCmd.AddCommand(grpcProgressCmd)
	grpcProgressCmd.AddCommand(progressUpdateCmd)
	progressUpdateCmd.Flags().String("id", "", "Manga identifier")
	progressUpdateCmd.Flags().Int("chapter", 0, "Current chapter")
}
