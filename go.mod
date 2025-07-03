module midjourney-proxy-go

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/gorilla/websocket v1.5.1
	github.com/joho/godotenv v1.5.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/viper v1.17.0
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.2
	gorm.io/gorm v1.25.5
	gorm.io/driver/sqlite v1.5.4
	gorm.io/driver/mysql v1.5.2
	gorm.io/driver/postgres v1.5.4
	go.mongodb.org/mongo-driver v1.13.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.5.0
	github.com/robfig/cron/v3 v3.0.1
	golang.org/x/crypto v0.17.0
	golang.org/x/time v0.5.0
)