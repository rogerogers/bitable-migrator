package cmd

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	appID     string
	appSecret string
)

var rootCmd = &cobra.Command{
	Use:          "bitable-migrator",
	Short:        "Declarative Schema-as-Code migration tool for Feishu/Lark Bitable",
	Long: `Bitable Migrator is a lightweight, declarative, and high-performance schema migration tool 
for Lark/Feishu Bitable (多维表格) written in Go, managing structures safely as code.`,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// 1. Initialize dotenv fallback
	_ = godotenv.Load()

	// 2. Initialize Viper environments
	viper.AutomaticEnv()
	_ = viper.BindEnv("app_id", "LARK_APP_ID")
	_ = viper.BindEnv("app_secret", "LARK_APP_SECRET")

	// 3. Define Global Flags
	rootCmd.PersistentFlags().StringVar(&appID, "app-id", "", "Lark App ID (falls back to LARK_APP_ID env)")
	rootCmd.PersistentFlags().StringVar(&appSecret, "app-secret", "", "Lark App Secret (falls back to LARK_APP_SECRET env)")

	// 4. Bind Global Flags to Viper
	_ = viper.BindPFlag("app_id", rootCmd.PersistentFlags().Lookup("app-id"))
	_ = viper.BindPFlag("app_secret", rootCmd.PersistentFlags().Lookup("app-secret"))
}
