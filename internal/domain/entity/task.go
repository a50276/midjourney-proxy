package entity

import (
	"time"
	"encoding/json"
	"gorm.io/gorm"
)

// TaskStatus 任务状态枚举
type TaskStatus string

const (
	TaskStatusNotStart   TaskStatus = "NOT_START"   // 未启动
	TaskStatusSubmitted  TaskStatus = "SUBMITTED"   // 已提交
	TaskStatusInProgress TaskStatus = "IN_PROGRESS" // 执行中
	TaskStatusFailure    TaskStatus = "FAILURE"     // 失败
	TaskStatusSuccess    TaskStatus = "SUCCESS"     // 成功
	TaskStatusCancel     TaskStatus = "CANCEL"      // 取消
)

// TaskAction 任务动作枚举
type TaskAction string

const (
	TaskActionImagine   TaskAction = "IMAGINE"   // 想象
	TaskActionUpscale   TaskAction = "UPSCALE"   // 放大
	TaskActionVariation TaskAction = "VARIATION" // 变化
	TaskActionReroll    TaskAction = "REROLL"    // 重新生成
	TaskActionDescribe  TaskAction = "DESCRIBE"  // 描述
	TaskActionBlend     TaskAction = "BLEND"     // 混合
	TaskActionShorten   TaskAction = "SHORTEN"   // 缩短
	TaskActionShow      TaskAction = "SHOW"      // 显示
	TaskActionPan       TaskAction = "PAN"       // 平移
	TaskActionZoom      TaskAction = "ZOOM"      // 缩放
	TaskActionVary      TaskAction = "VARY"      // 局部重绘
	TaskActionModal     TaskAction = "MODAL"     // 模态
	TaskActionAction    TaskAction = "ACTION"    // 行动
)

// BotType 机器人类型枚举
type BotType string

const (
	BotTypeMidjourney  BotType = "MID_JOURNEY"
	BotTypeNijijourney BotType = "NIJI_JOURNEY"
)

// GenerationSpeedMode 生成速度模式枚举
type GenerationSpeedMode string

const (
	SpeedModeFast  GenerationSpeedMode = "FAST"
	SpeedModeRelax GenerationSpeedMode = "RELAX"
	SpeedModeTurbo GenerationSpeedMode = "TURBO"
)

// CustomComponent 自定义组件
type CustomComponent struct {
	Type     int    `json:"type"`
	Style    int    `json:"style"`
	Label    string `json:"label"`
	Emoji    string `json:"emoji"`
	CustomID string `json:"custom_id"`
}

// AccountFilter 账号过滤器
type AccountFilter struct {
	InstanceID   string                `json:"instance_id,omitempty"`
	Mode         GenerationSpeedMode   `json:"mode,omitempty"`
	BotType      BotType               `json:"bot_type,omitempty"`
	RemixEnabled bool                  `json:"remix_enabled,omitempty"`
	Modes        []GenerationSpeedMode `json:"modes,omitempty"`
}

