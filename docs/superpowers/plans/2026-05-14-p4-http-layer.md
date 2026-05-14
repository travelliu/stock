# P4 — HTTP Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire all P3 services behind a gin HTTP API using the mtk response/error/trace pattern: `{requestID, code, message, data}` envelope, custom error type with codes, request-ID injection, i18n message maps, and recovery middleware. All business errors return HTTP 200 with the error code in the body.

**Architecture:** HTTP transport lives under `pkg/stockd/http/` and `pkg/stockd/utils/`. The `utils` package owns response envelope, error type, message maps, trace (request ID), and recovery — matching the mtk reference. Handlers are thin: bind input, call services, return `utils.HTTPRequestSuccess` or `utils.HTTPRequestFailedV4`/`V5`. The router composes services and mounts route groups under `/api`.

**Reference pattern:** `/root/code/gitlab/mogdb_en/mtk/pkg/mtkd/utils/response.go`, `error.go`, `msg.go`, `trace.go`, `recovery.go`, `http/router.go`.

**Tech Stack:** `gin-gonic/gin`, `gin-contrib/sessions`, `gin-contrib/static`, `swaggo/gin-swagger`, `swaggo/files`, `sirupsen/logrus`, `crypto/rand`, `crypto/tls`, `net/http`.

**Reference spec:** `docs/superpowers/specs/2026-05-14-go-vue-rewrite-design.md` §3 (HTTP API + Auth), §4.4 (Build & embed), §6.3 (Config), §7.2 (P4).

---

## File overview

| File | Responsibility |
|------|----------------|
| `pkg/stockd/utils/error.go` + `_test.go` | Custom `errors` struct, `New`/`Wrap`, `GetCodeAndData` |
| `pkg/stockd/utils/msg.go` | Error code constants, zh/en message maps, `GetMsg`/`GetErrMsg`/`GetSuccessMsg` |
| `pkg/stockd/utils/trace.go` + `_test.go` | `RequestID` middleware, `GetReqID`, `GetLang`, `ParseAcceptLanguage` |
| `pkg/stockd/utils/response.go` + `_test.go` | `HTTPResponse`, `HTTPRequestSuccess`, `HTTPRequestFailed`/`V4`/`V5` |
| `pkg/stockd/utils/recovery.go` | `Recovery()` gin middleware with panic logging |
| `pkg/stockd/http/middleware.go` | `ResolveUser` (soft auth, no abort), `AuthRequired`, `AdminRequired` |
| `pkg/stockd/http/router.go` | Compose services, mount route groups, apply middleware, SPA static mount, Swagger |
| `pkg/stockd/http/handler/*.go` + `_test.go` | All route handlers (thin wrappers around services) |
| `embed.go` (repo root) | `//go:embed all:web/dist` + `EmbedFolder()` |
| `cmd/stockd/main.go` | Real entrypoint: load config → DB → AutoMigrate → bootstrap → services → router → TLS/gin → graceful shutdown |

---

### Task 23: `pkg/stockd/utils/` — error, msg, trace, response, recovery

**Files:**
- Create: `pkg/stockd/utils/error.go`
- Create: `pkg/stockd/utils/error_test.go`
- Create: `pkg/stockd/utils/msg.go`
- Create: `pkg/stockd/utils/trace.go`
- Create: `pkg/stockd/utils/trace_test.go`
- Create: `pkg/stockd/utils/response.go`
- Create: `pkg/stockd/utils/response_test.go`
- Create: `pkg/stockd/utils/recovery.go`

- [ ] **Step 1: Write `pkg/stockd/utils/error.go`**

```go
package utils

import "fmt"

type errors struct {
	code    int
	message string
	data    []interface{}
}

type errorsInterface interface {
	Code() int
	Error() string
	Data() []interface{}
}

func GetCodeAndData(err error) (int, []interface{}) {
	if c, ok := err.(errorsInterface); ok {
		return c.Code(), c.Data()
	}
	return 0, nil
}

func New(code int, format string, messages ...interface{}) error {
	return Wrap(code, nil, format, messages...)
}

func Wrap(code int, err error, format string, messages ...interface{}) error {
	if err != nil {
		format = fmt.Sprintf("%s : %s", format, err.Error())
	}
	message := fmt.Sprintf(format, messages...)
	if format == "" {
		message = fmt.Sprintf(GetErrMsg(code, ""), messages...)
	}
	return &errors{code, message, messages}
}

func (e *errors) Code() int        { return e.code }
func (e *errors) Data() []interface{} { return e.data }

func (e *errors) Error() string {
	if e.message == "" {
		e.message = GetErrMsg(e.code, "")
	}
	return e.message
}
```

- [ ] **Step 2: Write `pkg/stockd/utils/error_test.go`**

