# Frontend Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `web/src` to match mtkd style (dark narrow sidebar, SCSS variables, i18n, `g-*` components, `apis/` split), unify all HTTP Req/Resp DTOs into `pkg/models`, and standardize all JSON fields to camelCase.

**Architecture:** Backend: extract all inline structs from handlers/cli into `pkg/models`, rename fields to standard camelCase. Frontend: full rewrite of `web/src` with new directory structure, vue-i18n, SCSS, component hierarchy.

**Tech Stack:** Go, Gin, GORM; Vue 3, TypeScript, Element Plus, Pinia, Vue Router, axios, vue-i18n, SCSS, `@element-plus/icons-vue`, Vitest, Playwright.

---

## File Structure

### Backend (modified)

| File | Action | Responsibility |
|---|---|---|
| `pkg/models/auth.go` | Create | `LoginReq` DTO |
| `pkg/models/me.go` | Create | `ChangePasswordReq`, `SetTushareTokenReq`, `IssueTokenReq`, `IssueTokenResp` |
| `pkg/models/admin.go` | Create | `CreateUserReq`, `PatchUserReq` |
| `pkg/models/draft_req.go` | Create | `UpsertDraftReq` |
| `pkg/models/doc.go` | Create | Package-level documentation |
| `pkg/models/user.go` | Modify | JSON tag: `userName` -> `username`, `id` omitempty removed |
| `pkg/models/portfolio.go` | Modify | JSON tag: `userID` -> `userId`, `id`/`addedAt` omitempty removed |
| `pkg/models/daily_bar.go` | Modify | Fix duplicate `Open` JSON tag |
| `pkg/models/api_token.go` | Modify | JSON tag: `userID` -> `userId` |
| `pkg/models/intraday_draft.go` | Modify | JSON tag: `userID` -> `userId`, `id` omitempty removed |
| `pkg/models/job_run.go` | Modify | JSON tag: `id` omitempty removed |
| `pkg/stockd/http/auth.go` | Modify | Replace `loginReq` with `models.LoginReq` |
| `pkg/stockd/http/me.go` | Modify | Replace inline structs with `models.*` |
| `pkg/stockd/http/admin.go` | Modify | Replace inline structs with `models.*` |
| `pkg/stockd/http/draft.go` | Modify | Replace inline struct with `models.UpsertDraftReq` |
| `pkg/cli/cmd/login.go` | Modify | Use `models.User` instead of inline struct |
| `pkg/cli/cmd/portfolio.go` | Modify | Use `[]*models.Portfolio` for GET |
| `pkg/models/models_test.go` | Modify | Add `TestModelJSONFields` |
| `pkg/stockd/http/auth_test.go` | Modify | Update assertions |
| `pkg/cli/client/client_test.go` | Modify | Add portfolio field alignment test |

### Frontend (created)

| File | Responsibility |
|---|---|
| `web/src/assets/css/index.scss` | Body reset, `g-*` classes, `:root` variables (up red, down green) |
| `web/src/assets/css/element-reset.scss` | Card radius, button height overrides |
| `web/src/assets/css/mixin.scss` | `@mixin scrollbar()` |
| `web/src/types/api.ts` | Single TS interface file, 1:1 with `pkg/models` |
| `web/src/apis/axios.ts` | Axios instance + interceptors (envelope unwrap + 401 redirect + lang header) |
| `web/src/apis/auth.ts` | login / logout / me |
| `web/src/apis/me.ts` | changePassword / setTushareToken / API tokens CRUD |
| `web/src/apis/portfolio.ts` | list / add / remove / updateNote |
| `web/src/apis/stocks.ts` | search / detail / bars |
| `web/src/apis/analysis.ts` | analysis with params |
| `web/src/apis/draft.ts` | draft CRUD |
| `web/src/apis/admin.ts` | users / sync |
| `web/src/stores/auth.ts` | User state + login/logout/fetchMe |
| `web/src/stores/lang.ts` | Current lang (zh/en) |
| `web/src/intl/index.ts` | vue-i18n instance |
| `web/src/intl/lang.ts` | Langs / ElementLangs |
| `web/src/intl/langs/zh/index.ts` | Full Chinese translations |
| `web/src/intl/langs/en/index.ts` | Keys only, empty values |
| `web/src/utils/message.ts` | `wMessage` wrapper (dedup within 3s) |
| `web/src/utils/storage.ts` | localStorage/sessionStorage wrapper |
| `web/src/utils/format.ts` | Price / change percentage formatting |
| `web/src/components/GIcon.vue` | Element Plus icon name wrapper |
| `web/src/components/GEllipsis.vue` | Text overflow ellipsis |
| `web/src/components/ConsoleMenu.vue` | Top-left logo + stock menu + bottom lang switch |
| `web/src/components/UserMenu.vue` | Bottom-left user dropdown |
| `web/src/components/StockBasicCard.vue` | Top stock info card |
| `web/src/components/DailyBarTable.vue` | Daily bar table (up red, down green) |
| `web/src/components/SpreadHistogram.vue` | Spread distribution histogram |
| `web/src/components/SpreadModelTable.vue` | Multi-window model table |
| `web/src/components/TradePlanTable.vue` | Trading plan / reverse prediction table |
| `web/src/components/DraftFormBlock.vue` | Today open/high/low/close input |
| `web/src/views/LoginView.vue` | Login page |
| `web/src/views/StockListView.vue` | Stock list (replaces PortfolioView) |
| `web/src/views/stock/StockDetailView.vue` | Shell + el-tabs |
| `web/src/views/stock/BasicTab.vue` | Tab 1: basic info + spreads |
| `web/src/views/stock/StatisticsTab.vue` | Tab 2: detailed statistics |
| `web/src/views/profile/ProfileView.vue` | Shell + left submenu |
| `web/src/views/profile/ProfileInfo.vue` | Read-only user info |
| `web/src/views/profile/ChangePassword.vue` | Change password |
| `web/src/views/profile/TushareToken.vue` | Set tushare token |
| `web/src/views/profile/ApiTokens.vue` | API token management |
| `web/src/views/admin/UsersView.vue` | User management (new style) |
| `web/src/views/admin/SyncView.vue` | Sync management (new style) |
| `web/src/views/NotFound.vue` | 404 page |
| `web/src/App.vue` | Root: el-config-provider + layout |
| `web/src/router/index.ts` | Route definitions |
| `web/src/main.ts` | Entry point |

### Frontend (deleted)

- `web/src/api/client.ts`
- `web/src/components/AnalysisPanel.vue`
- `web/src/components/HistoryPanel.vue`
- `web/src/components/DraftPanel.vue`
- `web/src/views/PortfolioView.vue`
- `web/src/views/SettingsView.vue`

---

## Task 1: Create pkg/models DTO files

**Files:**
- Create: `pkg/models/auth.go`
- Create: `pkg/models/me.go`
- Create: `pkg/models/admin.go`
- Create: `pkg/models/draft_req.go`
- Create: `pkg/models/doc.go`
- Test: `pkg/models/models_test.go`

- [ ] **Step 1: Write `pkg/models/auth.go`**

```go
package models

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
```

- [ ] **Step 2: Write `pkg/models/me.go`**

```go
package models

import "time"

type ChangePasswordReq struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type SetTushareTokenReq struct {
	Token string `json:"token"`
}

type IssueTokenReq struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type IssueTokenResp struct {
	Token    string    `json:"token"`
	Metadata *APIToken `json:"metadata"`
}
```

- [ ] **Step 3: Write `pkg/models/admin.go`**

```go
package models

type CreateUserReq struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	TushareToken string `json:"tushareToken,omitempty"`
}

type PatchUserReq struct {
	Role         *string `json:"role,omitempty"`
	Disabled     *bool   `json:"disabled,omitempty"`
	TushareToken *string `json:"tushareToken,omitempty"`
}
```

- [ ] **Step 4: Write `pkg/models/draft_req.go`**

```go
package models

type UpsertDraftReq struct {
	TsCode    string   `json:"tsCode"`
	TradeDate string   `json:"tradeDate"`
	Open      *float64 `json:"open,omitempty"`
	High      *float64 `json:"high,omitempty"`
	Low       *float64 `json:"low,omitempty"`
	Close     *float64 `json:"close,omitempty"`
}
```

- [ ] **Step 5: Write `pkg/models/doc.go`**

```go
// Package models is the single source of truth for all HTTP request/response DTOs.
// The frontend type definitions in web/src/types/api.ts MUST stay in sync;
// when you change a field here, update both sides in the same PR.
package models
```

- [ ] **Step 6: Run Go build to verify new files compile**

Run: `cd /root/code/github/travelliu/stock && go build ./pkg/models/...`
Expected: PASS (no output)

- [ ] **Step 7: Commit**

```bash
git add pkg/models/auth.go pkg/models/me.go pkg/models/admin.go pkg/models/draft_req.go pkg/models/doc.go
git commit -m "feat(models): add auth, me, admin, draft DTOs and package doc"
```

---

## Task 2: Rename JSON fields to standard camelCase

**Files:**
- Modify: `pkg/models/user.go`
- Modify: `pkg/models/portfolio.go`
- Modify: `pkg/models/daily_bar.go`
- Modify: `pkg/models/api_token.go`
- Modify: `pkg/models/intraday_draft.go`
- Modify: `pkg/models/job_run.go`
- Test: `pkg/models/models_test.go`

- [ ] **Step 1: Modify `pkg/models/user.go`**

```go
package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username,omitempty"`
	PasswordHash string    `gorm:"not null" json:"passwordHash,omitempty"`
	Role         string    `gorm:"size:16;not null" json:"role,omitempty"`
	TushareToken string    `gorm:"size:128" json:"tushareToken,omitempty"`
	Disabled     bool      `gorm:"not null;default:false" json:"disabled,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
```

Changes: `id,omitempty` -> `id`, `userName` -> `username`.

- [ ] **Step 2: Modify `pkg/models/portfolio.go`**

```go
package models

import "time"

type Portfolio struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	UserID  uint      `gorm:"uniqueIndex:idx_user_code;not null" json:"userId"`
	TsCode  string    `gorm:"uniqueIndex:idx_user_code;size:16;not null" json:"tsCode,omitempty"`
	Note    string    `gorm:"size:255" json:"note,omitempty"`
	AddedAt time.Time `json:"addedAt"`
}

type PortfolioReq struct {
	TsCode string `json:"tsCode"`
	Note   string `json:"note,omitempty"`
}
```

Changes: `id,omitempty` -> `id`, `userID` -> `userId`, `addedAt,omitempty` -> `addedAt`.

