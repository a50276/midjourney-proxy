package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"midjourney-proxy-go/internal/domain/entity"
	"midjourney-proxy-go/internal/infrastructure/discord"
	"midjourney-proxy-go/pkg/logger"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	db             *gorm.DB
	discordManager *discord.Manager
	logger         logger.Logger
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(db *gorm.DB, discordManager *discord.Manager, logger logger.Logger) *TaskHandler {
	return &TaskHandler{
		db:             db,
		discordManager: discordManager,
		logger:         logger,
	}
}

// SubmitImagineRequest 提交Imagine请求
type SubmitImagineRequest struct {
	Prompt        string                      `json:"prompt" binding:"required"`
	Base64Array   []string                    `json:"base64Array,omitempty"`
	BotType       string                      `json:"botType,omitempty"`
	State         string                      `json:"state,omitempty"`
	NotifyHook    string                      `json:"notifyHook,omitempty"`
	AccountFilter *entity.AccountFilter       `json:"accountFilter,omitempty"`
	Mode          entity.GenerationSpeedMode  `json:"mode,omitempty"`
}

// SubmitChangeRequest 提交变化请求
type SubmitChangeRequest struct {
	TaskID        string                `json:"taskId" binding:"required"`
	Action        entity.TaskAction     `json:"action" binding:"required"`
	Index         *int                  `json:"index,omitempty"`
	State         string                `json:"state,omitempty"`
	NotifyHook    string                `json:"notifyHook,omitempty"`
	AccountFilter *entity.AccountFilter `json:"accountFilter,omitempty"`
}

// SubmitDescribeRequest 提交描述请求
type SubmitDescribeRequest struct {
	Base64        string                `json:"base64,omitempty"`
	Link          string                `json:"link,omitempty"`
	BotType       string                `json:"botType,omitempty"`
	State         string                `json:"state,omitempty"`
	NotifyHook    string                `json:"notifyHook,omitempty"`
	AccountFilter *entity.AccountFilter `json:"accountFilter,omitempty"`
}

