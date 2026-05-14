package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockd/config"
)

func writeYAML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	return p
}

func TestLoad_HappyPath(t *testing.T) {
	p := writeYAML(t, `
server:
  listen: ":8443"
  base_url: "https://stock.example.com"
  session_secret: "12345678901234567890123456789012"
database:
  driver: sqlite
  dsn: "/tmp/stock.db"
tushare:
  default_token: "tok"
scheduler:
  enabled: true
logging:
  level: info
  format: json
`)
	cfg, err := config.Load(p)
	require.NoError(t, err)
	assert.Equal(t, ":8443", cfg.Server.Listen)
	assert.Equal(t, "sqlite", cfg.Database.Driver)
	assert.True(t, cfg.Scheduler.Enabled)
	assert.Equal(t, "0 22 * * 1-5", cfg.Scheduler.DailyFetchCron, "default cron")
	assert.Equal(t, "0 3 * * 0", cfg.Scheduler.StocklistSyncCron, "default cron")
}

func TestLoad_RejectsShortSecret(t *testing.T) {
	p := writeYAML(t, `
server:
  session_secret: "short"
database:
  driver: sqlite
  dsn: "/tmp/x.db"
`)
	_, err := config.Load(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session_secret")
}

func TestLoad_RejectsUnknownDriver(t *testing.T) {
	p := writeYAML(t, `
server:
  session_secret: "12345678901234567890123456789012"
database:
  driver: mssql
  dsn: "x"
`)
	_, err := config.Load(p)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "driver")
}
