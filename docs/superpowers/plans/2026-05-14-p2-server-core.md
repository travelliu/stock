# P2 — Server Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up `stockd`'s foundation: viper config loader, multi-driver GORM init + AutoMigrate, all data-model structs, the dual-auth middleware (session + Bearer token), and the first-run admin bootstrap with a `job_runs` audit table.

**Architecture:** Server-only packages live under `pkg/stockd/`. The CLI stays untouched. Each package is internally cohesive (config doesn't know about gorm, db doesn't know about HTTP, auth doesn't know about routing). Foreign keys are intentionally omitted from GORM tags — ownership is enforced in the service layer (P3).

**Tech Stack:** `gorm.io/gorm` + sqlite/mysql/postgres drivers, `spf13/viper`, `gin-gonic/gin`, `gin-contrib/sessions`, `golang.org/x/crypto/bcrypt`, `crypto/rand`, `crypto/sha256`, `sirupsen/logrus`.

**Reference spec:** `docs/superpowers/specs/2026-05-14-go-vue-rewrite-design.md` §2, §3.1, §6.3, §6.4, §7.2 (P2).

---

## File overview

| File | Responsibility |
|------|----------------|
| `pkg/stockd/config/config.go` | `Load(path string) (*Config, error)` — viper-backed loader + validation |
| `pkg/stockd/config/config_test.go` | Round-trip a YAML fixture; assert defaults; assert errors on missing required keys |
| `pkg/stockd/db/db.go` | `Open(*Config) (*gorm.DB, error)`; driver switch; AutoMigrate everything from `models.*` |
| `pkg/stockd/db/db_test.go` | sqlite-in-memory tests for `Open` + migration idempotency |
| `pkg/stockd/models/{user,api_token,stock,daily_bar,portfolio,intraday_draft,job_run}.go` | One file per table (matches spec §2) |
| `pkg/stockd/models/models_test.go` | Smoke tests: each model can be saved + reloaded |
| `pkg/stockd/auth/password.go` | `HashPassword`, `CheckPassword` (bcrypt) |
| `pkg/stockd/auth/token.go` | `GenerateAPIToken() (plain, hash string)`, `HashToken(plain) string` (sha256), `tk_` prefix verification |
| `pkg/stockd/auth/session.go` | Cookie-store factory `NewSessionStore(secret []byte) sessions.Store` |
| `pkg/stockd/auth/middleware.go` | `Middleware(db, store) gin.HandlerFunc` — extract user from Bearer or session, attach to context |
| `pkg/stockd/auth/auth_test.go` | bcrypt round-trip, token gen+verify, middleware happy/401 paths |
| `pkg/stockd/bootstrap/bootstrap.go` | `EnsureAdmin(db, logger) error` — seed `admin/<random>` when users table is empty |
| `pkg/stockd/bootstrap/bootstrap_test.go` | Idempotent: second call is a no-op |

The `job_run` model is defined here even though the scheduler that writes to it lives in P3 (task 22) — owning the table next to the others keeps AutoMigrate centralised.

---

### Task 10: `pkg/stockd/config` — viper loader + validation

**Files:**
- Create: `pkg/stockd/config/config.go`
- Test: `pkg/stockd/config/config_test.go`
- Create: `deploy/config.example.yaml` (also referenced from README later)

- [ ] **Step 1: Write the failing test**

Create `pkg/stockd/config/config_test.go`:
```go
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
```

- [ ] **Step 2: Run test (expect compile failure)**

Run: `go test ./pkg/stockd/config/... -v`
Expected: `undefined: config.Load`.

- [ ] **Step 3: Write the implementation**

Create `pkg/stockd/config/config.go`:
```go
// Package config loads stockd's YAML configuration via viper.
package config

import (
	"fmt"
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

type SchedulerConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	DailyFetchCron    string `mapstructure:"daily_fetch_cron"`
	StocklistSyncCron string `mapstructure:"stocklist_sync_cron"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	v.SetDefault("server.listen", ":8443")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("scheduler.enabled", true)
	v.SetDefault("scheduler.daily_fetch_cron", "0 22 * * 1-5")
	v.SetDefault("scheduler.stocklist_sync_cron", "0 3 * * 0")
	v.SetDefault("tushare.base_url", "http://api.tushare.pro")
	v.SetDefault("tushare.timeout", "30s")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

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
```

- [ ] **Step 4: Verify tests pass**

Run: `go test ./pkg/stockd/config/... -v`
Expected: PASS.

- [ ] **Step 5: Write `deploy/config.example.yaml`**

```yaml
server:
  listen: ":8443"
  base_url: "https://stock.example.com"
  session_secret: "change-me-this-must-be-at-least-32-bytes-long"
  tls:
    enabled: true
    cert_file: /etc/stockd/tls/fullchain.pem
    key_file:  /etc/stockd/tls/privkey.pem

database:
  driver: sqlite                     # sqlite | mysql | postgres
  dsn: "/var/lib/stockd/stock.db"

tushare:
  default_token: ""                  # required for sync to work
  base_url: "http://api.tushare.pro"
  timeout: 30s

scheduler:
  enabled: true
  daily_fetch_cron:    "0 22 * * 1-5"
  stocklist_sync_cron: "0 3 * * 0"

logging:
  level: info
  format: json
```

- [ ] **Step 6: Commit**

```bash
git add pkg/stockd/config/ deploy/config.example.yaml
git commit -m "feat(config): viper-backed config loader with validation"
```

---

### Task 11: `pkg/stockd/db` — multi-driver init + AutoMigrate harness

**Files:**
- Create: `pkg/stockd/db/db.go`
- Test: `pkg/stockd/db/db_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/stockd/db/db_test.go`:
```go
package db_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockd/config"
	"stock/pkg/stockd/db"
)

