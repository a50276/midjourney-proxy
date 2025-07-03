package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"midjourney-proxy-go/internal/api/middleware"
	"midjourney-proxy-go/internal/domain/entity"
	"midjourney-proxy-go/internal/infrastructure/config"
	"midjourney-proxy-go/pkg/logger"
)

// UserHandler 用户处理器
type UserHandler struct {
	db     *gorm.DB
	config *config.Config
	logger logger.Logger
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB, config *config.Config, logger logger.Logger) *UserHandler {
	return &UserHandler{
		db:     db,
		config: config,
		logger: logger,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Token     string `json:"token"`
		TokenType string `json:"token_type"`
		ExpiresIn int    `json:"expires_in"`
		User      *UserInfo `json:"user"`
	} `json:"data,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email,omitempty"`
	Role           string     `json:"role"`
	Enabled        bool       `json:"enabled"`
	TotalDrawCount int        `json:"total_draw_count"`
	DayDrawCount   int        `json:"day_draw_count"`
	DayDrawLimit   int        `json:"day_draw_limit"`
	TotalDrawLimit int        `json:"total_draw_limit"`
	ExpiredAt      *time.Time `json:"expired_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT token
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} LoginResponse
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 查找用户
	var user entity.User
	err := h.db.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40100,
				"message": "用户名或密码错误",
			})
		} else {
			h.logger.Errorf("Failed to query user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "登录失败",
			})
		}
		return
	}

	// 验证密码
	if err := user.CheckPassword(req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": "用户名或密码错误",
		})
		return
	}

	// 检查用户状态
	if !user.Enabled {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    40300,
			"message": "账号已被禁用",
		})
		return
	}

	if user.IsExpired() {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    40300,
			"message": "账号已过期",
		})
		return
	}

	// 生成JWT token
	token, err := middleware.GenerateJWTToken(
		user.ID,
		user.Role,
		h.config.Security.JWTSecret,
		h.config.Security.JWTExpireHours,
	)
	if err != nil {
		h.logger.Errorf("Failed to generate JWT token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "生成token失败",
		})
		return
	}

	// 返回登录成功
	response := LoginResponse{
		Code:    1,
		Message: "登录成功",
		Data: &struct {
			Token     string `json:"token"`
			TokenType string `json:"token_type"`
			ExpiresIn int    `json:"expires_in"`
			User      *UserInfo `json:"user"`
		}{
			Token:     token,
			TokenType: "Bearer",
			ExpiresIn: h.config.Security.JWTExpireHours * 3600,
			User: &UserInfo{
				ID:             user.ID,
				Username:       user.Username,
				Email:          user.Email,
				Role:           string(user.Role),
				Enabled:        user.Enabled,
				TotalDrawCount: user.TotalDrawCount,
				DayDrawCount:   user.DayDrawCount,
				DayDrawLimit:   user.DayDrawLimit,
				TotalDrawLimit: user.TotalDrawLimit,
				ExpiredAt:      user.ExpiredAt,
				CreatedAt:      user.CreatedAt,
			},
		},
	}

	h.logger.Infof("User %s logged in successfully", user.Username)
	c.JSON(http.StatusOK, response)
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求"
// @Success 200 {object} LoginResponse
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	if !h.config.App.EnableRegister {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    40300,
			"message": "注册功能已关闭",
		})
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser entity.User
	err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    40900,
			"message": "用户名或邮箱已存在",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		h.logger.Errorf("Failed to check existing user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "注册失败",
		})
		return
	}

	// 创建新用户
	user := entity.User{
		ID:       uuid.New().String(),
		Username: req.Username,
		Email:    req.Email,
		Role:     entity.RoleUser,
		Enabled:  true,
	}

	// 设置密码
	if err := user.SetPassword(req.Password); err != nil {
		h.logger.Errorf("Failed to set password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "注册失败",
		})
		return
	}

	// 保存用户
	if err := h.db.Create(&user).Error; err != nil {
		h.logger.Errorf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "注册失败",
		})
		return
	}

	// 生成JWT token
	token, err := middleware.GenerateJWTToken(
		user.ID,
		user.Role,
		h.config.Security.JWTSecret,
		h.config.Security.JWTExpireHours,
	)
	if err != nil {
		h.logger.Errorf("Failed to generate JWT token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "生成token失败",
		})
		return
	}

	// 返回注册成功
	response := LoginResponse{
		Code:    1,
		Message: "注册成功",
		Data: &struct {
			Token     string `json:"token"`
			TokenType string `json:"token_type"`
			ExpiresIn int    `json:"expires_in"`
			User      *UserInfo `json:"user"`
		}{
			Token:     token,
			TokenType: "Bearer",
			ExpiresIn: h.config.Security.JWTExpireHours * 3600,
			User: &UserInfo{
				ID:             user.ID,
				Username:       user.Username,
				Email:          user.Email,
				Role:           string(user.Role),
				Enabled:        user.Enabled,
				TotalDrawCount: user.TotalDrawCount,
				DayDrawCount:   user.DayDrawCount,
				DayDrawLimit:   user.DayDrawLimit,
				TotalDrawLimit: user.TotalDrawLimit,
				ExpiredAt:      user.ExpiredAt,
				CreatedAt:      user.CreatedAt,
			},
		},
	}

	h.logger.Infof("User %s registered successfully", user.Username)
	c.JSON(http.StatusCreated, response)
}

// RefreshToken 刷新token
func (h *UserHandler) RefreshToken(c *gin.Context) {
	// 获取当前用户
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": "用户未登录",
		})
		return
	}

	user := userInterface.(*entity.User)

	// 生成新的JWT token
	token, err := middleware.GenerateJWTToken(
		user.ID,
		user.Role,
		h.config.Security.JWTSecret,
		h.config.Security.JWTExpireHours,
	)
	if err != nil {
		h.logger.Errorf("Failed to generate JWT token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "刷新token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "刷新成功",
		"data": gin.H{
			"token":      token,
			"token_type": "Bearer",
			"expires_in": h.config.Security.JWTExpireHours * 3600,
		},
	})
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "登出成功",
	})
}

// List 用户列表（管理员）
func (h *UserHandler) List(c *gin.Context) {
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

	var users []entity.User
	var total int64

	query := h.db.Model(&entity.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	query.Count(&total)

	// 获取用户列表
	query.Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&users)

	// 转换为用户信息
	var userInfos []UserInfo
	for _, user := range users {
		userInfos = append(userInfos, UserInfo{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Role:           string(user.Role),
			Enabled:        user.Enabled,
			TotalDrawCount: user.TotalDrawCount,
			DayDrawCount:   user.DayDrawCount,
			DayDrawLimit:   user.DayDrawLimit,
			TotalDrawLimit: user.TotalDrawLimit,
			ExpiredAt:      user.ExpiredAt,
			CreatedAt:      user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data": gin.H{
			"list":  userInfos,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// Create 创建用户（管理员）
func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username       string              `json:"username" binding:"required"`
		Email          string              `json:"email" binding:"required,email"`
		Password       string              `json:"password" binding:"required,min=6"`
		Role           entity.UserRole     `json:"role"`
		Enabled        bool                `json:"enabled"`
		DayDrawLimit   int                 `json:"day_draw_limit"`
		TotalDrawLimit int                 `json:"total_draw_limit"`
		ExpiredAt      *time.Time          `json:"expired_at"`
		Remark         string              `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser entity.User
	err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    40900,
			"message": "用户名或邮箱已存在",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		h.logger.Errorf("Failed to check existing user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建用户失败",
		})
		return
	}

	// 创建用户
	user := entity.User{
		ID:             uuid.New().String(),
		Username:       req.Username,
		Email:          req.Email,
		Role:           req.Role,
		Enabled:        req.Enabled,
		DayDrawLimit:   req.DayDrawLimit,
		TotalDrawLimit: req.TotalDrawLimit,
		ExpiredAt:      req.ExpiredAt,
		Remark:         req.Remark,
	}

	if user.Role == "" {
		user.Role = entity.RoleUser
	}

	// 设置密码
	if err := user.SetPassword(req.Password); err != nil {
		h.logger.Errorf("Failed to set password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建用户失败",
		})
		return
	}

	// 保存用户
	if err := h.db.Create(&user).Error; err != nil {
		h.logger.Errorf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建用户失败",
		})
		return
	}

	h.logger.Infof("User %s created by admin", user.Username)
	c.JSON(http.StatusCreated, gin.H{
		"code":    1,
		"message": "创建成功",
		"data": UserInfo{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Role:           string(user.Role),
			Enabled:        user.Enabled,
			TotalDrawCount: user.TotalDrawCount,
			DayDrawCount:   user.DayDrawCount,
			DayDrawLimit:   user.DayDrawLimit,
			TotalDrawLimit: user.TotalDrawLimit,
			ExpiredAt:      user.ExpiredAt,
			CreatedAt:      user.CreatedAt,
		},
	})
}