```go
package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/stockd/utils"
)

func TestNew(t *testing.T) {
	err := utils.New(utils.ErrInvalidParam, "bad value %v", 42)
	assert.Equal(t, utils.ErrInvalidParam, err.(interface{ Code() int }).Code())
	assert.Contains(t, err.Error(), "bad value 42")
}

func TestWrap(t *testing.T) {
	inner := utils.New(utils.ErrUserNotFound, "user missing")
	wrapped := utils.Wrap(utils.ErrInvalidParam, inner, "outer")
	assert.Contains(t, wrapped.Error(), "outer")
	assert.Contains(t, wrapped.Error(), "user missing")
}

func TestGetCodeAndData(t *testing.T) {
	err := utils.New(utils.ErrInvalidParam, "msg", 1, 2)
	code, data := utils.GetCodeAndData(err)
	assert.Equal(t, utils.ErrInvalidParam, code)
	assert.Equal(t, []interface{}{1, 2}, data)
}

func TestGetCodeAndData_StandardError(t *testing.T) {
	code, data := utils.GetCodeAndData(assert.AnError)
	assert.Equal(t, 0, code)
	assert.Nil(t, data)
}
```

- [ ] **Step 3: Write `pkg/stockd/utils/msg.go`**

```go
package utils

import "strings"

const (
	LangZh = "zh"
	LangEn = "en"
)

const (
	SUCCESS = 200
	ERROR   = 500

	ErrInvalidParam    = 40001
	ErrUnauthorized    = 40002
	ErrForbidden       = 40003
	ErrUserNotFound    = 40004
	ErrInvalidPassword = 40005
	ErrUserDisabled    = 40006
	ErrStockNotFound   = 40007
	ErrDraftInvalid    = 40008
	ErrTokenInvalid    = 40009
	ErrTokenExpired    = 40010
	ErrDuplicateUser   = 40011
	ErrInvalidCode     = 40012
	ErrTaskRun         = 40013
	ErrTaskNoRunReport = 40014
)

var (
	defaultZhMsg = map[int]string{
		SUCCESS:            "成功",
		ERROR:              "系统异常，请联系管理员",
		ErrInvalidParam:    "参数错误",
		ErrUnauthorized:    "未认证",
		ErrForbidden:       "无权限",
		ErrUserNotFound:    "用户不存在",
		ErrInvalidPassword: "密码错误",
		ErrUserDisabled:    "账户已禁用",
		ErrStockNotFound:   "股票不存在",
		ErrDraftInvalid:    "草稿数据无效",
		ErrTokenInvalid:    "Token无效",
		ErrTokenExpired:    "Token已过期",
		ErrDuplicateUser:   "用户名已存在",
		ErrInvalidCode:     "股票代码错误",
		ErrTaskRun:         "任务正在运行",
		ErrTaskNoRunReport: "任务没有运行报告",
	}
	defaultEnMsg = map[int]string{
		SUCCESS:            "Succeed",
		ERROR:              "The system is abnormal, please contact the administrator",
		ErrInvalidParam:    "Invalid parameter",
		ErrUnauthorized:    "Unauthorized",
		ErrForbidden:       "Forbidden",
		ErrUserNotFound:    "User not found",
		ErrInvalidPassword: "Invalid password",
		ErrUserDisabled:    "Account disabled",
		ErrStockNotFound:   "Stock not found",
		ErrDraftInvalid:    "Invalid draft data",
		ErrTokenInvalid:    "Invalid token",
		ErrTokenExpired:    "Token expired",
		ErrDuplicateUser:   "Username already exists",
		ErrInvalidCode:     "Invalid stock code",
		ErrTaskRun:         "Task is running",
		ErrTaskNoRunReport: "Task has no run report",
	}

	defaultErrorMsg = map[string]map[int]string{
		LangZh: defaultZhMsg,
		LangEn: defaultEnMsg,
	}
	ErrorMsg = map[string]map[int]string{
		LangZh: defaultZhMsg,
		LangEn: defaultEnMsg,
	}
)

func GetErrMsg(code int, lang string) string     { return getMsg(code, lang, ERROR) }
func GetSuccessMsg(code int, lang string) string { return getMsg(code, lang, SUCCESS) }
func GetMsg(code int, lang string) string        { return getMsg(code, lang, 0) }

func getMsg(code int, lang string, status int) string {
	lang = strings.ToLower(lang)
	if lang == "" {
		lang = LangEn
	}
	msgMap, ok := ErrorMsg[lang]
	if !ok {
		msgMap = defaultErrorMsg[LangEn]
	}
	if m, ok := msgMap[code]; ok {
		return m
	}
	if status != 0 {
		if m, ok := msgMap[status]; ok {
			return m
		}
	}
	return ""
}
```

- [ ] **Step 4: Write `pkg/stockd/utils/trace.go`**

