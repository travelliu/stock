package http

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	root "stock"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/config"
	"stock/pkg/stockd/services/analysis"
	"stock/pkg/stockd/services/bars"
	"stock/pkg/stockd/services/draft"
	"stock/pkg/stockd/services/portfolio"
	"stock/pkg/stockd/services/scheduler"
	stocksvc "stock/pkg/stockd/services/stock"
	"stock/pkg/stockd/services/token"
	"stock/pkg/stockd/services/user"
	"stock/pkg/stockd/utils"
	"stock/pkg/tushare"
)

func NewRouter(gdb *gorm.DB, cfg *config.Config, sched *scheduler.Service) *gin.Engine {
	userSvc := user.New(gdb)
	tokenSvc := token.New(gdb)
	stockSvc := stocksvc.New(gdb)
	portfolioSvc := portfolio.New(gdb)
	draftSvc := draft.New(gdb)
	barsSvc := bars.New(gdb, tushare.NewClient(tushare.WithBaseURL(cfg.Tushare.BaseURL)))
	analysisSvc := analysis.New(gdb)

	h := NewHandler(userSvc, tokenSvc, stockSvc, portfolioSvc, draftSvc, barsSvc, analysisSvc, sched)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(utils.Recovery())
	r.Use(utils.RequestID())
	r.Use(utils.Language())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization", "Lang")
	r.Use(cors.New(corsConfig))

	store := auth.NewSessionStore([]byte(cfg.Server.SessionSecret))
	r.Use(sessions.Sessions(auth.SessionName, store))

	r.Use(auth.ResolveUser(gdb, cfg.Tushare.DefaultToken))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")

	api.POST("/auth/login", h.Login)
	api.POST("/auth/logout", h.Logout)
	api.GET("/auth/me", AuthRequired(), h.Me)

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

	me := api.Group("/me")
	me.Use(AuthRequired())
	me.GET("/tokens", h.ListTokens)
	me.POST("/tokens", h.IssueToken)
	me.DELETE("/tokens/:id", h.RevokeToken)
	me.PATCH("/tushare_token", h.SetTushareToken)
	me.POST("/password", h.ChangePassword)

	api.GET("/stocks", h.SearchStocks)
	api.GET("/stocks/"+tsCodeUrl, h.GetStock)

	pr := api.Group("/portfolio")
	pr.Use(AuthRequired())
	pr.GET("", h.ListPortfolio)
	pr.POST("", h.AddPortfolio)
	prTs := pr.Group("/" + tsCodeUrl)
	prTs.DELETE("", h.RemovePortfolio)
	prTs.PATCH("", h.UpdatePortfolioNote)

	br := api.Group("/bars")
	br.Use(AuthRequired())
	br.GET("/"+tsCodeUrl, h.QueryBars)

	dr := api.Group("/drafts")
	dr.Use(AuthRequired())
	dr.GET("/today", h.GetDraftToday)
	dr.PUT("", h.UpsertDraft)
	dr.DELETE("/:id", h.DeleteDraft)

	anr := api.Group("/analysis")
	anr.Use(AuthRequired())
	anr.GET("/"+tsCodeUrl, h.GetAnalysis)

	r.Use(static.Serve("/", root.EmbedFolder()))
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.JSON(404, utils.HTTPResponse{Code: 404, Message: "not found"})
			return
		}
		c.FileFromFS("/web/dist/index.html", http.FS(root.StaticDir))
	})

	return r
}