// Get 获取用户（管理员）
func (h *UserHandler) Get(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "用户ID不能为空",
		})
		return
	}

	var user entity.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    40400,
				"message": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "查询用户失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data": UserInfo{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Role:           string(user.Role),
			Enabled:        user.Enabled,
			TotalDrawCount: user.TotalDrawCount,
			DayDrawCount:   user.DayDrawCount,
			DayDrawLimit:   user.DayDrawLimit,
			TotalDrawLimit: user.TotalDrawLimit,
			ExpiredAt:      user.ExpiredAt,
			CreatedAt:      user.CreatedAt,
		},
	})
}

// Update 更新用户（管理员）
func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "用户ID不能为空",
		})
		return
	}

	var req struct {
		Email          *string             `json:"email,omitempty"`
		Password       *string             `json:"password,omitempty"`
		Role           *entity.UserRole    `json:"role,omitempty"`
		Enabled        *bool               `json:"enabled,omitempty"`
		DayDrawLimit   *int                `json:"day_draw_limit,omitempty"`
		TotalDrawLimit *int                `json:"total_draw_limit,omitempty"`
		ExpiredAt      *time.Time          `json:"expired_at,omitempty"`
		Remark         *string             `json:"remark,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 查找用户
	var user entity.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    40400,
				"message": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "查询用户失败",
			})
		}
		return
	}

	// 更新字段
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Password != nil {
		if err := user.SetPassword(*req.Password); err != nil {
			h.logger.Errorf("Failed to set password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "更新用户失败",
			})
			return
		}
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Enabled != nil {
		user.Enabled = *req.Enabled
	}
	if req.DayDrawLimit != nil {
		user.DayDrawLimit = *req.DayDrawLimit
	}
	if req.TotalDrawLimit != nil {
		user.TotalDrawLimit = *req.TotalDrawLimit
	}
	if req.ExpiredAt != nil {
		user.ExpiredAt = req.ExpiredAt
	}
	if req.Remark != nil {
		user.Remark = *req.Remark
	}

	// 保存更新
	if err := h.db.Save(&user).Error; err != nil {
		h.logger.Errorf("Failed to update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新用户失败",
		})
		return
	}

	h.logger.Infof("User %s updated by admin", user.Username)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "更新成功",
	})
}

// Delete 删除用户（管理员）
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "用户ID不能为空",
		})
		return
	}

	if err := h.db.Where("id = ?", userID).Delete(&entity.User{}).Error; err != nil {
		h.logger.Errorf("Failed to delete user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除用户失败",
		})
		return
	}

	h.logger.Infof("User %s deleted by admin", userID)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "删除成功",
	})
}