func TestOpen_SQLiteInMemory(t *testing.T) {
	cfg := &config.Config{Database: config.DatabaseConfig{Driver: "sqlite", DSN: "file::memory:?cache=shared"}}
	gdb, err := db.Open(cfg)
	require.NoError(t, err)
	assert.NotNil(t, gdb)
	// Migration ran: `users` table exists.
	var n int
	require.NoError(t, gdb.Raw("SELECT count(*) FROM users").Scan(&n).Error)
	assert.Equal(t, 0, n)
}

func TestOpen_RejectsUnknownDriver(t *testing.T) {
	cfg := &config.Config{Database: config.DatabaseConfig{Driver: "mssql", DSN: "x"}}
	_, err := db.Open(cfg)
	require.Error(t, err)
}
```

- [ ] **Step 2: Write `pkg/stockd/db/db.go`**

```go
// Package db opens the GORM connection and runs AutoMigrate for every models.
package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"stock/pkg/stockd/config"
	"stock/pkg/stockd/models"
)

// Open returns a configured *gorm.DB and runs AutoMigrate.
func Open(cfg *config.Config) (*gorm.DB, error) {
	var dialect gorm.Dialector
	switch cfg.Database.Driver {
	case "sqlite":
		dialect = sqlite.Open(cfg.Database.DSN)
	case "mysql":
		dialect = mysql.Open(cfg.Database.DSN)
	case "postgres":
		dialect = postgres.Open(cfg.Database.DSN)
	default:
		return nil, fmt.Errorf("unknown driver %q", cfg.Database.Driver)
	}
	gdb, err := gorm.Open(dialect, &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(gormlogger.Warn),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := AutoMigrate(gdb); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}
	return gdb, nil
}

