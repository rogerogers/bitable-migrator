package cmd

import (
	"github.com/spf13/cobra"
)

var syncConfigPath string

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize local YAML schema changes to Lark",
	Long:  "Compares local YAML configuration with online Bitable, applies additions/updates, and writes back allocated field IDs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSyncOrDiff(syncConfigPath, false)
	},
}

func init() {
	syncCmd.Flags().StringVarP(&syncConfigPath, "config", "c", "./bitable.yaml", "Path to the bitable.yaml config file")
	rootCmd.AddCommand(syncCmd)
}
