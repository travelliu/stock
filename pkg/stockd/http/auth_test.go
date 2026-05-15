package http_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	http2 "stock/pkg/stockd/http"
	"stock/pkg/stockd/services"
	"strings"
	"testing"
	
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	
	"stock/pkg/models"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/db"
	"stock/pkg/stockd/services/analysis"
	"stock/pkg/stockd/utils"
	"stock/pkg/tushare"
)

func setupAuthRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	
	r := gin.New()
	r.Use(utils.RequestID())
	r.Use(utils.Language())
	store := auth.NewSessionStore([]byte("12345678901234567890123456789012"))
	r.Use(sessions.Sessions(auth.SessionName, store))
	r.Use(auth.ResolveUser(gdb, ""))
	
	userSvc := services.New(gdb)
	tokenSvc := services.New(gdb)
	stockSvc := services.New(gdb)
	portfolioSvc := services.New(gdb)
	draftSvc := services.New(gdb)
	barsSvc := services.New(gdb, tushare.NewClient())
	analysisSvc := analysis.New(gdb, nil)
	predictionSvc := services.New(gdb)
	schedulerSvc := services.New(gdb)
	h := http2.NewHandler(userSvc, tokenSvc, stockSvc, portfolioSvc, draftSvc, barsSvc, analysisSvc, predictionSvc, schedulerSvc)
	r.POST("/api/auth/login", h.Login)
	r.POST("/api/auth/logout", h.Logout)
	r.GET("/api/auth/me", http2.AuthRequired(), h.Me)
	return r, gdb
}

func TestLogin(t *testing.T) {
	r, gdb := setupAuthRouter(t)
	pw, _ := auth.HashPassword("secret")
	require.NoError(t, gdb.Create(&models.User{Username: "alice", PasswordHash: pw, Role: "user"}).Error)
	
	body, _ := json.Marshal(models.LoginReq{Username: "alice", Password: "secret"})
	w := httptest.NewRecorder()
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
	
	body, _ := json.Marshal(models.LoginReq{Username: "alice", Password: "wrong"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
}
