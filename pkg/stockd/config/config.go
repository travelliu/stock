// Package config loads stockd's YAML configuration via viper.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Tushare   TushareConfig   `mapstructure:"tushare"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

type ServerConfig struct {
	Listen        string    `mapstructure:"listen"`
	BaseURL       string    `mapstructure:"base_url"`
	SessionSecret string    `mapstructure:"session_secret"`
	TLS           TLSConfig `mapstructure:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type TushareConfig struct {
	DefaultToken string        `mapstructure:"default_token"`
	BaseURL      string        `mapstructure:"base_url"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

func (t *TushareConfig) GetDefaultToken(s string) string {
	if s == "" {
		return t.DefaultToken
	}
	return s
}

type SchedulerConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	DailyFetchCron    string `mapstructure:"daily_fetch_cron"`
	StocklistSyncCron string `mapstructure:"stocklist_sync_cron"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Dir    string `mapstructure:"dir"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.AddConfigPath("./")
	v.AddConfigPath("./config")
	v.SetConfigName("config")
	v.SetDefault("server.listen", ":8443")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("scheduler.enabled", true)
	v.SetDefault("scheduler.daily_fetch_cron", "0 22 * * 1-5")
	v.SetDefault("scheduler.stocklist_sync_cron", "0 3 * * 0")
	v.SetDefault("tushare.base_url", "http://api.tushare.pro")
	v.SetDefault("tushare.timeout", "30s")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.dir", "logs")

	v.SetEnvPrefix("STOCKD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicit bindings for keys that may not appear in config file
	_ = v.BindEnv("database.dsn")
	_ = v.BindEnv("database.driver")
	_ = v.BindEnv("server.listen")
	_ = v.BindEnv("server.session_secret")
	_ = v.BindEnv("tushare.default_token")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	fmt.Println(v.ConfigFileUsed())
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(c *Config) error {
	if len(c.Server.SessionSecret) < 32 {
		return fmt.Errorf("server.session_secret must be at least 32 bytes (got %d)", len(c.Server.SessionSecret))
	}
	switch c.Database.Driver {
	case "sqlite", "mysql", "postgres":
	default:
		return fmt.Errorf("database.driver must be sqlite|mysql|postgres (got %q)", c.Database.Driver)
	}
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	return nil
}
