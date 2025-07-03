#!/bin/bash

# Midjourney Proxy 部署脚本
# 支持多种部署方式：Docker、源码编译等

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 命令未找到，请先安装 $1"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
Midjourney Proxy 部署脚本

用法: $0 [选项]

选项:
    -h, --help              显示此帮助信息
    -m, --mode MODE         部署模式 (docker|source|binary)
    -d, --db DATABASE       数据库类型 (sqlite|mysql|postgres|mongodb)
    -p, --port PORT         端口号 (默认: 8080)
    --dev                   开发模式
    --production            生产模式
    --skip-build            跳过构建步骤
    --clean                 清理环境后重新部署

示例:
    $0 -m docker -d mysql -p 8080
    $0 -m source --dev
    $0 --clean -m docker

EOF
}

# 默认配置
DEPLOY_MODE="docker"
DATABASE="sqlite"
PORT=8080
APP_MODE="production"
SKIP_BUILD=false
CLEAN=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -m|--mode)
            DEPLOY_MODE="$2"
            shift 2
            ;;
        -d|--db)
            DATABASE="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        --dev)
            APP_MODE="development"
            shift
            ;;
        --production)
            APP_MODE="production"
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

log_info "Midjourney Proxy 部署脚本"
log_info "部署模式: $DEPLOY_MODE"
log_info "数据库: $DATABASE"
log_info "端口: $PORT"
log_info "应用模式: $APP_MODE"

# 创建必要目录
create_directories() {
    log_info "创建必要目录..."
    mkdir -p data logs configs scripts nginx/ssl
}

# 清理环境
clean_environment() {
    if [ "$CLEAN" = true ]; then
        log_warning "正在清理环境..."
        docker-compose down -v 2>/dev/null || true
        docker system prune -f 2>/dev/null || true
        rm -rf data logs
        log_success "环境清理完成"
    fi
}

# 生成配置文件
generate_config() {
    log_info "生成配置文件..."
    
    cat > configs/app.yaml << EOF
app:
  name: "Midjourney Proxy"
  version: "1.0.0"
  mode: "${APP_MODE}"
  port: ${PORT}
  demo_mode: false
  enable_guest: true
  enable_register: false

log:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "./logs/app.log"

database:
  type: "${DATABASE}"
  sqlite:
    path: "./data/midjourney.db"
  mysql:
    host: "mysql"
    port: 3306
    username: "midjourney"
    password: "password123"
    database: "midjourney"
    charset: "utf8mb4"
  postgres:
    host: "postgres"
    port: 5432
    username: "midjourney"
    password: "password123"
    database: "midjourney"
    sslmode: "disable"
  mongodb:
    uri: "mongodb://admin:password123@mongodb:27017"
    database: "midjourney"

redis:
  enabled: true
  host: "redis"
  port: 6379
  password: "redis123456"
  database: 0

security:
  admin_token: "$(openssl rand -hex 16)"
  user_token: ""
  jwt_secret: "$(openssl rand -hex 32)"
  jwt_expire_hours: 24

discord:
  accounts: []

rate_limiting:
  enabled: true
  whitelist: ["127.0.0.1", "::1"]
  blacklist: []
  rules:
    "*/mj/submit/*":
      "3": 1
      "60": 6
      "600": 20
      "3600": 60
      "86400": 120
EOF

    log_success "配置文件生成完成: configs/app.yaml"
}

