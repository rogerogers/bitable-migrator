package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rogerogers/bitable-migrator/pkg/migrator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tablesAppToken string
)

var tablesCmd = &cobra.Command{
	Use:   "tables",
	Short: "List all tables in a Lark Bitable",
	Long:  "Fetches and lists all data tables in an existing online Feishu Bitable, displaying their ID, Name, and Revision.",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedAppID := viper.GetString("app_id")
		resolvedAppSecret := viper.GetString("app_secret")

		if resolvedAppID == "" || resolvedAppSecret == "" {
			return fmt.Errorf("Lark App ID and App Secret must be provided via flags (--app-id, --app-secret) or env variables (LARK_APP_ID, LARK_APP_SECRET)")
		}

		migr := migrator.NewMigrator(resolvedAppID, resolvedAppSecret)
		fmt.Printf("Fetching tables for Bitable App '%s'...\n", tablesAppToken)

		tables, err := migr.FetchOnlineTables(tablesAppToken)
		if err != nil {
			return err
		}

		if len(tables) == 0 {
			fmt.Println("No tables found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TABLE ID\tNAME\tREVISION")

		for _, t := range tables {
			var id, name, rev string
			if t.TableId != nil {
				id = *t.TableId
			}
			if t.Name != nil {
				name = *t.Name
			}
			if t.Revision != nil {
				rev = fmt.Sprintf("%d", *t.Revision)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", id, name, rev)
		}
		w.Flush()

		return nil
	},
}

func init() {
	tablesCmd.Flags().StringVar(&tablesAppToken, "app", "", "Feishu Bitable App Token (Required)")
	_ = tablesCmd.MarkFlagRequired("app")

	rootCmd.AddCommand(tablesCmd)
}
