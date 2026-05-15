package cmd

import (
	"fmt"
	"stock/pkg/models"

	"github.com/spf13/cobra"

	"stock/pkg/cli/client"
)

var portfolioCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "Manage your portfolio",
}

var portfolioListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked stocks",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		var res []*models.Portfolio
		if err := c.GET("/api/portfolio", &res); err != nil {
			return err
		}
		for _, p := range res {
			fmt.Printf("%s\t%s\n", p.TsCode, p.Note)
		}
		return nil
	},
}

var portfolioAddCmd = &cobra.Command{
	Use:   "add [ts_code]",
	Short: "Add a stock to portfolio",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		note, _ := cmd.Flags().GetString("note")
		return c.POST("/api/portfolio",
			&models.PortfolioReq{Note: note, TsCode: args[0]}, nil)
	},
}

var portfolioRmCmd = &cobra.Command{
	Use:   "rm [ts_code]",
	Short: "Remove a stock from portfolio",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		return c.DELETE("/api/portfolio/" + args[0])
	},
}

func init() {
	rootCmd.AddCommand(portfolioCmd)
	portfolioCmd.AddCommand(portfolioListCmd, portfolioAddCmd, portfolioRmCmd)
	portfolioAddCmd.Flags().String("note", "", "Optional note")
}