- [ ] **Step 3: Modify `pkg/models/daily_bar.go`**

```go
package models

type DailyBar struct {
	TsCode    string  `gorm:"primaryKey;size:16" json:"tsCode,omitempty"`
	TradeDate string  `gorm:"primaryKey;size:8" json:"tradeDate,omitempty"`
	Open      float64 `json:"open,omitempty"`
	High      float64 `json:"high,omitempty"`
	Low       float64 `json:"low,omitempty"`
	Close     float64 `json:"close,omitempty"`
	Vol       float64 `json:"vol,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	Spreads   Spreads `gorm:"embedded;embeddedPrefix:spread_" json:"spreads"`
}
```

Change: fix duplicate `json:"open" json:"open,omitempty"` to single `json:"open,omitempty"`. Also remove `omitempty` from `spreads` so it's always present (embedded struct).

- [ ] **Step 4: Modify `pkg/models/api_token.go`**

```go
package models

import "time"

type APIToken struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"userId"`
	Name       string     `gorm:"size:64;not null" json:"name"`
	TokenHash  string     `gorm:"uniqueIndex;size:64;not null" json:"tokenHash"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}
```

Changes: `userID` -> `userId`, `id` already had no omitempty but add explicit.

- [ ] **Step 5: Modify `pkg/models/intraday_draft.go`**

```go
package models

import "time"

type IntradayDraft struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_code_date;not null" json:"userId"`
	TsCode    string    `gorm:"uniqueIndex:idx_user_code_date;size:16;not null" json:"tsCode,omitempty"`
	TradeDate string    `gorm:"uniqueIndex:idx_user_code_date;size:8;not null" json:"tradeDate,omitempty"`
	Open      *float64  `json:"open,omitempty"`
	High      *float64  `json:"high,omitempty"`
	Low       *float64  `json:"low,omitempty"`
	Close     *float64  `json:"close,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}
```

Changes: `id,omitempty` -> `id`, `userID` -> `userId`.

- [ ] **Step 6: Modify `pkg/models/job_run.go`**

```go
package models

import "time"

type JobRun struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	JobName    string     `gorm:"size:64;index;not null" json:"jobName,omitempty"`
	StartedAt  time.Time  `gorm:"not null" json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Status     string     `gorm:"size:16;not null" json:"status,omitempty"` // "running" | "success" | "error"
	Message    string     `gorm:"type:text" json:"message,omitempty"`
}
```

Change: `id,omitempty` -> `id`.

- [ ] **Step 7: Run model tests to verify DB still works**

Run: `go test -race ./pkg/models/...`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add pkg/models/user.go pkg/models/portfolio.go pkg/models/daily_bar.go pkg/models/api_token.go pkg/models/intraday_draft.go pkg/models/job_run.go
git commit -m "refactor(models): standardize JSON fields to camelCase, remove omitempty on keys"
```

---

## Task 3: Replace inline structs in handlers and CLI

**Files:**
- Modify: `pkg/stockd/http/auth.go`
- Modify: `pkg/stockd/http/me.go`
- Modify: `pkg/stockd/http/admin.go`
- Modify: `pkg/stockd/http/draft.go`
- Modify: `pkg/cli/cmd/login.go`
- Modify: `pkg/cli/cmd/portfolio.go`
- Test: `pkg/stockd/http/auth_test.go`

- [ ] **Step 1: Modify `pkg/stockd/http/auth.go`**

```go
package http

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"stock/pkg/models"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

func (h *handler) Login(c *gin.Context) {
	var req models.LoginReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u, err := h.userSvc.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	sess := sessions.Default(c)
	sess.Set("uid", u.ID)
	_ = sess.Save()
	utils.HTTPRequestSuccess(c, http.StatusOK, u)
}

func (h *handler) Logout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	_ = sess.Save()
	utils.HTTPRequestSuccess(c, http.StatusOK, gin.H{"message": "logged out"})
}

func (h *handler) Me(c *gin.Context) {
	u := auth.User(c)
	if u == nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrUnauthorized)
		return
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, u)
}
```

- [ ] **Step 2: Modify `pkg/stockd/http/me.go`**

```go
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/models"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/services/token"
	"stock/pkg/stockd/utils"
)

func (h *handler) ListTokens(c *gin.Context) {
	u := auth.User(c)
	list, err := h.tokenSvc.List(c.Request.Context(), u.ID)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}

func (h *handler) IssueToken(c *gin.Context) {
	var req models.IssueTokenReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	plain, tok, err := h.tokenSvc.Issue(c.Request.Context(), token.IssueInput{
		UserID: u.ID, Name: req.Name, ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, models.IssueTokenResp{Token: plain, Metadata: tok})
}

func (h *handler) RevokeToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	u := auth.User(c)
	if err := h.tokenSvc.Revoke(c.Request.Context(), u.ID, uint(id)); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "revoked"})
}

func (h *handler) SetTushareToken(c *gin.Context) {
	var req models.SetTushareTokenReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.userSvc.SetTushareToken(c.Request.Context(), u.ID, req.Token); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "updated"})
}

func (h *handler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.userSvc.ChangePassword(c.Request.Context(), u.ID, req.Old, req.New); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "password changed"})
}
```

- [ ] **Step 3: Modify `pkg/stockd/http/admin.go`**

```go
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/models"
	"stock/pkg/stockd/services/user"
	"stock/pkg/stockd/utils"
)

func (h *handler) CreateUser(c *gin.Context) {
	var req models.CreateUserReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u, err := h.userSvc.Create(c.Request.Context(), user.CreateInput{
		Username: req.Username, Password: req.Password, Role: req.Role, TushareToken: req.TushareToken,
	})
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, u)
}

func (h *handler) ListUsers(c *gin.Context) {
	list, err := h.userSvc.List(c.Request.Context())
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, list)
}

func (h *handler) PatchUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req models.PatchUserReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	if req.Role != nil {
		_ = h.userSvc.SetRole(c.Request.Context(), uint(id), *req.Role)
	}
	if req.Disabled != nil {
		_ = h.userSvc.SetDisabled(c.Request.Context(), uint(id), *req.Disabled)
	}
	if req.TushareToken != nil {
		_ = h.userSvc.SetTushareToken(c.Request.Context(), uint(id), *req.TushareToken)
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, gin.H{"message": "updated"})
}

func (h *handler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.userSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, gin.H{"message": "deleted"})
}
```

- [ ] **Step 4: Modify `pkg/stockd/http/draft.go`**

```go
package http

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"stock/pkg/models"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/services/draft"
	"stock/pkg/stockd/utils"
)

func (h *handler) GetDraftToday(c *gin.Context) {
	u := auth.User(c)
	tradeDate := c.DefaultQuery("trade_date", time.Now().Format("20060102"))
	d, err := h.draftSvc.GetByDate(c.Request.Context(), u.ID, c.Query("ts_code"), tradeDate)
	if err != nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrInvalidParam)
		return
	}
	utils.HTTPRequestSuccess(c, 200, d)
}

func (h *handler) UpsertDraft(c *gin.Context) {
	var req models.UpsertDraftReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	d, err := h.draftSvc.Upsert(c.Request.Context(), draft.UpsertInput{
		UserID: u.ID, TsCode: req.TsCode, TradeDate: req.TradeDate,
		Open: req.Open, High: req.High, Low: req.Low, Close: req.Close,
	})
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, d)
}

func (h *handler) DeleteDraft(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	u := auth.User(c)
	if err := h.draftSvc.Delete(c.Request.Context(), u.ID, uint(id)); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "deleted"})
}
```

- [ ] **Step 5: Modify `pkg/cli/cmd/login.go`**

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"stock/pkg/cli/client"
	"stock/pkg/models"
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
		var me models.User
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
```

- [ ] **Step 6: Modify `pkg/cli/cmd/portfolio.go`**

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"stock/pkg/cli/client"
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
```

- [ ] **Step 7: Modify `pkg/stockd/http/auth_test.go`**

Update `TestLogin` and `TestLogin_BadPassword` to use `models.LoginReq` in request body and assert response contains `"username"` not `"userName"`.

Replace the two test bodies with:

```go
func TestLogin(t *testing.T) {
	r, gdb := setupAuthRouter(t)
	pw, _ := auth.HashPassword("secret")
	require.NoError(t, gdb.Create(&models.User{Username: "alice", PasswordHash: pw, Role: "user"}).Error)

	w := httptest.NewRecorder()
	body, _ := json.Marshal(models.LoginReq{Username: "alice", Password: "secret"})
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), `"username":"alice"`)
	assert.NotContains(t, w.Body.String(), `"userName"`)
}

func TestLogin_BadPassword(t *testing.T) {
	r, gdb := setupAuthRouter(t)
	pw, _ := auth.HashPassword("secret")
	require.NoError(t, gdb.Create(&models.User{Username: "alice", PasswordHash: pw, Role: "user"}).Error)

	w := httptest.NewRecorder()
	body, _ := json.Marshal(models.LoginReq{Username: "alice", Password: "wrong"})
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
}
```

Also add `import "encoding/json"` to the test file imports.

- [ ] **Step 8: Run backend tests**

