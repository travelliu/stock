package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"stock/pkg/stockctl/client"
)

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Manage intraday drafts",
}

var draftGetCmd = &cobra.Command{
	Use:   "get [ts_code]",
	Short: "Get today's draft",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		var res struct {
			Open  *float64 `json:"open"`
			High  *float64 `json:"high"`
			Low   *float64 `json:"low"`
			Close *float64 `json:"close"`
		}
		if err := c.GET("/api/drafts/today?ts_code="+args[0], &res); err != nil {
			return err
		}
		fmt.Printf("Open: %v\nHigh: %v\nLow: %v\nClose: %v\n",
			ptrStr(res.Open), ptrStr(res.High), ptrStr(res.Low), ptrStr(res.Close))
		return nil
	},
}

var draftSetCmd = &cobra.Command{
	Use:   "set [ts_code]",
	Short: "Set draft values",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		body := map[string]any{
			"ts_code":    args[0],
			"trade_date": today(),
		}
		if v, _ := cmd.Flags().GetFloat64("open"); v != 0 {
			body["open"] = v
		}
		if v, _ := cmd.Flags().GetFloat64("high"); v != 0 {
			body["high"] = v
		}
		if v, _ := cmd.Flags().GetFloat64("low"); v != 0 {
			body["low"] = v
		}
		if v, _ := cmd.Flags().GetFloat64("close"); v != 0 {
			body["close"] = v
		}
		return c.PUT("/api/drafts", body, nil)
	},
}

var draftClearCmd = &cobra.Command{
	Use:   "clear [ts_code]",
	Short: "Clear today's draft",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(cfg.ServerURL, cfg.Token)
		var res struct {
			ID uint `json:"id"`
		}
		if err := c.GET("/api/drafts/today?ts_code="+args[0], &res); err != nil {
			return err
		}
		return c.DELETE(fmt.Sprintf("/api/drafts/%d", res.ID))
	},
}

func ptrStr(p *float64) string {
	if p == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *p)
}

func today() string {
	return time.Now().Format("20060102")
}

func init() {
	rootCmd.AddCommand(draftCmd)
	draftCmd.AddCommand(draftGetCmd, draftSetCmd, draftClearCmd)
	draftSetCmd.Flags().Float64("open", 0, "Open price")
	draftSetCmd.Flags().Float64("high", 0, "High price")
	draftSetCmd.Flags().Float64("low", 0, "Low price")
	draftSetCmd.Flags().Float64("close", 0, "Close price")
}