// AutoMigrate creates/updates every table managed by stockd.
func AutoMigrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&models.User{},
		&models.APIToken{},
		&models.Stock{},
		&models.DailyBar{},
		&models.Portfolio{},
		&models.IntradayDraft{},
		&models.JobRun{},
	)
}
```

- [ ] **Step 3: Run tests (failure expected: models don't exist yet)**

Run: `go test ./pkg/stockd/db/... -v`
Expected: build error on `models.User` etc. → resolved by Task 12.

- [ ] **Step 4: Commit (defer green to Task 12)**

```bash
git add pkg/stockd/db/
git commit -m "feat(db): GORM multi-driver Open and AutoMigrate harness"
```

---

### Task 12: `pkg/stockd/models` — all GORM models

**Files:**
- Create: `pkg/stockd/models/user.go`
- Create: `pkg/stockd/models/api_token.go`
- Create: `pkg/stockd/models/stock.go`
- Create: `pkg/stockd/models/daily_bar.go`
- Create: `pkg/stockd/models/portfolio.go`
- Create: `pkg/stockd/models/intraday_draft.go`
- Create: `pkg/stockd/models/job_run.go`
- Test: `pkg/stockd/models/models_test.go`

All field definitions match spec §2 verbatim.

- [ ] **Step 1: Create `user.go`**

```go
package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;size:64;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"size:16;not null"`
	TushareToken string `gorm:"size:128"`
	Disabled     bool   `gorm:"not null;default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
```

- [ ] **Step 2: Create `api_token.go`**

```go
package models

import "time"

type APIToken struct {
	ID         uint   `gorm:"primaryKey"`
	UserID     uint   `gorm:"index;not null"`
	Name       string `gorm:"size:64;not null"`
	TokenHash  string `gorm:"uniqueIndex;size:64;not null"`
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}
```

- [ ] **Step 3: Create `stock.go`**

```go
package models

import "time"

type Stock struct {
	TsCode    string `gorm:"primaryKey;size:16"`
	Code      string `gorm:"index;size:8;not null"`
	Name      string `gorm:"size:32;not null"`
	Area      string `gorm:"size:16"`
	Industry  string `gorm:"size:32"`
	Market    string `gorm:"size:16"`
	Exchange  string `gorm:"size:8"`
	ListDate  string `gorm:"size:8"`
	Delisted  bool   `gorm:"not null;default:false"`
	UpdatedAt time.Time
}
```

- [ ] **Step 4: Create `daily_bar.go`**

```go
package models

type DailyBar struct {
	TsCode    string `gorm:"primaryKey;size:16"`
	TradeDate string `gorm:"primaryKey;size:8"`
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Vol       float64
	Amount    float64
	SpreadOH  float64
	SpreadOL  float64
	SpreadHL  float64
	SpreadOC  float64
	SpreadHC  float64
	SpreadLC  float64
}
```

- [ ] **Step 5: Create `portfolio.go`**

```go
package models

import "time"

type Portfolio struct {
	ID      uint      `gorm:"primaryKey"`
	UserID  uint      `gorm:"uniqueIndex:idx_user_code;not null"`
	TsCode  string    `gorm:"uniqueIndex:idx_user_code;size:16;not null"`
	Note    string    `gorm:"size:255"`
	AddedAt time.Time
}
```

- [ ] **Step 6: Create `intraday_draft.go`**

```go
package models

import "time"

type IntradayDraft struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_code_date;not null"`
	TsCode    string    `gorm:"uniqueIndex:idx_user_code_date;size:16;not null"`
	TradeDate string    `gorm:"uniqueIndex:idx_user_code_date;size:8;not null"`
	Open      *float64
	High      *float64
	Low       *float64
	Close     *float64
	UpdatedAt time.Time
}
```

- [ ] **Step 7: Create `job_run.go`**

```go
package models

import "time"

type JobRun struct {
	ID         uint      `gorm:"primaryKey"`
	JobName    string    `gorm:"size:64;index;not null"`
	StartedAt  time.Time `gorm:"not null"`
	FinishedAt *time.Time
	Status     string    `gorm:"size:16;not null"` // "running" | "success" | "error"
	Message    string    `gorm:"type:text"`
}
```

- [ ] **Step 8: Create the smoke test**

Create `pkg/stockd/models/models_test.go`:
```go
package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/db"
	"stock/pkg/stockd/models"
)

func openTestDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(0)"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestModelsRoundTrip(t *testing.T) {
	gdb := openTestDB(t)
	now := time.Now()
	cases := []any{
		&models.User{Username: "alice", PasswordHash: "x", Role: "user", CreatedAt: now, UpdatedAt: now},
		&models.APIToken{UserID: 1, Name: "cli", TokenHash: "deadbeef", CreatedAt: now},
		&models.Stock{TsCode: "600519.SH", Code: "600519", Name: "贵州茅台", Market: "主板", Exchange: "SSE", UpdatedAt: now},
		&models.DailyBar{TsCode: "600519.SH", TradeDate: "20250513", Open: 1620, High: 1655, Low: 1601, Close: 1632, Vol: 3500, Amount: 5e5},
		&models.Portfolio{UserID: 1, TsCode: "600519.SH", AddedAt: now},
		&models.IntradayDraft{UserID: 1, TsCode: "600519.SH", TradeDate: "20260514", UpdatedAt: now},
		&models.JobRun{JobName: "daily-fetch", StartedAt: now, Status: "running"},
	}
	for _, c := range cases {
		require.NoError(t, gdb.Create(c).Error)
	}
}
```

- [ ] **Step 9: Run tests**

Run: `go test ./pkg/stockd/... -v`
Expected: `pkg/stockd/config`, `pkg/stockd/db`, `pkg/stockd/model` all PASS.

- [ ] **Step 10: Commit**

```bash
git add pkg/stockd/models/
git commit -m "feat(model): add GORM models matching design spec §2"
```

---

### Task 13: `pkg/stockd/auth` — passwords, API tokens, session store, middleware

**Files:**
- Create: `pkg/stockd/auth/password.go`
- Create: `pkg/stockd/auth/token.go`
- Create: `pkg/stockd/auth/session.go`
- Create: `pkg/stockd/auth/middleware.go`
- Test: `pkg/stockd/auth/auth_test.go`

API-token format: plain prefix `stk_` + 32 url-safe characters (24 random bytes). DB stores sha256 of the plain string. Verification: extract by Bearer header, sha256 the rest, look up `api_tokens.token_hash`.

- [ ] **Step 1: Write `password.go`**

```go
// Package auth provides password hashing, API-token generation, session
// store factory, and the gin middleware that resolves the calling user.
package auth

import "golang.org/x/crypto/bcrypt"

const BcryptCost = 12

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
```

- [ ] **Step 2: Write `token.go`**

```go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const TokenPrefix = "stk_"

// GenerateAPIToken returns (plainText, sha256Hex). The plain text is shown to
// the user once; only the hash is persisted.
func GenerateAPIToken() (string, string, error) {
	var buf [24]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", "", err
	}
	plain := TokenPrefix + base64.RawURLEncoding.EncodeToString(buf[:])
	return plain, HashToken(plain), nil
}

// HashToken returns the lowercase-hex sha256 of the plain token.
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// ParseBearer returns the raw token portion of an Authorization header, or
// an error if the header is missing, malformed, or not a stockd token.
func ParseBearer(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("expected Bearer scheme")
	}
	tok := strings.TrimSpace(parts[1])
	if !strings.HasPrefix(tok, TokenPrefix) {
		return "", fmt.Errorf("token must start with %q", TokenPrefix)
	}
	return tok, nil
}
```

- [ ] **Step 3: Write `session.go`**

```go
package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

const SessionName = "stockd_session"

// NewSessionStore returns a signed cookie store with sensible defaults.
// secret must be >= 32 bytes (enforced by config validation).
func NewSessionStore(secret []byte) sessions.Store {
	store := cookie.NewStore(secret)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: 2, // http.SameSiteLaxMode
	})
	return store
}
```

- [ ] **Step 4: Write `middleware.go`**

```go
package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"stock/pkg/stockd/models"
)

const (
	ctxUserKey  = "stockd.user"
	ctxTokenKey = "stockd.tushare_token"
	sessionUserKey = "uid"
)

