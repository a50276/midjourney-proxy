package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"midjourney-proxy-go/internal/domain/entity"
	"midjourney-proxy-go/internal/infrastructure/discord"
	"midjourney-proxy-go/pkg/logger"
)

// AccountHandler 账号处理器
type AccountHandler struct {
	db             *gorm.DB
	discordManager *discord.Manager
	logger         logger.Logger
}

// NewAccountHandler 创建账号处理器
func NewAccountHandler(db *gorm.DB, discordManager *discord.Manager, logger logger.Logger) *AccountHandler {
	return &AccountHandler{
		db:             db,
		discordManager: discordManager,
		logger:         logger,
	}
}

// List 获取账号列表
func (h *AccountHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	var accounts []entity.DiscordAccount
	var total int64

	query := h.db.Model(&entity.DiscordAccount{})
	if keyword != "" {
		query = query.Where("channel_id LIKE ? OR guild_id LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	query.Count(&total)

	// 获取账号列表
	query.Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&accounts)

	// 更新运行状态信息
	instances := h.discordManager.GetAllInstances()
	for i := range accounts {
		if instance, exists := instances[accounts[i].ID]; exists {
			accounts[i].Running = instance.IsConnected()
			accounts[i].RunningCount = 0 // TODO: 实际的运行任务数
			accounts[i].QueueCount = 0   // TODO: 实际的队列任务数
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data": gin.H{
			"list":  accounts,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// Create 创建账号
func (h *AccountHandler) Create(c *gin.Context) {
	var req struct {
		ChannelID            string                       `json:"channel_id" binding:"required"`
		GuildID              string                       `json:"guild_id" binding:"required"`
		UserToken            string                       `json:"user_token" binding:"required"`
		BotToken             string                       `json:"bot_token"`
		UserAgent            string                       `json:"user_agent"`
		Enabled              bool                         `json:"enabled"`
		EnableMJ             bool                         `json:"enable_mj"`
		EnableNiji           bool                         `json:"enable_niji"`
		CoreSize             int                          `json:"core_size"`
		QueueSize            int                          `json:"queue_size"`
		MaxQueueSize         int                          `json:"max_queue_size"`
		TimeoutMinutes       int                          `json:"timeout_minutes"`
		Interval             float64                      `json:"interval"`
		Weight               int                          `json:"weight"`
		Sort                 int                          `json:"sort"`
		WorkTime             string                       `json:"work_time"`
		FishingTime          string                       `json:"fishing_time"`
		DayDrawLimit         int                          `json:"day_draw_limit"`
		Mode                 entity.GenerationSpeedMode  `json:"mode"`
		AllowModes           []entity.GenerationSpeedMode `json:"allow_modes"`
		RemixAutoSubmit      bool                         `json:"remix_auto_submit"`
		EnableAutoSetRelax   bool                         `json:"enable_auto_set_relax"`
		IsBlend              bool                         `json:"is_blend"`
		IsDescribe           bool                         `json:"is_describe"`
		IsShorten            bool                         `json:"is_shorten"`
		Remark               string                       `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查频道ID是否已存在
	var existingAccount entity.DiscordAccount
	err := h.db.Where("channel_id = ?", req.ChannelID).First(&existingAccount).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    40900,
			"message": "频道ID已存在",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		h.logger.Errorf("Failed to check existing account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建账号失败",
		})
		return
	}

	// 创建账号
	account := entity.DiscordAccount{
		ID:                 uuid.New().String(),
		ChannelID:          req.ChannelID,
		GuildID:            req.GuildID,
		UserToken:          req.UserToken,
		BotToken:           req.BotToken,
		UserAgent:          req.UserAgent,
		Enabled:            req.Enabled,
		EnableMJ:           req.EnableMJ,
		EnableNiji:         req.EnableNiji,
		CoreSize:           req.CoreSize,
		QueueSize:          req.QueueSize,
		MaxQueueSize:       req.MaxQueueSize,
		TimeoutMinutes:     req.TimeoutMinutes,
		Interval:           req.Interval,
		Weight:             req.Weight,
		Sort:               req.Sort,
		WorkTime:           req.WorkTime,
		FishingTime:        req.FishingTime,
		DayDrawLimit:       req.DayDrawLimit,
		Mode:               req.Mode,
		AllowModes:         req.AllowModes,
		RemixAutoSubmit:    req.RemixAutoSubmit,
		EnableAutoSetRelax: req.EnableAutoSetRelax,
		IsBlend:            req.IsBlend,
		IsDescribe:         req.IsDescribe,
		IsShorten:          req.IsShorten,
		Remark:             req.Remark,
	}

	// 设置默认值
	if account.UserAgent == "" {
		account.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
	if account.CoreSize <= 0 {
		account.CoreSize = 3
	}
	if account.QueueSize <= 0 {
		account.QueueSize = 10
	}
	if account.MaxQueueSize <= 0 {
		account.MaxQueueSize = 100
	}
	if account.TimeoutMinutes <= 0 {
		account.TimeoutMinutes = 5
	}
	if account.Interval <= 0 {
		account.Interval = 1.2
	}
	if account.DayDrawLimit == 0 {
		account.DayDrawLimit = -1
	}

	// 保存账号
	if err := h.db.Create(&account).Error; err != nil {
		h.logger.Errorf("Failed to create account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建账号失败",
		})
		return
	}

	h.logger.Infof("Discord account %s created", account.ChannelID)
	c.JSON(http.StatusCreated, gin.H{
		"code":    1,
		"message": "创建成功",
		"data":    account,
	})
}

// Get 获取账号
func (h *AccountHandler) Get(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "账号ID不能为空",
		})
		return
	}

	var account entity.DiscordAccount
	if err := h.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    40400,
				"message": "账号不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "查询账号失败",
			})
		}
		return
	}

	// 更新运行状态信息
	if instance := h.discordManager.GetInstance(account.ID); instance != nil {
		account.Running = instance.IsConnected()
		account.RunningCount = 0 // TODO: 实际的运行任务数
		account.QueueCount = 0   // TODO: 实际的队列任务数
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data":    account,
	})
}

// Update 更新账号
func (h *AccountHandler) Update(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "账号ID不能为空",
		})
		return
	}

	var req struct {
		UserToken            *string                       `json:"user_token,omitempty"`
		BotToken             *string                       `json:"bot_token,omitempty"`
		UserAgent            *string                       `json:"user_agent,omitempty"`
		Enabled              *bool                         `json:"enabled,omitempty"`
		EnableMJ             *bool                         `json:"enable_mj,omitempty"`
		EnableNiji           *bool                         `json:"enable_niji,omitempty"`
		CoreSize             *int                          `json:"core_size,omitempty"`
		QueueSize            *int                          `json:"queue_size,omitempty"`
		MaxQueueSize         *int                          `json:"max_queue_size,omitempty"`
		TimeoutMinutes       *int                          `json:"timeout_minutes,omitempty"`
		Interval             *float64                      `json:"interval,omitempty"`
		Weight               *int                          `json:"weight,omitempty"`
		Sort                 *int                          `json:"sort,omitempty"`
		WorkTime             *string                       `json:"work_time,omitempty"`
		FishingTime          *string                       `json:"fishing_time,omitempty"`
		DayDrawLimit         *int                          `json:"day_draw_limit,omitempty"`
		Mode                 *entity.GenerationSpeedMode  `json:"mode,omitempty"`
		AllowModes           *[]entity.GenerationSpeedMode `json:"allow_modes,omitempty"`
		RemixAutoSubmit      *bool                         `json:"remix_auto_submit,omitempty"`
		EnableAutoSetRelax   *bool                         `json:"enable_auto_set_relax,omitempty"`
		IsBlend              *bool                         `json:"is_blend,omitempty"`
		IsDescribe           *bool                         `json:"is_describe,omitempty"`
		IsShorten            *bool                         `json:"is_shorten,omitempty"`
		Remark               *string                       `json:"remark,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 查找账号
	var account entity.DiscordAccount
	if err := h.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    40400,
				"message": "账号不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "查询账号失败",
			})
		}
		return
	}

	// 更新字段
	if req.UserToken != nil {
		account.UserToken = *req.UserToken
	}
	if req.BotToken != nil {
		account.BotToken = *req.BotToken
	}
	if req.UserAgent != nil {
		account.UserAgent = *req.UserAgent
	}
	if req.Enabled != nil {
		account.Enabled = *req.Enabled
	}
	if req.EnableMJ != nil {
		account.EnableMJ = *req.EnableMJ
	}
	if req.EnableNiji != nil {
		account.EnableNiji = *req.EnableNiji
	}
	if req.CoreSize != nil {
		account.CoreSize = *req.CoreSize
	}
	if req.QueueSize != nil {
		account.QueueSize = *req.QueueSize
	}
	if req.MaxQueueSize != nil {
		account.MaxQueueSize = *req.MaxQueueSize
	}
	if req.TimeoutMinutes != nil {
		account.TimeoutMinutes = *req.TimeoutMinutes
	}
	if req.Interval != nil {
		account.Interval = *req.Interval
	}
	if req.Weight != nil {
		account.Weight = *req.Weight
	}
	if req.Sort != nil {
		account.Sort = *req.Sort
	}
	if req.WorkTime != nil {
		account.WorkTime = *req.WorkTime
	}
	if req.FishingTime != nil {
		account.FishingTime = *req.FishingTime
	}
	if req.DayDrawLimit != nil {
		account.DayDrawLimit = *req.DayDrawLimit
	}
	if req.Mode != nil {
		account.Mode = *req.Mode
	}
	if req.AllowModes != nil {
		account.AllowModes = *req.AllowModes
	}
	if req.RemixAutoSubmit != nil {
		account.RemixAutoSubmit = *req.RemixAutoSubmit
	}
	if req.EnableAutoSetRelax != nil {
		account.EnableAutoSetRelax = *req.EnableAutoSetRelax
	}
	if req.IsBlend != nil {
		account.IsBlend = *req.IsBlend
	}
	if req.IsDescribe != nil {
		account.IsDescribe = *req.IsDescribe
	}
	if req.IsShorten != nil {
		account.IsShorten = *req.IsShorten
	}
	if req.Remark != nil {
		account.Remark = *req.Remark
	}

	// 保存更新
	if err := h.db.Save(&account).Error; err != nil {
		h.logger.Errorf("Failed to update account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新账号失败",
		})
		return
	}

	h.logger.Infof("Discord account %s updated", account.ChannelID)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "更新成功",
	})
}

