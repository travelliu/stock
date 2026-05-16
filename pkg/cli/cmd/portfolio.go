package cmd

import (
	"sync"

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

		var portfolio []*models.Portfolio
		if err := c.GET("/api/portfolio", &portfolio); err != nil {
			return err
		}

		quotes := make(map[string]*models.RealtimeQuote, len(portfolio))
		var mu sync.Mutex
		var wg sync.WaitGroup
		for _, p := range portfolio {
			wg.Add(1)
			go func(code string) {
				defer wg.Done()
				var q models.RealtimeQuote
				if err := c.GET("/api/quotes/"+code, &q); err == nil {
					mu.Lock()
					quotes[code] = &q
					mu.Unlock()
				}
			}(p.Code)
		}
		wg.Wait()

		render.PortfolioTable(portfolio, quotes)
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
			&models.PortfolioReq{Note: note, Code: args[0]}, nil)
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