Run: `go test -race ./pkg/stockd/http/... ./pkg/cli/...`
Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add pkg/stockd/http/auth.go pkg/stockd/http/me.go pkg/stockd/http/admin.go pkg/stockd/http/draft.go pkg/cli/cmd/login.go pkg/cli/cmd/portfolio.go pkg/stockd/http/auth_test.go
git commit -m "refactor(handler,cli): replace inline structs with pkg/models DTOs"
```

---

## Task 4: Backend contract tests

**Files:**
- Modify: `pkg/models/models_test.go`
- Modify: `pkg/cli/client/client_test.go`

- [ ] **Step 1: Add `TestModelJSONFields` to `pkg/models/models_test.go`**

Append to the end of the file:

```go
func TestModelJSONFields(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		val     any
		want    []string
		notWant []string
	}{
		{
			name: "User",
			val: models.User{
				ID: 1, Username: "alice", PasswordHash: "h", Role: "user",
				TushareToken: "tk", Disabled: false, CreatedAt: now, UpdatedAt: now,
			},
			want:    []string{"id", "username", "role", "tushareToken", "disabled", "createdAt", "updatedAt"},
			notWant: []string{"userName", "passwordHash", "userID"},
		},
		{
			name: "Portfolio",
			val:  models.Portfolio{ID: 1, UserID: 2, TsCode: "600519.SH", Note: "n", AddedAt: now},
			want:    []string{"id", "userId", "tsCode", "note", "addedAt"},
			notWant: []string{"userID"},
		},
		{
			name: "APIToken",
			val:  models.APIToken{ID: 1, UserID: 2, Name: "cli", TokenHash: "h", CreatedAt: now},
			want:    []string{"id", "userId", "name", "tokenHash", "createdAt"},
			notWant: []string{"userID"},
		},
		{
			name: "IntradayDraft",
			val:  models.IntradayDraft{ID: 1, UserID: 2, TsCode: "x", TradeDate: "20250513", UpdatedAt: now},
			want:    []string{"id", "userId", "tsCode", "tradeDate", "updatedAt"},
			notWant: []string{"userID"},
		},
		{
			name: "DailyBar",
			val:  models.DailyBar{TsCode: "x", TradeDate: "20250513", Open: 10, High: 11, Low: 9, Close: 10, Spreads: models.Spreads{OH: 1, HL: 2}},
			want: []string{"tsCode", "tradeDate", "open", "high", "low", "close", "spreads"},
		},
		{
			name: "JobRun",
			val:  models.JobRun{ID: 1, JobName: "daily-fetch", StartedAt: now, Status: "success"},
			want:    []string{"id", "jobName", "startedAt", "status"},
			notWant: []string{"createdAt"},
		},
		{
			name: "LoginReq",
			val:  models.LoginReq{Username: "alice", Password: "secret"},
			want: []string{"username", "password"},
		},
		{
			name: "IssueTokenResp",
			val:  models.IssueTokenResp{Token: "tk", Metadata: &models.APIToken{ID: 1, Name: "x"}},
			want: []string{"token", "metadata"},
		},
	}
	for _, c := range cases {
		b, err := json.Marshal(c.val)
		require.NoError(t, err, c.name)
		s := string(b)
		for _, w := range c.want {
			assert.Contains(t, s, fmt.Sprintf(`"%s"`, w), "%s should contain %s", c.name, w)
		}
		for _, nw := range c.notWant {
			assert.NotContains(t, s, fmt.Sprintf(`"%s"`, nw), "%s should not contain %s", c.name, nw)
		}
	}
}
```

Add to imports: `"encoding/json"`, `"fmt"`.

- [ ] **Step 2: Add portfolio alignment test to `pkg/cli/client/client_test.go`**

Append to the end of the file:

```go
func TestGET_PortfolioFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"requestID":"test","code":200,"message":"ok","data":[{"id":1,"userId":2,"tsCode":"600519.SH","note":"茅台","addedAt":"2025-05-13T10:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "stk_test")
	var out []*models.Portfolio
	require.NoError(t, c.GET("/portfolio", &out))
	require.Len(t, out, 1)
	assert.Equal(t, uint(1), out[0].ID)
	assert.Equal(t, uint(2), out[0].UserID)
	assert.Equal(t, "600519.SH", out[0].TsCode)
	assert.Equal(t, "茅台", out[0].Note)
	assert.False(t, out[0].AddedAt.IsZero())
}
```

Add `"stock/pkg/models"` to imports.

- [ ] **Step 3: Run all backend tests**

Run: `go test -race ./pkg/models/... ./pkg/stockd/http/... ./pkg/cli/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/models/models_test.go pkg/cli/client/client_test.go
git commit -m "test(models,cli): add JSON field contract tests and portfolio alignment test"
```

---

## Task 5: Install frontend dependencies and create SCSS infrastructure

**Files:**
- Modify: `web/package.json`
- Create: `web/src/assets/css/index.scss`
- Create: `web/src/assets/css/element-reset.scss`
- Create: `web/src/assets/css/mixin.scss`
- Create: `web/src/assets/image/logo-mini.svg`

- [ ] **Step 1: Install dependencies**

Run:
```bash
cd /root/code/github/travelliu/stock/web
npm install vue-i18n@9 sass @element-plus/icons-vue
```

Expected: packages installed, `package.json` and `package-lock.json` updated.

- [ ] **Step 2: Create `web/src/assets/css/index.scss`**

```scss
:root {
  --color-up: #f56c6c;
  --color-down: #67c23a;
  --color-flat: #909399;
  --sider-bg: #20242b;
  --sider-bg-hover: #313741;
  --sider-text: #acb3bf;
  --sider-active: #ffd04b;
  --content-bg: #f5f7fa;
  --card-radius: 4px;
}

body {
  margin: 0;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
  background: var(--content-bg);
  color: #303133;
}

.g-up {
  color: var(--color-up);
}

.g-down {
  color: var(--color-down);
}

.g-flat {
  color: var(--color-flat);
}

.g-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
```

- [ ] **Step 3: Create `web/src/assets/css/element-reset.scss`**

```scss
.el-card {
  border-radius: var(--card-radius);
}

.el-button {
  border-radius: var(--card-radius);
}
```

- [ ] **Step 4: Create `web/src/assets/css/mixin.scss`**

```scss
@mixin scrollbar() {
  &::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }
  &::-webkit-scrollbar-thumb {
    background: #c0c4cc;
    border-radius: 3px;
  }
}
```

- [ ] **Step 5: Create `web/src/assets/image/logo-mini.svg`**

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" width="32" height="32">
  <rect x="8" y="24" width="10" height="32" fill="#409EFF"/>
  <rect x="27" y="14" width="10" height="42" fill="#409EFF"/>
  <rect x="46" y="8" width="10" height="48" fill="#409EFF"/>
</svg>
```

- [ ] **Step 6: Commit**

```bash
git add web/package.json web/package-lock.json web/src/assets/
git commit -m "chore(web): add vue-i18n, sass, icons-vue and SCSS infrastructure"
```

---

## Task 6: Create frontend type definitions, utils, and i18n

**Files:**
- Create: `web/src/types/api.ts`
- Create: `web/src/utils/message.ts`
- Create: `web/src/utils/storage.ts`
- Create: `web/src/utils/format.ts`
- Create: `web/src/intl/index.ts`
- Create: `web/src/intl/lang.ts`
- Create: `web/src/intl/langs/zh/index.ts`
- Create: `web/src/intl/langs/en/index.ts`

- [ ] **Step 1: Create `web/src/types/api.ts`**

```ts
// keep in sync with pkg/models
// @see pkg/models/foo.go::Foo

export interface User {
  id: number
  username: string
  role: string
  tushareToken: string
  disabled: boolean
  createdAt: string
  updatedAt: string
}

export interface Portfolio {
  id: number
  userId: number
  tsCode: string
  note: string
  addedAt: string
}

export interface PortfolioReq {
  tsCode: string
  note?: string
}

export interface Stock {
  tsCode: string
  code: string
  name: string
  area: string
  industry: string
  market: string
  exchange: string
  listDate: string
  delisted: boolean
  updatedAt: string
}

export interface Spreads {
  oh: number
  ol: number
  hl: number
  oc: number
  hc: number
  lc: number
}

export interface DailyBar {
  tsCode: string
  tradeDate: string
  open: number
  high: number
  low: number
  close: number
  vol: number
  amount: number
  spreads: Spreads
}

export interface AnalysisResult {
  tsCode: string
  stockName: string
  yesterdayClose?: number
  windows: string[]
  openPrice?: number
  actualHigh?: number
  actualLow?: number
  actualClose?: number
  windowMeans: Record<string, Record<string, number | null>>
  compositeMeans: Record<string, number>
  modelTable: ModelTable
  referenceTable: ReferenceTable
}

export interface ModelTable {
  headers: string[]
  rows: string[][]
}

export interface ReferenceTable {
  headers: string[]
  rows: string[][]
}

export interface IntradayDraft {
  id: number
  userId: number
  tsCode: string
  tradeDate: string
  open?: number
  high?: number
  low?: number
  close?: number
  updatedAt: string
}

export interface APIToken {
  id: number
  userId: number
  name: string
  tokenHash: string
  lastUsedAt: string | null
  expiresAt: string | null
  createdAt: string
}

export interface JobRun {
  id: number
  jobName: string
  startedAt: string
  finishedAt: string | null
  status: string
  message: string
}

export interface LoginReq {
  username: string
  password: string
}

export interface ChangePasswordReq {
  old: string
  new: string
}

export interface SetTushareTokenReq {
  token: string
}

export interface IssueTokenReq {
  name: string
  expiresAt?: string
}

export interface IssueTokenResp {
  token: string
  metadata: APIToken
}

export interface CreateUserReq {
  username: string
  password: string
  role: string
  tushareToken?: string
}

export interface PatchUserReq {
  role?: string
  disabled?: boolean
  tushareToken?: string
}
```

- [ ] **Step 2: Create `web/src/utils/message.ts`**

```ts
import { ElMessage } from 'element-plus'

const recent = new Map<string, number>()

export function wMessage(type: 'success' | 'error' | 'warning' | 'info', message: string) {
  const key = `${type}:${message}`
  const last = recent.get(key)
  const now = Date.now()
  if (last && now - last < 3000) {
    return
  }
  recent.set(key, now)
  ElMessage[type](message)
}
```

- [ ] **Step 3: Create `web/src/utils/storage.ts`**

```ts
export const storage = {
  get(key: string): string | null {
    return localStorage.getItem(key)
  },
  set(key: string, value: string) {
    localStorage.setItem(key, value)
  },
  remove(key: string) {
    localStorage.removeItem(key)
  },
}
```

- [ ] **Step 4: Create `web/src/utils/format.ts`**

```ts
export function fmtPrice(n: number): string {
  return n.toFixed(2)
}

export function fmtPct(cur: number, prev: number): string {
  if (prev === 0) return '0.00%'
  const v = ((cur - prev) / prev) * 100
  return `${v >= 0 ? '+' : ''}${v.toFixed(2)}%`
}

export function priceClass(cur: number, prev: number): string {
  if (cur > prev) return 'g-up'
  if (cur < prev) return 'g-down'
  return 'g-flat'
}
```

- [ ] **Step 5: Create `web/src/intl/lang.ts`**

```ts
export const Langs = {
  zh: 'zh',
  en: 'en',
} as const

export type Lang = typeof Langs[keyof typeof Langs]

export const ElementLangs: Record<Lang, string> = {
  zh: 'zh-CN',
  en: 'en',
}
```

- [ ] **Step 6: Create `web/src/intl/langs/zh/index.ts`**

