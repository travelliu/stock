package cmd

import (
	"github.com/spf13/cobra"

	"stock/pkg/cli/client"
	"stock/pkg/cli/render"
	"stock/pkg/models"
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

		var portfolio []*models.StockPortfolio
		if err := c.GET("/api/portfolio", &portfolio); err != nil {
			return err
		}

		render.PortfolioTable(portfolio)
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
			&models.StockPortfolioReq{Note: note, Code: args[0]}, nil)
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
