package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"stock/pkg/stockctl/client"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store API token in config",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("API Token (stk_...): ")
		tok, _ := reader.ReadString('\n')
		tok = strings.TrimSpace(tok)
		if tok == "" {
			return fmt.Errorf("token required")
		}
		c := client.New(cfg.ServerURL, tok)
		var me struct {
			Username string `json:"username"`
		}
		if err := c.GET("/api/auth/me", &me); err != nil {
			return fmt.Errorf("token validation failed: %w", err)
		}
		cfg.Token = tok
		if err := cfg.Save(cfgFile); err != nil {
			return err
		}
		fmt.Printf("Logged in as %s. Token saved.\n", me.Username)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
