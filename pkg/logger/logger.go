package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger 日志接口
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

type logger struct {
	entry *logrus.Entry
}

// New 创建新的日志实例
func New(level, format string) Logger {
	log := logrus.New()

	// 设置日志级别
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// 设置日志格式
	switch strings.ToLower(format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	return &logger{
		entry: logrus.NewEntry(log),
	}
}

// NewWithOutput 创建带输出的日志实例
func NewWithOutput(level, format, output, filePath string) Logger {
	log := logrus.New()

	// 设置日志级别
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// 设置日志格式
	switch strings.ToLower(format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置输出
	switch strings.ToLower(output) {
	case "file":
		if filePath != "" {
			// 确保目录存在
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err == nil {
				if file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
					log.SetOutput(file)
				}
			}
		}
	case "both":
		if filePath != "" {
			// 确保目录存在
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err == nil {
				if file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
					log.SetOutput(io.MultiWriter(os.Stdout, file))
				}
			}
		}
	default:
		log.SetOutput(os.Stdout)
	}

	return &logger{
		entry: logrus.NewEntry(log),
	}
}

func (l *logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l *logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l *logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{
		entry: l.entry.WithField(key, value),
	}
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	return &logger{
		entry: l.entry.WithFields(fields),
	}
}