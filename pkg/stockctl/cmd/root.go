package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"stock/pkg/stockctl/config"
)

var (
	cfgFile   string
	serverURL string
	token     string
	cfg       *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "stockctl",
	Short: "Remote CLI for stockd",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return err
		}
		if serverURL != "" {
			cfg.ServerURL = serverURL
		} else if v := os.Getenv("STOCKCTL_SERVER"); v != "" {
			cfg.ServerURL = v
		}
		if token != "" {
			cfg.Token = token
		} else if v := os.Getenv("STOCKCTL_TOKEN"); v != "" {
			cfg.Token = v
		}
		if cfg.ServerURL == "" {
			return fmt.Errorf("server URL required (set --server or STOCKCTL_SERVER or run 'stockctl login')")
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "stockd server URL")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "API bearer token")
}