```ts
export default {
  common: {
    confirm: '确认',
    cancel: '取消',
    save: '保存',
    delete: '删除',
    add: '添加',
    search: '搜索',
    close: '关闭',
    loading: '加载中...',
    error: '错误',
    success: '成功',
    empty: '暂无数据',
  },
  menu: {
    stock: '股票',
    profile: '个人中心',
    admin: '管理',
    users: '用户管理',
    sync: '数据同步',
    logout: '退出登录',
  },
  login: {
    title: 'Stock Analysis',
    username: '用户名',
    password: '密码',
    login: '登录',
    failed: '登录失败',
  },
  stockList: {
    title: '我的持仓',
    addStock: '添加股票',
    code: '代码',
    note: '备注',
    action: '操作',
    detail: '详情',
    selectStock: '请选择股票',
    added: '添加成功',
  },
  stockDetail: {
    basic: '基础与价差',
    statistics: '详细统计',
    industry: '行业',
    listDate: '上市日期',
    lastClose: '最近收盘',
    spreads: '价差分布',
    dailyBars: '日线数据',
    date: '日期',
    open: '开盘',
    high: '最高',
    low: '最低',
    close: '收盘',
    vol: '成交量',
    oh: '高-开',
    ol: '开-低',
    hl: '高-低',
    oc: '开-收',
    hc: '高-收',
    lc: '低-收',
    draft: '今日草稿',
    draftSave: '保存草稿',
    draftApply: '应用',
    modelTable: '价差模型',
    tradePlan: '交易计划',
    window: '时段',
  },
  profile: {
    title: '个人中心',
    info: '个人信息',
    username: '用户名',
    role: '角色',
    registeredAt: '注册时间',
    changePassword: '修改密码',
    oldPassword: '原密码',
    newPassword: '新密码',
    passwordChanged: '密码已修改',
    tushareToken: 'Tushare Token',
    tokenSaved: 'Token 已保存',
    apiTokens: 'API Tokens',
    tokenName: '名称',
    createdAt: '创建时间',
    createToken: '新建 Token',
    tokenCreated: '创建成功',
    revoke: '撤销',
    revoked: '已撤销',
    copyPrompt: '请立即复制，只会显示一次',
  },
  admin: {
    users: '用户管理',
    sync: '数据同步',
    createUser: '新建用户',
    status: '状态',
    enabled: '正常',
    disabled: '禁用',
    toggle: '切换',
    syncStocklist: '同步股票列表',
    syncBars: '同步行情数据',
    job: '任务',
    startedAt: '开始时间',
    finishedAt: '结束时间',
  },
}
```

- [ ] **Step 7: Create `web/src/intl/langs/en/index.ts`**

```ts
export default {
  common: {},
  menu: {},
  login: {},
  stockList: {},
  stockDetail: {},
  profile: {},
  admin: {},
}
```

- [ ] **Step 8: Create `web/src/intl/index.ts`**

```ts
import { createI18n } from 'vue-i18n'
import zh from './langs/zh/index'
import en from './langs/en/index'
import { storage } from '@/utils/storage'

const saved = storage.get('lang') || 'zh'

export const i18n = createI18n({
  legacy: false,
  locale: saved,
  fallbackLocale: 'zh',
  messages: { zh, en },
})
```

- [ ] **Step 9: Commit**

```bash
git add web/src/types web/src/utils web/src/intl
git commit -m "feat(web): add api types, utils, and i18n infrastructure"
```

---

## Task 7: Create API layer and stores

**Files:**
- Create: `web/src/apis/axios.ts`
- Create: `web/src/apis/auth.ts`
- Create: `web/src/apis/me.ts`
- Create: `web/src/apis/portfolio.ts`
- Create: `web/src/apis/stocks.ts`
- Create: `web/src/apis/analysis.ts`
- Create: `web/src/apis/draft.ts`
- Create: `web/src/apis/admin.ts`
- Create: `web/src/stores/auth.ts`
- Create: `web/src/stores/lang.ts`

- [ ] **Step 1: Create `web/src/apis/axios.ts`**

```ts
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useLangStore } from '@/stores/lang'
import router from '@/router'
import { wMessage } from '@/utils/message'

const $http = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

$http.interceptors.request.use((cfg) => {
  const lang = useLangStore().lang
  cfg.headers.lang = lang
  return cfg
})

$http.interceptors.response.use(
  (res) => {
    if (res.data && typeof res.data.code === 'number') {
      if (res.data.code !== 200) {
        wMessage('error', res.data.message || 'unknown error')
        return Promise.reject(new Error(res.data.message || 'unknown error'))
      }
      return res.data.data
    }
    return res.data
  },
  (err) => {
    if (err.response?.status === 401) {
      useAuthStore().logout()
      router.push('/login')
    } else {
      wMessage('error', err.message || '网络错误')
    }
    return Promise.reject(err)
  }
)

export { $http }
```

- [ ] **Step 2: Create `web/src/apis/auth.ts`**

```ts
import type { User, LoginReq } from '@/types/api'
import { $http } from './axios'

export const login = (req: LoginReq): Promise<User> => $http.post('/auth/login', req)
export const logout = (): Promise<void> => $http.post('/auth/logout')
export const me = (): Promise<User> => $http.get('/auth/me')
```

- [ ] **Step 3: Create `web/src/apis/me.ts`**

```ts
import type { ChangePasswordReq, SetTushareTokenReq, IssueTokenReq, IssueTokenResp, APIToken } from '@/types/api'
import { $http } from './axios'

export const changePassword = (req: ChangePasswordReq): Promise<void> => $http.post('/me/password', req)
export const setTushareToken = (req: SetTushareTokenReq): Promise<void> => $http.patch('/me/tushare_token', req)
export const listTokens = (): Promise<APIToken[]> => $http.get('/me/tokens')
export const issueToken = (req: IssueTokenReq): Promise<IssueTokenResp> => $http.post('/me/tokens', req)
export const revokeToken = (id: number): Promise<void> => $http.delete(`/me/tokens/${id}`)
```

- [ ] **Step 4: Create `web/src/apis/portfolio.ts`**

```ts
import type { Portfolio, PortfolioReq } from '@/types/api'
import { $http } from './axios'

export const listPortfolio = (): Promise<Portfolio[]> => $http.get('/portfolio')
export const addPortfolio = (req: PortfolioReq): Promise<void> => $http.post('/portfolio', req)
export const removePortfolio = (tsCode: string): Promise<void> => $http.delete(`/portfolio/${tsCode}`)
export const updatePortfolioNote = (tsCode: string, req: PortfolioReq): Promise<void> =>
  $http.patch(`/portfolio/${tsCode}`, req)
```

- [ ] **Step 5: Create `web/src/apis/stocks.ts`**

```ts
import type { Stock, DailyBar } from '@/types/api'
import { $http } from './axios'

export const searchStocks = (q: string, limit = 20): Promise<Stock[]> =>
  $http.get('/stocks', { params: { q, limit } })
export const getStock = (tsCode: string): Promise<Stock> => $http.get(`/stocks/${tsCode}`)
export const queryBars = (tsCode: string, from?: string, to?: string): Promise<DailyBar[]> =>
  $http.get(`/bars/${tsCode}`, { params: { from, to } })
```

- [ ] **Step 6: Create `web/src/apis/analysis.ts`**

```ts
import type { AnalysisResult } from '@/types/api'
import { $http } from './axios'

export interface AnalysisParams {
  actualOpen?: number
  actualHigh?: number
  actualLow?: number
  actualClose?: number
  withDraft?: boolean
}

export const getAnalysis = (tsCode: string, params?: AnalysisParams): Promise<AnalysisResult> => {
  const qs = new URLSearchParams()
  if (params?.actualOpen !== undefined) qs.set('actual_open', String(params.actualOpen))
  if (params?.actualHigh !== undefined) qs.set('actual_high', String(params.actualHigh))
  if (params?.actualLow !== undefined) qs.set('actual_low', String(params.actualLow))
  if (params?.actualClose !== undefined) qs.set('actual_close', String(params.actualClose))
  qs.set('with_draft', String(params?.withDraft ?? true))
  return $http.get(`/analysis/${tsCode}?${qs.toString()}`)
}
```

- [ ] **Step 7: Create `web/src/apis/draft.ts`**

```ts
import type { IntradayDraft } from '@/types/api'
import { $http } from './axios'

export const getDraftToday = (tsCode: string, tradeDate?: string): Promise<IntradayDraft> =>
  $http.get('/drafts/today', { params: { ts_code: tsCode, trade_date: tradeDate } })
export const upsertDraft = (body: Record<string, unknown>): Promise<IntradayDraft> =>
  $http.put('/drafts', body)
export const deleteDraft = (id: number): Promise<void> => $http.delete(`/drafts/${id}`)
```

- [ ] **Step 8: Create `web/src/apis/admin.ts`**

```ts
import type { User, JobRun, CreateUserReq, PatchUserReq } from '@/types/api'
import { $http } from './axios'

export const listUsers = (): Promise<User[]> => $http.get('/admin/users')
export const createUser = (req: CreateUserReq): Promise<User> => $http.post('/admin/users', req)
export const patchUser = (id: number, req: PatchUserReq): Promise<void> => $http.patch(`/admin/users/${id}`, req)
export const deleteUser = (id: number): Promise<void> => $http.delete(`/admin/users/${id}`)
export const syncStocklist = (): Promise<void> => $http.post('/admin/stocks/sync')
export const syncBars = (): Promise<void> => $http.post('/admin/bars/sync')
export const jobStatus = (job: string): Promise<JobRun> => $http.get('/admin/sync/status', { params: { job } })
```

- [ ] **Step 9: Create `web/src/stores/auth.ts`**

```ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '@/types/api'
import * as authApi from '@/apis/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)

  async function fetchMe() {
    try {
      user.value = await authApi.me()
    } catch {
      user.value = null
    }
  }

  async function login(username: string, password: string) {
    user.value = await authApi.login({ username, password })
    return user.value
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // ignore
    }
    user.value = null
  }

  return { user, fetchMe, login, logout }
})
```

- [ ] **Step 10: Create `web/src/stores/lang.ts`**

```ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { storage } from '@/utils/storage'
import type { Lang } from '@/intl/lang'

export const useLangStore = defineStore('lang', () => {
  const lang = ref<Lang>((storage.get('lang') as Lang) || 'zh')

  function setLang(v: Lang) {
    lang.value = v
    storage.set('lang', v)
  }

  return { lang, setLang }
})
```

- [ ] **Step 11: Commit**

```bash
git add web/src/apis web/src/stores
git commit -m "feat(web): add API layer and stores"
```

---

## Task 8: Create global components

**Files:**
- Create: `web/src/components/GIcon.vue`
- Create: `web/src/components/GEllipsis.vue`

- [ ] **Step 1: Create `web/src/components/GIcon.vue`**

