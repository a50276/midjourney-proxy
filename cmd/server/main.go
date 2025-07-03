package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"midjourney-proxy-go/internal/api"
	"midjourney-proxy-go/internal/infrastructure/config"
	"midjourney-proxy-go/internal/infrastructure/database"
	"midjourney-proxy-go/internal/infrastructure/discord"
	"midjourney-proxy-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// @title Midjourney Proxy API
// @version 1.0
// @description A powerful, complete, full-featured, completely free and open source Midjourney API project.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name GPL v3.0
// @license.url https://www.gnu.org/licenses/gpl-3.0.html

// @host localhost:8080
// @BasePath /
func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)

	// 初始化数据库
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}

	// 自动迁移
	if err := database.Migrate(db); err != nil {
		logger.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化Discord连接管理器
	discordManager := discord.NewManager(cfg.Discord, logger)

	// 设置Gin模式
	if cfg.App.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	router := api.NewRouter(cfg, db, discordManager, logger)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Infof("Server starting on port %d", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 启动Discord连接
	go func() {
		if err := discordManager.Start(); err != nil {
			logger.Errorf("Failed to start Discord manager: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	// 关闭Discord连接
	discordManager.Stop()

	logger.Info("Server exited")
}