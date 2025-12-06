package cmd

import (
	"fmt"
	"os"

	"github.com/ngocan-dev/mangahub_/cli/cmd/auth"
	"github.com/ngocan-dev/mangahub_/cli/cmd/backup"
	chatcmd "github.com/ngocan-dev/mangahub_/cli/cmd/chat"
	configcmd "github.com/ngocan-dev/mangahub_/cli/cmd/config"
	"github.com/ngocan-dev/mangahub_/cli/cmd/db"
	"github.com/ngocan-dev/mangahub_/cli/cmd/export"
	grpcCmd "github.com/ngocan-dev/mangahub_/cli/cmd/grpc"
	"github.com/ngocan-dev/mangahub_/cli/cmd/library"
	"github.com/ngocan-dev/mangahub_/cli/cmd/logs"
	"github.com/ngocan-dev/mangahub_/cli/cmd/manga"
	notifycmd "github.com/ngocan-dev/mangahub_/cli/cmd/notify"
	profilecmd "github.com/ngocan-dev/mangahub_/cli/cmd/profile"
	"github.com/ngocan-dev/mangahub_/cli/cmd/progress"
	"github.com/ngocan-dev/mangahub_/cli/cmd/server"
	"github.com/ngocan-dev/mangahub_/cli/cmd/stats"
	"github.com/ngocan-dev/mangahub_/cli/cmd/sync"
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
	rootCmd.AddCommand(export.ExportCmd)
	rootCmd.AddCommand(grpcCmd.GRPCCmd)
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(db.DBCmd)
	rootCmd.AddCommand(manga.MangaCmd)
	rootCmd.AddCommand(library.LibraryCmd)
	rootCmd.AddCommand(progress.ProgressCmd)
	rootCmd.AddCommand(chatcmd.ChatCmd)
	rootCmd.AddCommand(sync.SyncCmd)
	rootCmd.AddCommand(stats.StatsCmd)
	rootCmd.AddCommand(logs.LogsCmd)
	rootCmd.AddCommand(configcmd.ConfigCmd)
	rootCmd.AddCommand(profilecmd.ProfileCmd)
}
