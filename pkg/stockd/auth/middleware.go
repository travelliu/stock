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
