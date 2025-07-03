package entity

import (
	"time"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRole 用户角色枚举
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User 用户实体
type User struct {
	ID       string   `gorm:"column:id;primaryKey" json:"id"`
	Username string   `gorm:"column:username;uniqueIndex" json:"username"`
	Email    string   `gorm:"column:email;uniqueIndex" json:"email,omitempty"`
	Password string   `gorm:"column:password" json:"-"`
	Role     UserRole `gorm:"column:role;default:'user'" json:"role"`
	
	// 令牌信息
	Token      string `gorm:"column:token;uniqueIndex" json:"token,omitempty"`
	TokenType  string `gorm:"column:token_type;default:'bearer'" json:"token_type,omitempty"`
	
	// 状态信息
	Enabled   bool `gorm:"column:enabled;default:true" json:"enabled"`
	IsWhite   bool `gorm:"column:is_white;default:false" json:"is_white"`
	
	// 绘图统计
	TotalDrawCount int `gorm:"column:total_draw_count;default:0" json:"total_draw_count"`
	DayDrawCount   int `gorm:"column:day_draw_count;default:0" json:"day_draw_count"`
	
	// 绘图限制
	DayDrawLimit   int `gorm:"column:day_draw_limit;default:-1" json:"day_draw_limit"`
	TotalDrawLimit int `gorm:"column:total_draw_limit;default:-1" json:"total_draw_limit"`
	
	// 有效期
	ExpiredAt *time.Time `gorm:"column:expired_at" json:"expired_at,omitempty"`
	
	// 备注信息
	Remark string `gorm:"column:remark;type:text" json:"remark,omitempty"`
	
	// 时间戳
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// SetPassword 设置密码（加密）
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 检查密码
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// IsAdmin 是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsExpired 是否已过期
func (u *User) IsExpired() bool {
	if u.ExpiredAt == nil {
		return false
	}
	return u.ExpiredAt.Before(time.Now())
}

// CanDraw 是否可以绘图
func (u *User) CanDraw() bool {
	if !u.Enabled || u.IsExpired() {
		return false
	}
	
	// 检查总绘图限制
	if u.TotalDrawLimit > 0 && u.TotalDrawCount >= u.TotalDrawLimit {
		return false
	}
	
	// 检查日绘图限制
	if u.DayDrawLimit > 0 && u.DayDrawCount >= u.DayDrawLimit {
		return false
	}
	
	return true
}

// IncrementDrawCount 增加绘图次数
func (u *User) IncrementDrawCount() {
	u.TotalDrawCount++
	u.DayDrawCount++
}

// ResetDayDrawCount 重置日绘图次数
func (u *User) ResetDayDrawCount() {
	u.DayDrawCount = 0
}

// BannedWord 禁用词实体
type BannedWord struct {
	ID      string `gorm:"column:id;primaryKey" json:"id"`
	Word    string `gorm:"column:word;uniqueIndex" json:"word"`
	GroupID string `gorm:"column:group_id;index" json:"group_id,omitempty"`
	Enabled bool   `gorm:"column:enabled;default:true" json:"enabled"`
	Remark  string `gorm:"column:remark;type:text" json:"remark,omitempty"`
	
	// 时间戳
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (BannedWord) TableName() string {
	return "banned_words"
}

// Setting 系统设置实体
type Setting struct {
	ID    string `gorm:"column:id;primaryKey" json:"id"`
	Key   string `gorm:"column:key;uniqueIndex" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
	Type  string `gorm:"column:type;default:'string'" json:"type"` // string, number, boolean, json
	Group string `gorm:"column:group;index" json:"group,omitempty"`
	Title string `gorm:"column:title" json:"title,omitempty"`
	Description string `gorm:"column:description;type:text" json:"description,omitempty"`
	
	// 时间戳
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

// DomainTag 领域标签实体
type DomainTag struct {
	ID      string `gorm:"column:id;primaryKey" json:"id"`
	Name    string `gorm:"column:name;uniqueIndex" json:"name"`
	Enabled bool   `gorm:"column:enabled;default:true" json:"enabled"`
	Sort    int    `gorm:"column:sort;default:0" json:"sort"`
	Remark  string `gorm:"column:remark;type:text" json:"remark,omitempty"`
	
	// 时间戳
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (DomainTag) TableName() string {
	return "domain_tags"
}

// Message 消息实体
type Message struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	MessageID    string    `gorm:"column:message_id;index" json:"message_id"`
	ChannelID    string    `gorm:"column:channel_id;index" json:"channel_id"`
	GuildID      string    `gorm:"column:guild_id;index" json:"guild_id"`
	AuthorID     string    `gorm:"column:author_id;index" json:"author_id"`
	Content      string    `gorm:"column:content;type:text" json:"content,omitempty"`
	Type         int       `gorm:"column:type" json:"type"`
	Flags        int       `gorm:"column:flags" json:"flags,omitempty"`
	Reference    string    `gorm:"column:reference;type:text" json:"reference,omitempty"`
	Hash         string    `gorm:"column:hash;index" json:"hash,omitempty"`
	
	// 附件信息
	Attachments []map[string]interface{} `gorm:"column:attachments;type:json" json:"attachments,omitempty"`
	
	// 组件信息
	Components []Component `gorm:"column:components;type:json" json:"components,omitempty"`
	
	// 嵌入信息
	Embeds []map[string]interface{} `gorm:"column:embeds;type:json" json:"embeds,omitempty"`
	
	// 时间戳
	Timestamp time.Time      `gorm:"column:timestamp" json:"timestamp"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}