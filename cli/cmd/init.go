package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

// initCmd initializes local MangaHub directories and files.
var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize MangaHub configuration and data directories",
	Long:    "Create the ~/.mangahub directory with configuration, database, and log placeholders for MangaHub.",
	Example: `mangahub init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, err := config.ConfigDir()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			return err
		}

		cfgPath, err := config.DefaultPath()
		if err != nil {
			return err
		}
		if _, err := config.LoadWithOptions(config.LoadOptions{Path: cfgPath}); err != nil {
			return err
		}

		dataPath := filepath.Join(cfgDir, "data.db")
		if _, err := os.Stat(dataPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if file, createErr := os.Create(dataPath); createErr == nil {
				defer file.Close()
			} else {
				return createErr
			}
		}

		logsDir := filepath.Join(cfgDir, "logs")
		if err := os.MkdirAll(logsDir, 0o755); err != nil {
			return err
		}

		cmd.Println("MangaHub initialized at", cfgDir)
		cmd.Println("Config file:", cfgPath)
		return nil
	},
}