// SubmitResultVO 提交结果
type SubmitResultVO struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  string      `json:"result,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResult 成功结果
func SuccessResult(taskID string) SubmitResultVO {
	return SubmitResultVO{
		Code:    1,
		Message: "提交成功",
		Result:  taskID,
	}
}

// ErrorResult 错误结果
func ErrorResult(code int, message string) SubmitResultVO {
	return SubmitResultVO{
		Code:    code,
		Message: message,
	}
}

// SubmitImagine 提交Imagine任务
// @Summary 提交Imagine任务
// @Description 提交一个Imagine绘图任务
// @Tags 任务提交
// @Accept json
// @Produce json
// @Param request body SubmitImagineRequest true "Imagine请求"
// @Success 200 {object} SubmitResultVO
// @Router /api/mj/submit/imagine [post]
func (h *TaskHandler) SubmitImagine(c *gin.Context) {
	var req SubmitImagineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 获取客户端IP
	clientIP := c.ClientIP()

	// 创建任务
	task := &entity.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		BotType:       entity.BotTypeMidjourney,
		Action:        entity.TaskActionImagine,
		Status:        entity.TaskStatusNotStart,
		Prompt:        req.Prompt,
		Description:   "/imagine " + req.Prompt,
		State:         req.State,
		ClientIP:      clientIP,
		Mode:          req.Mode,
		AccountFilter: req.AccountFilter,
	}

	// 设置Bot类型
	if req.BotType == "NIJI_JOURNEY" {
		task.BotType = entity.BotTypeNijijourney
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务到数据库
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	// 获取可用的Discord实例
	instance := h.discordManager.GetAvailableInstance()
	if instance == nil {
		// 更新任务状态为失败
		task.Fail("没有可用的Discord实例")
		h.db.Save(task)
		c.JSON(http.StatusServiceUnavailable, ErrorResult(50300, "没有可用的Discord实例"))
		return
	}

	// 启动任务
	task.Start()
	task.InstanceID = instance.ID
	h.db.Save(task)

	// 异步提交到Discord
	go func() {
		h.logger.Infof("Submitting imagine task %s to Discord instance %s", task.ID, instance.ID)
		
		// TODO: 实际提交到Discord
		// 这里应该调用Discord API提交任务
		
		// 模拟处理时间
		time.Sleep(2 * time.Second)
		
		// 模拟成功
		task.Success()
		task.ImageURL = "https://example.com/image.jpg" // 模拟图片URL
		h.db.Save(task)
		
		h.logger.Infof("Task %s completed successfully", task.ID)
	}()

	h.logger.Infof("Task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// SubmitChange 提交变化任务
// @Summary 提交变化任务
// @Description 提交一个变化任务（放大、变化、重新生成）
// @Tags 任务提交
// @Accept json
// @Produce json
// @Param request body SubmitChangeRequest true "变化请求"
// @Success 200 {object} SubmitResultVO
// @Router /api/mj/submit/change [post]
func (h *TaskHandler) SubmitChange(c *gin.Context) {
	var req SubmitChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 查找父任务
	var parentTask entity.Task
	if err := h.db.Where("id = ?", req.TaskID).First(&parentTask).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResult(40400, "关联任务不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResult(50000, "查询任务失败"))
		}
		return
	}

	// 检查父任务状态
	if parentTask.Status != entity.TaskStatusSuccess {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "关联任务状态错误"))
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 创建新任务
	task := &entity.Task{
		ID:          uuid.New().String(),
		ParentID:    parentTask.ID,
		UserID:      userID,
		BotType:     parentTask.BotType,
		RealBotType: parentTask.RealBotType,
		Action:      req.Action,
		Status:      entity.TaskStatusNotStart,
		Prompt:      parentTask.Prompt,
		PromptEn:    parentTask.PromptEn,
		State:       req.State,
		ClientIP:    c.ClientIP(),
		InstanceID:  parentTask.InstanceID,
	}

	// 设置描述
	switch req.Action {
	case entity.TaskActionUpscale:
		task.Description = "/up " + req.TaskID + " U" + strconv.Itoa(*req.Index)
	case entity.TaskActionVariation:
		task.Description = "/up " + req.TaskID + " V" + strconv.Itoa(*req.Index)
	case entity.TaskActionReroll:
		task.Description = "/up " + req.TaskID + " R"
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create change task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	// 启动任务
	task.Start()
	h.db.Save(task)

	h.logger.Infof("Change task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// SubmitSimpleChange 提交简单变化
func (h *TaskHandler) SubmitSimpleChange(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
		State   string `json:"state,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 创建任务
	task := &entity.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		BotType:     entity.BotTypeMidjourney,
		Action:      entity.TaskActionAction,
		Status:      entity.TaskStatusNotStart,
		Description: "/action " + req.Content,
		State:       req.State,
		ClientIP:    c.ClientIP(),
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create simple change task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	h.logger.Infof("Simple change task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// SubmitDescribe 提交描述任务
func (h *TaskHandler) SubmitDescribe(c *gin.Context) {
	var req SubmitDescribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	if req.Base64 == "" && req.Link == "" {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "base64或link不能为空"))
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 创建任务
	task := &entity.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		BotType:     entity.BotTypeMidjourney,
		Action:      entity.TaskActionDescribe,
		Status:      entity.TaskStatusNotStart,
		Description: "/describe",
		State:       req.State,
		ClientIP:    c.ClientIP(),
	}

	// 设置Bot类型
	if req.BotType == "NIJI_JOURNEY" {
		task.BotType = entity.BotTypeNijijourney
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create describe task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	h.logger.Infof("Describe task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// SubmitBlend 提交混合任务
func (h *TaskHandler) SubmitBlend(c *gin.Context) {
	var req struct {
		Base64Array   []string              `json:"base64Array" binding:"required,min=2,max=5"`
		Dimensions    string                `json:"dimensions,omitempty"`
		State         string                `json:"state,omitempty"`
		NotifyHook    string                `json:"notifyHook,omitempty"`
		AccountFilter *entity.AccountFilter `json:"accountFilter,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 创建任务
	task := &entity.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		BotType:       entity.BotTypeMidjourney,
		Action:        entity.TaskActionBlend,
		Status:        entity.TaskStatusNotStart,
		Description:   "/blend",
		State:         req.State,
		ClientIP:      c.ClientIP(),
		AccountFilter: req.AccountFilter,
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create blend task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	// 启动任务
	task.Start()
	h.db.Save(task)

	h.logger.Infof("Blend task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// SubmitShorten 提交缩短任务
func (h *TaskHandler) SubmitShorten(c *gin.Context) {
	var req struct {
		Prompt        string                `json:"prompt" binding:"required"`
		State         string                `json:"state,omitempty"`
		NotifyHook    string                `json:"notifyHook,omitempty"`
		AccountFilter *entity.AccountFilter `json:"accountFilter,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 创建任务
	task := &entity.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		BotType:       entity.BotTypeMidjourney,
		Action:        entity.TaskActionShorten,
		Status:        entity.TaskStatusNotStart,
		Prompt:        req.Prompt,
		Description:   "/shorten " + req.Prompt,
		State:         req.State,
		ClientIP:      c.ClientIP(),
		AccountFilter: req.AccountFilter,
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create shorten task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	// 启动任务
	task.Start()
	h.db.Save(task)

	h.logger.Infof("Shorten task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// SubmitShow 提交显示任务
func (h *TaskHandler) SubmitShow(c *gin.Context) {
	var req struct {
		TaskID        string                `json:"taskId" binding:"required"`
		State         string                `json:"state,omitempty"`
		NotifyHook    string                `json:"notifyHook,omitempty"`
		AccountFilter *entity.AccountFilter `json:"accountFilter,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 查找关联任务
	var parentTask entity.Task
	if err := h.db.Where("id = ?", req.TaskID).First(&parentTask).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResult(40400, "关联任务不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResult(50000, "查询任务失败"))
		}
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 创建任务
	task := &entity.Task{
		ID:            uuid.New().String(),
		ParentID:      parentTask.ID,
		UserID:        userID,
		BotType:       parentTask.BotType,
		Action:        entity.TaskActionShow,
		Status:        entity.TaskStatusNotStart,
		Description:   "/show " + req.TaskID,
		State:         req.State,
		ClientIP:      c.ClientIP(),
		AccountFilter: req.AccountFilter,
		InstanceID:    parentTask.InstanceID,
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create show task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	// 启动任务
	task.Start()
	h.db.Save(task)

	h.logger.Infof("Show task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// SubmitAction 提交动作任务
func (h *TaskHandler) SubmitAction(c *gin.Context) {
	var req struct {
		TaskID     string `json:"taskId" binding:"required"`
		CustomID   string `json:"customId" binding:"required"`
		State      string `json:"state,omitempty"`
		NotifyHook string `json:"notifyHook,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 查找关联任务
	var parentTask entity.Task
	if err := h.db.Where("id = ?", req.TaskID).First(&parentTask).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResult(40400, "关联任务不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResult(50000, "查询任务失败"))
		}
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 创建任务
	task := &entity.Task{
		ID:          uuid.New().String(),
		ParentID:    parentTask.ID,
		UserID:      userID,
		BotType:     parentTask.BotType,
		Action:      entity.TaskActionAction,
		Status:      entity.TaskStatusNotStart,
		Description: "/action " + req.CustomID,
		State:       req.State,
		ClientIP:    c.ClientIP(),
		InstanceID:  parentTask.InstanceID,
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create action task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	// 启动任务
	task.Start()
	h.db.Save(task)

	h.logger.Infof("Action task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// SubmitModal 提交模态任务
func (h *TaskHandler) SubmitModal(c *gin.Context) {
	var req struct {
		TaskID     string            `json:"taskId" binding:"required"`
		Prompt     string            `json:"prompt" binding:"required"`
		MaskBase64 string            `json:"maskBase64,omitempty"`
		State      string            `json:"state,omitempty"`
		NotifyHook string            `json:"notifyHook,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 查找关联任务
	var parentTask entity.Task
	if err := h.db.Where("id = ?", req.TaskID).First(&parentTask).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResult(40400, "关联任务不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResult(50000, "查询任务失败"))
		}
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 创建任务
	task := &entity.Task{
		ID:          uuid.New().String(),
		ParentID:    parentTask.ID,
		UserID:      userID,
		BotType:     parentTask.BotType,
		Action:      entity.TaskActionModal,
		Status:      entity.TaskStatusNotStart,
		Prompt:      req.Prompt,
		Description: "/modal " + req.Prompt,
		State:       req.State,
		ClientIP:    c.ClientIP(),
		InstanceID:  parentTask.InstanceID,
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create modal task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建任务失败"))
		return
	}

	// 启动任务
	task.Start()
	h.db.Save(task)

	h.logger.Infof("Modal task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// UploadDiscordImages 上传Discord图片
func (h *TaskHandler) UploadDiscordImages(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResult(40000, "参数错误: "+err.Error()))
		return
	}

	// 获取用户信息
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// 获取可用的Discord实例
	instance := h.discordManager.GetAvailableInstance()
	if instance == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResult(50300, "没有可用的Discord实例"))
		return
	}

	// 创建上传任务
	task := &entity.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		BotType:     entity.BotTypeMidjourney,
		Action:      "UPLOAD",
		Status:      entity.TaskStatusNotStart,
		Description: "Upload images to Discord",
		ClientIP:    c.ClientIP(),
		InstanceID:  instance.ID,
	}

	// 设置提交时间
	now := time.Now()
	task.SubmitTime = &now

	// 保存任务
	if err := h.db.Create(task).Error; err != nil {
		h.logger.Errorf("Failed to create upload task: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResult(50000, "创建上传任务失败"))
		return
	}

	// 启动任务
	task.Start()
	h.db.Save(task)

	h.logger.Infof("Upload task %s submitted by user %s", task.ID, userID)
	c.JSON(http.StatusOK, SuccessResult(task.ID))
}

// GetTask 获取任务
// @Summary 获取任务
// @Description 根据ID获取任务信息
// @Tags 任务查询
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} entity.Task
// @Router /api/mj/task/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "任务ID不能为空",
		})
		return
	}

	var task entity.Task
	if err := h.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    40400,
				"message": "任务不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "查询任务失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

// FetchTask 获取任务状态
func (h *TaskHandler) FetchTask(c *gin.Context) {
	h.GetTask(c)
}

// ListTasks 列出任务
// @Summary 列出任务
// @Description 获取任务列表
// @Tags 任务查询
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} gin.H
// @Router /api/mj/task/list [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	// 获取用户ID
	userID := "guest"
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	var tasks []entity.Task
	var total int64

	query := h.db.Model(&entity.Task{})
	if userID != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	// 获取总数
	query.Count(&total)

	// 获取任务列表
	query.Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&tasks)

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data": gin.H{
			"list":  tasks,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// GetQueue 获取队列状态
func (h *TaskHandler) GetQueue(c *gin.Context) {
	// 统计各状态的任务数量
	var stats []struct {
		Status entity.TaskStatus `json:"status"`
		Count  int64             `json:"count"`
	}

	h.db.Model(&entity.Task{}).
		Select("status, count(*) as count").
		Where("status IN ?", []entity.TaskStatus{
			entity.TaskStatusNotStart,
			entity.TaskStatusSubmitted,
			entity.TaskStatusInProgress,
		}).
		Group("status").
		Find(&stats)

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "查询成功",
		"data":    stats,
	})
}

// AdminList 管理员任务列表
func (h *TaskHandler) AdminList(c *gin.Context) {
	h.ListTasks(c)
}

// AdminGet 管理员获取任务
func (h *TaskHandler) AdminGet(c *gin.Context) {
	h.GetTask(c)
}

// AdminDelete 管理员删除任务
func (h *TaskHandler) AdminDelete(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "任务ID不能为空",
		})
		return
	}

	if err := h.db.Where("id = ?", taskID).Delete(&entity.Task{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "删除成功",
	})
}

// AdminRetry 管理员重试任务
func (h *TaskHandler) AdminRetry(c *gin.Context) {
	// TODO: 实现重试任务逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    50100,
		"message": "功能暂未实现",
	})
}