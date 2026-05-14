package cmd

import (
	"encoding/json"
	"fmt"
	"stock/pkg/models"

	"github.com/spf13/cobra"

	"stock/pkg/cli/client"
	"stock/pkg/cli/render"
)

var stockCmd = &cobra.Command{
	Use:   "stock",
	Short: "Stock search and analysis",
}

var stockSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search stocks by code or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		var res []*models.Stock
		if err := c.GET("/api/stocks?q="+args[0], &res); err != nil {
			return err
		}
		for _, s := range res {
			fmt.Printf("%s\t%s\n", s.TsCode, s.Name)
		}
		return nil
	},
}

var stockAnalysisCmd = &cobra.Command{
	Use:   "analysis [ts_code]",
	Short: "Run spread analysis",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		format, _ := cmd.Flags().GetString("format")

		path := "/api/analysis/" + args[0]
		qs := ""
		if v, _ := cmd.Flags().GetFloat64("actual-open"); v != 0 {
			qs += fmt.Sprintf("&actual_open=%.2f", v)
		}
		if v, _ := cmd.Flags().GetFloat64("actual-high"); v != 0 {
			qs += fmt.Sprintf("&actual_high=%.2f", v)
		}
		if v, _ := cmd.Flags().GetFloat64("actual-low"); v != 0 {
			qs += fmt.Sprintf("&actual_low=%.2f", v)
		}
		if v, _ := cmd.Flags().GetFloat64("actual-close"); v != 0 {
			qs += fmt.Sprintf("&actual_close=%.2f", v)
		}
		if useDraft, _ := cmd.Flags().GetBool("use-draft"); useDraft {
			qs += "&with_draft=true"
		}
		if qs != "" {
			path += "?" + qs[1:]
		}

		if format == "json" {
			var raw json.RawMessage
			if err := c.GET(path, &raw); err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result models.AnalysisResult
		if err := c.GET(path, &result); err != nil {
			return err
		}
		render.AnalysisTable(result)
		return nil
	},
}

var stockHistoryCmd = &cobra.Command{
	Use:   "history [ts_code]",
	Short: "Show daily bar history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		path := fmt.Sprintf("/api/bars/%s?from=%s&to=%s", args[0], from, to)
		var res []*models.DailyBar
		if err := c.GET(path, &res); err != nil {
			return err
		}
		for _, b := range res {
			fmt.Printf("%s  O:%.2f H:%.2f L:%.2f C:%.2f\n", b.TradeDate, b.Open, b.High, b.Low, b.Close)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stockCmd)
	stockCmd.AddCommand(stockSearchCmd, stockAnalysisCmd, stockHistoryCmd)
	stockAnalysisCmd.Flags().String("format", "table", "Output format: table|json")
	stockAnalysisCmd.Flags().Float64("actual-open", 0, "Override open price")
	stockAnalysisCmd.Flags().Float64("actual-high", 0, "Override high price")
	stockAnalysisCmd.Flags().Float64("actual-low", 0, "Override low price")
	stockAnalysisCmd.Flags().Float64("actual-close", 0, "Override close price")
	stockAnalysisCmd.Flags().Bool("use-draft", false, "Use today's draft values")
	stockHistoryCmd.Flags().String("from", "", "Start date YYYYMMDD")
	stockHistoryCmd.Flags().String("to", "", "End date YYYYMMDD")
}
