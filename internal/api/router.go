package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	
	"midjourney-proxy-go/internal/api/handler"
	"midjourney-proxy-go/internal/api/middleware"
	"midjourney-proxy-go/internal/infrastructure/config"
	"midjourney-proxy-go/internal/infrastructure/discord"
	"midjourney-proxy-go/pkg/logger"
)

// NewRouter 创建路由器
func NewRouter(
	cfg *config.Config,
	db *gorm.DB,
	discordManager *discord.Manager,
	logger logger.Logger,
) *gin.Engine {
	// 创建Gin引擎
	router := gin.New()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// 静态文件服务
	router.Static("/static", "./web/dist")
	router.StaticFile("/", "./web/dist/index.html")
	router.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   cfg.App.Version,
		})
	})

	// 创建处理器
	taskHandler := handler.NewTaskHandler(db, discordManager, logger)
	accountHandler := handler.NewAccountHandler(db, discordManager, logger)
	userHandler := handler.NewUserHandler(db, cfg, logger)
	adminHandler := handler.NewAdminHandler(db, discordManager, cfg, logger)

	// API路由组
	api := router.Group("/api")
	{
		// 认证中间件
		authMiddleware := middleware.Auth(cfg.Security, db, logger)

		// 公开API
		public := api.Group("/public")
		{
			public.GET("/info", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"name":     cfg.App.Name,
					"version":  cfg.App.Version,
					"demo":     cfg.App.DemoMode,
					"register": cfg.App.EnableRegister,
					"guest":    cfg.App.EnableGuest,
				})
			})
		}

		// 用户认证
		auth := api.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			if cfg.App.EnableRegister {
				auth.POST("/register", userHandler.Register)
			}
			auth.POST("/refresh", authMiddleware, userHandler.RefreshToken)
			auth.POST("/logout", authMiddleware, userHandler.Logout)
		}

		// 任务提交API
		submit := api.Group("/mj/submit")
		if !cfg.App.EnableGuest {
			submit.Use(authMiddleware)
		}
		submit.Use(middleware.RateLimit(cfg.RateLimit, logger))
		{
			submit.POST("/imagine", taskHandler.SubmitImagine)
			submit.POST("/change", taskHandler.SubmitChange)
			submit.POST("/simple-change", taskHandler.SubmitSimpleChange)
			submit.POST("/describe", taskHandler.SubmitDescribe)
			submit.POST("/blend", taskHandler.SubmitBlend)
			submit.POST("/shorten", taskHandler.SubmitShorten)
			submit.POST("/show", taskHandler.SubmitShow)
			submit.POST("/action", taskHandler.SubmitAction)
			submit.POST("/modal", taskHandler.SubmitModal)
			submit.POST("/upload-discord-images", taskHandler.UploadDiscordImages)
		}

		// 任务查询API
		task := api.Group("/mj/task")
		if !cfg.App.EnableGuest {
			task.Use(authMiddleware)
		}
		{
			task.GET("/:id", taskHandler.GetTask)
			task.GET("/:id/fetch", taskHandler.FetchTask)
			task.GET("/list", taskHandler.ListTasks)
			task.GET("/queue", taskHandler.GetQueue)
		}

		// 管理员API
		admin := api.Group("/admin")
		admin.Use(authMiddleware)
		admin.Use(middleware.AdminOnly())
		{
			// 账号管理
			accounts := admin.Group("/accounts")
			{
				accounts.GET("", accountHandler.List)
				accounts.POST("", accountHandler.Create)
				accounts.GET("/:id", accountHandler.Get)
				accounts.PUT("/:id", accountHandler.Update)
				accounts.DELETE("/:id", accountHandler.Delete)
				accounts.POST("/:id/sync", accountHandler.Sync)
				accounts.POST("/:id/enable", accountHandler.Enable)
				accounts.POST("/:id/disable", accountHandler.Disable)
			}

			// 用户管理
			users := admin.Group("/users")
			{
				users.GET("", userHandler.List)
				users.POST("", userHandler.Create)
				users.GET("/:id", userHandler.Get)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}

			// 任务管理
			tasks := admin.Group("/tasks")
			{
				tasks.GET("", taskHandler.AdminList)
				tasks.GET("/:id", taskHandler.AdminGet)
				tasks.DELETE("/:id", taskHandler.AdminDelete)
				tasks.POST("/:id/retry", taskHandler.AdminRetry)
			}

			// 系统设置
			settings := admin.Group("/settings")
			{
				settings.GET("", adminHandler.GetSettings)
				settings.PUT("", adminHandler.UpdateSettings)
				settings.GET("/info", adminHandler.GetSystemInfo)
			}

			// 统计信息
			stats := admin.Group("/stats")
			{
				stats.GET("/overview", adminHandler.GetOverview)
				stats.GET("/tasks", adminHandler.GetTaskStats)
				stats.GET("/accounts", adminHandler.GetAccountStats)
			}

			// 禁用词管理
			bannedWords := admin.Group("/banned-words")
			{
				bannedWords.GET("", adminHandler.ListBannedWords)
				bannedWords.POST("", adminHandler.CreateBannedWord)
				bannedWords.PUT("/:id", adminHandler.UpdateBannedWord)
				bannedWords.DELETE("/:id", adminHandler.DeleteBannedWord)
			}

			// 领域标签管理
			domainTags := admin.Group("/domain-tags")
			{
				domainTags.GET("", adminHandler.ListDomainTags)
				domainTags.POST("", adminHandler.CreateDomainTag)
				domainTags.PUT("/:id", adminHandler.UpdateDomainTag)
				domainTags.DELETE("/:id", adminHandler.DeleteDomainTag)
			}
		}
	}

	// Swagger文档
	if cfg.App.Mode == "development" {
		swaggerFiles := ginSwagger.WrapHandler(swaggerFiles.Handler)
		router.GET("/swagger/*any", swaggerFiles)
		router.GET("/docs", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
		})
	}

	return router
}