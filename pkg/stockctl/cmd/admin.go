package cmd

import (
	"github.com/spf13/cobra"

	"stock/pkg/stockctl/client"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin operations",
}

var adminUserCreateCmd = &cobra.Command{
	Use:   "user create [username]",
	Short: "Create a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		role, _ := cmd.Flags().GetString("role")
		password, _ := cmd.Flags().GetString("password")
		return c.POST("/api/admin/users", map[string]string{
			"username": args[0],
			"password": password,
			"role":     role,
		}, nil)
	},
}

var adminSyncBarsCmd = &cobra.Command{
	Use:   "sync bars",
	Short: "Trigger daily bar sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		return c.POST("/api/admin/bars/sync", nil, nil)
	},
}

var adminSyncStocklistCmd = &cobra.Command{
	Use:   "sync stocklist",
	Short: "Trigger stocklist sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		return c.POST("/api/admin/stocks/sync", nil, nil)
	},
}

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminUserCreateCmd, adminSyncBarsCmd, adminSyncStocklistCmd)
	adminUserCreateCmd.Flags().String("role", "user", "Role: user|admin")
	adminUserCreateCmd.Flags().String("password", "", "Initial password (required)")
	_ = adminUserCreateCmd.MarkFlagRequired("password")
}
