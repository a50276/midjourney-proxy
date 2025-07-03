package handler

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"midjourney-proxy-go/internal/domain/entity"
	"midjourney-proxy-go/internal/infrastructure/config"
	"midjourney-proxy-go/internal/infrastructure/discord"
	"midjourney-proxy-go/pkg/logger"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	db             *gorm.DB
	discordManager *discord.Manager
	config         *config.Config
	logger         logger.Logger
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(db *gorm.DB, discordManager *discord.Manager, config *config.Config, logger logger.Logger) *AdminHandler {
	return &AdminHandler{
		db:             db,
		discordManager: discordManager,
		config:         config,
		logger:         logger,
	}
}

// GetSettings 获取系统设置
func (h *AdminHandler) GetSettings(c *gin.Context) {
	// TODO: 实现获取系统设置逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data":    h.config,
	})
}

// UpdateSettings 更新系统设置
func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	// TODO: 实现更新系统设置逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// GetSystemInfo 获取系统信息
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 获取Discord实例状态
	instances := h.discordManager.GetAllInstances()
	var connectedCount, totalCount int
	for _, instance := range instances {
		totalCount++
		if instance.IsConnected() {
			connectedCount++
		}
	}

	// 获取任务统计
	var taskStats struct {
		Total      int64 `json:"total"`
		Success    int64 `json:"success"`
		Failed     int64 `json:"failed"`
		InProgress int64 `json:"in_progress"`
	}

	h.db.Model(&entity.Task{}).Count(&taskStats.Total)
	h.db.Model(&entity.Task{}).Where("status = ?", entity.TaskStatusSuccess).Count(&taskStats.Success)
	h.db.Model(&entity.Task{}).Where("status = ?", entity.TaskStatusFailure).Count(&taskStats.Failed)
	h.db.Model(&entity.Task{}).Where("status IN ?", []entity.TaskStatus{
		entity.TaskStatusSubmitted,
		entity.TaskStatusInProgress,
	}).Count(&taskStats.InProgress)

	// 获取用户统计
	var userStats struct {
		Total   int64 `json:"total"`
		Enabled int64 `json:"enabled"`
		Admin   int64 `json:"admin"`
	}

	h.db.Model(&entity.User{}).Count(&userStats.Total)
	h.db.Model(&entity.User{}).Where("enabled = ?", true).Count(&userStats.Enabled)
	h.db.Model(&entity.User{}).Where("role = ?", entity.RoleAdmin).Count(&userStats.Admin)

	systemInfo := gin.H{
		"app": gin.H{
			"name":    h.config.App.Name,
			"version": h.config.App.Version,
			"mode":    h.config.App.Mode,
		},
		"runtime": gin.H{
			"go_version":   runtime.Version(),
			"goroutines":   runtime.NumGoroutine(),
			"memory_used":  bToMb(m.Alloc),
			"memory_total": bToMb(m.TotalAlloc),
			"memory_sys":   bToMb(m.Sys),
			"gc_runs":      m.NumGC,
		},
		"discord": gin.H{
			"total_instances":     totalCount,
			"connected_instances": connectedCount,
		},
		"statistics": gin.H{
			"tasks": taskStats,
			"users": userStats,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data":    systemInfo,
	})
}

// GetOverview 获取概览统计
func (h *AdminHandler) GetOverview(c *gin.Context) {
	// 获取今日任务统计
	var todayStats struct {
		TotalTasks    int64 `json:"total_tasks"`
		SuccessTasks  int64 `json:"success_tasks"`
		FailedTasks   int64 `json:"failed_tasks"`
		PendingTasks  int64 `json:"pending_tasks"`
	}

	today := "date(created_at) = date('now')"
	if h.config.Database.Type == "mysql" {
		today = "date(created_at) = curdate()"
	}

	h.db.Model(&entity.Task{}).Where(today).Count(&todayStats.TotalTasks)
	h.db.Model(&entity.Task{}).Where(today+" AND status = ?", entity.TaskStatusSuccess).Count(&todayStats.SuccessTasks)
	h.db.Model(&entity.Task{}).Where(today+" AND status = ?", entity.TaskStatusFailure).Count(&todayStats.FailedTasks)
	h.db.Model(&entity.Task{}).Where(today+" AND status IN ?", []entity.TaskStatus{
		entity.TaskStatusNotStart,
		entity.TaskStatusSubmitted,
		entity.TaskStatusInProgress,
	}).Count(&todayStats.PendingTasks)

	// 获取Discord实例状态
	instances := h.discordManager.GetAllInstances()
	var instanceStats []gin.H
	for _, instance := range instances {
		account := instance.GetAccount()
		instanceStats = append(instanceStats, gin.H{
			"id":         instance.ID,
			"channel_id": account.ChannelID,
			"enabled":    account.Enabled,
			"connected":  instance.IsConnected(),
			"last_ping":  instance.LastPing,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data": gin.H{
			"today_stats":    todayStats,
			"instance_stats": instanceStats,
		},
	})
}

// GetTaskStats 获取任务统计
func (h *AdminHandler) GetTaskStats(c *gin.Context) {
	// TODO: 实现任务统计逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// GetAccountStats 获取账号统计
func (h *AdminHandler) GetAccountStats(c *gin.Context) {
	// TODO: 实现账号统计逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// ListBannedWords 获取禁用词列表
func (h *AdminHandler) ListBannedWords(c *gin.Context) {
	var words []entity.BannedWord
	h.db.Order("created_at DESC").Find(&words)

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data":    words,
	})
}

// CreateBannedWord 创建禁用词
func (h *AdminHandler) CreateBannedWord(c *gin.Context) {
	// TODO: 实现创建禁用词逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// UpdateBannedWord 更新禁用词
func (h *AdminHandler) UpdateBannedWord(c *gin.Context) {
	// TODO: 实现更新禁用词逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// DeleteBannedWord 删除禁用词
func (h *AdminHandler) DeleteBannedWord(c *gin.Context) {
	// TODO: 实现删除禁用词逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// ListDomainTags 获取领域标签列表
func (h *AdminHandler) ListDomainTags(c *gin.Context) {
	var tags []entity.DomainTag
	h.db.Order("sort ASC, created_at DESC").Find(&tags)

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data":    tags,
	})
}

// CreateDomainTag 创建领域标签
func (h *AdminHandler) CreateDomainTag(c *gin.Context) {
	// TODO: 实现创建领域标签逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// UpdateDomainTag 更新领域标签
func (h *AdminHandler) UpdateDomainTag(c *gin.Context) {
	// TODO: 实现更新领域标签逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// DeleteDomainTag 删除领域标签
func (h *AdminHandler) DeleteDomainTag(c *gin.Context) {
	// TODO: 实现删除领域标签逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}

// bToMb 字节转MB
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}