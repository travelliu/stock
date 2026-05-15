package http

import (
	"net/http"
	"stock/pkg/stockd/services"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	root "stock"
	"stock/pkg/stockd/auth"

	"stock/pkg/stockd/services/analysis"
	"stock/pkg/stockd/utils"
)

func initGin(logger *logrus.Logger) *gin.Engine {
	gin.DisableConsoleColor()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		logger.Infof("%v %v %v %v", httpMethod, absolutePath, handlerName, nuHandlers)
	}
	router := gin.New()

	router.Use(utils.RequestID(), Logger(logger), Cors())
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(Recovery(logger))
	router.Use(utils.Language())

	return router
}

func NewRouter(svc *services.Service, logger *logrus.Logger) *gin.Engine {

	analysisSvc := analysis.New(svc.GetDB(), nil)

	h := NewHandler(svc, analysisSvc)

	r := initGin(logger)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization", "Lang")
	r.Use(cors.New(corsConfig))

	store := auth.NewSessionStore([]byte(svc.GetConfig().Server.SessionSecret))
	r.Use(sessions.Sessions(auth.SessionName, store))

	r.Use(auth.ResolveUser(svc.GetDB(), svc.GetConfig().Tushare.DefaultToken))

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
	api.GET("/stocks/"+codeUrl, h.GetStock)

	pr := api.Group("/portfolio")
	pr.Use(AuthRequired())
	pr.GET("", h.ListPortfolio)
	pr.POST("", h.AddPortfolio)
	prTs := pr.Group("/" + codeUrl)
	prTs.DELETE("", h.RemovePortfolio)
	prTs.PATCH("", h.UpdatePortfolioNote)

	br := api.Group("/bars")
	br.Use(AuthRequired())
	br.GET("/"+codeUrl, h.QueryBars)

	dr := api.Group("/drafts")
	dr.Use(AuthRequired())
	dr.GET("/today", h.GetDraftToday)
	dr.PUT("", h.UpsertDraft)
	dr.DELETE("/:id", h.DeleteDraft)

	anr := api.Group("/analysis")
	anr.Use(AuthRequired())
	anr.GET("/"+codeUrl, h.GetAnalysis)
	anr.POST("/recalc", h.RecalcPredictions)
	anr.GET("/predictions/"+codeUrl, h.ListPredictions)

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