// Middleware returns a gin handler that resolves the calling user from a
// Bearer token (CLI/skill) or session cookie (browser). On success it
// attaches *models.User and the effective Tushare token to the gin context.
func Middleware(gdb *gorm.DB, defaultTushareToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := resolveUser(c, gdb)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false, "error": err.Error(),
			})
			return
		}
		if user.Disabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false, "error": "account disabled",
			})
			return
		}
		c.Set(ctxUserKey, user)
		token := defaultTushareToken
		if user.TushareToken != "" {
			token = user.TushareToken
		}
		c.Set(ctxTokenKey, token)
		c.Next()
	}
}

func resolveUser(c *gin.Context, gdb *gorm.DB) (*models.User, error) {
	if h := c.GetHeader("Authorization"); h != "" {
		plain, err := ParseBearer(h)
		if err == nil {
			return userByToken(gdb, plain)
		}
	}
	sess := sessions.Default(c)
	if v := sess.Get(sessionUserKey); v != nil {
		if uid, ok := v.(uint); ok {
			return userByID(gdb, uid)
		}
	}
	return nil, errors.New("unauthenticated")
}

func userByToken(gdb *gorm.DB, plain string) (*models.User, error) {
	hash := HashToken(plain)
	var tok models.APIToken
	if err := gdb.Where("token_hash = ?", hash).First(&tok).Error; err != nil {
		return nil, errors.New("invalid token")
	}
	if tok.ExpiresAt != nil && tok.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	now := time.Now()
	gdb.Model(&tok).Update("last_used_at", &now)
	return userByID(gdb, tok.UserID)
}

func userByID(gdb *gorm.DB, id uint) (*models.User, error) {
	var u models.User
	if err := gdb.First(&u, id).Error; err != nil {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

// User returns the user attached to this request, or nil if absent.
func User(c *gin.Context) *models.User {
	if v, ok := c.Get(ctxUserKey); ok {
		return v.(*models.User)
	}
	return nil
}

// TushareTokenFor returns the effective tushare token for the request.
func TushareTokenFor(c *gin.Context) string {
	if v, ok := c.Get(ctxTokenKey); ok {
		return v.(string)
	}
	return ""
}
```

- [ ] **Step 5: Write the auth tests**

Create `pkg/stockd/auth/auth_test.go`:
```go
package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/db"
	"stock/pkg/stockd/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestPasswordRoundTrip(t *testing.T) {
	h, err := auth.HashPassword("hunter2")
	require.NoError(t, err)
	require.NoError(t, auth.CheckPassword(h, "hunter2"))
	assert.Error(t, auth.CheckPassword(h, "wrong"))
}

func TestGenerateAPIToken_Format(t *testing.T) {
	plain, hash, err := auth.GenerateAPIToken()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(plain, auth.TokenPrefix))
	assert.Equal(t, 64, len(hash), "sha256 hex is 64 chars")
	assert.Equal(t, hash, auth.HashToken(plain))
}

func TestParseBearer(t *testing.T) {
	tok, err := auth.ParseBearer("Bearer stk_AAA")
	require.NoError(t, err)
	assert.Equal(t, "stk_AAA", tok)

	_, err = auth.ParseBearer("")
	assert.Error(t, err)
	_, err = auth.ParseBearer("Token stk_x")
	assert.Error(t, err)
	_, err = auth.ParseBearer("Bearer xxx")
	assert.Error(t, err)
}

