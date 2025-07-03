package database

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"midjourney-proxy-go/internal/domain/entity"
	"midjourney-proxy-go/internal/infrastructure/config"
)

// New 创建数据库连接
func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.Type {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.SQLite.Path), &gorm.Config{})
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.MySQL.Username,
			cfg.MySQL.Password,
			cfg.MySQL.Host,
			cfg.MySQL.Port,
			cfg.MySQL.Database,
			cfg.MySQL.Charset,
		)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
			cfg.Postgres.Host,
			cfg.Postgres.Username,
			cfg.Postgres.Password,
			cfg.Postgres.Database,
			cfg.Postgres.Port,
			cfg.Postgres.SSLMode,
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// Migrate 执行数据库迁移
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.Task{},
		&entity.DiscordAccount{},
		&entity.BannedWord{},
		&entity.Setting{},
		&entity.DomainTag{},
		&entity.Message{},
	)
}