package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"midjourney-proxy-go/internal/domain/entity"
	"midjourney-proxy-go/internal/infrastructure/config"
	"midjourney-proxy-go/pkg/logger"
)

// Claims JWT载荷
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Auth 认证中间件
func Auth(securityCfg config.SecurityConfig, db *gorm.DB, logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 检查Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 检查token查询参数
			token := c.Query("token")
			if token != "" {
				authHeader = "Bearer " + token
			}
		}

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 检查是否是管理员token
		if tokenString == securityCfg.AdminToken {
			// 查找或创建管理员用户
			var adminUser entity.User
			err := db.Where("role = ? AND token = ?", entity.RoleAdmin, tokenString).First(&adminUser).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// 创建默认管理员用户
					adminUser = entity.User{
						ID:       "admin",
						Username: "admin",
						Role:     entity.RoleAdmin,
						Token:    tokenString,
						Enabled:  true,
					}
					if err := db.Create(&adminUser).Error; err != nil {
						logger.Errorf("Failed to create admin user: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{
							"code":    500,
							"message": "Internal server error",
						})
						c.Abort()
						return
					}
				} else {
					logger.Errorf("Failed to query admin user: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "Internal server error",
					})
					c.Abort()
					return
				}
			}

			c.Set("user", &adminUser)
			c.Set("user_id", adminUser.ID)
			c.Set("user_role", adminUser.Role)
			c.Next()
			return
		}

		// 检查是否是用户token
		if tokenString == securityCfg.UserToken && securityCfg.UserToken != "" {
			// 查找或创建用户
			var user entity.User
			err := db.Where("role = ? AND token = ?", entity.RoleUser, tokenString).First(&user).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// 创建默认用户
					user = entity.User{
						ID:       "user",
						Username: "user",
						Role:     entity.RoleUser,
						Token:    tokenString,
						Enabled:  true,
					}
					if err := db.Create(&user).Error; err != nil {
						logger.Errorf("Failed to create user: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{
							"code":    500,
							"message": "Internal server error",
						})
						c.Abort()
						return
					}
				} else {
					logger.Errorf("Failed to query user: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "Internal server error",
					})
					c.Abort()
					return
				}
			}

			c.Set("user", &user)
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
			c.Next()
			return
		}

		// 解析JWT token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(securityCfg.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			// 检查token是否过期
			if claims.ExpiresAt.Time.Before(time.Now()) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "Token expired",
				})
				c.Abort()
				return
			}

			// 查找用户
			var user entity.User
			err := db.Where("id = ?", claims.UserID).First(&user).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusUnauthorized, gin.H{
						"code":    401,
						"message": "User not found",
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "Internal server error",
					})
				}
				c.Abort()
				return
			}

			// 检查用户状态
			if !user.Enabled {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "User is disabled",
				})
				c.Abort()
				return
			}

			// 检查用户是否过期
			if user.IsExpired() {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "User is expired",
				})
				c.Abort()
				return
			}

			c.Set("user", &user)
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid token",
			})
			c.Abort()
		}
	})
}

// AdminOnly 管理员权限中间件
func AdminOnly() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		if role != entity.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// GenerateJWTToken 生成JWT token
func GenerateJWTToken(userID string, role entity.UserRole, secret string, expireHours int) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}