# Docker 部署
deploy_docker() {
    log_info "使用 Docker 部署..."
    
    check_command docker
    check_command docker-compose
    
    # 生成 Docker Compose override 文件
    cat > docker-compose.override.yml << EOF
version: '3.8'

services:
  midjourney-proxy:
    ports:
      - "${PORT}:8080"
    environment:
      - APP_MODE=${APP_MODE}
EOF
    
    # 根据数据库类型选择服务
    if [ "$DATABASE" != "sqlite" ]; then
        log_info "启动数据库服务: $DATABASE"
        docker-compose --profile $DATABASE up -d $DATABASE
        
        # 等待数据库启动
        log_info "等待数据库启动..."
        sleep 10
    fi
    
    # 构建并启动主应用
    if [ "$SKIP_BUILD" = false ]; then
        log_info "构建应用镜像..."
        docker-compose build midjourney-proxy
    fi
    
    log_info "启动应用服务..."
    docker-compose up -d midjourney-proxy redis
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 5
    
    # 检查服务状态
    if curl -s http://localhost:$PORT/health > /dev/null; then
        log_success "应用启动成功！"
        log_info "访问地址: http://localhost:$PORT"
        log_info "管理后台: http://localhost:$PORT"
        log_info "API文档: http://localhost:$PORT/swagger/index.html"
    else
        log_error "应用启动失败，请检查日志"
        docker-compose logs midjourney-proxy
        exit 1
    fi
}

# 源码部署
deploy_source() {
    log_info "使用源码部署..."
    
    check_command go
    check_command node
    check_command npm
    
    # 检查 Go 版本
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ $(echo "$GO_VERSION 1.21" | tr " " "\n" | sort -V | head -n1) != "1.21" ]]; then
        log_error "需要 Go 1.21 或更高版本"
        exit 1
    fi
    
    # 安装 Go 依赖
    log_info "安装 Go 依赖..."
    go mod tidy
    
    # 构建前端
    if [ "$SKIP_BUILD" = false ]; then
        log_info "构建前端..."
        cd web
        npm install
        npm run build
        cd ..
    fi
    
    # 生成 Swagger 文档
    if command -v swag &> /dev/null; then
        log_info "生成 API 文档..."
        swag init -g cmd/server/main.go
    fi
    
    # 编译应用
    log_info "编译应用..."
    go build -o midjourney-proxy cmd/server/main.go
    
    # 启动应用
    log_info "启动应用..."
    if [ "$APP_MODE" = "development" ]; then
        ./midjourney-proxy &
        APP_PID=$!
        echo $APP_PID > midjourney-proxy.pid
    else
        nohup ./midjourney-proxy > logs/app.log 2>&1 &
        APP_PID=$!
        echo $APP_PID > midjourney-proxy.pid
    fi
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 3
    
    # 检查服务状态
    if curl -s http://localhost:$PORT/health > /dev/null; then
        log_success "应用启动成功！PID: $APP_PID"
        log_info "访问地址: http://localhost:$PORT"
        log_info "停止应用: kill $APP_PID"
    else
        log_error "应用启动失败"
        exit 1
    fi
}

# 二进制部署
deploy_binary() {
    log_info "使用预编译二进制部署..."
    
    # 检查二进制文件是否存在
    if [ ! -f "midjourney-proxy" ]; then
        log_error "二进制文件不存在，请先编译或下载"
        exit 1
    fi
    
    # 设置权限
    chmod +x midjourney-proxy
    
    # 启动应用
    log_info "启动应用..."
    if [ "$APP_MODE" = "development" ]; then
        ./midjourney-proxy &
        APP_PID=$!
        echo $APP_PID > midjourney-proxy.pid
    else
        nohup ./midjourney-proxy > logs/app.log 2>&1 &
        APP_PID=$!
        echo $APP_PID > midjourney-proxy.pid
    fi
    
    log_success "应用启动成功！PID: $APP_PID"
}

# 主函数
main() {
    clean_environment
    create_directories
    generate_config
    
    case $DEPLOY_MODE in
        docker)
            deploy_docker
            ;;
        source)
            deploy_source
            ;;
        binary)
            deploy_binary
            ;;
        *)
            log_error "不支持的部署模式: $DEPLOY_MODE"
            exit 1
            ;;
    esac
    
    log_success "部署完成！"
    
    # 显示后续操作提示
    cat << EOF

后续操作:
1. 访问管理后台配置 Discord 账号
2. 查看应用日志: docker-compose logs -f (Docker模式) 或 tail -f logs/app.log
3. 停止应用: docker-compose down (Docker模式) 或 kill \$(cat midjourney-proxy.pid)

配置文件位置: configs/app.yaml
数据目录: data/
日志目录: logs/

EOF
}

# 执行主函数
main