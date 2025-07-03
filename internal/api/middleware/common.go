package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	"midjourney-proxy-go/internal/infrastructure/config"
	"midjourney-proxy-go/pkg/logger"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// IPLimiter IP限流器
type IPLimiter struct {
	ips    map[string]*rate.Limiter
	mu     sync.RWMutex
	config config.RateLimitConfig
}

// NewIPLimiter 创建IP限流器
func NewIPLimiter(config config.RateLimitConfig) *IPLimiter {
	return &IPLimiter{
		ips:    make(map[string]*rate.Limiter),
		config: config,
	}
}

// GetLimiter 获取IP的限流器
func (i *IPLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(1, 5) // 默认限制：每秒1个请求，突发5个
		i.ips[ip] = limiter
	}

	return limiter
}

// isIPInList 检查IP是否在列表中
func isIPInList(ip string, list []string) bool {
	for _, item := range list {
		if strings.Contains(item, "/") {
			// CIDR格式
			_, ipnet, err := net.ParseCIDR(item)
			if err == nil {
				if ipnet.Contains(net.ParseIP(ip)) {
					return true
				}
			}
		} else {
			// 直接IP比较
			if ip == item {
				return true
			}
		}
	}
	return false
}

// getClientIP 获取客户端IP
func getClientIP(c *gin.Context) string {
	// 检查X-Forwarded-For头
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 使用RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}

// RateLimit 限流中间件
func RateLimit(config config.RateLimitConfig, logger logger.Logger) gin.HandlerFunc {
	if !config.Enabled {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	limiter := NewIPLimiter(config)

	return gin.HandlerFunc(func(c *gin.Context) {
		ip := getClientIP(c)

		// 检查白名单
		if isIPInList(ip, config.Whitelist) {
			c.Next()
			return
		}

		// 检查黑名单
		if isIPInList(ip, config.Blacklist) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "IP is blacklisted",
			})
			c.Abort()
			return
		}

		// 检查路径规则
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// 构造匹配键
		matchKey := fmt.Sprintf("%s %s", method, path)
		
		// 查找匹配的规则
		var rateLimits map[string]int
		for pattern, limits := range config.Rules {
			if matchPattern(pattern, matchKey) {
				rateLimits = limits
				break
			}
		}

		if rateLimits != nil {
			// 应用限流规则
			ipLimiter := limiter.GetLimiter(ip)
			if !ipLimiter.Allow() {
				logger.Warnf("Rate limit exceeded for IP %s on path %s", ip, path)
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":    429,
					"message": "Rate limit exceeded",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// matchPattern 匹配模式
func matchPattern(pattern, path string) bool {
	// 简单的通配符匹配
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			return strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix)
		}
	}
	
	return pattern == path
}

// Recovery 错误恢复中间件
func Recovery(logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	})
}