package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用程序配置
type Config struct {
	App         AppConfig         `mapstructure:"app"`
	Log         LogConfig         `mapstructure:"log"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Discord     DiscordConfig     `mapstructure:"discord"`
	Translate   TranslateConfig   `mapstructure:"translate"`
	FaceSwap    FaceSwapConfig    `mapstructure:"face_swap"`
	Storage     StorageConfig     `mapstructure:"storage"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limiting"`
	Security    SecurityConfig    `mapstructure:"security"`
	Notification NotificationConfig `mapstructure:"notification"`
	Captcha     CaptchaConfig     `mapstructure:"captcha"`
}

// AppConfig 应用程序配置
type AppConfig struct {
	Name           string `mapstructure:"name"`
	Version        string `mapstructure:"version"`
	Mode           string `mapstructure:"mode"`
	Port           int    `mapstructure:"port"`
	DemoMode       bool   `mapstructure:"demo_mode"`
	EnableGuest    bool   `mapstructure:"enable_guest"`
	EnableRegister bool   `mapstructure:"enable_register"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string         `mapstructure:"type"`
	SQLite   SQLiteConfig   `mapstructure:"sqlite"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	MongoDB  MongoDBConfig  `mapstructure:"mongodb"`
}

// SQLiteConfig SQLite配置
type SQLiteConfig struct {
	Path string `mapstructure:"path"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// DiscordConfig Discord配置
type DiscordConfig struct {
	Accounts   []DiscordAccount   `mapstructure:"accounts"`
	Proxy      ProxyConfig        `mapstructure:"proxy"`
	NgDiscord  NgDiscordConfig    `mapstructure:"ng_discord"`
}

// DiscordAccount Discord账号配置
type DiscordAccount struct {
	ID                   string `mapstructure:"id"`
	ChannelID            string `mapstructure:"channel_id"`
	GuildID              string `mapstructure:"guild_id"`
	UserToken            string `mapstructure:"user_token"`
	BotToken             string `mapstructure:"bot_token"`
	UserAgent            string `mapstructure:"user_agent"`
	Enabled              bool   `mapstructure:"enabled"`
	EnableMJ             bool   `mapstructure:"enable_mj"`
	EnableNiji           bool   `mapstructure:"enable_niji"`
	CoreSize             int    `mapstructure:"core_size"`
	QueueSize            int    `mapstructure:"queue_size"`
	TimeoutMinutes       int    `mapstructure:"timeout_minutes"`
	Interval             float64 `mapstructure:"interval"`
	Weight               int    `mapstructure:"weight"`
	WorkTime             string `mapstructure:"work_time"`
	FishingTime          string `mapstructure:"fishing_time"`
	DayDrawLimit         int    `mapstructure:"day_draw_limit"`
	RemixAutoSubmit      bool   `mapstructure:"remix_auto_submit"`
	Mode                 string `mapstructure:"mode"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
}

// NgDiscordConfig NgDiscord配置
type NgDiscordConfig struct {
	Server       string `mapstructure:"server"`
	CDN          string `mapstructure:"cdn"`
	WSS          string `mapstructure:"wss"`
	ResumeWSS    string `mapstructure:"resume_wss"`
	UploadServer string `mapstructure:"upload_server"`
	SaveToLocal  bool   `mapstructure:"save_to_local"`
	CustomCDN    string `mapstructure:"custom_cdn"`
}

// TranslateConfig 翻译配置
type TranslateConfig struct {
	Way    string         `mapstructure:"way"`
	Baidu  BaiduConfig    `mapstructure:"baidu"`
	OpenAI OpenAIConfig   `mapstructure:"openai"`
}

// BaiduConfig 百度翻译配置
type BaiduConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIURL      string  `mapstructure:"api_url"`
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	Timeout     int     `mapstructure:"timeout"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// FaceSwapConfig 换脸配置
type FaceSwapConfig struct {
	Enabled        bool  `mapstructure:"enabled"`
	Token          string `mapstructure:"token"`
	CoreSize       int   `mapstructure:"core_size"`
	QueueSize      int   `mapstructure:"queue_size"`
	TimeoutMinutes int   `mapstructure:"timeout_minutes"`
	MaxFileSize    int64 `mapstructure:"max_file_size"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type  string      `mapstructure:"type"`
	Local LocalConfig `mapstructure:"local"`
	OSS   OSSConfig   `mapstructure:"oss"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	Path string `mapstructure:"path"`
	CDN  string `mapstructure:"cdn"`
}

// OSSConfig 阿里云OSS配置
type OSSConfig struct {
	BucketName      string `mapstructure:"bucket_name"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"`
	CustomCDN       string `mapstructure:"custom_cdn"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled   bool                            `mapstructure:"enabled"`
	Whitelist []string                        `mapstructure:"whitelist"`
	Blacklist []string                        `mapstructure:"blacklist"`
	Rules     map[string]map[string]int       `mapstructure:"rules"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	AdminToken     string `mapstructure:"admin_token"`
	UserToken      string `mapstructure:"user_token"`
	JWTSecret      string `mapstructure:"jwt_secret"`
	JWTExpireHours int    `mapstructure:"jwt_expire_hours"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Webhook string     `mapstructure:"webhook"`
	SMTP    SMTPConfig `mapstructure:"smtp"`
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
	ToEmail   string `mapstructure:"to_email"`
	EnableSSL bool   `mapstructure:"enable_ssl"`
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Server     string `mapstructure:"server"`
	NotifyHook string `mapstructure:"notify_hook"`
}

// Load 加载配置
func Load() (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	v.SetConfigName("app")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// 设置环境变量前缀
	v.SetEnvPrefix("MJ")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}