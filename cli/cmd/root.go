package cmd

import (
	"fmt"
	"os"

	"github.com/ngocan-dev/mangahub_/cli/cmd/auth"
	grpcCmd "github.com/ngocan-dev/mangahub_/cli/cmd/grpc"
	"github.com/ngocan-dev/mangahub_/cli/cmd/library"
	"github.com/ngocan-dev/mangahub_/cli/cmd/manga"
	notifycmd "github.com/ngocan-dev/mangahub_/cli/cmd/notify"
	"github.com/ngocan-dev/mangahub_/cli/cmd/progress"
	"github.com/ngocan-dev/mangahub_/cli/cmd/server"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	quiet   bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "mangahub",
	Short: "MangaHub CLI for managing manga libraries and progress",
	Long:  "MangaHub CLI provides commands to register, login, search manga, manage libraries, and track progress.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	defaultPath, _ := config.DefaultPath()
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultPath, "Custom config path (default: ~/.mangahub/config.json)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logs")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress all non-error output")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		config.SetRuntimeOptions(verbose, quiet)
		_, err := config.Load(cfgFile)
		return err
	}

	rootCmd.AddCommand(server.ServerCmd)
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(notifycmd.NotifyCmd)
	rootCmd.AddCommand(grpcCmd.GRPCCmd)
	rootCmd.AddCommand(manga.MangaCmd)
	rootCmd.AddCommand(library.LibraryCmd)
	rootCmd.AddCommand(progress.ProgressCmd)
}