```vue
<script setup lang="ts">
import { computed } from 'vue'
import * as Icons from '@element-plus/icons-vue'

const props = defineProps<{ name: string }>()
const iconComponent = computed(() => (Icons as Record<string, unknown>)[props.name])
</script>

<template>
  <el-icon>
    <component :is="iconComponent" />
  </el-icon>
</template>
```

- [ ] **Step 2: Create `web/src/components/GEllipsis.vue`**

```vue
<script setup lang="ts">
defineProps<{ line?: number }>()
</script>

<template>
  <div :style="{ overflow: 'hidden', textOverflow: 'ellipsis', display: '-webkit-box', WebkitLineClamp: line || 1, WebkitBoxOrient: 'vertical' }">
    <slot />
  </div>
</template>
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/GIcon.vue web/src/components/GEllipsis.vue
git commit -m "feat(web): add global g-icon and g-ellipsis components"
```

---

## Task 9: Create layout components and root App

**Files:**
- Create: `web/src/components/ConsoleMenu.vue`
- Create: `web/src/components/UserMenu.vue`
- Modify: `web/src/App.vue`

- [ ] **Step 1: Create `web/src/components/ConsoleMenu.vue`**

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLangStore } from '@/stores/lang'
import { i18n } from '@/intl'
import GIcon from './GIcon.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const lang = useLangStore()

const activeIndex = computed(() => route.path)

function switchLang() {
  const next = lang.lang === 'zh' ? 'en' : 'zh'
  lang.setLang(next)
  i18n.global.locale.value = next
}
</script>

<template>
  <div class="console-menu">
    <div class="logo" @click="router.push('/')">
      <img src="@/assets/image/logo-mini.svg" alt="logo" />
    </div>

    <div class="menu-center">
      <div class="menu-item" :class="{ active: activeIndex.startsWith('/stocks') }" @click="router.push('/stocks')">
        <GIcon name="TrendCharts" />
        <span>{{ $t('menu.stock') }}</span>
      </div>
    </div>

    <div class="menu-bottom">
      <div v-if="auth.user" class="menu-item" :class="{ active: activeIndex.startsWith('/profile') }" @click="router.push('/profile')">
        <GIcon name="User" />
        <span>{{ $t('menu.profile') }}</span>
      </div>
      <div class="menu-item" @click="switchLang">
        <GIcon name="Globe" />
        <span>{{ lang.lang === 'zh' ? '中' : 'EN' }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.console-menu {
  width: 65px;
  height: 100vh;
  background: var(--sider-bg);
  color: var(--sider-text);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
  flex-shrink: 0;
}
.logo {
  padding: 12px 0;
  cursor: pointer;
}
.menu-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 24px;
  gap: 8px;
}
.menu-bottom {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding-bottom: 8px;
}
.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 4px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  width: 52px;
  transition: background 0.2s;
}
.menu-item:hover {
  background: var(--sider-bg-hover);
}
.menu-item.active {
  color: var(--sider-active);
}
</style>
```

- [ ] **Step 2: Create `web/src/components/UserMenu.vue`**

```vue
<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import GIcon from './GIcon.vue'

const router = useRouter()
const auth = useAuthStore()

function doLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="user-menu">
    <el-dropdown trigger="click">
      <div class="user-trigger">
        <el-avatar :size="28" :icon="'UserFilled'" />
        <span class="username">{{ auth.user?.username }}</span>
        <GIcon name="ArrowDown" />
      </div>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="router.push('/profile')">{{ $t('menu.profile') }}</el-dropdown-item>
          <el-dropdown-item divided @click="doLogout">{{ $t('menu.logout') }}</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<style scoped lang="scss">
.user-menu {
  padding: 0 16px;
}
.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #303133;
}
.username {
  font-size: 14px;
}
</style>
```

- [ ] **Step 3: Modify `web/src/App.vue`**

```vue
<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ConsoleMenu from '@/components/ConsoleMenu.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

onMounted(() => {
  auth.fetchMe().then(() => {
    if (!auth.user && route.meta.requiresAuth) {
      router.push('/login')
    }
  })
})
</script>

<template>
  <div class="app-root">
    <template v-if="auth.user && route.path !== '/login'">
      <ConsoleMenu />
      <div class="main">
        <router-view />
      </div>
    </template>
    <template v-else>
      <router-view />
    </template>
  </div>
</template>

<style scoped lang="scss">
.app-root {
  display: flex;
  min-height: 100vh;
}
.main {
  flex: 1;
  padding: 16px;
  overflow: auto;
}
</style>
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/ConsoleMenu.vue web/src/components/UserMenu.vue web/src/App.vue
git commit -m "feat(web): add layout components and App.vue"
```

---

## Task 10: Create LoginView and StockListView

**Files:**
- Create: `web/src/views/LoginView.vue`
- Create: `web/src/views/StockListView.vue`

- [ ] **Step 1: Create `web/src/views/LoginView.vue`**

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const username = ref('')
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!username.value || !password.value) {
    wMessage('warning', '请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await authStore.login(username.value, password.value)
    router.push('/stocks')
  } catch (e: any) {
    wMessage('error', e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <el-card class="login-card" shadow="always">
      <h2 class="title">{{ $t('login.title') }}</h2>
      <el-form @submit.prevent="handleLogin">
        <el-form-item>
          <el-input v-model="username" :placeholder="$t('login.username')" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="password" type="password" :placeholder="$t('login.password')" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width: 100%" :loading="loading" @click="handleLogin">
            {{ $t('login.login') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  width: 100vw;
  background: var(--content-bg);
}
.login-card {
  width: 360px;
}
.title {
  text-align: center;
  margin-bottom: 24px;
  font-weight: 500;
}
</style>
```

- [ ] **Step 2: Create `web/src/views/StockListView.vue`**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { usePortfolioStore } from '@/stores/portfolio'
import { searchStocks } from '@/apis/stocks'
import type { Stock } from '@/types/api'

const router = useRouter()
const portfolioStore = usePortfolioStore()
const showAdd = ref(false)
const selectedStock = ref('')
const note = ref('')
const stockOptions = ref<{ value: string; label: string }[]>([])
const loadingAdd = ref(false)

onMounted(() => {
  portfolioStore.fetch()
})

async function searchStockOptions(query: string) {
  if (!query) {
    stockOptions.value = []
    return
  }
  try {
    const list = await searchStocks(query, 20)
    stockOptions.value = list.map((s: Stock) => ({
      value: s.tsCode,
      label: `${s.tsCode} ${s.name}`,
    }))
  } catch {
    stockOptions.value = []
  }
}

async function doAdd() {
  if (!selectedStock.value) {
    wMessage('warning', $t('stockList.selectStock'))
    return
  }
  loadingAdd.value = true
  try {
    await portfolioStore.add(selectedStock.value, note.value)
    wMessage('success', $t('stockList.added'))
    showAdd.value = false
    selectedStock.value = ''
    note.value = ''
  } finally {
    loadingAdd.value = false
  }
}

function goDetail(tsCode: string) {
  router.push(`/stocks/${tsCode}`)
}
</script>

