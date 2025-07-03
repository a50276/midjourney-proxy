# Midjourney Proxy - Go版本

一个功能强大、完整、全面且完全免费开源的 Midjourney API 代理项目的 Go 语言重构版本。

代理 Midjourney 的 Discord 频道，通过 API 绘图，支持图片、视频一键换脸。

## 🚀 特性

### ✅ 核心功能
- [x] 支持 Imagine 指令和相关动作 [V1/V2.../U1/U2.../R]
- [x] Imagine 时支持添加图片 base64，作为垫图
- [x] 支持 Blend (图片混合)、Describe (图生文) 指令、Shorten (提示词分析) 指令
- [x] 支持任务实时进度
- [x] 支持中文 prompt 翻译，需配置百度翻译、GPT 翻译
- [x] prompt 敏感词预检测，支持覆盖调整
- [x] 支持 user-token 连接 wss，可以获取错误信息和完整功能
- [x] 支持多账号配置，每个账号可设置对应的任务队列
- [x] 账号池持久化，动态维护
- [x] 支持获取账号 /info、/settings 信息
- [x] 支持 niji・journey Bot 和 Midjourney Bot
- [x] 内嵌管理后台页面，支持多语言
- [x] 支持MJ账号的增删改查功能
- [x] 支持MJ任务查询和管理
- [x] 提供功能齐全的绘图测试页面
- [x] 兼容支持市面上主流绘图客户端和 API 调用

### 🔧 技术特性
- [x] 使用 Go 语言重构，性能更优
- [x] 采用 Gin 框架，轻量高效
- [x] 支持多种数据库：SQLite、MySQL、PostgreSQL、MongoDB
- [x] 完整的 RESTful API 设计
- [x] JWT 认证和权限管理
- [x] 完善的日志系统
- [x] 请求限流和IP黑白名单
- [x] WebSocket 实时通信
- [x] 优雅的错误处理
- [x] 现代化的前端管理界面

### 🎨 前端特性
- [x] Vue.js 3 + TypeScript 前端
- [x] Element Plus UI 组件库
- [x] 响应式设计，支持移动端
- [x] 实时任务状态更新
- [x] 直观的账号管理界面
- [x] 完整的系统监控面板
- [x] 多语言支持

## 📦 安装部署

### 环境要求
- Go 1.21+
- Node.js 16+ (如需自行构建前端)

### 快速启动

#### 1. 二进制部署（推荐）

下载对应平台的二进制文件：

```bash
# Linux
wget https://github.com/your-repo/midjourney-proxy-go/releases/latest/download/midjourney-proxy-go-linux-amd64.tar.gz
tar -xzf midjourney-proxy-go-linux-amd64.tar.gz
cd midjourney-proxy-go

# 编辑配置文件
nano configs/app.yaml

# 启动服务
./midjourney-proxy-go
```

#### 2. Docker 部署

```bash
# 创建配置目录
mkdir -p /opt/midjourney-proxy/{data,logs,configs}

# 下载配置文件模板
wget -O /opt/midjourney-proxy/configs/app.yaml https://raw.githubusercontent.com/your-repo/midjourney-proxy-go/main/configs/app.yaml

# 编辑配置文件
nano /opt/midjourney-proxy/configs/app.yaml

# 启动容器
docker run -d \
  --name midjourney-proxy \
  -p 8080:8080 \
  -v /opt/midjourney-proxy/data:/app/data \
  -v /opt/midjourney-proxy/logs:/app/logs \
  -v /opt/midjourney-proxy/configs:/app/configs \
  --restart unless-stopped \
  your-repo/midjourney-proxy-go:latest
```

#### 3. 源码编译

```bash
# 克隆代码
git clone https://github.com/your-repo/midjourney-proxy-go.git
cd midjourney-proxy-go

# 编译
go mod download
go build -o midjourney-proxy-go cmd/server/main.go

# 启动
./midjourney-proxy-go
```

### 配置说明

主要配置文件：`configs/app.yaml`

```yaml
app:
  name: "Midjourney Proxy"
  version: "1.0.0"
  mode: "production"  # development, production
  port: 8080
  demo_mode: false
  enable_guest: true
  enable_register: false

database:
  type: "sqlite"  # sqlite, mysql, postgres, mongodb
  sqlite:
    path: "./data/midjourney.db"

security:
  admin_token: "your-admin-token"  # 管理员令牌
  user_token: ""                   # 用户令牌
  jwt_secret: "your-jwt-secret"
  jwt_expire_hours: 24

discord:
  accounts: []  # Discord账号配置

# 更多配置项请参考 configs/app.yaml 文件
```

## 🎯 使用说明

