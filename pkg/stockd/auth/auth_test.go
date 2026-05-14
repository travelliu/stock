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

	"stock/pkg/models"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/db"

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
