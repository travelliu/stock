package cmd

import (
	"encoding/json"
	"fmt"
	"stock/pkg/models"
	"strings"

	"github.com/spf13/cobra"

	"stock/pkg/cli/client"
	"stock/pkg/cli/render"
)

var stockCmd = &cobra.Command{
	Use:   "stock",
	Short: "StockBasicInfo search and analysis",
}

var stockSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search stocks by code or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		var res []*models.StockBasicInfo
		if err := c.GET("/api/stocks?q="+args[0], &res); err != nil {
			return err
		}
		for _, s := range res {
			fmt.Printf("%s\t%s\n", s.TsCode, s.Name)
		}
		return nil
	},
}

var stockFetchCmd = &cobra.Command{
	Use:   "fetch [ts_code]",
	Short: "Search stocks by code or name",
	// Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		if err := c.POST("/api/admin/bars/sync", nil, nil); err != nil {
			return err
		}
		return nil
	},
}

var stockAnalysisCmd = &cobra.Command{
	Use:   "analysis [ts_code]",
	Short: "RunStockAnalysis spread analysis",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		format, _ := cmd.Flags().GetString("format")

		tsCode := args[0]
		plainCode := tsCode
		if idx := strings.Index(tsCode, "."); idx > 0 {
			plainCode = tsCode[:idx]
		}

		// Fetch cached realtime+analysis in one call.
		var ra models.StockRealtimeAndAnalysis
		hasRA := c.GET("/api/stocks/"+plainCode+"/quote", &ra) == nil

		hasCustomFlags := cmd.Flags().Changed("actual-open") ||
			cmd.Flags().Changed("actual-high") ||
			cmd.Flags().Changed("actual-low") ||
			cmd.Flags().Changed("actual-close")

		// If no custom price flags and cache has analysis, use it directly.
		if !hasCustomFlags && hasRA && ra.StockAnalysisResult != nil {
			if format == "json" {
				return nil
			}
			render.AnalysisTable(*ra.StockAnalysisResult)
			return nil
		}

		// Build fresh analysis request with user overrides or quote prices as defaults.
		path := "/api/stocks/" + tsCode + "/analysis"
		qs := ""
		appendParam := func(flag, key string, quoteVal float64) {
			if cmd.Flags().Changed(flag) {
				v, _ := cmd.Flags().GetFloat64(flag)
				qs += fmt.Sprintf("&%s=%.2f", key, v)
			} else if hasRA && ra.StockRealtime != nil && quoteVal != 0 {
				qs += fmt.Sprintf("&%s=%.2f", key, quoteVal)
			}
		}

		var openDefault float64
		if hasRA && ra.StockRealtime != nil {
			openDefault = ra.StockRealtime.Open
			if openDefault == 0 {
				openDefault = ra.StockRealtime.Price
			}
		}
		var high, low, price float64
		if hasRA && ra.StockRealtime != nil {
			high = ra.StockRealtime.High
			low = ra.StockRealtime.Low
			price = ra.StockRealtime.Price
		}
		appendParam("actual-open", "actual_open", openDefault)
		appendParam("actual-high", "actual_high", high)
		appendParam("actual-low", "actual_low", low)
		appendParam("actual-close", "actual_close", price)

		if qs != "" {
			path += "?" + qs[1:]
		}

		if format == "json" {
			var raw json.RawMessage
			if err := c.GET(path, &raw); err != nil {
				return err
			}
			return nil
		}

		var result models.StockAnalysisResult
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
		path := fmt.Sprintf("/api/stocks/%s/bars?from=%s&to=%s", args[0], from, to)
		var res *models.BarsPage
		if err := c.GET(path, &res); err != nil {
			return err
		}
		render.BarsTable(res.Items)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stockCmd)
	stockCmd.AddCommand(stockSearchCmd, stockAnalysisCmd, stockHistoryCmd, stockPredictionsCmd, stockRecalcCmd, stockFetchCmd)
	stockAnalysisCmd.Flags().String("format", "table", "Output format: table|json")
	stockAnalysisCmd.Flags().Float64("actual-open", 0, "Override open price")
	stockAnalysisCmd.Flags().Float64("actual-high", 0, "Override high price")
	stockAnalysisCmd.Flags().Float64("actual-low", 0, "Override low price")
	stockAnalysisCmd.Flags().Float64("actual-close", 0, "Override close price")
	stockHistoryCmd.Flags().String("from", "", "Start date YYYYMMDD")
	stockHistoryCmd.Flags().String("to", "", "End date YYYYMMDD")
	stockPredictionsCmd.Flags().String("from", "", "Start date YYYYMMDD")
	stockPredictionsCmd.Flags().String("to", "", "End date YYYYMMDD")
	stockPredictionsCmd.Flags().Int("limit", 30, "Max records")
	stockRecalcCmd.Flags().String("ts-code", "", "Recalc specific stock only")
}

var stockPredictionsCmd = &cobra.Command{
	Use:   "predictions [ts_code]",
	Short: "Show prediction history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		limit, _ := cmd.Flags().GetInt("limit")
		path := fmt.Sprintf("/api/stocks/%s/predictions?limit=%d", args[0], limit)
		if from != "" {
			path += "&from=" + from
		}
		if to != "" {
			path += "&to=" + to
		}
		var preds []models.StockAnalysisPrediction
		if err := c.GET(path, &preds); err != nil {
			return err
		}
		render.PredictionsTable(args[0], "", preds)
		return nil
	},
}

var stockRecalcCmd = &cobra.Command{
	Use:   "recalc",
	Short: "Recalculate predictions for portfolio stocks",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		tsCode, _ := cmd.Flags().GetString("ts-code")
		path := "/api/stocks/analysis/recalc"
		if tsCode != "" {
			path += "?ts_code=" + tsCode
		}
		var res struct {
			Updated int `json:"updated"`
		}
		if err := c.POST(path, nil, &res); err != nil {
			return err
		}
		fmt.Printf("Updated %d predictions\n", res.Updated)
		return nil
	},
}