```go
package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	ContextKeyReqID         ContextKey = "requestID"
	HTTPHeaderNameRequestID            = "X-Request-ID"
)

func genRequestID() string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}

func GetReqID(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyReqID).(string); ok {
		return v
	}
	return ""
}

func AttachReqID(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyReqID, genRequestID())
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := AttachReqID(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetLang(c *gin.Context) string {
	lang := c.GetString("lang")
	if lang == "" {
		lang = c.GetHeader("Lang")
	}
	if lang == "" {
		lang = c.GetHeader("lang")
	}
	if lang == "" {
		lang = ParseAcceptLanguage(c.GetHeader("Accept-Language"))
	}
	return lang
}

func ParseAcceptLanguage(s string) string {
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		q := strings.Split(part, ";")
		if len(q) == 1 {
			return q[0]
		}
		qp := strings.Split(q[1], "=")
		if len(qp) >= 2 && qp[1] == "1" {
			return q[0]
		}
	}
	return ""
}

func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("lang", GetLang(c))
		c.Next()
	}
}
```

- [ ] **Step 5: Write `pkg/stockd/utils/trace_test.go`**

```go
package utils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"stock/pkg/stockd/utils"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(utils.RequestID())
	r.GET("/", func(c *gin.Context) {
		assert.NotEmpty(t, utils.GetReqID(c.Request.Context()))
		c.Status(200)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 200, w.Code)
}

func TestParseAcceptLanguage(t *testing.T) {
	assert.Equal(t, "zh-CN", utils.ParseAcceptLanguage("zh-CN,en;q=0.9"))
	assert.Equal(t, "en", utils.ParseAcceptLanguage("en"))
	assert.Equal(t, "", utils.ParseAcceptLanguage(""))
}
```

- [ ] **Step 6: Write `pkg/stockd/utils/response.go`**

```go
package utils

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HTTPResponse struct {
	RequestID string      `json:"requestID"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

func httpResponse(requestID string, code int, msg string, data interface{}) *HTTPResponse {
	return &HTTPResponse{RequestID: requestID, Code: code, Message: msg, Data: data}
}

func getCaller() string {
	pc, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d:%s", file, line, runtime.FuncForPC(pc).Name())
}

func HTTPRequestFailed(c *gin.Context, err error) {
	HTTPRequestFailedV4(c, err, ERROR)
}

func HTTPRequestFailedV4(c *gin.Context, err error, code int, data ...interface{}) {
	if err != nil {
		logrus.WithField(string(ContextKeyReqID), GetReqID(c.Request.Context())).
			Errorf("%s -> Error: %s", getCaller(), err)
	}
	var (
		errCode int
		errData []interface{}
		msg     string
	)
	errCode, errData = GetCodeAndData(err)
	if errCode == 0 {
		errCode = code
		errData = data
	}
	if errCode == 0 {
		errCode = ERROR
	}
	if errData == nil {
		errData = data
	}
	if errCode == 600 {
		errCode = ERROR
		msg = err.Error()
	} else {
		msg = GetMsg(errCode, GetLang(c))
	}
	if len(errData) > 0 && (errCode != ERROR && errCode != 400) && strings.Contains(msg, "%") {
		msg = fmt.Sprintf(msg, errData...)
		msg = strings.TrimSpace(msg)
	}
	if err != nil && !strings.Contains(err.Error(), "%!(EXTRA") && err.Error() != msg {
		msg += " " + err.Error()
	}
	c.JSON(http.StatusOK, httpResponse(GetReqID(c.Request.Context()), errCode, msg, nil))
}

func HTTPRequestFailedV5(c *gin.Context, err error) {
	c.JSON(http.StatusOK, httpResponse(GetReqID(c.Request.Context()), ERROR, err.Error(), ""))
}

func HTTPRequestSuccess(c *gin.Context, code int, data interface{}) {
	c.JSON(http.StatusOK, httpResponse(
		GetReqID(c.Request.Context()), code,
		GetSuccessMsg(code, GetLang(c)), data,
	))
}
```

- [ ] **Step 7: Write `pkg/stockd/utils/response_test.go`**

```go
package utils_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockd/utils"
)

func TestHTTPRequestSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	utils.HTTPRequestSuccess(c, 200, gin.H{"key": "val"})
	assert.Equal(t, http.StatusOK, w.Code)
	var r utils.HTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &r))
	assert.Equal(t, 200, r.Code)
	assert.NotEmpty(t, r.RequestID)
}

func TestHTTPRequestFailedV4(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	utils.HTTPRequestFailedV4(c, errors.New("boom"), 500)
	assert.Equal(t, http.StatusOK, w.Code)
	var r utils.HTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &r))
	assert.Equal(t, 500, r.Code)
	assert.NotEmpty(t, r.Message)
}

func TestHTTPRequestFailedV4_Code600(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	utils.HTTPRequestFailedV4(c, errors.New("parse error"), 600)
	var r utils.HTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &r))
	assert.Equal(t, 500, r.Code)
	assert.Contains(t, r.Message, "parse error")
}
```

- [ ] **Step 8: Write `pkg/stockd/utils/recovery.go`**

```go
package utils

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				if !brokenPipe {
					logrus.Errorf("[Recovery] %s panic recovered:\n%v",
						time.Now().Format("2006/01/02 - 15:04:05"), err)
				}
				c.JSON(http.StatusInternalServerError, HTTPResponse{Code: ERROR, Message: fmt.Sprintf("%v", err)})
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

- [ ] **Step 9: Run tests + commit**