### 1. 访问管理后台

启动服务后，访问：`http://localhost:8080`

默认管理员令牌：`admin`

### 2. 配置 Discord 账号

1. 登录管理后台
2. 进入"账号管理"页面
3. 添加 Discord 账号信息：
   - Channel ID：Discord 频道 ID
   - Guild ID：Discord 服务器 ID
   - User Token：Discord 用户令牌
   - 其他配置项

### 3. API 调用

#### 提交绘图任务

```bash
curl -X POST "http://localhost:8080/api/mj/submit/imagine" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin" \
  -d '{
    "prompt": "a beautiful landscape"
  }'
```

#### 查询任务状态

```bash
curl -X GET "http://localhost:8080/api/mj/task/{task_id}" \
  -H "Authorization: Bearer admin"
```

#### 获取任务列表

```bash
curl -X GET "http://localhost:8080/api/mj/task/list?page=1&size=20" \
  -H "Authorization: Bearer admin"
```

### 4. API 文档

服务启动后，访问 Swagger 文档：`http://localhost:8080/swagger/index.html`

## 🔌 API 接口

### 任务提交 API

- `POST /api/mj/submit/imagine` - 提交 Imagine 任务
- `POST /api/mj/submit/change` - 提交变化任务（U1/U2/V1/V2/R）
- `POST /api/mj/submit/describe` - 提交图生文任务
- `POST /api/mj/submit/blend` - 提交图片混合任务
- `POST /api/mj/submit/shorten` - 提交提示词分析任务

### 任务查询 API

- `GET /api/mj/task/{id}` - 获取任务详情
- `GET /api/mj/task/list` - 获取任务列表
- `GET /api/mj/task/queue` - 获取队列状态

### 管理员 API

- `GET /api/admin/accounts` - 账号管理
- `GET /api/admin/users` - 用户管理
- `GET /api/admin/tasks` - 任务管理
- `GET /api/admin/settings` - 系统设置
- `GET /api/admin/stats/*` - 统计信息

## 🎨 前端界面

### 主要页面

1. **首页** - 系统概览和快速操作
2. **绘图测试** - 在线测试绘图功能
3. **任务管理** - 查看和管理所有绘图任务
4. **账号管理** - Discord 账号配置和监控
5. **用户管理** - 用户账号管理
6. **系统设置** - 系统参数配置
7. **统计监控** - 系统运行状态监控

### 界面特色

- **现代化设计** - 采用 Element Plus 设计风格
- **响应式布局** - 完美支持桌面和移动设备
- **实时更新** - WebSocket 实时推送任务状态
- **多语言支持** - 中英文界面切换
- **暗色主题** - 支持明暗主题切换

## 🔧 开发说明

### 项目结构

```
midjourney-proxy-go/
├── cmd/server/          # 应用程序入口
├── internal/            # 内部代码
│   ├── api/            # API 层
│   │   ├── handler/    # 处理器
│   │   └── middleware/ # 中间件
│   ├── domain/         # 领域层
│   │   └── entity/     # 实体模型
│   ├── infrastructure/ # 基础设施层
│   │   ├── config/     # 配置
│   │   ├── database/   # 数据库
│   │   └── discord/    # Discord 集成
│   └── application/    # 应用层
├── pkg/                # 公共包
│   └── logger/         # 日志包
├── configs/            # 配置文件
├── web/                # 前端代码
├── docs/               # 文档
└── scripts/            # 脚本
```

### 开发环境设置

```bash
# 安装依赖
go mod tidy

# 安装开发工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成 API 文档
swag init -g cmd/server/main.go

# 运行开发服务器
go run cmd/server/main.go
```

### 前端开发

```bash
cd web
npm install
npm run dev
```

## 📝 更新日志

### v1.0.0
- ✅ 完成 Go 语言重构
- ✅ 实现完整的 API 接口
- ✅ 创建现代化前端界面
- ✅ 支持多种数据库
- ✅ 完善的权限管理
- ✅ 实时状态更新

## 🤝 贡献

欢迎提交 Pull Request 和 Issue！

### 开发流程

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 GPL v3.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## ⚠️ 免责声明

- 本项目仅供学习和研究使用
- 请遵守相关法律法规，不得用于违法用途
- 用户需自行承担使用风险

## 🙏 致谢

感谢原 .NET 版本项目的贡献者们，本项目基于其优秀的设计理念进行 Go 语言重构。

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- GitHub Issues: [提交问题](https://github.com/your-repo/midjourney-proxy-go/issues)
- 邮箱: your-email@example.com

---

**⭐ 如果这个项目对你有帮助，请给它一个 Star！**