func TestMiddleware_BearerToken(t *testing.T) {
	gdb := openDB(t)
	pw, _ := auth.HashPassword("x")
	user := models.User{Username: "alice", PasswordHash: pw, Role: "user"}
	require.NoError(t, gdb.Create(&user).Error)
	plain, hash, _ := auth.GenerateAPIToken()
	require.NoError(t, gdb.Create(&models.APIToken{UserID: user.ID, Name: "cli", TokenHash: hash}).Error)

	r := gin.New()
	store := auth.NewSessionStore([]byte("12345678901234567890123456789012"))
	r.Use(sessions.Sessions(auth.SessionName, store))
	r.Use(auth.Middleware(gdb, ""))
	r.GET("/me", func(c *gin.Context) {
		u := auth.User(c)
		c.JSON(http.StatusOK, gin.H{"username": u.Username})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+plain)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alice")

	w = httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/me", nil)
	r.ServeHTTP(w, req2)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

- [ ] **Step 6: Run tests**

Run: `go test ./pkg/stockd/auth/... -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add pkg/stockd/auth/
git commit -m "feat(auth): password hashing, API token issuance, session store, dual-auth middleware"
```

---

### Task 14: First-run bootstrap + `job_runs` plumbing

**Files:**
- Create: `pkg/stockd/bootstrap/bootstrap.go`
- Test: `pkg/stockd/bootstrap/bootstrap_test.go`

The bootstrap policy (spec §6.4): when the `users` table is empty, generate a 24-char random password, hash it, and seed `admin/<password>` with role `admin`. The plain password is **printed to stderr and the configured logger** so an operator can capture it.

- [ ] **Step 1: Write the test**

Create `pkg/stockd/bootstrap/bootstrap_test.go`:
```go
package bootstrap_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockd/bootstrap"
	"stock/pkg/stockd/db"
	"stock/pkg/stockd/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestEnsureAdmin_SeedsWhenEmpty(t *testing.T) {
	gdb := openDB(t)
	logger := logrus.New()
	plain, err := bootstrap.EnsureAdmin(gdb, logger)
	require.NoError(t, err)
	assert.NotEmpty(t, plain, "seeded password should be returned")

	var n int64
	require.NoError(t, gdb.Model(&models.User{}).Count(&n).Error)
	assert.Equal(t, int64(1), n)

	var u models.User
	require.NoError(t, gdb.First(&u, "username = ?", "admin").Error)
	assert.Equal(t, "admin", u.Role)
}

func TestEnsureAdmin_NoopWhenUsersExist(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.User{Username: "u", PasswordHash: "h", Role: "user"}).Error)
	plain, err := bootstrap.EnsureAdmin(gdb, logrus.New())
	require.NoError(t, err)
	assert.Empty(t, plain, "should not seed when users already exist")
}
```

- [ ] **Step 2: Write `bootstrap.go`**

```go
// Package bootstrap performs one-time tasks at server startup.
package bootstrap

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/models"
)

// EnsureAdmin seeds admin/<random-password> when the users table is empty.
// On seed, the plain password is returned AND logged at WARN level so the
// operator can capture it. Returns "" if no seeding occurred.
func EnsureAdmin(gdb *gorm.DB, logger *logrus.Logger) (string, error) {
	var n int64
	if err := gdb.Model(&models.User{}).Count(&n).Error; err != nil {
		return "", fmt.Errorf("count users: %w", err)
	}
	if n > 0 {
		return "", nil
	}
	plain, err := generatePassword(24)
	if err != nil {
		return "", err
	}
	hash, err := auth.HashPassword(plain)
	if err != nil {
		return "", err
	}
	admin := &models.User{Username: "admin", PasswordHash: hash, Role: "admin"}
	if err := gdb.Create(admin).Error; err != nil {
		return "", fmt.Errorf("create admin: %w", err)
	}
	logger.WithFields(logrus.Fields{
		"username": "admin",
		"password": plain,
	}).Warn("seeded initial admin user — change this password immediately")
	return plain, nil
}

func generatePassword(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf)[:n], nil
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./pkg/stockd/... -v`
Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add pkg/stockd/bootstrap/
git commit -m "feat(bootstrap): seed initial admin user with random password on empty DB"
```

---

## Exit criterion

- [ ] `go test ./pkg/stockd/...` green (config, db, model, auth, bootstrap)
- [ ] Importing `pkg/stockd/db` creates all 7 tables in an in-memory SQLite
- [ ] Bootstrap seeds admin when DB is empty and is idempotent

## Hand-off

Next: [P3 — Services](./2026-05-14-p3-services.md). P3 layers business logic (user/token/stock/portfolio/draft/bars/analysis/scheduler) on top of the models defined here.