```bash
go test ./pkg/stockd/utils/... -v
git add pkg/stockd/utils/
git commit -m "feat(utils): response envelope, error codes, trace, recovery (mtk pattern)"
```

---

### Task 24: Update P2 auth middleware to use utils response format

**Files:**
- Modify: `pkg/stockd/auth/middleware.go`

The P2 middleware aborts with raw `gin.H`. Update it to use `utils.HTTPRequestFailedV4` and add a `ResolveUser` variant that does not abort.

- [ ] **Step 1: Replace `pkg/stockd/auth/middleware.go`**

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
	"stock/pkg/stockd/utils"
)

const (
	ctxUserKey     = "stockd.user"
	ctxTokenKey    = "stockd.tushare_token"
	sessionUserKey = "uid"
)

// ResolveUser returns a gin handler that resolves the calling user from Bearer
// or session and attaches *models.User + effective token to the context.
// It does NOT abort on failure — downstream middleware/handlers decide.
func ResolveUser(gdb *gorm.DB, defaultTushareToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := resolveUser(c, gdb)
		if user != nil {
			c.Set(ctxUserKey, user)
			token := defaultTushareToken
			if user.TushareToken != "" {
				token = user.TushareToken
			}
			c.Set(ctxTokenKey, token)
		}
		c.Next()
	}
}