// Delete 删除账号
func (h *AccountHandler) Delete(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "账号ID不能为空",
		})
		return
	}

	// 查找账号
	var account entity.DiscordAccount
	if err := h.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    40400,
				"message": "账号不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "查询账号失败",
			})
		}
		return
	}

	// 从Discord管理器中移除
	h.discordManager.RemoveAccount(accountID)

	// 删除账号
	if err := h.db.Where("id = ?", accountID).Delete(&entity.DiscordAccount{}).Error; err != nil {
		h.logger.Errorf("Failed to delete account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除账号失败",
		})
		return
	}

	h.logger.Infof("Discord account %s deleted", account.ChannelID)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "删除成功",
	})
}

// Sync 同步账号信息
func (h *AccountHandler) Sync(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "账号ID不能为空",
		})
		return
	}

	// 查找账号
	var account entity.DiscordAccount
	if err := h.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    40400,
				"message": "账号不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "查询账号失败",
			})
		}
		return
	}

	// TODO: 实现同步账号信息逻辑
	// 1. 连接Discord获取最新信息
	// 2. 更新账号设置
	// 3. 保存到数据库

	h.logger.Infof("Discord account %s sync requested", account.ChannelID)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "同步成功",
	})
}

// Enable 启用账号
func (h *AccountHandler) Enable(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "账号ID不能为空",
		})
		return
	}

	// 更新账号状态
	if err := h.db.Model(&entity.DiscordAccount{}).Where("id = ?", accountID).Update("enabled", true).Error; err != nil {
		h.logger.Errorf("Failed to enable account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "启用账号失败",
		})
		return
	}

	h.logger.Infof("Discord account %s enabled", accountID)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "启用成功",
	})
}

// Disable 禁用账号
func (h *AccountHandler) Disable(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "账号ID不能为空",
		})
		return
	}

	// 更新账号状态
	if err := h.db.Model(&entity.DiscordAccount{}).Where("id = ?", accountID).Update("enabled", false).Error; err != nil {
		h.logger.Errorf("Failed to disable account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "禁用账号失败",
		})
		return
	}

	h.logger.Infof("Discord account %s disabled", accountID)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "禁用成功",
	})
}