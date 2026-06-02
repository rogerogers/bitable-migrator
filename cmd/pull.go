package cmd

import (
	"fmt"

	"github.com/rogerogers/bitable-migrator/pkg/migrator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	pullAppToken   string
	pullTableID    string
	pullOutputPath string
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Generate a new local YAML schema from an existing online Lark Bitable table",
	Long:  "Fetches the structure of an existing online Feishu Bitable table and exports it as a clean local YAML configuration template.",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedAppID := viper.GetString("app_id")
		resolvedAppSecret := viper.GetString("app_secret")

		if resolvedAppID == "" || resolvedAppSecret == "" {
			return fmt.Errorf("Lark App ID and App Secret must be provided via flags (--app-id, --app-secret) or env variables (LARK_APP_ID, LARK_APP_SECRET)")
		}

		migr := migrator.NewMigrator(resolvedAppID, resolvedAppSecret)
		fmt.Printf("Pulling Bitable schema into '%s'...\n", pullOutputPath)
		err := migr.Pull(pullAppToken, pullTableID, pullOutputPath)
		if err != nil {
			return err
		}
		fmt.Println("Successfully completed pull!")
		return nil
	},
}

func init() {
	pullCmd.Flags().StringVar(&pullAppToken, "app", "", "Feishu Bitable App Token (Required)")
	pullCmd.Flags().StringVar(&pullTableID, "table", "", "Feishu Table ID (Required)")
	pullCmd.Flags().StringVarP(&pullOutputPath, "output", "o", "./bitable.yaml", "Destination path for generated yaml schema")

	_ = pullCmd.MarkFlagRequired("app")
	_ = pullCmd.MarkFlagRequired("table")

	rootCmd.AddCommand(pullCmd)
}
