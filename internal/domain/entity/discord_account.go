package entity

import (
	"encoding/json"
	"time"
	"gorm.io/gorm"
)

// Component Discord组件
type Component struct {
	Type       int                    `json:"type"`
	Components []ComponentDetail      `json:"components,omitempty"`
	CustomID   string                 `json:"custom_id,omitempty"`
	Style      int                    `json:"style,omitempty"`
	Label      string                 `json:"label,omitempty"`
	Emoji      map[string]interface{} `json:"emoji,omitempty"`
	Disabled   bool                   `json:"disabled,omitempty"`
}

// ComponentDetail 组件详情
type ComponentDetail struct {
	Type     int                    `json:"type"`
	CustomID interface{}            `json:"custom_id,omitempty"`
	Style    int                    `json:"style,omitempty"`
	Label    string                 `json:"label,omitempty"`
	Emoji    map[string]interface{} `json:"emoji,omitempty"`
	Disabled bool                   `json:"disabled,omitempty"`
}

// DiscordAccount Discord账号实体
type DiscordAccount struct {
	ID          string          `gorm:"column:id;primaryKey" json:"id"`
	ChannelID   string          `gorm:"column:channel_id;uniqueIndex" json:"channel_id"`
	GuildID     string          `gorm:"column:guild_id" json:"guild_id"`
	
	// 私信频道
	PrivateChannelID  string `gorm:"column:private_channel_id" json:"private_channel_id,omitempty"`
	NijiBotChannelID  string `gorm:"column:niji_bot_channel_id" json:"niji_bot_channel_id,omitempty"`
	
	// Token信息
	UserToken string `gorm:"column:user_token;type:text" json:"user_token"`
	BotToken  string `gorm:"column:bot_token;type:text" json:"bot_token,omitempty"`
	UserAgent string `gorm:"column:user_agent;type:text" json:"user_agent"`
	
	// 启用状态
	Enabled    bool `gorm:"column:enabled;default:true" json:"enabled"`
	EnableMJ   bool `gorm:"column:enable_mj;default:true" json:"enable_mj"`
	EnableNiji bool `gorm:"column:enable_niji;default:false" json:"enable_niji"`
	
	// 快速模式设置
	EnableFastToRelax bool `gorm:"column:enable_fast_to_relax;default:false" json:"enable_fast_to_relax"`
	EnableRelaxToFast bool `gorm:"column:enable_relax_to_fast;default:false" json:"enable_relax_to_fast"`
	FastExhausted     bool `gorm:"column:fast_exhausted;default:false" json:"fast_exhausted"`
	
	// 锁定状态
	Lock           bool   `gorm:"column:lock;default:false" json:"lock"`
	DisabledReason string `gorm:"column:disabled_reason;type:text" json:"disabled_reason,omitempty"`
	
	// 邀请链接
	PermanentInvitationLink string `gorm:"column:permanent_invitation_link;size:2000" json:"permanent_invitation_link,omitempty"`
	
	// CloudFlare验证
	CfHashCreated *time.Time `gorm:"column:cf_hash_created" json:"cf_hash_created,omitempty"`
	CfHashURL     string     `gorm:"column:cf_hash_url;type:text" json:"cf_hash_url,omitempty"`
	CfURL         string     `gorm:"column:cf_url;type:text" json:"cf_url,omitempty"`
	
	// 赞助者信息
	IsSponsor     bool   `gorm:"column:is_sponsor;default:false" json:"is_sponsor"`
	SponsorUserID string `gorm:"column:sponsor_user_id" json:"sponsor_user_id,omitempty"`
	
	// 并发设置
	CoreSize       int     `gorm:"column:core_size;default:3" json:"core_size"`
	QueueSize      int     `gorm:"column:queue_size;default:10" json:"queue_size"`
	MaxQueueSize   int     `gorm:"column:max_queue_size;default:100" json:"max_queue_size"`
	TimeoutMinutes int     `gorm:"column:timeout_minutes;default:5" json:"timeout_minutes"`
	
	// 间隔设置
	Interval         float64 `gorm:"column:interval;default:1.2" json:"interval"`
	AfterIntervalMin float64 `gorm:"column:after_interval_min;default:1.2" json:"after_interval_min"`
	AfterIntervalMax float64 `gorm:"column:after_interval_max;default:1.2" json:"after_interval_max"`
	
	// 备注和赞助商
	Remark  string `gorm:"column:remark;type:text" json:"remark,omitempty"`
	Sponsor string `gorm:"column:sponsor;type:text" json:"sponsor,omitempty"`
	
	// 信息更新时间
	InfoUpdated *time.Time `gorm:"column:info_updated" json:"info_updated,omitempty"`
	
	// 权重和排序
	Weight int `gorm:"column:weight;default:0" json:"weight"`
	Sort   int `gorm:"column:sort;default:0" json:"sort"`
	
	// 时间配置
	WorkTime    string `gorm:"column:work_time" json:"work_time,omitempty"`
	FishingTime string `gorm:"column:fishing_time" json:"fishing_time,omitempty"`
	
	// Remix设置
	RemixAutoSubmit bool `gorm:"column:remix_auto_submit;default:false" json:"remix_auto_submit"`
	
	// 生成模式
	Mode                GenerationSpeedMode   `gorm:"column:mode" json:"mode,omitempty"`
	AllowModesData      string                `gorm:"column:allow_modes;type:text" json:"-"`
	AllowModes          []GenerationSpeedMode `gorm:"-" json:"allow_modes,omitempty"`
	EnableAutoSetRelax  bool                  `gorm:"column:enable_auto_set_relax;default:false" json:"enable_auto_set_relax"`
	
	// MJ组件
	ComponentsData       string     `gorm:"column:components;type:text" json:"-"`
	Components           []Component `gorm:"-" json:"components,omitempty"`
	SettingsMessageID    string     `gorm:"column:settings_message_id" json:"settings_message_id,omitempty"`
	
	// Niji组件
	NijiComponentsData    string     `gorm:"column:niji_components;type:text" json:"-"`
	NijiComponents        []Component `gorm:"-" json:"niji_components,omitempty"`
	NijiSettingsMessageID string     `gorm:"column:niji_settings_message_id" json:"niji_settings_message_id,omitempty"`
	
	// 功能开关
	IsBlend    bool `gorm:"column:is_blend;default:true" json:"is_blend"`
	IsDescribe bool `gorm:"column:is_describe;default:true" json:"is_describe"`
	IsShorten  bool `gorm:"column:is_shorten;default:true" json:"is_shorten"`
	
	// 自动登录
	LoginAccount   string     `gorm:"column:login_account" json:"login_account,omitempty"`
	LoginPassword  string     `gorm:"column:login_password" json:"login_password,omitempty"`
	Login2FA       string     `gorm:"column:login_2fa" json:"login_2fa,omitempty"`
	IsAutoLogining bool       `gorm:"column:is_auto_logining;default:false" json:"is_auto_logining"`
	LoginStart     *time.Time `gorm:"column:login_start" json:"login_start,omitempty"`
	LoginEnd       *time.Time `gorm:"column:login_end" json:"login_end,omitempty"`
	LoginMessage   string     `gorm:"column:login_message;size:2000" json:"login_message,omitempty"`
	
	// 绘图限制
	DayDrawLimit int `gorm:"column:day_draw_limit;default:-1" json:"day_draw_limit"`
	DayDrawCount int `gorm:"column:day_draw_count;default:0" json:"day_draw_count"`
	
	// 垂直领域
	IsVerticalDomain     bool   `gorm:"column:is_vertical_domain;default:false" json:"is_vertical_domain"`
	VerticalDomainIDsData string `gorm:"column:vertical_domain_ids;type:text" json:"-"`
	VerticalDomainIDs    []string `gorm:"-" json:"vertical_domain_ids,omitempty"`
	
	// 子频道
	SubChannelsData   string            `gorm:"column:sub_channels;type:text" json:"-"`
	SubChannels       []string          `gorm:"-" json:"sub_channels,omitempty"`
	SubChannelValues  map[string]string `gorm:"column:sub_channel_values;type:json" json:"sub_channel_values,omitempty"`
	
	// 运行状态（仅用于显示）
	RunningCount int  `gorm:"-" json:"running_count"`
	QueueCount   int  `gorm:"-" json:"queue_count"`
	Running      bool `gorm:"-" json:"running"`
	
	// 扩展属性
	Properties map[string]interface{} `gorm:"column:properties;type:json" json:"properties,omitempty"`
	
	// 时间戳
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (DiscordAccount) TableName() string {
	return "discord_accounts"
}

// BeforeSave GORM钩子，保存前序列化复杂字段
func (d *DiscordAccount) BeforeSave(tx *gorm.DB) error {
	// 序列化AllowModes
	if d.AllowModes != nil {
		data, err := json.Marshal(d.AllowModes)
		if err != nil {
			return err
		}
		d.AllowModesData = string(data)
	}
	
	// 序列化Components
	if d.Components != nil {
		data, err := json.Marshal(d.Components)
		if err != nil {
			return err
		}
		d.ComponentsData = string(data)
	}
	
	// 序列化NijiComponents
	if d.NijiComponents != nil {
		data, err := json.Marshal(d.NijiComponents)
		if err != nil {
			return err
		}
		d.NijiComponentsData = string(data)
	}
	
	// 序列化VerticalDomainIDs
	if d.VerticalDomainIDs != nil {
		data, err := json.Marshal(d.VerticalDomainIDs)
		if err != nil {
			return err
		}
		d.VerticalDomainIDsData = string(data)
	}
	
	// 序列化SubChannels
	if d.SubChannels != nil {
		data, err := json.Marshal(d.SubChannels)
		if err != nil {
			return err
		}
		d.SubChannelsData = string(data)
	}
	
	return nil
}

// AfterFind GORM钩子，查询后反序列化复杂字段
func (d *DiscordAccount) AfterFind(tx *gorm.DB) error {
	// 反序列化AllowModes
	if d.AllowModesData != "" {
		var modes []GenerationSpeedMode
		if err := json.Unmarshal([]byte(d.AllowModesData), &modes); err == nil {
			d.AllowModes = modes
		}
	}
	
	// 反序列化Components
	if d.ComponentsData != "" {
		var components []Component
		if err := json.Unmarshal([]byte(d.ComponentsData), &components); err == nil {
			d.Components = components
		}
	}
	
	// 反序列化NijiComponents
	if d.NijiComponentsData != "" {
		var components []Component
		if err := json.Unmarshal([]byte(d.NijiComponentsData), &components); err == nil {
			d.NijiComponents = components
		}
	}
	
	// 反序列化VerticalDomainIDs
	if d.VerticalDomainIDsData != "" {
		var ids []string
		if err := json.Unmarshal([]byte(d.VerticalDomainIDsData), &ids); err == nil {
			d.VerticalDomainIDs = ids
		}
	}
	
	// 反序列化SubChannels
	if d.SubChannelsData != "" {
		var channels []string
		if err := json.Unmarshal([]byte(d.SubChannelsData), &channels); err == nil {
			d.SubChannels = channels
		}
	}
	
	return nil
}

// IsAcceptNewTask 是否接受新任务
func (d *DiscordAccount) IsAcceptNewTask() bool {
	if !d.Enabled || d.Lock {
		return false
	}
	
	// 检查绘图限制
	if d.DayDrawLimit > 0 && d.DayDrawCount >= d.DayDrawLimit {
		return false
	}
	
	// TODO: 检查工作时间和摸鱼时间
	
	return true
}

// IsContinueDrawing 是否允许继续绘图
func (d *DiscordAccount) IsContinueDrawing() bool {
	return d.DayDrawLimit <= -1 || d.DayDrawCount < d.DayDrawLimit
}

// GetMJButtons 获取MJ按钮
func (d *DiscordAccount) GetMJButtons() []CustomComponent {
	var buttons []CustomComponent
	for _, comp := range d.Components {
		if comp.Type != 1 { // 不是ActionRow
			continue
		}
		for _, detail := range comp.Components {
			if detail.CustomID != nil {
				buttons = append(buttons, CustomComponent{
					Type:     detail.Type,
					Style:    detail.Style,
					Label:    detail.Label,
					CustomID: detail.CustomID.(string),
				})
			}
		}
	}
	return buttons
}

// GetNijiButtons 获取Niji按钮
func (d *DiscordAccount) GetNijiButtons() []CustomComponent {
	var buttons []CustomComponent
	for _, comp := range d.NijiComponents {
		if comp.Type != 1 { // 不是ActionRow
			continue
		}
		for _, detail := range comp.Components {
			if detail.CustomID != nil {
				buttons = append(buttons, CustomComponent{
					Type:     detail.Type,
					Style:    detail.Style,
					Label:    detail.Label,
					CustomID: detail.CustomID.(string),
				})
			}
		}
	}
	return buttons
}

// IsMJRemixOn MJ是否开启Remix模式
func (d *DiscordAccount) IsMJRemixOn() bool {
	buttons := d.GetMJButtons()
	for _, btn := range buttons {
		if btn.Label == "Remix mode" && btn.Style == 3 {
			return true
		}
	}
	return false
}

// IsMJFastModeOn MJ是否开启快速模式
func (d *DiscordAccount) IsMJFastModeOn() bool {
	buttons := d.GetMJButtons()
	for _, btn := range buttons {
		if (btn.Label == "Fast mode" || btn.Label == "Turbo mode") && btn.Style == 3 {
			return true
		}
	}
	return false
}

// IsNijiRemixOn Niji是否开启Remix模式
func (d *DiscordAccount) IsNijiRemixOn() bool {
	buttons := d.GetNijiButtons()
	for _, btn := range buttons {
		if btn.Label == "Remix mode" && btn.Style == 3 {
			return true
		}
	}
	return false
}

// IsNijiFastModeOn Niji是否开启快速模式
func (d *DiscordAccount) IsNijiFastModeOn() bool {
	buttons := d.GetNijiButtons()
	for _, btn := range buttons {
		if (btn.Label == "Fast mode" || btn.Label == "Turbo mode") && btn.Style == 3 {
			return true
		}
	}
	return false
}

// GetDisplay 获取显示信息
func (d *DiscordAccount) GetDisplay() map[string]interface{} {
	display := map[string]interface{}{
		"id":              d.ID,
		"channel_id":      d.ChannelID,
		"enabled":         d.Enabled,
		"lock":            d.Lock,
		"running":         d.Running,
		"running_count":   d.RunningCount,
		"queue_count":     d.QueueCount,
		"core_size":       d.CoreSize,
		"queue_size":      d.QueueSize,
		"day_draw_count":  d.DayDrawCount,
		"day_draw_limit":  d.DayDrawLimit,
		"weight":          d.Weight,
		"sort":            d.Sort,
		"mj_remix_on":     d.IsMJRemixOn(),
		"mj_fast_mode_on": d.IsMJFastModeOn(),
		"niji_remix_on":   d.IsNijiRemixOn(),
		"niji_fast_mode_on": d.IsNijiFastModeOn(),
	}
	
	// 添加扩展属性
	if d.Properties != nil {
		for k, v := range d.Properties {
			display[k] = v
		}
	}
	
	return display
}