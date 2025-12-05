package cmd

import (
	"fmt"
	"os"

	"github.com/ngocan-dev/mangahub_/cli/cmd/auth"
	"github.com/ngocan-dev/mangahub_/cli/cmd/backup"
	"github.com/ngocan-dev/mangahub_/cli/cmd/chat"
	configcmd "github.com/ngocan-dev/mangahub_/cli/cmd/config"
	"github.com/ngocan-dev/mangahub_/cli/cmd/db"
	"github.com/ngocan-dev/mangahub_/cli/cmd/export"
	"github.com/ngocan-dev/mangahub_/cli/cmd/grpc"
	"github.com/ngocan-dev/mangahub_/cli/cmd/library"
	"github.com/ngocan-dev/mangahub_/cli/cmd/manga"
	"github.com/ngocan-dev/mangahub_/cli/cmd/notify"
	"github.com/ngocan-dev/mangahub_/cli/cmd/profile"
	"github.com/ngocan-dev/mangahub_/cli/cmd/progress"
	"github.com/ngocan-dev/mangahub_/cli/cmd/server"
	"github.com/ngocan-dev/mangahub_/cli/cmd/stats"
	syncCmd "github.com/ngocan-dev/mangahub_/cli/cmd/sync"
	"github.com/ngocan-dev/mangahub_/cli/cmd/update"
	"github.com/ngocan-dev/mangahub_/cli/config"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "mangahub",
	Short: "MangaHub CLI for managing manga libraries and progress",
	Long:  "MangaHub CLI provides commands to manage manga, libraries, synchronization, and more.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.mangahub/config.yaml)")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return config.LoadConfig(cfgFile)
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(server.ServerCmd)
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(manga.MangaCmd)
	rootCmd.AddCommand(library.LibraryCmd)
	rootCmd.AddCommand(progress.ProgressCmd)
	rootCmd.AddCommand(syncCmd.SyncCmd)
	rootCmd.AddCommand(notify.NotifyCmd)
	rootCmd.AddCommand(grpc.GRPCCmd)
	rootCmd.AddCommand(chat.ChatCmd)
	rootCmd.AddCommand(stats.StatsCmd)
	rootCmd.AddCommand(export.ExportCmd)
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(db.DBCmd)
	rootCmd.AddCommand(configcmd.ConfigCmd)
	rootCmd.AddCommand(profile.ProfileCmd)
	rootCmd.AddCommand(update.UpdateCmd)
}