// Task 任务实体
type Task struct {
	ID          string          `gorm:"column:id;primaryKey" json:"id"`
	ParentID    string          `gorm:"column:parent_id;index" json:"parent_id,omitempty"`
	UserID      string          `gorm:"column:user_id;index" json:"user_id"`
	BotType     BotType         `gorm:"column:bot_type" json:"bot_type"`
	RealBotType *BotType        `gorm:"column:real_bot_type" json:"real_bot_type,omitempty"`
	IsWhite     bool            `gorm:"column:is_white;default:false" json:"is_white"`
	
	// 消息相关
	Nonce                   string `gorm:"column:nonce" json:"nonce,omitempty"`
	InteractionMetadataID   string `gorm:"column:interaction_metadata_id" json:"interaction_metadata_id,omitempty"`
	MessageID               string `gorm:"column:message_id" json:"message_id,omitempty"`
	RemixModalMessageID     string `gorm:"column:remix_modal_message_id" json:"remix_modal_message_id,omitempty"`
	RemixAutoSubmit         bool   `gorm:"column:remix_auto_submit;default:false" json:"remix_auto_submit"`
	RemixModaling           bool   `gorm:"column:remix_modaling;default:false" json:"remix_modaling"`
	
	// 实例相关
	InstanceID    string `gorm:"column:instance_id;index" json:"instance_id"`
	SubInstanceID string `gorm:"column:sub_instance_id" json:"sub_instance_id,omitempty"`
	
	// 任务信息
	Action      TaskAction `gorm:"column:action" json:"action"`
	Status      TaskStatus `gorm:"column:status;index" json:"status"`
	Prompt      string     `gorm:"column:prompt;type:text" json:"prompt,omitempty"`
	PromptEn    string     `gorm:"column:prompt_en;type:text" json:"prompt_en,omitempty"`
	PromptFull  string     `gorm:"column:prompt_full;type:text" json:"prompt_full,omitempty"`
	Description string     `gorm:"column:description;type:text" json:"description,omitempty"`
	State       string     `gorm:"column:state" json:"state,omitempty"`
	
	// 时间相关
	SubmitTime *time.Time `gorm:"column:submit_time;index" json:"submit_time,omitempty"`
	StartTime  *time.Time `gorm:"column:start_time" json:"start_time,omitempty"`
	FinishTime *time.Time `gorm:"column:finish_time" json:"finish_time,omitempty"`
	
	// 结果相关
	ImageURL     string `gorm:"column:image_url;size:1024" json:"image_url,omitempty"`
	ThumbnailURL string `gorm:"column:thumbnail_url;size:1024" json:"thumbnail_url,omitempty"`
	Progress     string `gorm:"column:progress" json:"progress,omitempty"`
	FailReason   string `gorm:"column:fail_reason;type:text" json:"fail_reason,omitempty"`
	
	// 按钮和组件
	Buttons []CustomComponent `gorm:"column:buttons;type:json" json:"buttons,omitempty"`
	
	// 种子和图片信息
	Seed          string `gorm:"column:seed" json:"seed,omitempty"`
	SeedMessageID string `gorm:"column:seed_message_id" json:"seed_message_id,omitempty"`
	JobID         string `gorm:"column:job_id" json:"job_id,omitempty"`
	
	// 网络相关
	ClientIP string `gorm:"column:client_ip;index" json:"client_ip,omitempty"`
	
	// 换脸相关
	IsReplicate      bool   `gorm:"column:is_replicate;default:false" json:"is_replicate"`
	ReplicateSource  string `gorm:"column:replicate_source;size:1024" json:"replicate_source,omitempty"`
	ReplicateTarget  string `gorm:"column:replicate_target;size:1024" json:"replicate_target,omitempty"`
	
	// 生成模式
	Mode GenerationSpeedMode `gorm:"column:mode" json:"mode,omitempty"`
	
	// 过滤器
	AccountFilterData string `gorm:"column:account_filter;type:text" json:"-"`
	AccountFilter     *AccountFilter `gorm:"-" json:"account_filter,omitempty"`
	
	// 原始内容信息
	URL         string `gorm:"column:url;size:1024" json:"url,omitempty"`
	ProxyURL    string `gorm:"column:proxy_url;size:1024" json:"proxy_url,omitempty"`
	Height      *int   `gorm:"column:height" json:"height,omitempty"`
	Width       *int   `gorm:"column:width" json:"width,omitempty"`
	Size        *int64 `gorm:"column:size" json:"size,omitempty"`
	ContentType string `gorm:"column:content_type;size:200" json:"content_type,omitempty"`
	
	// 扩展属性
	Properties map[string]interface{} `gorm:"column:properties;type:json" json:"properties,omitempty"`
	
	// 时间戳
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// BeforeSave GORM钩子，保存前序列化复杂字段
func (t *Task) BeforeSave(tx *gorm.DB) error {
	if t.AccountFilter != nil {
		data, err := json.Marshal(t.AccountFilter)
		if err != nil {
			return err
		}
		t.AccountFilterData = string(data)
	}
	return nil
}

// AfterFind GORM钩子，查询后反序列化复杂字段
func (t *Task) AfterFind(tx *gorm.DB) error {
	if t.AccountFilterData != "" {
		var filter AccountFilter
		if err := json.Unmarshal([]byte(t.AccountFilterData), &filter); err == nil {
			t.AccountFilter = &filter
		}
	}
	return nil
}

// Start 启动任务
func (t *Task) Start() {
	now := time.Now()
	t.StartTime = &now
	t.Status = TaskStatusSubmitted
	t.Progress = "0%"
}

// Success 任务成功
func (t *Task) Success() {
	now := time.Now()
	t.FinishTime = &now
	t.Status = TaskStatusSuccess
	t.Progress = "100%"
}

// Fail 任务失败
func (t *Task) Fail(reason string) {
	now := time.Now()
	t.FinishTime = &now
	t.Status = TaskStatusFailure
	t.FailReason = reason
	t.Progress = ""
}

// Cancel 取消任务
func (t *Task) Cancel() {
	now := time.Now()
	t.FinishTime = &now
	t.Status = TaskStatusCancel
	t.Progress = ""
}

// SetProperty 设置属性
func (t *Task) SetProperty(key string, value interface{}) {
	if t.Properties == nil {
		t.Properties = make(map[string]interface{})
	}
	t.Properties[key] = value
}

// GetProperty 获取属性
func (t *Task) GetProperty(key string) (interface{}, bool) {
	if t.Properties == nil {
		return nil, false
	}
	value, exists := t.Properties[key]
	return value, exists
}

// IsFinished 判断任务是否已完成
func (t *Task) IsFinished() bool {
	return t.Status == TaskStatusSuccess || t.Status == TaskStatusFailure || t.Status == TaskStatusCancel
}

// GetDisplayStatus 获取显示状态
func (t *Task) GetDisplayStatus() string {
	return string(t.Status)
}

// GetDisplayAction 获取显示动作
func (t *Task) GetDisplayAction() string {
	return string(t.Action)
}