// Middleware is the strict variant: aborts with 401 if user cannot be resolved.
func Middleware(gdb *gorm.DB, defaultTushareToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := resolveUser(c, gdb)
		if err != nil {
			utils.HTTPRequestFailedV4(c, err, utils.ErrUnauthorized)
			c.Abort()
			return
		}
		if user.Disabled {
			utils.HTTPRequestFailedV4(c, nil, utils.ErrUserDisabled)
			c.Abort()
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

func User(c *gin.Context) *models.User {
	if v, ok := c.Get(ctxUserKey); ok {
		return v.(*models.User)
	}
	return nil
}

func TushareTokenFor(c *gin.Context) string {
	if v, ok := c.Get(ctxTokenKey); ok {
		return v.(string)
	}
	return ""
}
```

- [ ] **Step 2: Run auth tests**

```bash
go test ./pkg/stockd/auth/... -v
git add pkg/stockd/auth/middleware.go
git commit -m "refactor(auth): use utils.HTTPRequestFailedV4, add ResolveUser soft variant"
```

---

### Task 25: Auth routes (`/api/auth/*`)

**Files:**
- Create: `pkg/stockd/http/handler/auth.go`
- Create: `pkg/stockd/http/handler/auth_test.go`

Handler pattern follows mtk: single `handler` struct, methods use `c.BindJSON`, return `utils.HTTPRequestSuccess`/`utils.HTTPRequestFailedV4`.

- [ ] **Step 1: Write `pkg/stockd/http/handler/auth.go`**

```go
package handler

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/services/user"
	"stock/pkg/stockd/utils"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *handler) Login(c *gin.Context) {
	var req loginReq
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

- [ ] **Step 2: Write `pkg/stockd/http/handler/auth_test.go`**

```go
package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/db"
	"stock/pkg/stockd/http/handler"
	"stock/pkg/stockd/models"
	"stock/pkg/stockd/services/analysis"
	"stock/pkg/stockd/services/bars"
	"stock/pkg/stockd/services/draft"
	"stock/pkg/stockd/services/portfolio"
	"stock/pkg/stockd/services/scheduler"
	"stock/pkg/stockd/services/stock"
	"stock/pkg/stockd/services/token"
	"stock/pkg/stockd/services/user"
	"stock/pkg/stockd/utils"
	"stock/pkg/tushare"
)

func setupAuthRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))

	r := gin.New()
	r.Use(utils.RequestID())
	r.Use(utils.Language())
	store := auth.NewSessionStore([]byte("12345678901234567890123456789012"))
	r.Use(sessions.Sessions(auth.SessionName, store))
	r.Use(auth.ResolveUser(gdb, ""))

	userSvc := user.New(gdb)
	tokenSvc := token.New(gdb)
	stockSvc := stock.New(gdb)
	portfolioSvc := portfolio.New(gdb)
	draftSvc := draft.New(gdb)
	barsSvc := bars.New(gdb, tushare.NewClient())
	analysisSvc := analysis.New(gdb)
	schedulerSvc := scheduler.New(gdb)
	h := handler.NewHandler(userSvc, tokenSvc, stockSvc, portfolioSvc, draftSvc, barsSvc, analysisSvc, schedulerSvc)
	r.POST("/api/auth/login", h.Login)
	r.POST("/api/auth/logout", h.Logout)
	r.GET("/api/auth/me", handler.AuthRequired(), h.Me)
	return r, gdb
}

func TestLogin(t *testing.T) {
	r, gdb := setupAuthRouter(t)
	pw, _ := auth.HashPassword("secret")
	require.NoError(t, gdb.Create(&models.User{Username: "alice", PasswordHash: pw, Role: "user"}).Error)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"alice","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
}

func TestLogin_BadPassword(t *testing.T) {
	r, gdb := setupAuthRouter(t)
	pw, _ := auth.HashPassword("secret")
	require.NoError(t, gdb.Create(&models.User{Username: "alice", PasswordHash: pw, Role: "user"}).Error)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"alice","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
}
```

Note: See Task 34 for the full `handler` struct definition and `NewHandler` signature.

- [ ] **Step 3: Run tests + commit**

```bash
go test ./pkg/stockd/http/handler/... -run TestLogin -v
git add pkg/stockd/http/handler/auth.go pkg/stockd/http/handler/auth_test.go
git commit -m "feat(http/auth): login/logout/me with mtk response format"
```

---

### Task 26: Admin user routes (`/api/admin/users/*`)

**Files:**
- Modify: `pkg/stockd/http/handler/admin.go` (was created in old P4, now rewritten)
- Create: `pkg/stockd/http/handler/admin_test.go`

- [ ] **Step 1: Write admin methods on `handler` struct**

Add to `pkg/stockd/http/handler/handler.go` (see Task 34 for the struct definition):

```go
func (h *handler) CreateUser(c *gin.Context) {
	var req struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Role         string `json:"role"`
		TushareToken string `json:"tushare_token,omitempty"`
	}
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
	var req struct {
		Role         *string `json:"role,omitempty"`
		Disabled     *bool   `json:"disabled,omitempty"`
		TushareToken *string `json:"tushare_token,omitempty"`
	}
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

Note: `user.CreateInput` needs `TushareToken` field. If P3's `user.Service.Create` doesn't have it, add it.

Also add `SetRole` to `pkg/stockd/services/user/user.go` if missing:
```go
func (s *Service) SetRole(ctx context.Context, id uint, role string) error {
	if role != "user" && role != "admin" {
		return fmt.Errorf("role must be user|admin")
	}
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}
```

- [ ] **Step 2: Run tests + commit**

```bash
go test ./pkg/stockd/http/handler/... -run TestAdmin -v
git add pkg/stockd/http/handler/admin.go pkg/stockd/http/handler/admin_test.go
git add pkg/stockd/services/user/user.go
git commit -m "feat(http/admin): admin user CRUD with mtk response format"
```

---

### Task 27: Self-service routes (`/api/me/*`)

Add to `pkg/stockd/http/handler/me.go` (methods on `handler` struct).

- [ ] **Step 1: Write me methods**

```go
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

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
	var req struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
	}
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
	utils.HTTPRequestSuccess(c, 200, gin.H{"token": plain, "metadata": tok})
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
	var req struct{ Token string `json:"token"` }
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
	var req struct {
		Old string `json:"old"`
		New string `json:"new"`
	}
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

- [ ] **Step 2: Commit**

```bash
git add pkg/stockd/http/handler/me.go pkg/stockd/http/handler/me_test.go
git commit -m "feat(http/me): self-service endpoints with mtk response format"
```

---

### Task 28: Stock + portfolio routes

- [ ] **Step 1: Write stock methods**

```go
func (h *handler) SearchStocks(c *gin.Context) {
	q := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := h.stockSvc.Search(c.Request.Context(), q, limit)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}

func (h *handler) GetStock(c *gin.Context) {
	tsCode := c.Param("tsCode")
	s, err := h.stockSvc.Get(c.Request.Context(), tsCode)
	if err != nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrStockNotFound)
		return
	}
	utils.HTTPRequestSuccess(c, 200, s)
}
```

- [ ] **Step 2: Write portfolio methods**

```go
func (h *handler) ListPortfolio(c *gin.Context) {
	u := auth.User(c)
	list, err := h.portfolioSvc.List(c.Request.Context(), u.ID)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}

func (h *handler) AddPortfolio(c *gin.Context) {
	var req struct {
		TsCode string `json:"ts_code"`
		Note   string `json:"note,omitempty"`
	}
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.portfolioSvc.Add(c.Request.Context(), u.ID, req.TsCode, req.Note); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "added"})
}

func (h *handler) RemovePortfolio(c *gin.Context) {
	u := auth.User(c)
	if err := h.portfolioSvc.Remove(c.Request.Context(), u.ID, c.Param("tsCode")); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "removed"})
}

func (h *handler) UpdatePortfolioNote(c *gin.Context) {
	var req struct{ Note string `json:"note"` }
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.portfolioSvc.UpdateNote(c.Request.Context(), u.ID, c.Param("tsCode"), req.Note); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "updated"})
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/stockd/http/handler/stock.go pkg/stockd/http/handler/portfolio.go
git add pkg/stockd/http/handler/stock_test.go pkg/stockd/http/handler/portfolio_test.go
git commit -m "feat(http/stock,http/portfolio): search/get/portfolio CRUD with mtk format"
```

---

### Task 29: Bars + draft routes

- [ ] **Step 1: Write bars methods**

```go
func (h *handler) QueryBars(c *gin.Context) {
	list, err := h.barsSvc.Query(c.Request.Context(), c.Param("tsCode"), c.Query("from"), c.Query("to"))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}
```

- [ ] **Step 2: Write draft methods**

```go
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
	var req struct {
		TsCode    string   `json:"ts_code"`
		TradeDate string   `json:"trade_date"`
		Open      *float64 `json:"open,omitempty"`
		High      *float64 `json:"high,omitempty"`
		Low       *float64 `json:"low,omitempty"`
		Close     *float64 `json:"close,omitempty"`
	}
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

- [ ] **Step 3: Commit**

```bash
git add pkg/stockd/http/handler/bars.go pkg/stockd/http/handler/draft.go
git add pkg/stockd/http/handler/bars_test.go pkg/stockd/http/handler/draft_test.go
git commit -m "feat(http/bars,http/draft): bars query and draft endpoints with mtk format"
```

---

### Task 30: Analysis route

- [ ] **Step 1: Write analysis method**

```go
func (h *handler) GetAnalysis(c *gin.Context) {
	u := auth.User(c)
	in := analysis.Input{UserID: u.ID, TsCode: c.Param("tsCode")}

	if v := c.Query("actual_open"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		in.OpenPrice = &f
	}
	if v := c.Query("actual_high"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		in.ActualHigh = &f
	}
	if v := c.Query("actual_low"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		in.ActualLow = &f
	}
	if v := c.Query("actual_close"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		in.ActualClose = &f
	}
	in.WithDraft = c.DefaultQuery("with_draft", "true") == "true"

	res, err := h.analysisSvc.Run(c.Request.Context(), in)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, res)
}
```

- [ ] **Step 2: Commit**

```bash
git add pkg/stockd/http/handler/analysis.go pkg/stockd/http/handler/analysis_test.go
git commit -m "feat(http/analysis): analysis endpoint with mtk response format"
```

---

### Task 31: Admin sync routes

- [ ] **Step 1: Write sync methods**

```go
func (h *handler) SyncStocklist(c *gin.Context) {
	token := auth.TushareTokenFor(c)
	n, err := h.stockSvc.SyncFromTushare(c.Request.Context(), token)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"synced": n})
}

func (h *handler) SyncBars(c *gin.Context) {
	if err := h.schedulerSvc.Trigger(c.Request.Context(), "daily-fetch"); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "daily-fetch triggered"})
}

func (h *handler) ImportCSV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	defer file.Close()
	n, err := h.stockSvc.ImportCSV(c.Request.Context(), file)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"imported": n})
}

func (h *handler) JobStatus(c *gin.Context) {
	name := c.Query("job")
	if name == "" {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrInvalidParam)
		return
	}
	jr, err := h.schedulerSvc.LastRun(c.Request.Context(), name)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, jr)
}
```

- [ ] **Step 2: Commit**

```bash
git add pkg/stockd/http/handler/sync.go pkg/stockd/http/handler/sync_test.go
git commit -m "feat(http/sync): sync triggers and CSV import with mtk response format"
```

---

### Task 32: Swagger annotations

Same as before but response types reference `utils.HTTPResponse`.

- [ ] **Step 1: Update Makefile swagger target**

```makefile
swagger:
	@command -v swag >/dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g pkg/stockd/http/router.go -o docs/swagger
```

- [ ] **Step 2: Add swaggo annotations**

Example for `Login`:
```go
// Login godoc
// @Summary      User login
// @Description  Authenticate with username/password and set session cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  loginReq  true  "Credentials"
// @Success      200   {object}  utils.HTTPResponse
// @Router       /api/auth/login [post]
func (h *handler) Login(c *gin.Context) { ... }
```

- [ ] **Step 3: Run `make swagger` and commit**

```bash
make swagger
git add Makefile pkg/stockd/http/handler/
git add docs/swagger/
git commit -m "docs(swagger): add swaggo annotations for mtk response shape"
```

---

### Task 33: Static SPA mount + embed.go update

Same as before. The embed.go and router SPA fallback don't change.

- [ ] **Step 1: Replace stub `embed.go`**

```go
package stock

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/static"
)

//go:embed all:web/dist
var StaticDir embed.FS

type embedFS struct{ http.FileSystem }

func (e embedFS) Exists(prefix, filepath string) bool {
	if _, err := e.Open(filepath); err != nil {
		return false
	}
	return true
}

func EmbedFolder() static.ServeFileSystem {
	sub, _ := fs.Sub(StaticDir, "web/dist")
	return embedFS{http.FS(sub)}
}
```

- [ ] **Step 2: Commit**

```bash
git add embed.go
git commit -m "feat(embed): serve Vue SPA from embedded web/dist with fallback"
```

---

### Task 34: TLS + graceful shutdown + router (single handler struct)

**Files:**
- Create: `pkg/stockd/http/handler/handler.go`
- Create: `pkg/stockd/http/middleware.go`
- Create: `pkg/stockd/http/router.go`
- Modify: `cmd/stockd/main.go`

- [ ] **Step 1: Write `pkg/stockd/http/handler/handler.go`**

```go
// Package handler implements HTTP handlers for the stockd API.
package handler

import (
	"stock/pkg/stockd/services/analysis"
	"stock/pkg/stockd/services/bars"
	"stock/pkg/stockd/services/draft"
	"stock/pkg/stockd/services/portfolio"
	"stock/pkg/stockd/services/scheduler"
	"stock/pkg/stockd/services/stock"
	"stock/pkg/stockd/services/token"
	"stock/pkg/stockd/services/user"
)

type handler struct {
	userSvc      *user.Service
	tokenSvc     *token.Service
	stockSvc     *stock.Service
	portfolioSvc *portfolio.Service
	draftSvc     *draft.Service
	barsSvc      *bars.Service
	analysisSvc  *analysis.Service
	schedulerSvc *scheduler.Service
}

func NewHandler(
	userSvc *user.Service,
	tokenSvc *token.Service,
	stockSvc *stock.Service,
	portfolioSvc *portfolio.Service,
	draftSvc *draft.Service,
	barsSvc *bars.Service,
	analysisSvc *analysis.Service,
	schedulerSvc *scheduler.Service,
) *handler {
	return &handler{
		userSvc: userSvc, tokenSvc: tokenSvc, stockSvc: stockSvc,
		portfolioSvc: portfolioSvc, draftSvc: draftSvc, barsSvc: barsSvc,
		analysisSvc: analysisSvc, schedulerSvc: schedulerSvc,
	}
}
```

- [ ] **Step 2: Write `pkg/stockd/http/middleware.go`**

```go
package http

import (
	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

// AuthRequired aborts with ErrUnauthorized when no authenticated user is present.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.User(c) == nil {
			utils.HTTPRequestFailedV4(c, nil, utils.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminRequired aborts with ErrForbidden when the user is not an admin.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := auth.User(c)
		if u == nil {
			utils.HTTPRequestFailedV4(c, nil, utils.ErrUnauthorized)
			c.Abort()
			return
		}
		if u.Role != "admin" {
			utils.HTTPRequestFailedV4(c, nil, utils.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 3: Write `pkg/stockd/http/router.go`**

```go
package http

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"stock"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/config"
	"stock/pkg/stockd/http/handler"
	"stock/pkg/stockd/services/analysis"
	"stock/pkg/stockd/services/bars"
	"stock/pkg/stockd/services/draft"
	"stock/pkg/stockd/services/portfolio"
	"stock/pkg/stockd/services/scheduler"
	"stock/pkg/stockd/services/stock"
	"stock/pkg/stockd/services/token"
	"stock/pkg/stockd/services/user"
	"stock/pkg/stockd/utils"
	"stock/pkg/tushare"
)

func NewRouter(gdb *gorm.DB, cfg *config.Config, sched *scheduler.Service) *gin.Engine {
	// Services
	userSvc := user.New(gdb)
	tokenSvc := token.New(gdb)
	stockSvc := stock.New(gdb)
	portfolioSvc := portfolio.New(gdb)
	draftSvc := draft.New(gdb)
	barsSvc := bars.New(gdb, tushare.NewClient(tushare.WithBaseURL(cfg.Tushare.BaseURL)))
	analysisSvc := analysis.New(gdb)

	h := handler.NewHandler(userSvc, tokenSvc, stockSvc, portfolioSvc, draftSvc, barsSvc, analysisSvc, sched)

	// Router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(utils.Recovery())
	r.Use(utils.RequestID())
	r.Use(utils.Language())

	// CORS (allows Vue dev server at :5173 and carries credentials/custom headers)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization", "Lang")
	r.Use(cors.New(corsConfig))

	// Session store
	store := auth.NewSessionStore([]byte(cfg.Server.SessionSecret))
	r.Use(sessions.Sessions(auth.SessionName, store))

	// Soft auth resolver (attaches user if present, never aborts)
	r.Use(auth.ResolveUser(gdb, cfg.Tushare.DefaultToken))

	// Swagger (public)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")

	// Auth routes (public)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/logout", h.Logout)
	api.GET("/auth/me", AuthRequired(), h.Me)

	// Admin routes
	adm := api.Group("/admin")
	adm.Use(AuthRequired(), AdminRequired())
	adm.POST("/users", h.CreateUser)
	adm.GET("/users", h.ListUsers)
	adm.PATCH("/users/:id", h.PatchUser)
	adm.DELETE("/users/:id", h.DeleteUser)
	adm.POST("/stocks/sync", h.SyncStocklist)
	adm.POST("/stocks/import-csv", h.ImportCSV)
	adm.POST("/bars/sync", h.SyncBars)
	adm.GET("/sync/status", h.JobStatus)

	// Self-service routes
	me := api.Group("/me")
	me.Use(AuthRequired())
	me.GET("/tokens", h.ListTokens)
	me.POST("/tokens", h.IssueToken)
	me.DELETE("/tokens/:id", h.RevokeToken)
	me.PATCH("/tushare_token", h.SetTushareToken)
	me.POST("/password", h.ChangePassword)

	// Stock catalog (public read)
	api.GET("/stocks", h.SearchStocks)
	api.GET("/stocks/:tsCode", h.GetStock)

	// Portfolio
	pr := api.Group("/portfolio")
	pr.Use(AuthRequired())
	pr.GET("", h.ListPortfolio)
	pr.POST("", h.AddPortfolio)
	pr.DELETE("/:tsCode", h.RemovePortfolio)
	pr.PATCH("/:tsCode", h.UpdatePortfolioNote)

	// Bars
	br := api.Group("/bars")
	br.Use(AuthRequired())
	br.GET("/:tsCode", h.QueryBars)

	// Drafts
	dr := api.Group("/drafts")
	dr.Use(AuthRequired())
	dr.GET("/today", h.GetDraftToday)
	dr.PUT("", h.UpsertDraft)
	dr.DELETE("/:id", h.DeleteDraft)

	// Analysis
	anr := api.Group("/analysis")
	anr.Use(AuthRequired())
	anr.GET("/:tsCode", h.GetAnalysis)

	// Static SPA mount
	r.Use(static.Serve("/", stock.EmbedFolder()))
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.JSON(404, utils.HTTPResponse{Code: 404, Message: "not found"})
			return
		}
		c.FileFromFS("/web/dist/index.html", http.FS(stock.StaticDir))
	})

	return r
}
```

- [ ] **Step 4: Write `cmd/stockd/main.go`**

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"stock/pkg/stockd/bootstrap"
	"stock/pkg/stockd/config"
	"stock/pkg/stockd/db"
	httpkg "stock/pkg/stockd/http"
	"stock/pkg/stockd/services/bars"
	"stock/pkg/stockd/services/portfolio"
	"stock/pkg/stockd/services/scheduler"
	"stock/pkg/stockd/services/stock"
	"stock/pkg/tushare"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("stockd %s (built %s)\n", Version, BuildTime)
		return
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfgPath := os.Getenv("STOCKD_CONFIG")
	if cfgPath == "" {
		cfgPath = "/etc/stockd/config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.WithError(err).Fatal("config load failed")
	}

	lvl, _ := logrus.ParseLevel(cfg.Logging.Level)
	logger.SetLevel(lvl)

	gdb, err := db.Open(cfg)
	if err != nil {
		logger.WithError(err).Fatal("database open failed")
	}

	if _, err := bootstrap.EnsureAdmin(gdb, logger); err != nil {
		logger.WithError(err).Fatal("bootstrap failed")
	}

	sched := scheduler.New(gdb)
	if cfg.Scheduler.Enabled {
		tc := tushare.NewClient(tushare.WithBaseURL(cfg.Tushare.BaseURL))
		barsSvc := bars.New(gdb, tc)
		stockSvc := stock.New(gdb)
		portfolioSvc := portfolio.New(gdb)

		sched.RegisterCron("daily-fetch", cfg.Scheduler.DailyFetchCron, func(ctx context.Context) error {
			codes, err := portfolioSvc.DistinctTsCodes(ctx)
			if err != nil {
				return err
			}
			for _, code := range codes {
				if _, err := barsSvc.Sync(ctx, cfg.Tushare.DefaultToken, code); err != nil {
					logger.WithError(err).WithField("ts_code", code).Error("daily sync failed")
				}
			}
			return nil
		})
		sched.RegisterCron("stocklist-sync", cfg.Scheduler.StocklistSyncCron, func(ctx context.Context) error {
			_, err := stockSvc.SyncFromTushare(ctx, cfg.Tushare.DefaultToken)
			return err
		})
		sched.Start()
		defer sched.Stop()
	}

	router := httpkg.NewRouter(gdb, cfg, sched)

	srv := &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: router,
	}

	go func() {
		logger.WithField("addr", cfg.Server.Listen).Info("starting server")
		var err error
		if cfg.Server.TLS.Enabled {
			err = srv.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("shutdown error")
	}
}
```

- [ ] **Step 5: Build and run**

```bash
make build
./bin/stockd version
```

Expected: `stockd <version> (built <time>)`.

- [ ] **Step 6: Commit**

```bash
git add cmd/stockd/main.go pkg/stockd/http/router.go pkg/stockd/http/middleware.go
git add pkg/stockd/http/handler/handler.go
git commit -m "feat(stockd): real main with TLS, scheduler, graceful shutdown, mtk response format"
```

---

## Exit criterion

- [ ] `go test ./pkg/stockd/...` green (utils, auth, http, handler)
- [ ] `make build` produces `bin/stockd` that starts and serves `/api/auth/login`, `/api/stocks`, `/swagger/index.html`
- [ ] All API responses have shape `{requestID, code, message, data}` with HTTP 200 (including errors)
- [ ] Recovery middleware catches panics and returns `{code: 500, message: ...}`
- [ ] Request ID is present in every response
- [ ] SPA fallback works
- [ ] TLS mode verified

## Hand-off

Next phases can run in parallel:
- [P5 — CLI + Skill](./2026-05-14-p5-cli-skill.md)
- [P6 — Frontend](./2026-05-14-p6-frontend.md)
