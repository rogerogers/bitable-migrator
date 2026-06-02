package cmd

import (
	"fmt"

	"github.com/rogerogers/bitable-migrator/pkg/migrator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var diffConfigPath string

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Preview schema differences without applying changes (Dry Run)",
	Long:  "Compares local YAML configuration with online Bitable and prints planned modifications without modifying anything online or locally.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSyncOrDiff(diffConfigPath, true)
	},
}

func init() {
	diffCmd.Flags().StringVarP(&diffConfigPath, "config", "c", "./bitable.yaml", "Path to the bitable.yaml config file")
	rootCmd.AddCommand(diffCmd)
}

func runSyncOrDiff(config string, dryRun bool) error {
	resolvedAppID := viper.GetString("app_id")
	resolvedAppSecret := viper.GetString("app_secret")

	if resolvedAppID == "" || resolvedAppSecret == "" {
		return fmt.Errorf("Lark App ID and App Secret must be provided via flags (--app-id, --app-secret) or env variables (LARK_APP_ID, LARK_APP_SECRET)")
	}

	migr := migrator.NewMigrator(resolvedAppID, resolvedAppSecret)
	action := "sync"
	if dryRun {
		action = "diff"
	}
	fmt.Printf("Starting %s using config '%s'...\n", action, config)
	err := migr.Sync(config, dryRun)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully completed %s!\n", action)
	return nil
}
