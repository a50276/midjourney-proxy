# 多阶段构建 Dockerfile
# 阶段1: 构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# 复制前端依赖文件
COPY web/package*.json ./
RUN npm ci --only=production

# 复制前端源码并构建
COPY web/ ./
RUN npm run build

# 阶段2: 构建后端
FROM golang:1.21-alpine AS backend-builder

# 安装必要工具
RUN apk add --no-cache git

WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 复制前端构建产物
COPY --from=frontend-builder /app/web/dist ./web/dist

# 编译Go应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 阶段3: 运行时镜像
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/main .
COPY --from=backend-builder /app/configs ./configs
COPY --from=backend-builder /app/web/dist ./web/dist

# 创建数据目录
RUN mkdir -p /app/data /app/logs

# 设置权限
RUN chmod +x ./main

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行应用
CMD ["./main"]