<template>
  <div>
    <div class="header-bar">
      <h2>{{ $t('stockList.title') }}</h2>
      <el-button type="primary" @click="showAdd = true">{{ $t('stockList.addStock') }}</el-button>
    </div>

    <el-table :data="portfolioStore.items" style="margin-top: 16px">
      <el-table-column prop="tsCode" :label="$t('stockList.code')" />
      <el-table-column prop="note" :label="$t('stockList.note')" />
      <el-table-column :label="$t('stockList.action')" width="140">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row.tsCode)">{{ $t('stockList.detail') }}</el-button>
          <el-button link type="danger" @click="portfolioStore.remove(row.tsCode)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showAdd" :title="$t('stockList.addStock')" width="400px">
      <el-form @submit.prevent="doAdd">
        <el-form-item :label="$t('stockList.code')">
          <el-select-v2
            v-model="selectedStock"
            :options="stockOptions"
            :placeholder="$t('stockList.selectStock')"
            clearable
            filterable
            remote
            :remote-method="searchStockOptions"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('stockList.note')">
          <el-input v-model="note" :placeholder="$t('common.empty')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loadingAdd" @click="doAdd">{{ $t('common.add') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

- [ ] **Step 3: Update portfolio store to use new API and types**

Modify `web/src/stores/portfolio.ts`:

```ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Portfolio } from '@/types/api'
import { listPortfolio, addPortfolio, removePortfolio } from '@/apis/portfolio'

export const usePortfolioStore = defineStore('portfolio', () => {
  const items = ref<Portfolio[]>([])

  async function fetch() {
    items.value = await listPortfolio()
  }

  async function add(tsCode: string, note: string) {
    await addPortfolio({ tsCode, note })
    await fetch()
  }

  async function remove(tsCode: string) {
    await removePortfolio(tsCode)
    await fetch()
  }

  return { items, fetch, add, remove }
})
```

- [ ] **Step 4: Run frontend type check**

Run: `cd /root/code/github/travelliu/stock/web && npx vue-tsc --noEmit`
Expected: PASS (no errors)

- [ ] **Step 5: Commit**

```bash
git add web/src/views/LoginView.vue web/src/views/StockListView.vue web/src/stores/portfolio.ts
git commit -m "feat(web): add LoginView and StockListView"
```

---

## Task 11: Create StockDetailView and BasicTab

**Files:**
- Create: `web/src/views/stock/StockDetailView.vue`
- Create: `web/src/views/stock/BasicTab.vue`
- Create: `web/src/components/StockBasicCard.vue`
- Create: `web/src/components/DailyBarTable.vue`
- Create: `web/src/components/SpreadHistogram.vue`

- [ ] **Step 1: Create `web/src/components/StockBasicCard.vue`**

```vue
<script setup lang="ts">
import { computed } from 'vue'
import type { Stock, DailyBar } from '@/types/api'
import { fmtPrice, fmtPct, priceClass } from '@/utils/format'

const props = defineProps<{ stock: Stock; lastBar?: DailyBar }>()

const changeClass = computed(() => {
  if (!props.lastBar) return 'g-flat'
  return priceClass(props.lastBar.close, props.lastBar.open)
})

const changePct = computed(() => {
  if (!props.lastBar) return '--'
  return fmtPct(props.lastBar.close, props.lastBar.open)
})
</script>

<template>
  <el-card>
    <div class="stock-card">
      <div class="name-block">
        <h3>{{ stock.name }} <span class="code">{{ stock.tsCode }}</span></h3>
        <div class="meta">
          <el-tag size="small">{{ stock.industry }}</el-tag>
          <span>{{ $t('stockDetail.listDate') }}: {{ stock.listDate }}</span>
        </div>
      </div>
      <div v-if="lastBar" class="price-block">
        <div class="price" :class="changeClass">{{ fmtPrice(lastBar.close) }}</div>
        <div class="pct" :class="changeClass">{{ changePct }}</div>
      </div>
    </div>
  </el-card>
</template>

<style scoped lang="scss">
.stock-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.name-block h3 {
  margin: 0 0 8px 0;
}
.code {
  font-size: 14px;
  color: #909399;
  font-weight: normal;
}
.meta {
  display: flex;
  gap: 12px;
  align-items: center;
  font-size: 13px;
  color: #606266;
}
.price-block {
  text-align: right;
}
.price {
  font-size: 28px;
  font-weight: bold;
}
.pct {
  font-size: 14px;
}
</style>
```

- [ ] **Step 2: Create `web/src/components/DailyBarTable.vue`**

```vue
<script setup lang="ts">
import type { DailyBar } from '@/types/api'
import { fmtPrice, priceClass } from '@/utils/format'

const props = defineProps<{ bars: DailyBar[] }>()

const spreadKeys = ['oh', 'ol', 'hl', 'oc', 'hc', 'lc'] as const
</script>

<template>
  <el-table :data="bars" height="500" size="small">
    <el-table-column prop="tradeDate" :label="$t('stockDetail.date')" width="100" />
    <el-table-column prop="open" :label="$t('stockDetail.open')" width="80">
      <template #default="{ row }">
        <span :class="priceClass(row.close, row.open)">{{ fmtPrice(row.open) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="high" :label="$t('stockDetail.high')" width="80">
      <template #default="{ row }">
        <span :class="priceClass(row.close, row.open)">{{ fmtPrice(row.high) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="low" :label="$t('stockDetail.low')" width="80">
      <template #default="{ row }">
        <span :class="priceClass(row.close, row.open)">{{ fmtPrice(row.low) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="close" :label="$t('stockDetail.close')" width="80">
      <template #default="{ row }">
        <span :class="priceClass(row.close, row.open)">{{ fmtPrice(row.close) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="vol" :label="$t('stockDetail.vol')" width="100" />
    <el-table-column v-for="k in spreadKeys" :key="k" :label="$t(`stockDetail.${k}`)" width="80">
      <template #default="{ row }">
        {{ fmtPrice(row.spreads[k]) }}
      </template>
    </el-table-column>
  </el-table>
</template>
```

- [ ] **Step 3: Create `web/src/components/SpreadHistogram.vue`**

```vue
<script setup lang="ts">
import { computed } from 'vue'
import type { DailyBar } from '@/types/api'

const props = defineProps<{ bars: DailyBar[] }>()

const labels = [
  { key: 'oh' as const, name: '高-开' },
  { key: 'ol' as const, name: '开-低' },
  { key: 'hl' as const, name: '高-低' },
  { key: 'oc' as const, name: '开-收' },
  { key: 'hc' as const, name: '高-收' },
  { key: 'lc' as const, name: '低-收' },
]

const stats = computed(() => {
  if (!props.bars.length) return []
  const max = Math.max(...labels.map(l => Math.max(...props.bars.map(b => b.spreads[l.key]))))
  return labels.map(l => {
    const avg = props.bars.reduce((s, b) => s + b.spreads[l.key], 0) / props.bars.length
    const pct = max > 0 ? (avg / max) * 100 : 0
    return { name: l.name, avg, pct }
  })
})
</script>

<template>
  <el-card>
    <template #header>{{ $t('stockDetail.spreads') }}</template>
    <div class="histogram">
      <div v-for="s in stats" :key="s.name" class="bar-row">
        <span class="label">{{ s.name }}</span>
        <el-progress :percentage="Math.round(s.pct)" :stroke-width="16" :show-text="false" />
        <span class="value">{{ s.avg.toFixed(2) }}</span>
      </div>
    </div>
  </el-card>
</template>

<style scoped lang="scss">
.histogram {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.bar-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.label {
  width: 60px;
  font-size: 13px;
  flex-shrink: 0;
}
.value {
  width: 50px;
  text-align: right;
  font-size: 13px;
  flex-shrink: 0;
}
</style>
```

- [ ] **Step 4: Create `web/src/views/stock/BasicTab.vue`**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { getStock, queryBars } from '@/apis/stocks'
import type { Stock, DailyBar } from '@/types/api'
import StockBasicCard from '@/components/StockBasicCard.vue'
import SpreadHistogram from '@/components/SpreadHistogram.vue'
import DailyBarTable from '@/components/DailyBarTable.vue'

const props = defineProps<{ tsCode: string }>()

const stock = ref<Stock | null>(null)
const bars = ref<DailyBar[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [s, b] = await Promise.all([
      getStock(props.tsCode),
      queryBars(props.tsCode),
    ])
    stock.value = s
    bars.value = b.slice(-30).reverse()
  } catch (e: any) {
    wMessage('error', e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <StockBasicCard v-if="stock" :stock="stock" :last-bar="bars[0]" />
    <div style="margin-top: 16px">
      <SpreadHistogram :bars="bars" />
    </div>
    <div style="margin-top: 16px">
      <DailyBarTable :bars="bars" />
    </div>
  </div>
</template>
```

- [ ] **Step 5: Create `web/src/views/stock/StockDetailView.vue`**

```vue
<script setup lang="ts">
import { useRoute } from 'vue-router'
import BasicTab from './BasicTab.vue'
import StatisticsTab from './StatisticsTab.vue'

const route = useRoute()
const tsCode = route.params.tsCode as string
</script>

<template>
  <div>
    <el-tabs type="border-card">
      <el-tab-pane :label="$t('stockDetail.basic')">
        <BasicTab :ts-code="tsCode" />
      </el-tab-pane>
      <el-tab-pane :label="$t('stockDetail.statistics')">
        <StatisticsTab :ts-code="tsCode" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
```

- [ ] **Step 6: Commit**

```bash
git add web/src/views/stock/StockDetailView.vue web/src/views/stock/BasicTab.vue web/src/components/StockBasicCard.vue web/src/components/DailyBarTable.vue web/src/components/SpreadHistogram.vue
git commit -m "feat(web): add StockDetailView shell and BasicTab with subcomponents"
```

---

## Task 12: Create StatisticsTab and subcomponents

**Files:**
- Create: `web/src/views/stock/StatisticsTab.vue`
- Create: `web/src/components/DraftFormBlock.vue`
- Create: `web/src/components/SpreadModelTable.vue`
- Create: `web/src/components/TradePlanTable.vue`

- [ ] **Step 1: Create `web/src/components/DraftFormBlock.vue`**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { getDraftToday, upsertDraft, deleteDraft } from '@/apis/draft'

const props = defineProps<{ tsCode: string }>()
const emit = defineEmits<{
  apply: [params: { open?: number; high?: number; low?: number; close?: number }]
}>()

const form = ref<{ open?: number; high?: number; low?: number; close?: number }>({})
const draftId = ref<number | null>(null)
const loading = ref(false)

async function load() {
  try {
    const d = await getDraftToday(props.tsCode)
    draftId.value = d.id
    form.value = {
      open: d.open ?? undefined,
      high: d.high ?? undefined,
      low: d.low ?? undefined,
      close: d.close ?? undefined,
    }
  } catch {
    draftId.value = null
    form.value = {}
  }
}

async function save() {
  loading.value = true
  try {
    const today = new Date().toISOString().slice(0, 10).replace(/-/g, '')
    const body: Record<string, unknown> = { tsCode: props.tsCode, tradeDate: today }
    if (form.value.open !== undefined) body.open = form.value.open
    if (form.value.high !== undefined) body.high = form.value.high
    if (form.value.low !== undefined) body.low = form.value.low
    if (form.value.close !== undefined) body.close = form.value.close
    await upsertDraft(body)
    wMessage('success', '草稿已保存')
    await load()
  } catch (e: any) {
    wMessage('error', e.message || '保存失败')
  } finally {
    loading.value = false
  }
}

async function clear() {
  if (!draftId.value) {
    form.value = {}
    return
  }
  try {
    await deleteDraft(draftId.value)
    draftId.value = null
    form.value = {}
    wMessage('success', '草稿已清除')
  } catch (e: any) {
    wMessage('error', e.message || '清除失败')
  }
}

function apply() {
  emit('apply', { ...form.value })
}

onMounted(load)
</script>

<template>
  <el-card>
    <template #header>{{ $t('stockDetail.draft') }}</template>
    <el-form inline @submit.prevent="save">
      <el-form-item :label="$t('stockDetail.open')">
        <el-input-number v-model="form.open" :precision="2" :controls="false" />
      </el-form-item>
      <el-form-item :label="$t('stockDetail.high')">
        <el-input-number v-model="form.high" :precision="2" :controls="false" />
      </el-form-item>
      <el-form-item :label="$t('stockDetail.low')">
        <el-input-number v-model="form.low" :precision="2" :controls="false" />
      </el-form-item>
      <el-form-item :label="$t('stockDetail.close')">
        <el-input-number v-model="form.close" :precision="2" :controls="false" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="save">{{ $t('stockDetail.draftSave') }}</el-button>
        <el-button @click="apply">{{ $t('stockDetail.draftApply') }}</el-button>
        <el-button @click="clear">{{ $t('common.delete') }}</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>
```

- [ ] **Step 2: Create `web/src/components/SpreadModelTable.vue`**

```vue
<script setup lang="ts">
import type { AnalysisResult } from '@/types/api'

const props = defineProps<{ result?: AnalysisResult | null }>()
</script>

<template>
  <el-card v-if="result">
    <template #header>{{ $t('stockDetail.modelTable') }}</template>
    <el-table :data="result.modelTable.rows" size="small" border>
      <el-table-column v-for="(h, i) in result.modelTable.headers" :key="i" :label="h">
        <template #default="{ row }">
          {{ row[i] }}
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>
```

- [ ] **Step 3: Create `web/src/components/TradePlanTable.vue`**

```vue
<script setup lang="ts">
import type { AnalysisResult } from '@/types/api'

const props = defineProps<{ result?: AnalysisResult | null }>()
</script>

<template>
  <el-card v-if="result" style="margin-top: 16px">
    <template #header>{{ $t('stockDetail.tradePlan') }}</template>
    <el-table :data="result.referenceTable.rows" size="small" border>
      <el-table-column v-for="(h, i) in result.referenceTable.headers" :key="i" :label="h">
        <template #default="{ row }">
          {{ row[i] }}
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>
```

- [ ] **Step 4: Create `web/src/views/stock/StatisticsTab.vue`**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { getAnalysis } from '@/apis/analysis'
import type { AnalysisResult } from '@/types/api'
import DraftFormBlock from '@/components/DraftFormBlock.vue'
import SpreadModelTable from '@/components/SpreadModelTable.vue'
import TradePlanTable from '@/components/TradePlanTable.vue'

const props = defineProps<{ tsCode: string }>()

const result = ref<AnalysisResult | null>(null)
const loading = ref(false)

async function runAnalysis(params?: { open?: number; high?: number; low?: number; close?: number }) {
  loading.value = true
  try {
    result.value = await getAnalysis(props.tsCode, {
      actualOpen: params?.open,
      actualHigh: params?.high,
      actualLow: params?.low,
      actualClose: params?.close,
      withDraft: true,
    })
  } catch (e: any) {
    wMessage('error', e.message || '分析失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => runAnalysis())
</script>

<template>
  <div v-loading="loading">
    <DraftFormBlock :ts-code="tsCode" @apply="runAnalysis" />
    <div style="margin-top: 16px">
      <SpreadModelTable :result="result" />
    </div>
    <TradePlanTable :result="result" />
  </div>
</template>
```

- [ ] **Step 5: Commit**

```bash
git add web/src/views/stock/StatisticsTab.vue web/src/components/DraftFormBlock.vue web/src/components/SpreadModelTable.vue web/src/components/TradePlanTable.vue
git commit -m "feat(web): add StatisticsTab with DraftFormBlock, model and trade plan tables"
```

---

## Task 13: Create Profile views

**Files:**
- Create: `web/src/views/profile/ProfileView.vue`
- Create: `web/src/views/profile/ProfileInfo.vue`
- Create: `web/src/views/profile/ChangePassword.vue`
- Create: `web/src/views/profile/TushareToken.vue`
- Create: `web/src/views/profile/ApiTokens.vue`

- [ ] **Step 1: Create `web/src/views/profile/ProfileView.vue`**

```vue
<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const menuItems = [
  { label: 'profile.info', path: '/profile' },
  { label: 'profile.changePassword', path: '/profile/password' },
  { label: 'profile.tushareToken', path: '/profile/token' },
  { label: 'profile.apiTokens', path: '/profile/api-tokens' },
]

const activeIndex = route.path
</script>

<template>
  <div class="profile-layout">
    <el-menu :default-active="activeIndex" :router="true" class="profile-menu">
      <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
        {{ $t(item.label) }}
      </el-menu-item>
    </el-menu>
    <div class="profile-content">
      <router-view />
    </div>
  </div>
</template>

<style scoped lang="scss">
.profile-layout {
  display: flex;
  gap: 16px;
}
.profile-menu {
  width: 200px;
  flex-shrink: 0;
}
.profile-content {
  flex: 1;
}
</style>
```

- [ ] **Step 2: Create `web/src/views/profile/ProfileInfo.vue`**

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const user = computed(() => auth.user)
</script>

<template>
  <el-card v-if="user">
    <el-descriptions :title="$t('profile.info')" :column="1">
      <el-descriptions-item :label="$t('profile.username')">{{ user.username }}</el-descriptions-item>
      <el-descriptions-item :label="$t('profile.role')">{{ user.role }}</el-descriptions-item>
      <el-descriptions-item :label="$t('profile.registeredAt')">{{ user.createdAt }}</el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>
```

- [ ] **Step 3: Create `web/src/views/profile/ChangePassword.vue`**

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { wMessage } from '@/utils/message'
import { changePassword } from '@/apis/me'

const form = ref({ old: '', new: '' })
const loading = ref(false)

async function submit() {
  if (!form.value.old || !form.value.new) {
    wMessage('warning', '请输入密码')
    return
  }
  loading.value = true
  try {
    await changePassword({ old: form.value.old, new: form.value.new })
    wMessage('success', $t('profile.passwordChanged'))
    form.value = { old: '', new: '' }
  } catch (e: any) {
    wMessage('error', e.message || '修改失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-card>
    <template #header>{{ $t('profile.changePassword') }}</template>
    <el-form @submit.prevent="submit" style="max-width: 400px">
      <el-form-item :label="$t('profile.oldPassword')">
        <el-input v-model="form.old" type="password" />
      </el-form-item>
      <el-form-item :label="$t('profile.newPassword')">
        <el-input v-model="form.new" type="password" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">{{ $t('common.save') }}</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>
```

- [ ] **Step 4: Create `web/src/views/profile/TushareToken.vue`**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { setTushareToken } from '@/apis/me'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const token = ref('')
const loading = ref(false)

onMounted(async () => {
  await auth.fetchMe()
  token.value = auth.user?.tushareToken || ''
})

async function submit() {
  loading.value = true
  try {
    await setTushareToken({ token: token.value })
    wMessage('success', $t('profile.tokenSaved'))
  } catch (e: any) {
    wMessage('error', e.message || '保存失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-card>
    <template #header>{{ $t('profile.tushareToken') }}</template>
    <el-form @submit.prevent="submit" style="max-width: 400px">
      <el-form-item label="Token">
        <el-input v-model="token" :placeholder="$t('common.empty')" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">{{ $t('common.save') }}</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>
```

- [ ] **Step 5: Create `web/src/views/profile/ApiTokens.vue`**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { listTokens, issueToken, revokeToken } from '@/apis/me'
import type { APIToken } from '@/types/api'

const tokens = ref<APIToken[]>([])
const showIssue = ref(false)
const showNewToken = ref(false)
const newTokenName = ref('')
const issuedToken = ref('')
const loading = ref(false)

async function load() {
  tokens.value = await listTokens()
}

async function create() {
  if (!newTokenName.value) return
  loading.value = true
  try {
    const res = await issueToken({ name: newTokenName.value })
    issuedToken.value = res.token
    showIssue.value = false
    showNewToken.value = true
    newTokenName.value = ''
    await load()
  } catch (e: any) {
    wMessage('error', e.message || '创建失败')
  } finally {
    loading.value = false
  }
}

async function revoke(id: number) {
  try {
    await revokeToken(id)
    wMessage('success', $t('profile.revoked'))
    await load()
  } catch (e: any) {
    wMessage('error', e.message || '撤销失败')
  }
}

onMounted(load)
</script>

<template>
  <el-card>
    <template #header>{{ $t('profile.apiTokens') }}</template>
    <el-table :data="tokens" style="margin-bottom: 16px">
      <el-table-column prop="name" :label="$t('profile.tokenName')" />
      <el-table-column prop="createdAt" :label="$t('profile.createdAt')" />
      <el-table-column :label="$t('common.action')" width="100">
        <template #default="{ row }">
          <el-button link type="danger" @click="revoke(row.id)">{{ $t('profile.revoke') }}</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-button type="primary" @click="showIssue = true">{{ $t('profile.createToken') }}</el-button>

    <el-dialog v-model="showIssue" :title="$t('profile.createToken')" width="400px">
      <el-form @submit.prevent="create">
        <el-form-item :label="$t('profile.tokenName')">
          <el-input v-model="newTokenName" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showIssue = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loading" @click="create">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showNewToken" :title="$t('profile.apiTokens')" width="400px" :close-on-click-modal="false">
      <p>{{ $t('profile.copyPrompt') }}</p>
      <el-input v-model="issuedToken" readonly />
      <template #footer>
        <el-button @click="showNewToken = false">{{ $t('common.close') }}</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>
```

- [ ] **Step 6: Commit**

```bash
git add web/src/views/profile/
git commit -m "feat(web): add ProfileView with info, password, token, api-tokens sub-pages"
```

---

## Task 14: Create Admin views

**Files:**
- Modify: `web/src/views/admin/UsersView.vue`
- Modify: `web/src/views/admin/SyncView.vue`
- Create: `web/src/views/NotFound.vue`

- [ ] **Step 1: Modify `web/src/views/admin/UsersView.vue`**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { listUsers, createUser, patchUser, deleteUser } from '@/apis/admin'
import type { User, CreateUserReq } from '@/types/api'

const users = ref<User[]>([])
const showCreate = ref(false)
const newUser = ref<CreateUserReq>({ username: '', password: '', role: 'user' })
const loading = ref(false)

async function fetchUsers() {
  users.value = await listUsers()
}

async function doCreate() {
  loading.value = true
  try {
    await createUser(newUser.value)
    wMessage('success', '创建成功')
    showCreate.value = false
    newUser.value = { username: '', password: '', role: 'user' }
    await fetchUsers()
  } catch (e: any) {
    wMessage('error', e.message || '创建失败')
  } finally {
    loading.value = false
  }
}

async function toggle(row: User) {
  try {
    await patchUser(row.id, { disabled: !row.disabled })
    await fetchUsers()
  } catch (e: any) {
    wMessage('error', e.message || '操作失败')
  }
}

async function remove(row: User) {
  try {
    await deleteUser(row.id)
    await fetchUsers()
  } catch (e: any) {
    wMessage('error', e.message || '删除失败')
  }
}

onMounted(fetchUsers)
</script>

<template>
  <div>
    <div class="header-bar">
      <h2>{{ $t('admin.users') }}</h2>
      <el-button type="primary" @click="showCreate = true">{{ $t('admin.createUser') }}</el-button>
    </div>

    <el-table :data="users" style="margin-top: 16px">
      <el-table-column prop="username" :label="$t('profile.username')" />
      <el-table-column prop="role" :label="$t('profile.role')" />
      <el-table-column :label="$t('admin.status')">
        <template #default="{ row }">
          {{ row.disabled ? $t('admin.disabled') : $t('admin.enabled') }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('common.action')" width="180">
        <template #default="{ row }">
          <el-button link @click="toggle(row)">{{ row.disabled ? $t('admin.enabled') : $t('admin.disabled') }}</el-button>
          <el-button link type="danger" @click="remove(row)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showCreate" :title="$t('admin.createUser')" width="400px">
      <el-form @submit.prevent="doCreate">
        <el-form-item :label="$t('profile.username')">
          <el-input v-model="newUser.username" />
        </el-form-item>
        <el-form-item :label="$t('profile.newPassword')">
          <el-input v-model="newUser.password" type="password" />
        </el-form-item>
        <el-form-item :label="$t('profile.role')">
          <el-select v-model="newUser.role">
            <el-option :label="$t('admin.enabled')" value="user" />
            <el-option :label="$t('admin.users')" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loading" @click="doCreate">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

- [ ] **Step 2: Modify `web/src/views/admin/SyncView.vue`**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { syncStocklist, syncBars, jobStatus } from '@/apis/admin'
import type { JobRun } from '@/types/api'

const jobs = ref<JobRun[]>([])
const jobNames = ['daily-fetch', 'stocklist-sync']

async function loadJobs() {
  jobs.value = []
  for (const name of jobNames) {
    try {
      const j = await jobStatus(name)
      jobs.value.push(j)
    } catch {
      // ignore
    }
  }
}

async function doSyncStocklist() {
  try {
    await syncStocklist()
    wMessage('success', $t('admin.syncStocklist'))
  } catch (e: any) {
    wMessage('error', e.message || '同步失败')
  }
}

async function doSyncBars() {
  try {
    await syncBars()
    wMessage('success', $t('admin.syncBars'))
  } catch (e: any) {
    wMessage('error', e.message || '同步失败')
  }
}

onMounted(loadJobs)
</script>

<template>
  <div>
    <h2>{{ $t('admin.sync') }}</h2>
    <el-button type="primary" @click="doSyncStocklist">{{ $t('admin.syncStocklist') }}</el-button>
    <el-button type="primary" @click="doSyncBars">{{ $t('admin.syncBars') }}</el-button>

    <h3 style="margin-top: 24px">{{ $t('admin.job') }}</h3>
    <el-table :data="jobs">
      <el-table-column prop="jobName" :label="$t('admin.job')" />
      <el-table-column prop="status" :label="$t('admin.status')" />
      <el-table-column prop="startedAt" :label="$t('admin.startedAt')" />
      <el-table-column prop="finishedAt" :label="$t('admin.finishedAt')" />
    </el-table>
  </div>
</template>
```

- [ ] **Step 3: Create `web/src/views/NotFound.vue`**

```vue
<template>
  <div class="not-found">
    <el-result icon="error" title="404" sub-title="页面不存在">
      <template #extra>
        <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
      </template>
    </el-result>
  </div>
</template>

<style scoped lang="scss">
.not-found {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}
</style>
```

- [ ] **Step 4: Commit**

```bash
git add web/src/views/admin/UsersView.vue web/src/views/admin/SyncView.vue web/src/views/NotFound.vue
git commit -m "feat(web): add admin views and NotFound"
```

---

## Task 15: Wire up router and main entry

**Files:**
- Modify: `web/src/router/index.ts`
- Modify: `web/src/main.ts`

- [ ] **Step 1: Rewrite `web/src/router/index.ts`**

```ts
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/views/LoginView.vue') },
    { path: '/', redirect: '/stocks' },
    {
      path: '/stocks',
      component: () => import('@/views/StockListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/stocks/:tsCode',
      component: () => import('@/views/stock/StockDetailView.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'StockBasic', component: () => import('@/views/stock/BasicTab.vue') },
        { path: 'statistics', name: 'StockStatistics', component: () => import('@/views/stock/StatisticsTab.vue') },
      ],
    },
    {
      path: '/profile',
      component: () => import('@/views/profile/ProfileView.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', component: () => import('@/views/profile/ProfileInfo.vue') },
        { path: 'password', component: () => import('@/views/profile/ChangePassword.vue') },
        { path: 'token', component: () => import('@/views/profile/TushareToken.vue') },
        { path: 'api-tokens', component: () => import('@/views/profile/ApiTokens.vue') },
      ],
    },
    {
      path: '/admin/users',
      component: () => import('@/views/admin/UsersView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/admin/sync',
      component: () => import('@/views/admin/SyncView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    { path: '/:pathMatch(.*)*', component: () => import('@/views/NotFound.vue') },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.user) {
    next('/login')
  } else if (to.meta.requiresAdmin && auth.user?.role !== 'admin') {
    next('/stocks')
  } else {
    next()
  }
})

export default router
```

- [ ] **Step 2: Rewrite `web/src/main.ts`**

```ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import '@/assets/css/index.scss'
import '@/assets/css/element-reset.scss'

import App from './App.vue'
import router from './router'
import { i18n } from './intl'
import GIcon from './components/GIcon.vue'
import GEllipsis from './components/GEllipsis.vue'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
app.use(i18n)
app.component('GIcon', GIcon)
app.component('GEllipsis', GEllipsis)
app.mount('#app')
```

- [ ] **Step 3: Run frontend type check**

Run: `cd /root/code/github/travelliu/stock/web && npx vue-tsc --noEmit`
Expected: PASS (no errors)

- [ ] **Step 4: Commit**

```bash
git add web/src/router/index.ts web/src/main.ts
git commit -m "feat(web): wire up router and main entry with i18n and global components"
```

---

## Task 16: Frontend tests

**Files:**
- Create: `web/src/utils/format.spec.ts`
- Create: `web/src/apis/auth.spec.ts`
- Modify: `web/e2e/smoke.spec.ts`

- [ ] **Step 1: Create `web/src/utils/format.spec.ts`**

```ts
import { describe, it, expect } from 'vitest'
import { fmtPrice, fmtPct, priceClass } from './format'

describe('format', () => {
  it('fmtPrice', () => {
    expect(fmtPrice(10.5)).toBe('10.50')
    expect(fmtPrice(0)).toBe('0.00')
  })

  it('fmtPct', () => {
    expect(fmtPct(11, 10)).toBe('+10.00%')
    expect(fmtPct(9, 10)).toBe('-10.00%')
    expect(fmtPct(10, 0)).toBe('0.00%')
  })

  it('priceClass', () => {
    expect(priceClass(11, 10)).toBe('g-up')
    expect(priceClass(9, 10)).toBe('g-down')
    expect(priceClass(10, 10)).toBe('g-flat')
  })
})
```

- [ ] **Step 2: Create `web/src/apis/auth.spec.ts`**

```ts
import { describe, it, expect, vi } from 'vitest'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import { login } from './auth'

const mock = new MockAdapter(axios)

describe('auth api', () => {
  it('login returns user with camelCase fields', async () => {
    mock.onPost('/api/auth/login').reply(200, {
      code: 200,
      message: 'ok',
      data: { id: 1, username: 'alice', role: 'user', tushareToken: 'tk', createdAt: '2025-01-01', updatedAt: '2025-01-01' },
    })
    const user = await login({ username: 'alice', password: 'secret' })
    expect(user.username).toBe('alice')
    expect(user.tushareToken).toBe('tk')
  })
})
```

- [ ] **Step 3: Install axios-mock-adapter**

Run:
```bash
cd /root/code/github/travelliu/stock/web
npm install -D axios-mock-adapter
```

- [ ] **Step 4: Modify `web/e2e/smoke.spec.ts`**

```ts
import { test, expect } from '@playwright/test'

test.describe('smoke', () => {
  test('login -> add stock -> detail tabs', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[placeholder="用户名"]', 'admin')
    const adminPassword = process.env.ADMIN_PASSWORD || 'changeme-see-server-logs'
    await page.fill('input[type="password"]', adminPassword)
    await page.click('button:has-text("登录")')
    await page.waitForURL('/stocks')

    await page.click('button:has-text("添加股票")')
    await page.click('.el-select-v2__wrapper')
    await page.fill('.el-select-v2__wrapper input', '600519')
    await page.click('.el-select-dropdown__item:has-text("600519")')
    await page.click('button:has-text("添加")')

    await page.click('text=详情')
    await page.waitForURL(/\/stocks\//)

    await page.click('text=基础与价差')
    await expect(page.locator('text=日线数据')).toBeVisible()

    await page.click('text=详细统计')
    await expect(page.locator('text=价差模型')).toBeVisible()
  })
})
```

- [ ] **Step 5: Run vitest**

Run: `cd /root/code/github/travelliu/stock/web && npm run test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add web/src/utils/format.spec.ts web/src/apis/auth.spec.ts web/e2e/smoke.spec.ts web/package.json web/package-lock.json
git commit -m "test(web): add format and auth api unit tests, update e2e smoke test"
```

---

## Task 17: Clean up old frontend files

**Files:**
- Delete: `web/src/api/client.ts`
- Delete: `web/src/components/AnalysisPanel.vue`
- Delete: `web/src/components/HistoryPanel.vue`
- Delete: `web/src/components/DraftPanel.vue`
- Delete: `web/src/views/PortfolioView.vue`
- Delete: `web/src/views/SettingsView.vue`
- Delete: `web/src/stores/auth.ts` (old version)

- [ ] **Step 1: Delete old files**

Run:
```bash
cd /root/code/github/travelliu/stock/web/src
rm -rf api/client.ts components/AnalysisPanel.vue components/HistoryPanel.vue components/DraftPanel.vue views/PortfolioView.vue views/SettingsView.vue
```

- [ ] **Step 2: Verify build still passes**

Run: `cd /root/code/github/travelliu/stock/web && npx vue-tsc --noEmit`
Expected: PASS

Run: `cd /root/code/github/travelliu/stock/web && npm run build`
Expected: PASS (dist/ created)

- [ ] **Step 3: Run all Go tests**

Run: `cd /root/code/github/travelliu/stock && go test -race ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore(web): remove old frontend files after rewrite"
```

---

## Self-Review

**1. Spec coverage check:**

| Spec Section | Task(s) | Status |
|---|---|---|
| 2.1 pkg/models 新增 auth.go / me.go | Task 1 | Covered |
| 2.2 字段重命名 | Task 2 | Covered |
| 2.3 Inline -> models 迁移 | Task 3 | Covered (auth, me, admin, draft; CLI login + portfolio) |
| 2.4 后端测试 | Task 4 | Covered |
| 2.5 doc.go | Task 1 | Covered |
| 3.1 目录结构 | Tasks 5-16 | Covered |
| 3.2 整体布局 | Task 9 | Covered (65px 侧栏) |
| 3.3 配色与 SCSS 变量 | Task 5 | Covered |
| 3.4 路由设计 | Task 15 | Covered |
| 3.5 详情页 2 Tab | Tasks 11-12 | Covered |
| 3.6 全局组件 g-* | Task 8 | Covered |
| 3.7 国际化 | Tasks 6-7 | Covered |
| 4.1 调用栈 | Tasks 6-7 | Covered |
| 4.2 apis 层签名 | Task 7 | Covered |
| 4.3 axios 拦截器 | Task 7 | Covered |
| 5.1 后端测试 | Task 4 | Covered |
| 5.2 前端测试 | Task 16 | Covered |
| 7. 旧->新映射 | Tasks 10-14 | Covered |

**2. Placeholder scan:** No TBD, TODO, or "add appropriate error handling" found. All steps contain complete code.

**3. Type consistency check:** All types match between `pkg/models` and `web/src/types/api.ts`. Field names (`userId`, `username`, `tsCode`, etc.) are consistent across backend and frontend.
